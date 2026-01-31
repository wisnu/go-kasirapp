package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `mapstructure:"app"`
	DB  DBConfig  `mapstructure:"db"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type DBConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Load .env file
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	// Enable env override (important!)
	v.AutomaticEnv()

	// Explicit bindings (recommended for .env)
	_ = v.BindEnv("APP_NAME")
	_ = v.BindEnv("APP_ENV")
	_ = v.BindEnv("APP_PORT")

	_ = v.BindEnv("DB_DRIVER")
	_ = v.BindEnv("DB_HOST")
	_ = v.BindEnv("DB_PORT")
	_ = v.BindEnv("DB_USER")
	_ = v.BindEnv("DB_PASSWORD")
	_ = v.BindEnv("DB_NAME")
	_ = v.BindEnv("DB_SSLMODE")

	// .env is optional (prod often uses real env vars)
	_ = v.ReadInConfig()

	cfg := &Config{
		App: AppConfig{
			Name: v.GetString("APP_NAME"),
			Env:  v.GetString("APP_ENV"),
			Port: v.GetInt("APP_PORT"),
		},
		DB: DBConfig{
			Driver:   v.GetString("DB_DRIVER"),
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetInt("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
	}

	return cfg, nil

}
