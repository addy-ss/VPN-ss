package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(router *gin.Engine, logger *logrus.Logger) {
	// 创建VPN处理器
	vpnHandler := NewVPNHandler(nil, logger)

	// API版本组
	v1 := router.Group("/api/v1")
	{
		// VPN管理端点
		vpn := v1.Group("/vpn")
		{
			vpn.POST("/start", vpnHandler.StartVPN)
			vpn.POST("/stop", vpnHandler.StopVPN)
			vpn.GET("/status", vpnHandler.GetVPNStatus)
			vpn.POST("/config/generate", vpnHandler.GenerateConfig)
			vpn.GET("/methods", vpnHandler.GetSupportedMethods)
		}

		// 健康检查
		v1.GET("/health", vpnHandler.HealthCheck)
	}

	// 根路径重定向到API文档
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "VPS VPN Service",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health": "/api/v1/health",
				"vpn": gin.H{
					"start":           "/api/v1/vpn/start",
					"stop":            "/api/v1/vpn/stop",
					"status":          "/api/v1/vpn/status",
					"generate_config": "/api/v1/vpn/config/generate",
					"methods":         "/api/v1/vpn/methods",
				},
			},
		})
	})
}
