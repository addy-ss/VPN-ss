package main

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"server"`

	Local struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"local"`

	Security struct {
		Password string `mapstructure:"password"`
		Method   string `mapstructure:"method"`
		Timeout  int    `mapstructure:"timeout"`
	} `mapstructure:"security"`

	Log struct {
		Level string `mapstructure:"level"`
		File  string `mapstructure:"file"`
	} `mapstructure:"log"`
}

var AppConfig Config

func LoadClientConfig(configPath string) (*ClientConfig, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setClientDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return nil, err
	}

	// 转换为ClientConfig
	clientConfig := &ClientConfig{
		ServerHost: AppConfig.Server.Host,
		ServerPort: AppConfig.Server.Port,
		LocalPort:  AppConfig.Local.Port,
		Password:   AppConfig.Security.Password,
		Method:     AppConfig.Security.Method,
		Timeout:    AppConfig.Security.Timeout,
	}

	return clientConfig, nil
}

func setClientDefaults() {
	// 设置默认值
	viper.SetDefault("server.host", "127.0.0.1")
	viper.SetDefault("server.port", 8388)
	viper.SetDefault("local.port", 1080)
	viper.SetDefault("security.password", "13687401432Fan!")
	viper.SetDefault("security.method", "aes-256-gcm")
	viper.SetDefault("security.timeout", 300)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "")
}
