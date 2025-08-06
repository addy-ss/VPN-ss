package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

type ClientConfig struct {
	ServerHost string
	ServerPort int
	LocalPort  int
	Password   string
	Method     string
	Timeout    int
}

type ProxyClient struct {
	config *ClientConfig
	logger *logrus.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

func main() {
	// 解析命令行参数
	var (
		serverHost = flag.String("server", "127.0.0.1", "服务器地址")
		serverPort = flag.Int("port", 8388, "服务器端口")
		localPort  = flag.Int("local", 1080, "本地监听端口")
		password   = flag.String("password", "13687401432Fan!", "密码")
		method     = flag.String("method", "aes-256-gcm", "加密方法")
		timeout    = flag.Int("timeout", 300, "超时时间(秒)")
	)
	flag.Parse()

	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// 创建配置
	config := &ClientConfig{
		ServerHost: *serverHost,
		ServerPort: *serverPort,
		LocalPort:  *localPort,
		Password:   *password,
		Method:     *method,
		Timeout:    *timeout,
	}

	// 创建客户端
	client := NewProxyClient(config, logger)

	// 启动客户端
	if err := client.Start(); err != nil {
		logger.Fatalf("启动客户端失败: %v", err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭客户端...")
	client.Stop()
	logger.Info("客户端已关闭")
}

func NewProxyClient(config *ClientConfig, logger *logrus.Logger) *ProxyClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProxyClient{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *ProxyClient) Start() error {
	// 启动本地监听
	addr := fmt.Sprintf(":%d", c.config.LocalPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听端口失败 %s: %v", addr, err)
	}
	defer listener.Close()

	c.logger.Infof("客户端代理启动，监听端口 %s", addr)
	c.logger.Infof("所有请求将通过 %s:%d 转发", c.config.ServerHost, c.config.ServerPort)

	// 接受连接
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			conn, err := listener.Accept()
			if err != nil {
				c.logger.Errorf("接受连接失败: %v", err)
				continue
			}
			go c.handleConnection(conn)
		}
	}
}

func (c *ProxyClient) Stop() {
	c.cancel()
}

func (c *ProxyClient) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	clientIP := localConn.RemoteAddr().String()
	c.logger.Infof("新连接来自 %s", clientIP)

	// 处理SOCKS5握手
	if err := c.handleSOCKS5Handshake(localConn); err != nil {
		c.logger.Errorf("SOCKS5握手失败: %v", err)
		return
	}
	c.logger.Infof("SOCKS5握手成功")

	// 读取SOCKS5请求
	target, err := c.readSOCKS5Request(localConn)
	if err != nil {
		c.logger.Errorf("读取SOCKS5请求失败: %v", err)
		return
	}
	c.logger.Infof("目标地址: %s", target)

	// 连接服务器
	serverAddr := fmt.Sprintf("%s:%d", c.config.ServerHost, c.config.ServerPort)
	c.logger.Infof("正在连接服务器: %s", serverAddr)
	serverConn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		c.logger.Errorf("连接服务器失败: %v", err)
		return
	}
	defer serverConn.Close()
	c.logger.Infof("服务器连接成功")

	// 处理代理连接
	if err := c.handleProxy(localConn, serverConn, target); err != nil {
		c.logger.Errorf("处理代理连接失败: %v", err)
	}
}

func (c *ProxyClient) handleSOCKS5Handshake(conn net.Conn) error {
	// 读取SOCKS5版本和方法数量
	buf := make([]byte, 2)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return fmt.Errorf("读取SOCKS5版本失败: %v", err)
	}

	if buf[0] != 0x05 {
		return fmt.Errorf("不支持的SOCKS版本: %d", buf[0])
	}

	methodCount := buf[1]
	methods := make([]byte, methodCount)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("读取认证方法失败: %v", err)
	}

	// 检查是否支持无认证方法
	supported := false
	for _, method := range methods {
		if method == 0x00 {
			supported = true
			break
		}
	}

	if !supported {
		// 回复不支持
		conn.Write([]byte{0x05, 0xFF})
		return fmt.Errorf("不支持的认证方法")
	}

	// 回复选择无认证方法
	conn.Write([]byte{0x05, 0x00})
	return nil
}

func (c *ProxyClient) readSOCKS5Request(conn net.Conn) (string, error) {
	// 读取SOCKS5请求头
	buf := make([]byte, 4)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return "", fmt.Errorf("读取SOCKS5请求头失败: %v", err)
	}

	if buf[0] != 0x05 {
		return "", fmt.Errorf("不支持的SOCKS版本: %d", buf[0])
	}

	if buf[1] != 0x01 {
		return "", fmt.Errorf("不支持的SOCKS命令: %d", buf[1])
	}

	// 解析目标地址
	addrType := buf[3]
	var target string

	switch addrType {
	case 0x01: // IPv4
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("读取IPv4地址失败: %v", err)
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return "", fmt.Errorf("读取端口失败: %v", err)
		}
		portNum := binary.BigEndian.Uint16(port)
		target = fmt.Sprintf("%s:%d", net.IP(addr).String(), portNum)

	case 0x03: // 域名
		domainLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLen); err != nil {
			return "", fmt.Errorf("读取域名长度失败: %v", err)
		}
		domain := make([]byte, domainLen[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", fmt.Errorf("读取域名失败: %v", err)
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return "", fmt.Errorf("读取端口失败: %v", err)
		}
		portNum := binary.BigEndian.Uint16(port)
		target = fmt.Sprintf("%s:%d", string(domain), portNum)

	case 0x04: // IPv6
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("读取IPv6地址失败: %v", err)
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return "", fmt.Errorf("读取端口失败: %v", err)
		}
		portNum := binary.BigEndian.Uint16(port)
		target = fmt.Sprintf("[%s]:%d", net.IP(addr).String(), portNum)

	default:
		return "", fmt.Errorf("不支持的地址类型: %d", addrType)
	}

	// 发送SOCKS5成功响应
	response := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	conn.Write(response)

	return target, nil
}

func (c *ProxyClient) constructTargetData(target string) []byte {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		// 如果解析失败，假设是域名
		host = target
		port = "80"
	}

	// 解析端口
	portNum := 80
	if port != "" {
		if p, err := net.LookupPort("tcp", port); err == nil {
			portNum = p
		}
	}

	// 构造地址数据
	var addrData []byte

	// 检查是否为IP地址
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			// IPv4
			addrData = append(addrData, 0x01) // 地址类型
			addrData = append(addrData, ip.To4()...)
		} else {
			// IPv6
			addrData = append(addrData, 0x04) // 地址类型
			addrData = append(addrData, ip...)
		}
	} else {
		// 域名
		addrData = append(addrData, 0x03) // 地址类型
		addrData = append(addrData, byte(len(host)))
		addrData = append(addrData, []byte(host)...)
	}

	// 添加端口
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(portNum))
	addrData = append(addrData, portBytes...)

	return addrData
}

func (c *ProxyClient) handleProxy(localConn, serverConn net.Conn, target string) error {
	c.logger.Infof("开始处理代理连接，目标: %s", target)

	// 创建加密器
	encryptor, err := c.createEncryptor(serverConn)
	if err != nil {
		return fmt.Errorf("创建加密器失败: %v", err)
	}
	c.logger.Infof("加密器创建成功")

	// 构造目标地址数据
	targetData := c.constructTargetData(target)
	c.logger.Infof("目标地址数据长度: %d", len(targetData))

	// 生成nonce并加密目标地址数据
	nonce := make([]byte, encryptor.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("生成nonce失败: %v", err)
	}
	encryptedTarget := encryptor.Seal(nil, nonce, targetData, nil)
	c.logger.Infof("加密后目标地址数据长度: %d", len(encryptedTarget))

	// 发送nonce和加密的目标地址数据
	if _, err := serverConn.Write(nonce); err != nil {
		return fmt.Errorf("发送nonce失败: %v", err)
	}
	length := len(encryptedTarget)
	lengthBuf := []byte{byte(length >> 8), byte(length)}
	if _, err := serverConn.Write(lengthBuf); err != nil {
		return fmt.Errorf("发送长度失败: %v", err)
	}
	if _, err := serverConn.Write(encryptedTarget); err != nil {
		return fmt.Errorf("发送加密目标地址失败: %v", err)
	}
	c.logger.Infof("目标地址数据发送成功")

	// 双向转发数据
	errChan := make(chan error, 2)

	// 从本地到服务器（加密）
	go func() {
		c.logger.Infof("开始本地到服务器数据转发")
		buffer := make([]byte, 8192)
		for {
			n, err := localConn.Read(buffer)
			if err != nil {
				c.logger.Errorf("读取本地数据失败: %v", err)
				errChan <- err
				return
			}

			// 生成nonce并加密数据
			nonce := make([]byte, encryptor.NonceSize())
			if _, err := rand.Read(nonce); err != nil {
				c.logger.Errorf("生成nonce失败: %v", err)
				errChan <- err
				return
			}
			encryptedData := encryptor.Seal(nil, nonce, buffer[:n], nil)

			// 写入nonce
			if _, err := serverConn.Write(nonce); err != nil {
				c.logger.Errorf("写入nonce失败: %v", err)
				errChan <- err
				return
			}

			// 写入长度
			length := len(encryptedData)
			lengthBuf := []byte{byte(length >> 8), byte(length)}
			if _, err := serverConn.Write(lengthBuf); err != nil {
				c.logger.Errorf("写入数据长度失败: %v", err)
				errChan <- err
				return
			}

			// 写入加密数据
			if _, err := serverConn.Write(encryptedData); err != nil {
				c.logger.Errorf("写入加密数据失败: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// 从服务器到本地（解密）
	go func() {
		c.logger.Infof("开始服务器到本地数据转发")
		for {
			// 读取nonce
			nonce := make([]byte, encryptor.NonceSize())
			if _, err := io.ReadFull(serverConn, nonce); err != nil {
				c.logger.Errorf("读取nonce失败: %v", err)
				errChan <- err
				return
			}

			// 读取长度
			lengthBuf := make([]byte, 2)
			if _, err := io.ReadFull(serverConn, lengthBuf); err != nil {
				c.logger.Errorf("读取数据长度失败: %v", err)
				errChan <- err
				return
			}

			length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
			if length <= 0 || length > 65535 {
				c.logger.Errorf("无效的加密数据长度: %d", length)
				errChan <- fmt.Errorf("无效的加密数据长度: %d", length)
				return
			}

			// 读取加密数据
			encryptedData := make([]byte, length)
			if _, err := io.ReadFull(serverConn, encryptedData); err != nil {
				c.logger.Errorf("读取加密数据失败: %v", err)
				errChan <- err
				return
			}

			// 解密数据
			decryptedData, err := encryptor.Open(nil, nonce, encryptedData, nil)
			if err != nil {
				c.logger.Errorf("解密数据失败: %v", err)
				errChan <- fmt.Errorf("解密数据失败: %v", err)
				return
			}

			// 写入本地连接
			if _, err := localConn.Write(decryptedData); err != nil {
				c.logger.Errorf("写入本地数据失败: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出错
	selectErr := <-errChan
	c.logger.Infof("代理连接结束: %v", selectErr)
	return selectErr
}

func (c *ProxyClient) createEncryptor(conn net.Conn) (cipher.AEAD, error) {
	// 生成随机salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("生成salt失败: %v", err)
	}

	// 写入salt到服务器
	if _, err := conn.Write(salt); err != nil {
		return nil, fmt.Errorf("写入salt失败: %v", err)
	}

	// 生成密钥
	key := pbkdf2.Key([]byte(c.config.Password), salt, 10000, 32, sha256.New)

	// 根据配置选择加密方法
	switch c.config.Method {
	case "chacha20-poly1305":
		return chacha20poly1305.New(key)
	case "aes-256-gcm":
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	default:
		return nil, fmt.Errorf("不支持的加密方法: %s", c.config.Method)
	}
}
