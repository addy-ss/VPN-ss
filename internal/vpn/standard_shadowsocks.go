package vpn

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"github.com/sirupsen/logrus"
)

// StandardShadowsocksServer 标准Shadowsocks服务器
type StandardShadowsocksServer struct {
	config          *Config
	listener        net.Listener
	logger          *logrus.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	useSecondServer bool
}

// NewStandardShadowsocksServer 创建标准Shadowsocks服务器
func NewStandardShadowsocksServer(config *Config, logger *logrus.Logger) *StandardShadowsocksServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &StandardShadowsocksServer{
		config:          config,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		useSecondServer: config.SecondServerEnabled,
	}
}

// Start 启动标准Shadowsocks服务器
func (s *StandardShadowsocksServer) Start() error {
	// 创建加密器
	cipher, err := core.PickCipher(s.config.Method, nil, s.config.Password)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	// 尝试启动监听，如果端口被占用则尝试下一个端口
	originalPort := s.config.Port
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		port := originalPort + i
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			if i == 0 {
				s.logger.Warnf("Failed to listen on %s: %v, trying next port", addr, err)
			}
			continue
		}

		s.listener = listener
		s.config.Port = port // 更新实际使用的端口

		s.logger.Infof("Standard Shadowsocks server listening on %s", addr)
		s.logger.Infof("Method: %s, Password: %s", s.config.Method, s.config.Password)

		// 接受连接
		go s.acceptConnections(cipher)
		return nil
	}

	return fmt.Errorf("failed to find available port after %d attempts", maxRetries)
}

// Stop 停止服务器
func (s *StandardShadowsocksServer) Stop() {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
}

// acceptConnections 接受连接
func (s *StandardShadowsocksServer) acceptConnections(cipher core.Cipher) {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				s.logger.Errorf("Failed to accept connection: %v", err)
				continue
			}
			go s.handleConnection(conn, cipher)
		}
	}
}

// handleConnection 处理连接
func (s *StandardShadowsocksServer) handleConnection(conn net.Conn, cipher core.Cipher) {
	defer conn.Close()

	clientIP := conn.RemoteAddr().String()
	s.logger.Infof("New standard Shadowsocks connection from %s", clientIP)

	// 创建加密连接
	ssconn := cipher.StreamConn(conn)
	defer ssconn.Close()

	// 读取SOCKS5请求
	tgt, err := socks.ReadAddr(ssconn)
	if err != nil {
		s.logger.Errorf("Failed to read target address: %v", err)
		return
	}

	s.logger.Infof("Received target: %s, useSecondServer: %v", tgt.String(), s.useSecondServer)

	// 如果启用了第二个服务端，转发到第二个服务端
	if s.useSecondServer {
		s.logger.Infof("Forwarding to second server: %s:%d", s.config.SecondServerHost, s.config.SecondServerPort)
		s.forwardToSecondServer(ssconn, tgt.String(), cipher)
		return
	}

	s.logger.Infof("Connecting directly to target: %s", tgt.String())

	// 直接连接目标服务器
	target, err := net.Dial("tcp", tgt.String())
	if err != nil {
		s.logger.Errorf("Failed to connect to target %s: %v", tgt.String(), err)
		return
	}
	defer target.Close()

	s.logger.Infof("Connected to target: %s", tgt.String())

	// 双向转发数据
	s.forwardData(ssconn, target, clientIP)
}

// forwardData 转发数据
func (s *StandardShadowsocksServer) forwardData(src, dst net.Conn, clientIP string) {
	errChan := make(chan error, 2)

	// 从源到目标
	go func() {
		buffer := make([]byte, 8192)
		for {
			n, err := src.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}
			_, err = dst.Write(buffer[:n])
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 从目标到源
	go func() {
		buffer := make([]byte, 8192)
		for {
			n, err := dst.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}
			_, err = src.Write(buffer[:n])
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出错
	select {
	case err := <-errChan:
		if err.Error() != "EOF" {
			s.logger.Errorf("Connection error from %s: %v", clientIP, err)
		}
	case <-s.ctx.Done():
		return
	}
}

// forwardToSecondServer 转发到第二个服务端
func (s *StandardShadowsocksServer) forwardToSecondServer(clientConn net.Conn, target string, cipher core.Cipher) {
	// 连接到第二个服务端
	secondServerAddr := fmt.Sprintf("%s:%d", s.config.SecondServerHost, s.config.SecondServerPort)
	secondServerConn, err := net.DialTimeout("tcp", secondServerAddr, 10*time.Second)
	if err != nil {
		s.logger.Errorf("Failed to connect to second server %s: %v", secondServerAddr, err)
		return
	}
	defer secondServerConn.Close()

	s.logger.Infof("Connected to second server %s, forwarding target: %s", secondServerAddr, target)

	// 创建到第二个服务端的加密连接
	secondCipher, err := core.PickCipher(s.config.SecondServerMethod, nil, s.config.SecondServerPassword)
	if err != nil {
		s.logger.Errorf("Failed to create cipher for second server: %v", err)
		return
	}

	secondSSConn := secondCipher.StreamConn(secondServerConn)
	defer secondSSConn.Close()

	// 发送目标地址到第二个服务端
	if err := s.writeTargetAddress(secondSSConn, target); err != nil {
		s.logger.Errorf("Failed to write target to second server: %v", err)
		return
	}

	// 双向转发数据：客户端 <-> 第一个服务端 <-> 第二个服务端 <-> 目标
	s.forwardBetweenServers(clientConn, secondSSConn)
}

// forwardBetweenServers 在两个服务端之间转发数据
func (s *StandardShadowsocksServer) forwardBetweenServers(clientConn, secondServerConn net.Conn) {
	errChan := make(chan error, 2)

	// 从客户端到第二个服务端
	go func() {
		buffer := make([]byte, 8192)
		for {
			n, err := clientConn.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}
			_, err = secondServerConn.Write(buffer[:n])
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 从第二个服务端到客户端
	go func() {
		buffer := make([]byte, 8192)
		for {
			n, err := secondServerConn.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}
			_, err = clientConn.Write(buffer[:n])
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出错
	select {
	case err := <-errChan:
		if err.Error() != "EOF" {
			s.logger.Errorf("Connection error between servers: %v", err)
		}
	case <-s.ctx.Done():
		return
	}
}

// writeTargetAddress 写入目标地址到连接
func (s *StandardShadowsocksServer) writeTargetAddress(conn net.Conn, target string) error {
	// 解析目标地址
	host, portStr, err := net.SplitHostPort(target)
	if err != nil {
		return fmt.Errorf("invalid target address: %v", err)
	}

	// 解析端口
	port, err := net.LookupPort("tcp", portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %v", err)
	}

	// 构建SOCKS5地址格式
	var addr []byte
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			// IPv4
			addr = make([]byte, 7)
			addr[0] = 1 // IPv4
			copy(addr[1:5], ip.To4())
			addr[5] = byte(port >> 8)
			addr[6] = byte(port)
		} else {
			// IPv6
			addr = make([]byte, 19)
			addr[0] = 4 // IPv6
			copy(addr[1:17], ip)
			addr[17] = byte(port >> 8)
			addr[18] = byte(port)
		}
	} else {
		// 域名
		addr = make([]byte, 4+len(host))
		addr[0] = 3 // 域名
		addr[1] = byte(len(host))
		copy(addr[2:2+len(host)], host)
		addr[2+len(host)] = byte(port >> 8)
		addr[3+len(host)] = byte(port)
	}

	// 写入地址
	_, err = conn.Write(addr)
	return err
}
