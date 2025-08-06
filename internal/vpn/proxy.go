package vpn

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

type ProxyServer struct {
	config   *Config
	listener net.Listener
	logger   *logrus.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

type Config struct {
	Method   string
	Password string
	Port     int
	Timeout  int
}

func NewProxyServer(config *Config, logger *logrus.Logger) *ProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProxyServer{
		config: config,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (p *ProxyServer) Start() error {
	// 启动监听
	addr := fmt.Sprintf(":%d", p.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}
	p.listener = listener

	p.logger.Infof("Shadowsocks server listening on %s", addr)

	// 接受连接
	go p.acceptConnections()

	return nil
}

func (p *ProxyServer) Stop() {
	p.cancel()
	if p.listener != nil {
		p.listener.Close()
	}
}

func (p *ProxyServer) acceptConnections() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			conn, err := p.listener.Accept()
			if err != nil {
				p.logger.Errorf("Failed to accept connection: %v", err)
				continue
			}
			go p.handleConnection(conn)
		}
	}
}

func (p *ProxyServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientIP := conn.RemoteAddr().String()
	p.logger.Infof("New connection from %s", clientIP)

	// 设置连接属性（如果是TCP连接）
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetLinger(0) // 立即关闭连接
	}

	// 设置初始超时时间（给客户端更多时间发送数据）
	initialTimeout := time.Duration(30) * time.Second
	conn.SetDeadline(time.Now().Add(initialTimeout))

	// 处理代理连接
	if err := p.handleProxy(conn); err != nil {
		// 记录详细的错误信息
		p.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"error":     err.Error(),
			"time":      time.Now().UTC(),
		}).Errorf("Failed to handle proxy connection from %s: %v", clientIP, err)
	} else {
		p.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"duration":  time.Since(time.Now()),
		}).Infof("Proxy connection from %s completed successfully", clientIP)
	}
}

func (p *ProxyServer) handleProxy(conn net.Conn) error {
	// 设置最大重试次数
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// 创建解密器
		decryptor, err := p.createDecryptor(conn)
		if err != nil {
			lastErr = fmt.Errorf("failed to create decryptor (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				p.logger.Warnf("Retrying connection (attempt %d/%d): %v", attempt, maxRetries, err)
				continue
			}
			return lastErr
		}

		// 读取并解密目标地址
		target, err := p.readDecryptedTarget(conn, decryptor)
		if err != nil {
			lastErr = fmt.Errorf("failed to read target (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				p.logger.Warnf("Retrying connection (attempt %d/%d): %v", attempt, maxRetries, err)
				continue
			}
			return lastErr
		}

		// 连接目标服务器
		targetConn, err := net.DialTimeout("tcp", target, 10*time.Second)
		if err != nil {
			lastErr = fmt.Errorf("failed to connect to target %s (attempt %d/%d): %v", target, attempt, maxRetries, err)
			if attempt < maxRetries {
				p.logger.Warnf("Retrying connection (attempt %d/%d): %v", attempt, maxRetries, err)
				continue
			}
			return lastErr
		}
		defer targetConn.Close()

		// 双向转发数据
		return p.forwardEncrypted(conn, targetConn, decryptor)
	}

	return lastErr
}

func (p *ProxyServer) createDecryptor(conn net.Conn) (cipher.AEAD, error) {
	// 读取salt
	salt := make([]byte, 32)
	if _, err := io.ReadFull(conn, salt); err != nil {
		return nil, fmt.Errorf("failed to read salt: %v", err)
	}
	p.logger.Infof("收到salt: %x", salt[:8]) // 只显示前8字节用于调试

	// 生成密钥
	key := pbkdf2.Key([]byte(p.config.Password), salt, 10000, 32, sha256.New)

	// 根据配置选择加密方法
	switch p.config.Method {
	case "chacha20-poly1305":
		return chacha20poly1305.New(key)
	case "aes-256-gcm":
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(block)
	default:
		return nil, fmt.Errorf("unsupported cipher method: %s", p.config.Method)
	}
}

func (p *ProxyServer) readDecryptedTarget(conn net.Conn, decryptor cipher.AEAD) (string, error) {
	// 设置读取超时，防止长时间等待
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 读取加密数据长度
	lengthBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		if err == io.EOF {
			return "", fmt.Errorf("connection closed by client before reading length: %v", err)
		}
		// 检查是否是超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "", fmt.Errorf("timeout reading length field: %v", err)
		}
		return "", fmt.Errorf("failed to read length field: %v", err)
	}

	length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
	if length <= 0 || length > 65535 { // 增加最大长度限制到65535字节
		return "", fmt.Errorf("invalid encrypted data length: %d", length)
	}

	// 重置读取超时，给更多时间读取数据
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 读取加密数据
	encryptedData := make([]byte, length)
	bytesRead := 0
	for bytesRead < length {
		n, err := conn.Read(encryptedData[bytesRead:])
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("connection closed by client while reading encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
			}
			// 检查是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return "", fmt.Errorf("timeout reading encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
			}
			return "", fmt.Errorf("failed to read encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
		}
		bytesRead += n
	}

	// 重置读取超时，恢复正常模式
	conn.SetReadDeadline(time.Time{})

	// 验证加密数据的完整性
	if len(encryptedData) == 0 {
		return "", fmt.Errorf("received empty encrypted data")
	}

	// 检查数据长度是否合理（至少包含nonce和最小密文）
	minLength := decryptor.NonceSize() + decryptor.Overhead()
	if len(encryptedData) < minLength {
		return "", fmt.Errorf("encrypted data too short: got %d bytes, need at least %d bytes", len(encryptedData), minLength)
	}

	// 添加panic恢复
	defer func() {
		if r := recover(); r != nil {
			p.logger.WithFields(logrus.Fields{
				"panic": r,
				"data_length": len(encryptedData),
				"min_length": minLength,
			}).Error("Panic during decryption, this may indicate corrupted data or protocol mismatch")
		}
	}()

	// 提取nonce
	nonceSize := decryptor.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", fmt.Errorf("encrypted data too short for nonce: got %d bytes, need at least %d bytes", len(encryptedData), nonceSize)
	}
	
	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]
	
	// 解密数据
	decryptedData, err := decryptor.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data (length: %d): %v", len(encryptedData), err)
	}

	// 解析目标地址
	if len(decryptedData) < 3 {
		return "", fmt.Errorf("invalid decrypted data length")
	}

	addrType := decryptedData[0]
	var target string

	switch addrType {
	case 1: // IPv4
		if len(decryptedData) < 7 {
			return "", fmt.Errorf("invalid IPv4 address length")
		}
		addr := decryptedData[1:5]
		target = net.IP(addr).String()
	case 3: // Domain name
		if len(decryptedData) < 4 {
			return "", fmt.Errorf("invalid domain name length")
		}
		domainLen := decryptedData[1]
		if len(decryptedData) < int(domainLen)+4 {
			return "", fmt.Errorf("invalid domain name data length")
		}
		domain := decryptedData[2 : 2+domainLen]
		target = string(domain)
	case 4: // IPv6
		if len(decryptedData) < 19 {
			return "", fmt.Errorf("invalid IPv6 address length")
		}
		addr := decryptedData[1:17]
		target = net.IP(addr).String()
	default:
		return "", fmt.Errorf("unsupported address type: %d", addrType)
	}

	// 读取端口
	portStart := 0
	switch addrType {
	case 1: // IPv4
		portStart = 5
	case 3: // Domain name
		portStart = 2 + int(decryptedData[1])
	case 4: // IPv6
		portStart = 17
	}

	if len(decryptedData) < portStart+2 {
		return "", fmt.Errorf("invalid port data length")
	}

	port := int(decryptedData[portStart])<<8 | int(decryptedData[portStart+1])

	return fmt.Sprintf("%s:%d", target, port), nil
}

func (p *ProxyServer) forwardEncrypted(src, dst net.Conn, decryptor cipher.AEAD) error {
	errChan := make(chan error, 2)

	// 从源到目标（解密）
	go func() {
		for {
			// 读取长度
			lengthBuf := make([]byte, 2)
			if _, err := io.ReadFull(src, lengthBuf); err != nil {
				errChan <- err
				return
			}

			length := int(lengthBuf[0])<<8 | int(lengthBuf[1])
			if length <= 0 || length > 65535 { // 增加最大长度限制到65535字节
				errChan <- fmt.Errorf("invalid encrypted data length: %d", length)
				return
			}

			// 读取加密数据
			encryptedData := make([]byte, length)
			if _, err := io.ReadFull(src, encryptedData); err != nil {
				errChan <- err
				return
			}

			// 提取nonce
			nonceSize := decryptor.NonceSize()
			if len(encryptedData) < nonceSize {
				errChan <- fmt.Errorf("encrypted data too short for nonce: got %d bytes, need at least %d bytes", len(encryptedData), nonceSize)
				return
			}
			
			nonce := encryptedData[:nonceSize]
			ciphertext := encryptedData[nonceSize:]
			
			// 解密数据
			decryptedData, err := decryptor.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				errChan <- fmt.Errorf("failed to decrypt data: %v", err)
				return
			}

			// 写入目标
			if _, err := dst.Write(decryptedData); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 从目标到源（加密）
	go func() {
		buffer := make([]byte, 8192) // 增加缓冲区大小到8192字节
		for {
			n, err := dst.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}

			// 加密数据
			encryptedData := decryptor.Seal(nil, nil, buffer[:n], nil)

			// 写入长度
			length := len(encryptedData)
			lengthBuf := []byte{byte(length >> 8), byte(length)}
			if _, err := src.Write(lengthBuf); err != nil {
				errChan <- err
				return
			}

			// 写入加密数据
			if _, err := src.Write(encryptedData); err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出错
	err := <-errChan
	return err
}

// 生成Shadowsocks配置
func (p *ProxyServer) GenerateConfig() string {
	// 生成随机密码
	password := make([]byte, 32)
	rand.Read(password)
	passwordStr := base64.StdEncoding.EncodeToString(password)

	config := fmt.Sprintf(`{
		"server": "0.0.0.0",
		"server_port": %d,
		"password": "%s",
		"method": "%s",
		"timeout": %d
	}`, p.config.Port, passwordStr, p.config.Method, p.config.Timeout)

	return config
}
