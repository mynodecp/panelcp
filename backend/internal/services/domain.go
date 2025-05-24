package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/mynodecp/mynodecp/backend/internal/models"
)

// DomainService handles domain-related operations
type DomainService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewDomainService creates a new domain service
func NewDomainService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *DomainService {
	return &DomainService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CreateDomain creates a new domain
func (s *DomainService) CreateDomain(ctx context.Context, userID uuid.UUID, name string) (*models.Domain, error) {
	// Check if domain already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Domain{}).
		Where("name = ?", name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check domain existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("domain already exists")
	}

	// Create document root path
	documentRoot := filepath.Join("/var/www", name, "public_html")

	domain := &models.Domain{
		UserID:       userID,
		Name:         name,
		DocumentRoot: documentRoot,
		IsActive:     true,
		PHPVersion:   "8.2",
	}

	if err := s.db.WithContext(ctx).Create(domain).Error; err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	// Create default DNS records
	if err := s.createDefaultDNSRecords(ctx, domain.ID, name); err != nil {
		s.logger.Error("Failed to create default DNS records", zap.Error(err))
	}

	// Create document root directory (this would be done by a system service)
	s.logger.Info("Domain created", zap.String("domain", name), zap.String("user_id", userID.String()))

	return domain, nil
}

// GetDomain retrieves a domain by ID
func (s *DomainService) GetDomain(ctx context.Context, domainID uuid.UUID) (*models.Domain, error) {
	var domain models.Domain
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Subdomains").
		Preload("DNSRecords").
		Preload("SSLCertificates").
		Where("id = ?", domainID).
		First(&domain).Error; err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return &domain, nil
}

// GetUserDomains retrieves all domains for a user
func (s *DomainService) GetUserDomains(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Domain, int64, error) {
	var domains []*models.Domain
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&models.Domain{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count domains: %w", err)
	}

	// Get domains with pagination
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(&domains).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get domains: %w", err)
	}

	return domains, total, nil
}

// UpdateDomain updates domain information
func (s *DomainService) UpdateDomain(ctx context.Context, domainID uuid.UUID, updates map[string]interface{}) (*models.Domain, error) {
	var domain models.Domain
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&domain).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	// Reload domain with relationships
	if err := s.db.WithContext(ctx).
		Preload("User").
		Preload("Subdomains").
		Preload("DNSRecords").
		Preload("SSLCertificates").
		Where("id = ?", domainID).
		First(&domain).Error; err != nil {
		return nil, fmt.Errorf("failed to reload domain: %w", err)
	}

	return &domain, nil
}

// DeleteDomain soft deletes a domain
func (s *DomainService) DeleteDomain(ctx context.Context, domainID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).Delete(&models.Domain{}).Error; err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	s.logger.Info("Domain deleted", zap.String("domain_id", domainID.String()))
	return nil
}

// CreateSubdomain creates a new subdomain
func (s *DomainService) CreateSubdomain(ctx context.Context, domainID uuid.UUID, name string) (*models.Subdomain, error) {
	// Check if domain exists
	var domain models.Domain
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Check if subdomain already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Subdomain{}).
		Where("domain_id = ? AND name = ?", domainID, name).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check subdomain existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("subdomain already exists")
	}

	// Create document root path
	documentRoot := filepath.Join("/var/www", domain.Name, "subdomains", name)

	subdomain := &models.Subdomain{
		DomainID:     domainID,
		Name:         name,
		DocumentRoot: documentRoot,
		IsActive:     true,
	}

	if err := s.db.WithContext(ctx).Create(subdomain).Error; err != nil {
		return nil, fmt.Errorf("failed to create subdomain: %w", err)
	}

	// Create DNS record for subdomain
	dnsRecord := &models.DNSRecord{
		DomainID: domainID,
		Type:     "A",
		Name:     name,
		Value:    "127.0.0.1", // This would be the server's IP
		TTL:      3600,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(dnsRecord).Error; err != nil {
		s.logger.Error("Failed to create DNS record for subdomain", zap.Error(err))
	}

	return subdomain, nil
}

// GetSubdomains retrieves all subdomains for a domain
func (s *DomainService) GetSubdomains(ctx context.Context, domainID uuid.UUID) ([]*models.Subdomain, error) {
	var subdomains []*models.Subdomain
	if err := s.db.WithContext(ctx).
		Where("domain_id = ?", domainID).
		Find(&subdomains).Error; err != nil {
		return nil, fmt.Errorf("failed to get subdomains: %w", err)
	}

	return subdomains, nil
}

// UpdateSubdomain updates subdomain information
func (s *DomainService) UpdateSubdomain(ctx context.Context, subdomainID uuid.UUID, updates map[string]interface{}) (*models.Subdomain, error) {
	var subdomain models.Subdomain
	if err := s.db.WithContext(ctx).Where("id = ?", subdomainID).First(&subdomain).Error; err != nil {
		return nil, fmt.Errorf("subdomain not found: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&subdomain).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update subdomain: %w", err)
	}

	return &subdomain, nil
}

// DeleteSubdomain deletes a subdomain
func (s *DomainService) DeleteSubdomain(ctx context.Context, subdomainID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", subdomainID).Delete(&models.Subdomain{}).Error; err != nil {
		return fmt.Errorf("failed to delete subdomain: %w", err)
	}

	return nil
}

// GetDomainStats retrieves domain statistics
func (s *DomainService) GetDomainStats(ctx context.Context, domainID uuid.UUID) (map[string]interface{}, error) {
	var domain models.Domain
	if err := s.db.WithContext(ctx).Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Count subdomains
	var subdomainCount int64
	s.db.WithContext(ctx).Model(&models.Subdomain{}).Where("domain_id = ?", domainID).Count(&subdomainCount)

	// Count email accounts
	var emailCount int64
	s.db.WithContext(ctx).Model(&models.EmailAccount{}).Where("domain_id = ?", domainID).Count(&emailCount)

	// Count databases
	var databaseCount int64
	s.db.WithContext(ctx).Model(&models.Database{}).Where("domain_id = ?", domainID).Count(&databaseCount)

	stats := map[string]interface{}{
		"disk_usage":       domain.DiskUsage,
		"bandwidth_usage":  domain.BandwidthUsage,
		"disk_quota":       domain.DiskQuota,
		"bandwidth_quota":  domain.BandwidthQuota,
		"subdomain_count":  subdomainCount,
		"email_count":      emailCount,
		"database_count":   databaseCount,
		"has_ssl":          domain.HasSSL,
		"php_version":      domain.PHPVersion,
	}

	return stats, nil
}

// createDefaultDNSRecords creates default DNS records for a new domain
func (s *DomainService) createDefaultDNSRecords(ctx context.Context, domainID uuid.UUID, domainName string) error {
	defaultRecords := []models.DNSRecord{
		{
			DomainID: domainID,
			Type:     "A",
			Name:     "@",
			Value:    "127.0.0.1", // This would be the server's IP
			TTL:      3600,
			IsActive: true,
		},
		{
			DomainID: domainID,
			Type:     "A",
			Name:     "www",
			Value:    "127.0.0.1", // This would be the server's IP
			TTL:      3600,
			IsActive: true,
		},
		{
			DomainID: domainID,
			Type:     "MX",
			Name:     "@",
			Value:    "mail." + domainName,
			TTL:      3600,
			Priority: &[]int{10}[0],
			IsActive: true,
		},
	}

	for _, record := range defaultRecords {
		if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create DNS record: %w", err)
		}
	}

	return nil
}
