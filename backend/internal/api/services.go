package api

import (
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/mynodecp/mynodecp/backend/internal/auth"
	"github.com/mynodecp/mynodecp/backend/internal/services"
)

// Services holds all API services
type Services struct {
	Auth     *auth.Service
	User     *services.UserService
	Domain   *services.DomainService
	Email    *services.EmailService
	Database *services.DatabaseService
	File     *services.FileService
	System   *services.SystemService
	Backup   *services.BackupService
	SSL      *services.SSLService
	DNS      *services.DNSService
}

// NewServices creates a new Services instance
func NewServices(db *gorm.DB, redis *redis.Client, authService *auth.Service, logger *zap.Logger) *Services {
	return &Services{
		Auth:     authService,
		User:     services.NewUserService(db, redis, logger),
		Domain:   services.NewDomainService(db, redis, logger),
		Email:    services.NewEmailService(db, redis, logger),
		Database: services.NewDatabaseService(db, redis, logger),
		File:     services.NewFileService(db, redis, logger),
		System:   services.NewSystemService(db, redis, logger),
		Backup:   services.NewBackupService(db, redis, logger),
		SSL:      services.NewSSLService(db, redis, logger),
		DNS:      services.NewDNSService(db, redis, logger),
	}
}
