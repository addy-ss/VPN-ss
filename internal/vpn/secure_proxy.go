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

	// 设置超时
	timeout := time.Duration(p.config.Timeout) * time.Second
	conn.SetDeadline(time.Now().Add(timeout))

	// 处理加密代理连接
	if err := p.handleSecureProxy(conn, clientIP); err != nil {
		p.logger.Errorf("Failed to handle secure proxy connection: %v", err)
		// 记录可疑活动
		p.auditLogger.LogSuspiciousActivity(clientIP, "proxy_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (p *SecureProxyServer) handleSecureProxy(conn net.Conn, clientIP string) error {
	// 读取加密的目标地址
	encryptedTarget, err := p.readEncryptedTarget(conn)
	if err != nil {
		return fmt.Errorf("failed to read encrypted target: %v", err)
	}

	// 解密目标地址
	target, err := p.cryptoManager.Decrypt(encryptedTarget)
	if err != nil {
		return fmt.Errorf("failed to decrypt target: %v", err)
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
	// 读取加密数据长度
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}

	length := int(lengthBuf[0])<<24 | int(lengthBuf[1])<<16 | int(lengthBuf[2])<<8 | int(lengthBuf[3])
	if length <= 0 || length > 4096 { // 限制最大长度
		return nil, fmt.Errorf("invalid encrypted data length: %d", length)
	}

	// 读取加密数据
	encryptedData := make([]byte, length)
	if _, err := io.ReadFull(conn, encryptedData); err != nil {
		return nil, err
	}

	return encryptedData, nil
}

func (p *SecureProxyServer) forwardEncrypted(src, dst net.Conn, clientIP string) error {
	errChan := make(chan error, 2)

	// 从源到目标（解密）
	go func() {
		buffer := make([]byte, 4096)
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
		buffer := make([]byte, 4096)
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
