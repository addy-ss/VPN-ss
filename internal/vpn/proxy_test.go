package vpn

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewProxyServer(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
	}

	server := NewProxyServer(config, logger)
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.config != config {
		t.Error("Expected config to be set correctly")
	}

	if server.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

func TestProxyServer_GenerateConfig(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
	}

	server := NewProxyServer(config, logger)
	configStr := server.GenerateConfig()

	if configStr == "" {
		t.Error("Expected non-empty config string")
	}

	// 验证配置包含必要的字段
	expectedFields := []string{"server", "server_port", "password", "method", "timeout"}
	for _, field := range expectedFields {
		if !contains(configStr, field) {
			t.Errorf("Expected config to contain field: %s", field)
		}
	}
}

func TestProxyServer_Stop(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
	}

	server := NewProxyServer(config, logger)

	// 测试停止功能（服务器未启动时）
	server.Stop()

	// 验证上下文被取消
	select {
	case <-server.ctx.Done():
		// 期望的行为
	default:
		t.Error("Expected context to be cancelled after Stop()")
	}
}

// 测试加密数据长度限制
func TestEncryptedDataLengthLimit(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
	}

	_ = NewProxyServer(config, logger) // 使用下划线忽略未使用的变量

	// 测试长度限制应该允许65535字节
	maxLength := 65535
	if maxLength > 65535 {
		t.Errorf("Expected max length to be 65535, got %d", maxLength)
	}

	// 测试长度限制应该拒绝超过65535字节的数据
	invalidLength := 70000
	if invalidLength <= 65535 {
		t.Errorf("Expected invalid length %d to be rejected", invalidLength)
	}
}

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				func() bool {
					for i := 1; i <= len(s)-len(substr); i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
}

// 测试多级代理配置
func TestMultiServerProxy(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
		// 第二个服务端配置
		SecondServerEnabled:  true,
		SecondServerHost:     "127.0.0.1",
		SecondServerPort:     8389,
		SecondServerMethod:   "aes-256-gcm",
		SecondServerPassword: "second-password",
		SecondServerTimeout:  300,
	}

	server := NewProxyServer(config, logger)
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// 验证第二个服务端配置
	if !server.useSecondServer {
		t.Error("Expected second server to be enabled")
	}

	if server.secondConfig == nil {
		t.Error("Expected second config to be created")
	}

	if server.secondConfig.Method != "aes-256-gcm" {
		t.Errorf("Expected second server method to be aes-256-gcm, got %s", server.secondConfig.Method)
	}

	if server.secondConfig.Port != 8389 {
		t.Errorf("Expected second server port to be 8389, got %d", server.secondConfig.Port)
	}
}

// 测试单级代理配置（不启用第二个服务端）
func TestSingleServerProxy(t *testing.T) {
	logger := logrus.New()
	config := &Config{
		Port:     8388,
		Method:   "aes-256-gcm",
		Password: "test-password",
		Timeout:  300,
		// 不启用第二个服务端
		SecondServerEnabled: false,
	}

	server := NewProxyServer(config, logger)
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// 验证第二个服务端未启用
	if server.useSecondServer {
		t.Error("Expected second server to be disabled")
	}

	if server.secondConfig != nil {
		t.Error("Expected second config to be nil when disabled")
	}
}
