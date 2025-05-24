package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Domain represents a domain in the hosting system
type Domain struct {
	ID              uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	Name            string    `json:"name" gorm:"uniqueIndex;not null"`
	DocumentRoot    string    `json:"document_root"`
	IsActive        bool      `json:"is_active" gorm:"default:true"`
	HasSSL          bool      `json:"has_ssl" gorm:"default:false"`
	SSLAutoRenew    bool      `json:"ssl_auto_renew" gorm:"default:true"`
	PHPVersion      string    `json:"php_version" gorm:"default:'8.2'"`
	DiskUsage       int64     `json:"disk_usage" gorm:"default:0"`
	BandwidthUsage  int64     `json:"bandwidth_usage" gorm:"default:0"`
	DiskQuota       int64     `json:"disk_quota" gorm:"default:1073741824"` // 1GB default
	BandwidthQuota  int64     `json:"bandwidth_quota" gorm:"default:10737418240"` // 10GB default
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User            User              `json:"user" gorm:"foreignKey:UserID"`
	Subdomains      []Subdomain       `json:"subdomains" gorm:"foreignKey:DomainID"`
	DNSRecords      []DNSRecord       `json:"dns_records" gorm:"foreignKey:DomainID"`
	SSLCertificates []SSLCertificate  `json:"ssl_certificates" gorm:"foreignKey:DomainID"`
	EmailAccounts   []EmailAccount    `json:"email_accounts" gorm:"foreignKey:DomainID"`
	Databases       []Database        `json:"databases" gorm:"foreignKey:DomainID"`
}

// Subdomain represents a subdomain
type Subdomain struct {
	ID           uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID     uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	Name         string    `json:"name" gorm:"not null"`
	DocumentRoot string    `json:"document_root"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Domain Domain `json:"domain" gorm:"foreignKey:DomainID"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID  uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	Type      string    `json:"type" gorm:"not null"` // A, AAAA, CNAME, MX, TXT, etc.
	Name      string    `json:"name" gorm:"not null"`
	Value     string    `json:"value" gorm:"not null"`
	TTL       int       `json:"ttl" gorm:"default:3600"`
	Priority  *int      `json:"priority,omitempty"` // For MX records
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Domain Domain `json:"domain" gorm:"foreignKey:DomainID"`
}

// SSLCertificate represents an SSL certificate
type SSLCertificate struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	DomainID    uuid.UUID  `json:"domain_id" gorm:"type:char(36);not null"`
	Type        string     `json:"type" gorm:"not null"` // letsencrypt, custom, self-signed
	Certificate string     `json:"-" gorm:"type:text"`
	PrivateKey  string     `json:"-" gorm:"type:text"`
	Chain       string     `json:"-" gorm:"type:text"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	AutoRenew   bool       `json:"auto_renew" gorm:"default:true"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RenewedAt   *time.Time `json:"renewed_at"`

	// Relationships
	Domain Domain `json:"domain" gorm:"foreignKey:DomainID"`
}

// EmailAccount represents an email account
type EmailAccount struct {
	ID           uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID     uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	Username     string    `json:"username" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	QuotaMB      int       `json:"quota_mb" gorm:"default:1024"` // 1GB default
	UsedMB       int       `json:"used_mb" gorm:"default:0"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Domain       Domain         `json:"domain" gorm:"foreignKey:DomainID"`
	Aliases      []EmailAlias   `json:"aliases" gorm:"foreignKey:EmailAccountID"`
	Forwarders   []EmailForwarder `json:"forwarders" gorm:"foreignKey:EmailAccountID"`
}

// EmailAlias represents an email alias
type EmailAlias struct {
	ID             uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID       uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	EmailAccountID *uuid.UUID `json:"email_account_id,omitempty" gorm:"type:char(36)"`
	Alias          string    `json:"alias" gorm:"not null"`
	Destination    string    `json:"destination" gorm:"not null"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Domain       Domain        `json:"domain" gorm:"foreignKey:DomainID"`
	EmailAccount *EmailAccount `json:"email_account,omitempty" gorm:"foreignKey:EmailAccountID"`
}

// EmailForwarder represents an email forwarder
type EmailForwarder struct {
	ID             uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID       uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	EmailAccountID *uuid.UUID `json:"email_account_id,omitempty" gorm:"type:char(36)"`
	Source         string    `json:"source" gorm:"not null"`
	Destination    string    `json:"destination" gorm:"not null"`
	IsActive       bool      `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Domain       Domain        `json:"domain" gorm:"foreignKey:DomainID"`
	EmailAccount *EmailAccount `json:"email_account,omitempty" gorm:"foreignKey:EmailAccountID"`
}

// Database represents a database
type Database struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DomainID  uuid.UUID `json:"domain_id" gorm:"type:char(36);not null"`
	Name      string    `json:"name" gorm:"not null"`
	Type      string    `json:"type" gorm:"not null"` // mysql, postgresql
	SizeMB    int64     `json:"size_mb" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Domain        Domain         `json:"domain" gorm:"foreignKey:DomainID"`
	DatabaseUsers []DatabaseUser `json:"database_users" gorm:"foreignKey:DatabaseID"`
}

// DatabaseUser represents a database user
type DatabaseUser struct {
	ID           uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	DatabaseID   uuid.UUID `json:"database_id" gorm:"type:char(36);not null"`
	Username     string    `json:"username" gorm:"not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Privileges   string    `json:"privileges" gorm:"type:text"` // JSON array of privileges
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Database Database `json:"database" gorm:"foreignKey:DatabaseID"`
}

// BeforeCreate hooks
func (d *Domain) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (s *Subdomain) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (d *DNSRecord) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (s *SSLCertificate) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (e *EmailAccount) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (e *EmailAlias) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (e *EmailForwarder) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (d *Database) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (d *DatabaseUser) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
