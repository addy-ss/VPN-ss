package api

import (
	"net/http"

	"vps/internal/vpn"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type VPNHandler struct {
	proxyServer *vpn.ProxyServer
	logger      *logrus.Logger
}

func NewVPNHandler(proxyServer *vpn.ProxyServer, logger *logrus.Logger) *VPNHandler {
	return &VPNHandler{
		proxyServer: proxyServer,
		logger:      logger,
	}
}

// 启动VPN服务
func (h *VPNHandler) StartVPN(c *gin.Context) {
	var request struct {
		Port     int    `json:"port" binding:"required"`
		Method   string `json:"method" binding:"required"`
		Password string `json:"password" binding:"required"`
		Timeout  int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 设置默认值
	if request.Timeout == 0 {
		request.Timeout = 300
	}

	config := &vpn.Config{
		Port:     request.Port,
		Method:   request.Method,
		Password: request.Password,
		Timeout:  request.Timeout,
	}

	h.proxyServer = vpn.NewProxyServer(config, h.logger)

	if err := h.proxyServer.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start VPN server",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "VPN server started successfully",
		"port":    request.Port,
		"method":  request.Method,
	})
}

// 停止VPN服务
func (h *VPNHandler) StopVPN(c *gin.Context) {
	if h.proxyServer != nil {
		h.proxyServer.Stop()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "VPN server stopped successfully",
	})
}

// 获取VPN状态
func (h *VPNHandler) GetVPNStatus(c *gin.Context) {
	status := "stopped"
	if h.proxyServer != nil {
		status = "running"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// 生成Shadowsocks配置
func (h *VPNHandler) GenerateConfig(c *gin.Context) {
	var request struct {
		Port     int    `json:"port"`
		Method   string `json:"method"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request parameters",
		})
		return
	}

	// 设置默认值
	if request.Port == 0 {
		request.Port = 8388
	}
	if request.Method == "" {
		request.Method = "aes-256-gcm"
	}

	config := &vpn.Config{
		Port:     request.Port,
		Method:   request.Method,
		Password: request.Password,
		Timeout:  300,
	}

	tempServer := vpn.NewProxyServer(config, h.logger)
	configStr := tempServer.GenerateConfig()

	c.JSON(http.StatusOK, gin.H{
		"config": configStr,
	})
}

// 获取支持的加密方法
func (h *VPNHandler) GetSupportedMethods(c *gin.Context) {
	methods := []string{
		"aes-256-gcm",
		"chacha20-poly1305",
		"aes-128-gcm",
		"aes-192-gcm",
	}

	c.JSON(http.StatusOK, gin.H{
		"methods": methods,
	})
}

// 健康检查
func (h *VPNHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "vps-vpn",
	})
}

// 诊断连接问题
func (h *VPNHandler) DiagnoseConnection(c *gin.Context) {
	var request struct {
		Host string `json:"host" binding:"required"`
		Port int    `json:"port" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 创建诊断工具
	diagnostics := vpn.NewConnectionDiagnostics(h.logger)

	// 测试连接
	if err := diagnostics.TestConnection(request.Host, request.Port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Connection test failed",
			"details": err.Error(),
			"host":    request.Host,
			"port":    request.Port,
		})
		return
	}

	// 获取连接统计
	stats := diagnostics.GetConnectionStats()

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection test successful",
		"host":    request.Host,
		"port":    request.Port,
		"stats":   stats,
	})
}

// 获取连接统计信息
func (h *VPNHandler) GetConnectionStats(c *gin.Context) {
	diagnostics := vpn.NewConnectionDiagnostics(h.logger)
	stats := diagnostics.GetConnectionStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
