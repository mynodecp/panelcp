package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger instance
func New() *zap.Logger {
	config := zap.NewProductionConfig()
	
	// Set log level from environment
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		if parsedLevel, err := zapcore.ParseLevel(level); err == nil {
			config.Level = zap.NewAtomicLevelAt(parsedLevel)
		}
	}

	// Configure encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = ""

	// Build logger
	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}

	return logger
}

// NewDevelopment creates a development logger
func NewDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	
	logger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}

	return logger
}

// WithFields creates a logger with predefined fields
func WithFields(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// RequestLogger creates a logger for HTTP requests
func RequestLogger(logger *zap.Logger, requestID, method, path string) *zap.Logger {
	return logger.With(
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
	)
}

// ServiceLogger creates a logger for services
func ServiceLogger(logger *zap.Logger, service string) *zap.Logger {
	return logger.With(zap.String("service", service))
}
