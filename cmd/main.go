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
	var standardSSServer *vpn.StandardShadowsocksServer
	if config.AppConfig.Shadowsocks.Enabled {
		// 标准Shadowsocks配置（用于标准客户端）
		standardConfig := &vpn.Config{
			Port:     config.AppConfig.Shadowsocks.Port, // 使用配置的端口
			Method:   config.AppConfig.Shadowsocks.Method,
			Password: config.AppConfig.Shadowsocks.Password,
			Timeout:  config.AppConfig.Shadowsocks.Timeout,
			// 第二个服务端配置
			SecondServerEnabled:  config.AppConfig.SecondServer.Enabled,
			SecondServerHost:     config.AppConfig.SecondServer.Host,
			SecondServerPort:     config.AppConfig.SecondServer.Port,
			SecondServerMethod:   config.AppConfig.SecondServer.Method,
			SecondServerPassword: config.AppConfig.SecondServer.Password,
			SecondServerTimeout:  config.AppConfig.SecondServer.Timeout,
		}

		// 自定义协议配置（用于我们的客户端）
		customConfig := &vpn.Config{
			Port:     config.AppConfig.Shadowsocks.Port + 1, // 使用下一个端口
			Method:   config.AppConfig.Shadowsocks.Method,
			Password: config.AppConfig.Shadowsocks.Password,
			Timeout:  config.AppConfig.Shadowsocks.Timeout,
			// 第二个服务端配置
			SecondServerEnabled:  config.AppConfig.SecondServer.Enabled,
			SecondServerHost:     config.AppConfig.SecondServer.Host,
			SecondServerPort:     config.AppConfig.SecondServer.Port,
			SecondServerMethod:   config.AppConfig.SecondServer.Method,
			SecondServerPassword: config.AppConfig.SecondServer.Password,
			SecondServerTimeout:  config.AppConfig.SecondServer.Timeout,
		}

		// 使用标准Shadowsocks服务器（兼容标准客户端）
		standardSSServer = vpn.NewStandardShadowsocksServer(standardConfig, logger)
		if err := standardSSServer.Start(); err != nil {
			logger.Errorf("Failed to start standard Shadowsocks server: %v", err)
		} else {
			logger.Infof("Standard Shadowsocks server started on port %d", standardConfig.Port)
		}

		// 同时启动自定义协议服务器（用于我们的客户端）
		proxyServer = vpn.NewProxyServer(customConfig, logger)
		if err := proxyServer.Start(); err != nil {
			logger.Errorf("Failed to start custom Shadowsocks server: %v", err)
		} else {
			logger.Infof("Custom Shadowsocks server started on port %d", customConfig.Port)
			if config.AppConfig.SecondServer.Enabled {
				logger.Infof("Second server forwarding enabled: %s:%d", config.AppConfig.SecondServer.Host, config.AppConfig.SecondServer.Port)
			}
		}
	}

	// 启动HTTP服务器
	go func() {
		logger.Infof("HTTP server starting on %s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)

		// 尝试启动HTTP服务器，如果端口被占用则尝试下一个端口
		originalPort := config.AppConfig.Server.Port
		maxRetries := 10

		for i := 0; i < maxRetries; i++ {
			port := originalPort + i
			addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, port)

			// 创建新的服务器实例
			httpServer := &http.Server{
				Addr:    addr,
				Handler: router,
			}

			if err := httpServer.ListenAndServe(); err != nil {
				if i == 0 {
					logger.Warnf("Failed to start HTTP server on %s: %v, trying next port", addr, err)
				}
				continue
			}

			// 如果成功启动，更新配置
			config.AppConfig.Server.Port = port
			logger.Infof("HTTP server started successfully on %s", addr)
			return
		}

		logger.Fatalf("Failed to start HTTP server after %d attempts", maxRetries)
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
	if standardSSServer != nil {
		standardSSServer.Stop()
		logger.Info("Standard Shadowsocks server stopped")
	}
	if proxyServer != nil {
		proxyServer.Stop()
		logger.Info("Custom Shadowsocks server stopped")
	}

	// 停止HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}

	logger.Info("Server exited")
}
