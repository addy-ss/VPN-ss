package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Shadowsocks ShadowsocksConfig `mapstructure:"shadowsocks"`
	Log         LogConfig         `mapstructure:"log"`
}

type ServerConfig struct {
	Port     int    `mapstructure:"port"`
	Host     string `mapstructure:"host"`
	Mode     string `mapstructure:"mode"` // debug or release
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type ShadowsocksConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Method   string `mapstructure:"method"` // 加密方法: aes-256-gcm, chacha20-poly1305, etc.
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	Timeout  int    `mapstructure:"timeout"`
}

type LogConfig struct {
	Level string `mapstructure:"level"` // debug, info, warn, error
	File  string `mapstructure:"file"`
}

var AppConfig Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		return err
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("shadowsocks.enabled", true)
	viper.SetDefault("shadowsocks.method", "aes-256-gcm")
	viper.SetDefault("shadowsocks.port", 8388)
	viper.SetDefault("shadowsocks.timeout", 300)

	viper.SetDefault("log.level", "info")
}
