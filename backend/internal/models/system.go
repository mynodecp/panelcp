package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileManager represents file manager entries
type FileManager struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:char(36);not null"`
	DomainID    *uuid.UUID `json:"domain_id,omitempty" gorm:"type:char(36)"`
	Path        string    `json:"path" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // file, directory
	Size        int64     `json:"size" gorm:"default:0"`
	Permissions string    `json:"permissions" gorm:"default:'644'"`
	Owner       string    `json:"owner"`
	Group       string    `json:"group"`
	MimeType    string    `json:"mime_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Domain *Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
}

// CronJob represents a cron job
type CronJob struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:char(36);not null"`
	DomainID    *uuid.UUID `json:"domain_id,omitempty" gorm:"type:char(36)"`
	Name        string     `json:"name" gorm:"not null"`
	Command     string     `json:"command" gorm:"not null"`
	Schedule    string     `json:"schedule" gorm:"not null"` // Cron expression
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastRunAt   *time.Time `json:"last_run_at"`
	NextRunAt   *time.Time `json:"next_run_at"`
	LastStatus  string     `json:"last_status"` // success, failed, running
	LastOutput  string     `json:"last_output" gorm:"type:text"`
	RunCount    int        `json:"run_count" gorm:"default:0"`
	FailCount   int        `json:"fail_count" gorm:"default:0"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Domain *Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
}

// Backup represents a backup
type Backup struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:char(36);not null"`
	DomainID    *uuid.UUID `json:"domain_id,omitempty" gorm:"type:char(36)"`
	Type        string     `json:"type" gorm:"not null"` // full, files, database
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	FilePath    string     `json:"file_path"`
	SizeMB      int64      `json:"size_mb" gorm:"default:0"`
	Status      string     `json:"status" gorm:"default:'pending'"` // pending, running, completed, failed
	Progress    int        `json:"progress" gorm:"default:0"` // 0-100
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Domain *Domain `json:"domain,omitempty" gorm:"foreignKey:DomainID"`
}

// SystemMetric represents system metrics
type SystemMetric struct {
	ID        uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Type      string    `json:"type" gorm:"not null"` // cpu, memory, disk, network
	Value     float64   `json:"value" gorm:"not null"`
	Unit      string    `json:"unit" gorm:"not null"` // percent, bytes, etc.
	Metadata  string    `json:"metadata" gorm:"type:text"` // JSON metadata
	CreatedAt time.Time `json:"created_at"`
}

// ServerResource represents server resource usage
type ServerResource struct {
	ID               uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	CPUUsage         float64   `json:"cpu_usage"`
	MemoryUsage      int64     `json:"memory_usage"`
	MemoryTotal      int64     `json:"memory_total"`
	DiskUsage        int64     `json:"disk_usage"`
	DiskTotal        int64     `json:"disk_total"`
	NetworkInBytes   int64     `json:"network_in_bytes"`
	NetworkOutBytes  int64     `json:"network_out_bytes"`
	LoadAverage1     float64   `json:"load_average_1"`
	LoadAverage5     float64   `json:"load_average_5"`
	LoadAverage15    float64   `json:"load_average_15"`
	ActiveConnections int      `json:"active_connections"`
	ProcessCount     int       `json:"process_count"`
	CreatedAt        time.Time `json:"created_at"`
}

// ServiceStatus represents the status of system services
type ServiceStatus struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	ServiceName string    `json:"service_name" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null"` // running, stopped, failed
	PID         *int      `json:"pid,omitempty"`
	Memory      int64     `json:"memory" gorm:"default:0"`
	CPU         float64   `json:"cpu" gorm:"default:0"`
	Uptime      int64     `json:"uptime" gorm:"default:0"` // seconds
	LastChecked time.Time `json:"last_checked"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SecurityEvent represents security events
type SecurityEvent struct {
	ID          uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:char(36)"`
	Type        string     `json:"type" gorm:"not null"` // login_failed, brute_force, suspicious_activity
	Severity    string     `json:"severity" gorm:"not null"` // low, medium, high, critical
	Source      string     `json:"source" gorm:"not null"` // web, ssh, ftp, etc.
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
	Description string     `json:"description" gorm:"type:text"`
	Metadata    string     `json:"metadata" gorm:"type:text"` // JSON metadata
	IsResolved  bool       `json:"is_resolved" gorm:"default:false"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	ResolvedBy  *uuid.UUID `json:"resolved_by,omitempty" gorm:"type:char(36)"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relationships
	User       *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	ResolvedByUser *User `json:"resolved_by_user,omitempty" gorm:"foreignKey:ResolvedBy"`
}

// BeforeCreate hooks
func (f *FileManager) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

func (c *CronJob) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

func (b *Backup) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

func (s *SystemMetric) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *ServerResource) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *ServiceStatus) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *SecurityEvent) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
