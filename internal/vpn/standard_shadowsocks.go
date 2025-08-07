package vpn

import (
	"context"
	"fmt"
	"net"

	"github.com/shadowsocks/go-shadowsocks2/core"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"github.com/sirupsen/logrus"
)

// StandardShadowsocksServer 标准Shadowsocks服务器
type StandardShadowsocksServer struct {
	config   *Config
	listener net.Listener
	logger   *logrus.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStandardShadowsocksServer 创建标准Shadowsocks服务器
func NewStandardShadowsocksServer(config *Config, logger *logrus.Logger) *StandardShadowsocksServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &StandardShadowsocksServer{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动标准Shadowsocks服务器
func (s *StandardShadowsocksServer) Start() error {
	// 创建加密器
	cipher, err := core.PickCipher(s.config.Method, nil, s.config.Password)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	// 启动监听
	addr := fmt.Sprintf(":%d", s.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}
	s.listener = listener

	s.logger.Infof("Standard Shadowsocks server listening on %s", addr)
	s.logger.Infof("Method: %s, Password: %s", s.config.Method, s.config.Password)

	// 接受连接
	go s.acceptConnections(cipher)

	return nil
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

	// 连接目标服务器
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
