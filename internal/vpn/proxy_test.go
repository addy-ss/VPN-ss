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
