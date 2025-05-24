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

// EmailService handles email-related operations
type EmailService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewEmailService creates a new email service
func NewEmailService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *EmailService {
	return &EmailService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CreateEmailAccount creates a new email account
func (s *EmailService) CreateEmailAccount(ctx context.Context, domainID uuid.UUID, username, password string, quotaMB int) (*models.EmailAccount, error) {
	// Check if domain exists
	var domain models.Domain
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Check if email account already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.EmailAccount{}).
		Where("domain_id = ? AND username = ?", domainID, username).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check email account existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("email account already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	emailAccount := &models.EmailAccount{
		DomainID:     domainID,
		Username:     username,
		PasswordHash: string(hashedPassword),
		QuotaMB:      quotaMB,
		IsActive:     true,
	}

	if err := s.db.WithContext(ctx).Create(emailAccount).Error; err != nil {
		return nil, fmt.Errorf("failed to create email account: %w", err)
	}

	s.logger.Info("Email account created", 
		zap.String("email", username+"@"+domain.Name),
		zap.String("domain_id", domainID.String()))

	return emailAccount, nil
}

// GetEmailAccounts retrieves all email accounts for a domain
func (s *EmailService) GetEmailAccounts(ctx context.Context, domainID uuid.UUID) ([]*models.EmailAccount, error) {
	var emailAccounts []*models.EmailAccount
	if err := s.db.WithContext(ctx).
		Preload("Domain").
		Where("domain_id = ?", domainID).
		Find(&emailAccounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get email accounts: %w", err)
	}

	return emailAccounts, nil
}

// UpdateEmailAccount updates email account information
func (s *EmailService) UpdateEmailAccount(ctx context.Context, accountID uuid.UUID, updates map[string]interface{}) (*models.EmailAccount, error) {
	var account models.EmailAccount
	if err := s.db.WithContext(ctx).Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, fmt.Errorf("email account not found: %w", err)
	}

	// Hash password if it's being updated
	if password, ok := updates["password"]; ok {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password.(string)), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password_hash"] = string(hashedPassword)
		delete(updates, "password")
	}

	if err := s.db.WithContext(ctx).Model(&account).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update email account: %w", err)
	}

	return &account, nil
}

// DeleteEmailAccount deletes an email account
func (s *EmailService) DeleteEmailAccount(ctx context.Context, accountID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", accountID).Delete(&models.EmailAccount{}).Error; err != nil {
		return fmt.Errorf("failed to delete email account: %w", err)
	}

	return nil
}

// CreateEmailAlias creates a new email alias
func (s *EmailService) CreateEmailAlias(ctx context.Context, domainID uuid.UUID, alias, destination string) (*models.EmailAlias, error) {
	emailAlias := &models.EmailAlias{
		DomainID:    domainID,
		Alias:       alias,
		Destination: destination,
		IsActive:    true,
	}

	if err := s.db.WithContext(ctx).Create(emailAlias).Error; err != nil {
		return nil, fmt.Errorf("failed to create email alias: %w", err)
	}

	return emailAlias, nil
}

// GetEmailAliases retrieves all email aliases for a domain
func (s *EmailService) GetEmailAliases(ctx context.Context, domainID uuid.UUID) ([]*models.EmailAlias, error) {
	var aliases []*models.EmailAlias
	if err := s.db.WithContext(ctx).
		Where("domain_id = ?", domainID).
		Find(&aliases).Error; err != nil {
		return nil, fmt.Errorf("failed to get email aliases: %w", err)
	}

	return aliases, nil
}

// DeleteEmailAlias deletes an email alias
func (s *EmailService) DeleteEmailAlias(ctx context.Context, aliasID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", aliasID).Delete(&models.EmailAlias{}).Error; err != nil {
		return fmt.Errorf("failed to delete email alias: %w", err)
	}

	return nil
}
