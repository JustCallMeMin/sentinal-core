package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	AppEnv      string `mapstructure:"APP_ENV"`
	Port        string `mapstructure:"SERVER_PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisAddr   string `mapstructure:"REDIS_ADDR"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
}

// Validate ensures all required configuration is present and valid
func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.DatabaseURL, validation.Required, is.URL),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.AppEnv, validation.In("development", "production", "test")),
	)
}

// LoadConfig reads configuration from .env file or environment variables
func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")

	// Read from .env file if it exists
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if .env doesn't exist, we rely on environment variables
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
