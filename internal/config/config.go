package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OpenIM   OpenIMConfig   `mapstructure:"openim"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type PostgresConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type OpenIMConfig struct {
	APIURL      string `mapstructure:"api_url"`
	AdminUserID string `mapstructure:"admin_user_id"`
	Secret      string `mapstructure:"secret"`
}

func Load(cfgFile string) (*Config, error) {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, viper.Unmarshal(&cfg)
}
