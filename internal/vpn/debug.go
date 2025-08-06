package vpn

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// ConnectionDiagnostics 连接诊断工具
type ConnectionDiagnostics struct {
	logger *logrus.Logger
}

// NewConnectionDiagnostics 创建连接诊断工具
func NewConnectionDiagnostics(logger *logrus.Logger) *ConnectionDiagnostics {
	return &ConnectionDiagnostics{
		logger: logger,
	}
}

// DiagnoseConnection 诊断连接问题
func (cd *ConnectionDiagnostics) DiagnoseConnection(clientIP string, errorMsg string) {
	cd.logger.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"error":     errorMsg,
		"timestamp": time.Now().UTC(),
	}).Warn("Connection diagnosis triggered")

	// 分析错误类型
	if cd.isEOFError(errorMsg) {
		cd.logger.WithFields(logrus.Fields{
			"client_ip":  clientIP,
			"issue":      "unexpected_eof",
			"suggestion": "Client connection was closed unexpectedly. This may be due to network issues, client timeout, or malformed requests.",
		}).Info("EOF error detected")
	}

	if cd.isTimeoutError(errorMsg) {
		cd.logger.WithFields(logrus.Fields{
			"client_ip":  clientIP,
			"issue":      "timeout",
			"suggestion": "Connection timed out. Consider increasing timeout values or checking network stability.",
		}).Info("Timeout error detected")
	}

	if cd.isLengthError(errorMsg) {
		cd.logger.WithFields(logrus.Fields{
			"client_ip":  clientIP,
			"issue":      "invalid_length",
			"suggestion": "Invalid data length received. This may indicate a protocol mismatch or corrupted data.",
		}).Info("Length error detected")
	}
}

// DiagnoseDecryptionError 诊断解密错误
func (cd *ConnectionDiagnostics) DiagnoseDecryptionError(clientIP string, errorMsg string, dataLength int) {
	cd.logger.WithFields(logrus.Fields{
		"client_ip":  clientIP,
		"error":      errorMsg,
		"data_length": dataLength,
		"timestamp":  time.Now().UTC(),
	}).Error("Decryption error detected")

	// 分析可能的错误原因
	if dataLength == 0 {
		cd.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"issue":     "empty_data",
			"suggestion": "Received empty data. Check if client is sending data correctly.",
		}).Warn("Empty data received")
	} else if dataLength < 16 { // 最小AEAD数据长度
		cd.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"issue":     "data_too_short",
			"suggestion": "Data too short for valid encryption. Check protocol compatibility.",
		}).Warn("Data too short for decryption")
	} else {
		cd.logger.WithFields(logrus.Fields{
			"client_ip": clientIP,
			"issue":     "decryption_failed",
			"suggestion": "Decryption failed. Check password, encryption method, and data integrity.",
		}).Warn("Decryption failed")
	}
}

// LogPanicRecovery 记录panic恢复信息
func (cd *ConnectionDiagnostics) LogPanicRecovery(clientIP string, panic interface{}, dataLength int) {
	cd.logger.WithFields(logrus.Fields{
		"client_ip":  clientIP,
		"panic":      panic,
		"data_length": dataLength,
		"timestamp":  time.Now().UTC(),
	}).Error("Panic recovered during decryption")

	// 提供具体的建议
	cd.logger.WithFields(logrus.Fields{
		"client_ip": clientIP,
		"action":    "immediate",
		"suggestion": "Server panic recovered. Consider restarting the service and checking client compatibility.",
	}).Warn("Server stability concern")
}

// isEOFError 检查是否是EOF错误
func (cd *ConnectionDiagnostics) isEOFError(errorMsg string) bool {
	return stringContains(errorMsg, "EOF") || stringContains(errorMsg, "connection closed")
}

// isTimeoutError 检查是否是超时错误
func (cd *ConnectionDiagnostics) isTimeoutError(errorMsg string) bool {
	return stringContains(errorMsg, "timeout") || stringContains(errorMsg, "deadline exceeded")
}

// isLengthError 检查是否是长度错误
func (cd *ConnectionDiagnostics) isLengthError(errorMsg string) bool {
	return stringContains(errorMsg, "invalid") && stringContains(errorMsg, "length")
}

// TestConnection 测试连接
func (cd *ConnectionDiagnostics) TestConnection(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", addr, err)
	}
	defer conn.Close()

	cd.logger.WithFields(logrus.Fields{
		"host": host,
		"port": port,
	}).Info("Connection test successful")

	return nil
}

// GetConnectionStats 获取连接统计信息
func (cd *ConnectionDiagnostics) GetConnectionStats() map[string]interface{} {
	// 这里可以添加连接统计信息的收集
	// 例如：成功连接数、失败连接数、平均响应时间等
	return map[string]interface{}{
		"timestamp": time.Now().UTC(),
		"status":    "monitoring_enabled",
	}
}

// stringContains 检查字符串是否包含子字符串
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
