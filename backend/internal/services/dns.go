package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/mynodecp/mynodecp/backend/internal/models"
)

// DNSService handles DNS record operations
type DNSService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewDNSService creates a new DNS service
func NewDNSService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *DNSService {
	return &DNSService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CreateDNSRecord creates a new DNS record
func (s *DNSService) CreateDNSRecord(ctx context.Context, domainID uuid.UUID, recordType, name, value string, ttl int, priority *int) (*models.DNSRecord, error) {
	record := &models.DNSRecord{
		DomainID: domainID,
		Type:     recordType,
		Name:     name,
		Value:    value,
		TTL:      ttl,
		Priority: priority,
		IsActive: true,
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	return record, nil
}

// GetDNSRecords retrieves all DNS records for a domain
func (s *DNSService) GetDNSRecords(ctx context.Context, domainID uuid.UUID) ([]*models.DNSRecord, error) {
	var records []*models.DNSRecord
	if err := s.db.WithContext(ctx).
		Where("domain_id = ?", domainID).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}

	return records, nil
}

// UpdateDNSRecord updates a DNS record
func (s *DNSService) UpdateDNSRecord(ctx context.Context, recordID uuid.UUID, updates map[string]interface{}) (*models.DNSRecord, error) {
	var record models.DNSRecord
	if err := s.db.WithContext(ctx).Where("id = ?", recordID).First(&record).Error; err != nil {
		return nil, fmt.Errorf("DNS record not found: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&record).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	return &record, nil
}

// DeleteDNSRecord deletes a DNS record
func (s *DNSService) DeleteDNSRecord(ctx context.Context, recordID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", recordID).Delete(&models.DNSRecord{}).Error; err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}
