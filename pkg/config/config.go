package config

import (
	"go-core-api/pkg/logger"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	JWT struct {
		Secret            string `mapstructure:"secret"`
		AccessExpiration  int    `mapstructure:"access_expiration"`
		RefreshExpiration int    `mapstructure:"refresh_expiration"`
	} `mapstructure:"jwt"`
	Mailer struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		From     string `mapstructure:"from"`
	} `mapstructure:"mailer"`
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile("config/config.yaml") // Đường dẫn tời file config
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Fatal("❌ Không thể đọc file config: %v", zap.Error(err))
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		logger.Fatal("❌ Không thể parse config: %v", zap.Error(err))
	}

	logger.Info("✅ Đã load config thành công!")
}
