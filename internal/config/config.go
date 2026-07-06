// Package config 负责加载并解析 config.yaml，支持环境变量覆盖。
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Mail      MailConfig      `mapstructure:"mail"`
	I18N      I18NConfig      `mapstructure:"i18n"`
	DB        DBConfig        `mapstructure:"db"`
}

type AppConfig struct {
	Port int    `mapstructure:"port"`
	Env  string `mapstructure:"env"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	AccessTTL  string `mapstructure:"access_ttl"`  // 如 "15m"
	RefreshTTL string `mapstructure:"refresh_ttl"` // 如 "168h"
}

type RateLimitConfig struct {
	PerSecond int `mapstructure:"per_second"`
}

type MailConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}

type I18NConfig struct {
	Default string `mapstructure:"default"`
}

type DBConfig struct {
	Path string `mapstructure:"path"`
}

// Load 从指定路径加载配置，未设置的项使用默认值，并允许环境变量覆盖。
// 环境变量约定：将配置点号键中的 "." 替换为 "_"，如 APP_PORT、JWT_SECRET。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	// 默认值
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.env", "dev")
	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "168h")
	v.SetDefault("rate_limit.per_second", 1)
	v.SetDefault("i18n.default", "zh")
	v.SetDefault("db.path", "./data/lychee.db")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}
