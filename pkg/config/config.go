package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	AppEnv      string `mapstructure:"APP_ENV"`
	Port        string `mapstructure:"SERVER_PORT"`
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	RedisAddr   string `mapstructure:"REDIS_ADDR"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	JWTSecret   string `mapstructure:"JWT_SECRET"`
	JWTExpiry   int    `mapstructure:"JWT_EXPIRY_HOURS"` // Expiry in hours
	DBMaxConns  int32  `mapstructure:"DB_MAX_CONNS"`
	DBMinConns  int32  `mapstructure:"DB_MIN_CONNS"`
	DBMaxIdle   string `mapstructure:"DB_MAX_IDLE_TIME"`
}

// Validate ensures all required configuration is present and valid
func (c *Config) Validate() error {
	return validation.ValidateStruct(c,
		// Required and must look like a connection string (postgres://... or similar)
		validation.Field(&c.DatabaseURL, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.AppEnv, validation.In("development", "production", "test")),
		validation.Field(&c.JWTSecret, validation.When(c.AppEnv == "production", validation.Required)),
	)
}

// LoadConfig reads configuration from .env file or environment variables
func LoadConfig() (*Config, error) {
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("JWT_SECRET", "sentinal-dev-secret-key-change-me")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("DB_MAX_CONNS", 20)
	viper.SetDefault("DB_MIN_CONNS", 5)
	viper.SetDefault("DB_MAX_IDLE_TIME", "15m")

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
