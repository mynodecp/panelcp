package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	HTTPPort    int    `mapstructure:"http_port"`
	GRPCPort    int    `mapstructure:"grpc_port"`
	Environment string `mapstructure:"environment"`
	Version     string `mapstructure:"version"`
	Domain      string `mapstructure:"domain"`
	TLSEnabled  bool   `mapstructure:"tls_enabled"`
	CertFile    string `mapstructure:"cert_file"`
	KeyFile     string `mapstructure:"key_file"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	SSLMode         string        `mapstructure:"ssl_mode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret           string        `mapstructure:"jwt_secret"`
	JWTExpiration       time.Duration `mapstructure:"jwt_expiration"`
	RefreshExpiration   time.Duration `mapstructure:"refresh_expiration"`
	PasswordMinLength   int           `mapstructure:"password_min_length"`
	PasswordRequireUpper bool         `mapstructure:"password_require_upper"`
	PasswordRequireLower bool         `mapstructure:"password_require_lower"`
	PasswordRequireDigit bool         `mapstructure:"password_require_digit"`
	PasswordRequireSpecial bool       `mapstructure:"password_require_special"`
	TwoFactorEnabled    bool          `mapstructure:"two_factor_enabled"`
	SessionTimeout      time.Duration `mapstructure:"session_timeout"`
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	RateLimitEnabled    bool          `mapstructure:"rate_limit_enabled"`
	RateLimitRequests   int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow     time.Duration `mapstructure:"rate_limit_window"`
	CORSEnabled         bool          `mapstructure:"cors_enabled"`
	CORSAllowedOrigins  []string      `mapstructure:"cors_allowed_origins"`
	CSRFEnabled         bool          `mapstructure:"csrf_enabled"`
	HSTSEnabled         bool          `mapstructure:"hsts_enabled"`
	HSTSMaxAge          int           `mapstructure:"hsts_max_age"`
	ContentTypeNosniff  bool          `mapstructure:"content_type_nosniff"`
	XFrameOptions       string        `mapstructure:"x_frame_options"`
	XSSProtection       bool          `mapstructure:"xss_protection"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./backend/configs")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set default values
	setDefaults()

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.http_port", 8080)
	viper.SetDefault("server.grpc_port", 9090)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.version", "1.0.0")
	viper.SetDefault("server.domain", "localhost")
	viper.SetDefault("server.tls_enabled", false)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.username", "mynodecp")
	viper.SetDefault("database.password", "mynodecp")
	viper.SetDefault("database.database", "mynodecp")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.ssl_mode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 2)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-super-secret-jwt-key-change-this-in-production")
	viper.SetDefault("auth.jwt_expiration", "15m")
	viper.SetDefault("auth.refresh_expiration", "7d")
	viper.SetDefault("auth.password_min_length", 8)
	viper.SetDefault("auth.password_require_upper", true)
	viper.SetDefault("auth.password_require_lower", true)
	viper.SetDefault("auth.password_require_digit", true)
	viper.SetDefault("auth.password_require_special", true)
	viper.SetDefault("auth.two_factor_enabled", true)
	viper.SetDefault("auth.session_timeout", "24h")

	// Security defaults
	viper.SetDefault("security.rate_limit_enabled", true)
	viper.SetDefault("security.rate_limit_requests", 100)
	viper.SetDefault("security.rate_limit_window", "1m")
	viper.SetDefault("security.cors_enabled", true)
	viper.SetDefault("security.cors_allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("security.csrf_enabled", true)
	viper.SetDefault("security.hsts_enabled", true)
	viper.SetDefault("security.hsts_max_age", 31536000)
	viper.SetDefault("security.content_type_nosniff", true)
	viper.SetDefault("security.x_frame_options", "DENY")
	viper.SetDefault("security.xss_protection", true)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 28)
	viper.SetDefault("logging.compress", true)
}

// validate validates the configuration
func validate(config *Config) error {
	if config.Server.HTTPPort <= 0 || config.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", config.Server.HTTPPort)
	}

	if config.Server.GRPCPort <= 0 || config.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", config.Server.GRPCPort)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "your-super-secret-jwt-key-change-this-in-production" {
		if config.Server.Environment == "production" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}

	return nil
}
