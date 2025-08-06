package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vps/config"
	"vps/internal/api"
	"vps/internal/vpn"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// 加载配置
	if err := config.LoadConfig("config.yaml"); err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	if config.AppConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由器
	router := gin.Default()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 设置路由
	api.SetupRoutes(router, logger)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port),
		Handler: router,
	}

	// 启动Shadowsocks服务（如果启用）
	var proxyServer *vpn.ProxyServer
	if config.AppConfig.Shadowsocks.Enabled {
		ssConfig := &vpn.Config{
			Port:     config.AppConfig.Shadowsocks.Port,
			Method:   config.AppConfig.Shadowsocks.Method,
			Password: config.AppConfig.Shadowsocks.Password,
			Timeout:  config.AppConfig.Shadowsocks.Timeout,
		}

		proxyServer = vpn.NewProxyServer(ssConfig, logger)
		if err := proxyServer.Start(); err != nil {
			logger.Errorf("Failed to start Shadowsocks server: %v", err)
		} else {
			logger.Infof("Shadowsocks server started on port %d", config.AppConfig.Shadowsocks.Port)
		}
	}

	// 启动HTTP服务器
	go func() {
		logger.Infof("HTTP server starting on %s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止Shadowsocks服务
	if proxyServer != nil {
		proxyServer.Stop()
		logger.Info("Shadowsocks server stopped")
	}

	// 停止HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}
