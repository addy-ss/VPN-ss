package vpn

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"time"

	"vps/internal/security"

	"github.com/sirupsen/logrus"
)

// SecureProxyServer 安全代理服务器
type SecureProxyServer struct {
	config        *Config
	listener      net.Listener
	logger        *logrus.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	cryptoManager *security.CryptoManager
	auditLogger   *security.AuditLogger
}

// NewSecureProxyServer 创建安全代理服务器
func NewSecureProxyServer(config *Config, logger *logrus.Logger, auditLogger *security.AuditLogger) *SecureProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &SecureProxyServer{
		config:      config,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		auditLogger: auditLogger,
	}
}

func (p *SecureProxyServer) Start() error {
	// 创建加密管理器
	cryptoManager, err := security.NewCryptoManager(p.config.Password, p.config.Method)
	if err != nil {
		return fmt.Errorf("failed to create crypto manager: %v", err)
	}
	p.cryptoManager = cryptoManager

	// 启动监听
	addr := fmt.Sprintf(":%d", p.config.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}
	p.listener = listener

	p.logger.Infof("Secure Shadowsocks server listening on %s", addr)

	// 接受连接
	go p.acceptConnections()

	return nil
}

func (p *SecureProxyServer) Stop() {
	p.cancel()
	if p.listener != nil {
		p.listener.Close()
	}
}

func (p *SecureProxyServer) acceptConnections() {
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
			go p.handleSecureConnection(conn)
		}
	}
}

func (p *SecureProxyServer) handleSecureConnection(conn net.Conn) {
	defer conn.Close()

	clientIP := security.GetClientIP(conn.RemoteAddr().String())
	p.logger.Infof("New secure connection from %s", clientIP)

	// 设置连接属性（如果是TCP连接）
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetLinger(0) // 立即关闭连接
	}

	// 设置初始超时时间（给客户端更多时间发送数据）
	initialTimeout := time.Duration(30) * time.Second
	conn.SetDeadline(time.Now().Add(initialTimeout))

	// 处理加密代理连接
	if err := p.handleSecureProxy(conn, clientIP); err != nil {
		// 记录详细的错误信息
		p.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"error":     err.Error(),
			"time":      time.Now().UTC(),
		}).Errorf("Failed to handle secure proxy connection from %s: %v", clientIP, err)

		// 记录可疑活动
		p.auditLogger.LogSuspiciousActivity(clientIP, "proxy_error", map[string]interface{}{
			"error": err.Error(),
			"time":  time.Now().UTC(),
		})
	} else {
		p.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"duration":  time.Since(time.Now()),
		}).Infof("Secure proxy connection from %s completed successfully", clientIP)
	}
}

func (p *SecureProxyServer) handleSecureProxy(conn net.Conn, clientIP string) error {
	// 读取加密的目标地址
	encryptedTarget, err := p.readEncryptedTarget(conn)
	if err != nil {
		return fmt.Errorf("failed to read encrypted target: %v", err)
	}

	// 重置读取超时，恢复正常模式
	conn.SetReadDeadline(time.Time{})

	// 验证加密数据的完整性
	if len(encryptedTarget) == 0 {
		return fmt.Errorf("received empty encrypted data")
	}

	// 检查数据长度是否合理（至少包含nonce和最小密文）
	minLength := p.cryptoManager.GetNonceSize() + p.cryptoManager.GetOverhead()
	if len(encryptedTarget) < minLength {
		return fmt.Errorf("encrypted data too short: got %d bytes, need at least %d bytes", len(encryptedTarget), minLength)
	}

	// 添加panic恢复
	defer func() {
		if r := recover(); r != nil {
			p.logger.WithFields(logrus.Fields{
				"panic": r,
				"data_length": len(encryptedTarget),
				"min_length": minLength,
			}).Error("Panic during decryption, this may indicate corrupted data or protocol mismatch")
		}
	}()

	// 解密目标地址
	target, err := p.cryptoManager.Decrypt(encryptedTarget)
	if err != nil {
		return fmt.Errorf("failed to decrypt target (length: %d): %v", len(encryptedTarget), err)
	}

	// 连接目标服务器
	targetConn, err := net.DialTimeout("tcp", string(target), 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to target %s: %v", string(target), err)
	}
	defer targetConn.Close()

	// 记录VPN连接
	p.auditLogger.LogVPNStart("anonymous", clientIP, map[string]interface{}{
		"target": string(target),
		"method": p.config.Method,
	})

	// 创建加密隧道
	return p.forwardEncrypted(conn, targetConn, clientIP)
}

func (p *SecureProxyServer) readEncryptedTarget(conn net.Conn) ([]byte, error) {
	// 设置读取超时，防止长时间等待
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 读取加密数据长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("connection closed by client before reading length: %v", err)
		}
		// 检查是否是超时错误
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("timeout reading length field: %v", err)
		}
		return nil, fmt.Errorf("failed to read length field: %v", err)
	}

	length := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])
	if length <= 0 || length > 65535 { // 增加最大长度限制到65535字节
		return nil, fmt.Errorf("invalid encrypted data length: %d", length)
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
				return nil, fmt.Errorf("connection closed by client while reading encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
			}
			// 检查是否是超时错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, fmt.Errorf("timeout reading encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
			}
			return nil, fmt.Errorf("failed to read encrypted data (read %d/%d bytes): %v", bytesRead, length, err)
		}
		bytesRead += n
	}

	// 重置读取超时，恢复正常模式
	conn.SetReadDeadline(time.Time{})

	return encryptedData, nil
}

func (p *SecureProxyServer) forwardEncrypted(src, dst net.Conn, clientIP string) error {
	errChan := make(chan error, 2)

	// 从源到目标（解密）
	go func() {
		buffer := make([]byte, 8192) // 增加缓冲区大小到8192字节
		for {
			n, err := src.Read(buffer)
			if err != nil {
				errChan <- err
				return
			}

			// 解密数据
			decrypted, err := p.cryptoManager.Decrypt(buffer[:n])
			if err != nil {
				errChan <- fmt.Errorf("failed to decrypt data: %v", err)
				return
			}

			// 写入目标
			_, err = dst.Write(decrypted)
			if err != nil {
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
			encrypted, err := p.cryptoManager.Encrypt(buffer[:n])
			if err != nil {
				errChan <- fmt.Errorf("failed to encrypt data: %v", err)
				return
			}

			// 写入长度和数据
			length := len(encrypted)
			lengthBuf := []byte{
				byte(length >> 24),
				byte(length >> 16),
				byte(length >> 8),
				byte(length),
			}

			_, err = src.Write(lengthBuf)
			if err != nil {
				errChan <- err
				return
			}

			_, err = src.Write(encrypted)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	// 等待任一方向出错
	err := <-errChan
	return err
}

// GenerateSecureConfig 生成安全配置
func (p *SecureProxyServer) GenerateSecureConfig() string {
	// 生成强密码
	password, err := security.GenerateStrongPassword(32)
	if err != nil {
		p.logger.Errorf("Failed to generate password: %v", err)
		password = "default-secure-password"
	}

	// 生成随机盐值
	salt := make([]byte, 32)
	rand.Read(salt)

	config := fmt.Sprintf(`{
		"server": "0.0.0.0",
		"server_port": %d,
		"password": "%s",
		"method": "%s",
		"timeout": %d,
		"security": {
			"encryption": true,
			"salt": "%s",
			"iterations": 10000
		}
	}`, p.config.Port, password, p.config.Method, p.config.Timeout, base64.StdEncoding.EncodeToString(salt))

	return config
}

// GetSecurityInfo 获取安全信息
func (p *SecureProxyServer) GetSecurityInfo() map[string]interface{} {
	return map[string]interface{}{
		"encryption_enabled": true,
		"method":             p.config.Method,
		"salt":               base64.StdEncoding.EncodeToString(p.cryptoManager.GetSalt()),
		"iterations":         10000,
		"key_derivation":     "PBKDF2-SHA256",
		"aead_mode":          true,
	}
}
