package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthManager 认证管理器
type AuthManager struct {
	secretKey []byte
	logger    *logrus.Logger
	users     map[string]*User
}

// User 用户信息
type User struct {
	Username     string    `json:"username"`
	PasswordHash string    `json:"password_hash"`
	Salt         []byte    `json:"salt"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	LastLogin    time.Time `json:"last_login"`
	IsActive     bool      `json:"is_active"`
}

// Claims JWT声明
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewAuthManager 创建认证管理器
func NewAuthManager(secretKey string, logger *logrus.Logger) *AuthManager {
	if secretKey == "" {
		// 生成随机密钥
		key := make([]byte, 32)
		rand.Read(key)
		secretKey = base64.StdEncoding.EncodeToString(key)
	}

	return &AuthManager{
		secretKey: []byte(secretKey),
		logger:    logger,
		users:     make(map[string]*User),
	}
}

// CreateUser 创建用户
func (am *AuthManager) CreateUser(username, password, role string) error {
	if am.users[username] != nil {
		return fmt.Errorf("user already exists")
	}

	// 生成盐值
	salt := make([]byte, 32)
	rand.Read(salt)

	// 哈希密码
	passwordHash := HashPassword(password, salt)

	user := &User{
		Username:     username,
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         role,
		CreatedAt:    time.Now(),
		IsActive:     true,
	}

	am.users[username] = user
	am.logger.Infof("Created user: %s with role: %s", username, role)

	return nil
}

// AuthenticateUser 认证用户
func (am *AuthManager) AuthenticateUser(username, password string) (bool, error) {
	user, exists := am.users[username]
	if !exists {
		return false, fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return false, fmt.Errorf("user account is disabled")
	}

	// 验证密码
	if !VerifyPassword(password, user.Salt, user.PasswordHash) {
		am.logger.Warnf("Failed login attempt for user: %s", username)
		return false, fmt.Errorf("invalid password")
	}

	// 更新最后登录时间
	user.LastLogin = time.Now()
	am.logger.Infof("Successful login for user: %s", username)

	return true, nil
}

// GenerateToken 生成JWT令牌
func (am *AuthManager) GenerateToken(username string) (string, error) {
	user, exists := am.users[username]
	if !exists {
		return "", fmt.Errorf("user not found")
	}

	claims := &Claims{
		Username: username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(am.secretKey)
}

// ValidateToken 验证JWT令牌
func (am *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// 检查用户是否仍然存在且活跃
		user, exists := am.users[claims.Username]
		if !exists || !user.IsActive {
			return nil, fmt.Errorf("user not found or inactive")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// HasPermission 检查权限
func (am *AuthManager) HasPermission(username, permission string) bool {
	user, exists := am.users[username]
	if !exists || !user.IsActive {
		return false
	}

	// 简单的基于角色的权限系统
	switch user.Role {
	case "admin":
		return true // 管理员拥有所有权限
	case "user":
		switch permission {
		case "vpn:read", "vpn:start", "vpn:stop":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// GetUser 获取用户信息
func (am *AuthManager) GetUser(username string) (*User, error) {
	user, exists := am.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateUserRole 更新用户角色
func (am *AuthManager) UpdateUserRole(username, newRole string) error {
	user, exists := am.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Role = newRole
	am.logger.Infof("Updated user role: %s -> %s", username, newRole)
	return nil
}

// DisableUser 禁用用户
func (am *AuthManager) DisableUser(username string) error {
	user, exists := am.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.IsActive = false
	am.logger.Infof("Disabled user: %s", username)
	return nil
}

// EnableUser 启用用户
func (am *AuthManager) EnableUser(username string) error {
	user, exists := am.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.IsActive = true
	am.logger.Infof("Enabled user: %s", username)
	return nil
}

// ListUsers 列出所有用户
func (am *AuthManager) ListUsers() []*User {
	users := make([]*User, 0, len(am.users))
	for _, user := range am.users {
		users = append(users, user)
	}
	return users
}

// ChangePassword 修改密码
func (am *AuthManager) ChangePassword(username, oldPassword, newPassword string) error {
	user, exists := am.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// 验证旧密码
	if !VerifyPassword(oldPassword, user.Salt, user.PasswordHash) {
		return fmt.Errorf("invalid old password")
	}

	// 生成新的盐值和哈希
	newSalt := make([]byte, 32)
	rand.Read(newSalt)
	newPasswordHash := HashPassword(newPassword, newSalt)

	// 更新用户信息
	user.Salt = newSalt
	user.PasswordHash = newPasswordHash

	am.logger.Infof("Password changed for user: %s", username)
	return nil
}
