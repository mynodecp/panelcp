package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mynodecp/mynodecp/backend/internal/models"
)

// DatabaseService handles database-related operations
type DatabaseService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewDatabaseService creates a new database service
func NewDatabaseService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *DatabaseService {
	return &DatabaseService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CreateDatabase creates a new database
func (s *DatabaseService) CreateDatabase(ctx context.Context, domainID uuid.UUID, name, dbType string) (*models.Database, error) {
	// Check if domain exists
	var domain models.Domain
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Check if database already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Database{}).
		Where("domain_id = ? AND name = ?", domainID, name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("database already exists")
	}

	database := &models.Database{
		DomainID: domainID,
		Name:     name,
		Type:     dbType,
	}

	if err := s.db.WithContext(ctx).Create(database).Error; err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	s.logger.Info("Database created", 
		zap.String("database", name),
		zap.String("type", dbType),
		zap.String("domain_id", domainID.String()))

	return database, nil
}

// GetDatabases retrieves all databases for a domain
func (s *DatabaseService) GetDatabases(ctx context.Context, domainID uuid.UUID) ([]*models.Database, error) {
	var databases []*models.Database
	if err := s.db.WithContext(ctx).
		Preload("DatabaseUsers").
		Where("domain_id = ?", domainID).
		Find(&databases).Error; err != nil {
		return nil, fmt.Errorf("failed to get databases: %w", err)
	}

	return databases, nil
}

// DeleteDatabase deletes a database
func (s *DatabaseService) DeleteDatabase(ctx context.Context, databaseID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", databaseID).Delete(&models.Database{}).Error; err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	return nil
}

// CreateDatabaseUser creates a new database user
func (s *DatabaseService) CreateDatabaseUser(ctx context.Context, databaseID uuid.UUID, username, password string, privileges []string) (*models.DatabaseUser, error) {
	// Check if database exists
	var database models.Database
	if err := s.db.WithContext(ctx).Where("id = ?", databaseID).First(&database).Error; err != nil {
		return nil, fmt.Errorf("database not found: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Convert privileges to JSON string (simplified)
	privilegesJSON := fmt.Sprintf(`["%s"]`, privileges[0])

	dbUser := &models.DatabaseUser{
		DatabaseID:   databaseID,
		Username:     username,
		PasswordHash: string(hashedPassword),
		Privileges:   privilegesJSON,
	}

	if err := s.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create database user: %w", err)
	}

	return dbUser, nil
}

// GetDatabaseUsers retrieves all users for a database
func (s *DatabaseService) GetDatabaseUsers(ctx context.Context, databaseID uuid.UUID) ([]*models.DatabaseUser, error) {
	var users []*models.DatabaseUser
	if err := s.db.WithContext(ctx).
		Where("database_id = ?", databaseID).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get database users: %w", err)
	}

	return users, nil
}

// DeleteDatabaseUser deletes a database user
func (s *DatabaseService) DeleteDatabaseUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", userID).Delete(&models.DatabaseUser{}).Error; err != nil {
		return fmt.Errorf("failed to delete database user: %w", err)
	}

	return nil
}
