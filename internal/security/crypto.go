package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
)

// CryptoManager 加密管理器
type CryptoManager struct {
	key        []byte
	aead       cipher.AEAD
	method     string
	salt       []byte
	iterations int
}

// NewCryptoManager 创建加密管理器
func NewCryptoManager(password string, method string) (*CryptoManager, error) {
	cm := &CryptoManager{
		method:     method,
		iterations: 10000, // PBKDF2迭代次数
	}

	// 生成随机盐值
	cm.salt = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, cm.salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	// 使用PBKDF2派生密钥
	cm.key = pbkdf2.Key([]byte(password), cm.salt, cm.iterations, 32, sha256.New)

	// 创建AEAD加密器
	switch method {
	case "aes-256-gcm":
		block, err := aes.NewCipher(cm.key)
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher: %v", err)
		}
		cm.aead, err = cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCM: %v", err)
		}
	case "chacha20-poly1305":
		var err error
		cm.aead, err = chacha20poly1305.New(cm.key)
		if err != nil {
			return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported encryption method: %s", method)
	}

	return cm, nil
}

// Encrypt 加密数据
func (cm *CryptoManager) Encrypt(plaintext []byte) ([]byte, error) {
	// 生成随机nonce
	nonce := make([]byte, cm.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	// 加密数据
	ciphertext := cm.aead.Seal(nil, nonce, plaintext, nil)

	// 返回格式: [nonce][ciphertext]
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)

	return result, nil
}

// Decrypt 解密数据
func (cm *CryptoManager) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < cm.aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// 分离nonce和密文
	nonce := ciphertext[:cm.aead.NonceSize()]
	ciphertext = ciphertext[cm.aead.NonceSize():]

	// 解密数据
	plaintext, err := cm.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return plaintext, nil
}

// GetSalt 获取盐值（用于密钥交换）
func (cm *CryptoManager) GetSalt() []byte {
	return cm.salt
}

// GetMethod 获取加密方法
func (cm *CryptoManager) GetMethod() string {
	return cm.method
}

// GenerateStrongPassword 生成强密码
func GenerateStrongPassword(length int) (string, error) {
	if length < 16 {
		length = 32 // 最小长度
	}

	// 包含大小写字母、数字和特殊字符
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := 0; i < length; i++ {
		randomByte := make([]byte, 1)
		if _, err := io.ReadFull(rand.Reader, randomByte); err != nil {
			return "", fmt.Errorf("failed to generate random byte: %v", err)
		}
		password[i] = chars[int(randomByte[0])%len(chars)]
	}

	return string(password), nil
}

// HashPassword 哈希密码
func HashPassword(password string, salt []byte) string {
	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash)
}

// VerifyPassword 验证密码
func VerifyPassword(password string, salt []byte, hashedPassword string) bool {
	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	expectedHash := base64.StdEncoding.EncodeToString(hash)
	return hashedPassword == expectedHash
}
