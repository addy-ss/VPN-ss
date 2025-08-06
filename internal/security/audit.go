package security

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// AuditEvent 审计事件
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	EventType string                 `json:"event_type"`
	Username  string                 `json:"username,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Action    string                 `json:"action,omitempty"`
	Status    string                 `json:"status"` // success, failure, warning
	Details   map[string]interface{} `json:"details,omitempty"`
	RiskLevel string                 `json:"risk_level"` // low, medium, high, critical
	SessionID string                 `json:"session_id,omitempty"`
}

// AuditLogger 审计日志记录器
type AuditLogger struct {
	logger *logrus.Logger
	events chan *AuditEvent
}

// NewAuditLogger 创建审计日志记录器
func NewAuditLogger(logger *logrus.Logger) *AuditLogger {
	al := &AuditLogger{
		logger: logger,
		events: make(chan *AuditEvent, 1000), // 缓冲通道
	}

	// 启动事件处理协程
	go al.processEvents()

	return al
}

// LogEvent 记录审计事件
func (al *AuditLogger) LogEvent(event *AuditEvent) {
	select {
	case al.events <- event:
		// 事件已发送到通道
	default:
		// 通道已满，记录警告
		al.logger.Warn("Audit event channel is full, dropping event")
	}
}

// LogLogin 记录登录事件
func (al *AuditLogger) LogLogin(username, ipAddress, userAgent string, success bool) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "login",
		Username:  username,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Action:    "login",
		Status:    "success",
		RiskLevel: "low",
	}

	if !success {
		event.Status = "failure"
		event.RiskLevel = "medium"
		event.Details = map[string]interface{}{
			"reason": "invalid_credentials",
		}
	}

	al.LogEvent(event)
}

// LogLogout 记录登出事件
func (al *AuditLogger) LogLogout(username, ipAddress string) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "logout",
		Username:  username,
		IPAddress: ipAddress,
		Action:    "logout",
		Status:    "success",
		RiskLevel: "low",
	}

	al.LogEvent(event)
}

// LogVPNStart 记录VPN启动事件
func (al *AuditLogger) LogVPNStart(username, ipAddress string, config map[string]interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "vpn_start",
		Username:  username,
		IPAddress: ipAddress,
		Resource:  "vpn_service",
		Action:    "start",
		Status:    "success",
		RiskLevel: "medium",
		Details:   config,
	}

	al.LogEvent(event)
}

// LogVPNStop 记录VPN停止事件
func (al *AuditLogger) LogVPNStop(username, ipAddress string) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "vpn_stop",
		Username:  username,
		IPAddress: ipAddress,
		Resource:  "vpn_service",
		Action:    "stop",
		Status:    "success",
		RiskLevel: "low",
	}

	al.LogEvent(event)
}

// LogAccessDenied 记录访问被拒绝事件
func (al *AuditLogger) LogAccessDenied(username, ipAddress, resource, reason string) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "access_denied",
		Username:  username,
		IPAddress: ipAddress,
		Resource:  resource,
		Action:    "access",
		Status:    "failure",
		RiskLevel: "high",
		Details: map[string]interface{}{
			"reason": reason,
		},
	}

	al.LogEvent(event)
}

// LogSuspiciousActivity 记录可疑活动
func (al *AuditLogger) LogSuspiciousActivity(ipAddress, activity string, details map[string]interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "suspicious_activity",
		IPAddress: ipAddress,
		Action:    activity,
		Status:    "warning",
		RiskLevel: "high",
		Details:   details,
	}

	al.LogEvent(event)
}

// LogConfigurationChange 记录配置变更
func (al *AuditLogger) LogConfigurationChange(username, ipAddress, configType string, oldValue, newValue interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "config_change",
		Username:  username,
		IPAddress: ipAddress,
		Resource:  "configuration",
		Action:    "modify",
		Status:    "success",
		RiskLevel: "medium",
		Details: map[string]interface{}{
			"config_type": configType,
			"old_value":   oldValue,
			"new_value":   newValue,
		},
	}

	al.LogEvent(event)
}

// LogSecurityAlert 记录安全告警
func (al *AuditLogger) LogSecurityAlert(alertType, description string, details map[string]interface{}) {
	event := &AuditEvent{
		Timestamp: time.Now(),
		EventType: "security_alert",
		Action:    alertType,
		Status:    "warning",
		RiskLevel: "critical",
		Details: map[string]interface{}{
			"description": description,
			"details":     details,
		},
	}

	al.LogEvent(event)
}

// processEvents 处理审计事件
func (al *AuditLogger) processEvents() {
	for event := range al.events {
		// 将事件转换为JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			al.logger.Errorf("Failed to marshal audit event: %v", err)
			continue
		}

		// 根据风险级别选择日志级别
		var logLevel logrus.Level
		switch event.RiskLevel {
		case "critical":
			logLevel = logrus.ErrorLevel
		case "high":
			logLevel = logrus.WarnLevel
		case "medium":
			logLevel = logrus.InfoLevel
		default:
			logLevel = logrus.DebugLevel
		}

		// 记录审计日志
		al.logger.WithFields(logrus.Fields{
			"audit_event": string(eventJSON),
			"event_type":  event.EventType,
			"risk_level":  event.RiskLevel,
		}).Log(logLevel, fmt.Sprintf("Audit: %s - %s", event.EventType, event.Action))
	}
}

// GetClientIP 获取客户端IP地址
func GetClientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// IsPrivateIP 检查是否为私有IP地址
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 检查私有IP范围
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// ValidateIPAddress 验证IP地址格式
func ValidateIPAddress(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}
