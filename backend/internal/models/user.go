package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID                uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	Username          string     `json:"username" gorm:"uniqueIndex;not null"`
	Email             string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash      string     `json:"-" gorm:"not null"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	IsActive          bool       `json:"is_active" gorm:"default:true"`
	IsEmailVerified   bool       `json:"is_email_verified" gorm:"default:false"`
	IsTwoFactorEnabled bool      `json:"is_two_factor_enabled" gorm:"default:false"`
	TwoFactorSecret   string     `json:"-"`
	LastLoginAt       *time.Time `json:"last_login_at"`
	LastLoginIP       string     `json:"last_login_ip"`
	FailedLoginCount  int        `json:"failed_login_count" gorm:"default:0"`
	LockedUntil       *time.Time `json:"locked_until"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Roles    []Role    `json:"roles" gorm:"many2many:user_roles"`
	Sessions []Session `json:"-" gorm:"foreignKey:UserID"`
	Domains  []Domain  `json:"domains" gorm:"foreignKey:UserID"`
}

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"not null"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Users       []User       `json:"-" gorm:"many2many:user_roles"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null"`
	DisplayName string    `json:"display_name" gorm:"not null"`
	Description string    `json:"description"`
	Resource    string    `json:"resource" gorm:"not null"`
	Action      string    `json:"action" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Roles []Role `json:"-" gorm:"many2many:role_permissions"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    uuid.UUID `json:"user_id" gorm:"type:char(36);primary_key"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:char(36);primary_key"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"type:char(36);primary_key"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:char(36);primary_key"`
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	Role       Role       `json:"role" gorm:"foreignKey:RoleID"`
	Permission Permission `json:"permission" gorm:"foreignKey:PermissionID"`
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID  `json:"id" gorm:"type:char(36);primary_key"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:char(36);not null"`
	Token        string     `json:"-" gorm:"uniqueIndex;not null"`
	RefreshToken string     `json:"-" gorm:"uniqueIndex;not null"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
	ExpiresAt    time.Time  `json:"expires_at"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uuid.UUID `json:"id" gorm:"type:char(36);primary_key"`
	UserID     *uuid.UUID `json:"user_id" gorm:"type:char(36)"`
	Action     string    `json:"action" gorm:"not null"`
	Resource   string    `json:"resource" gorm:"not null"`
	ResourceID *string   `json:"resource_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Details    string    `json:"details" gorm:"type:text"`
	Success    bool      `json:"success" gorm:"default:true"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for User model
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Role model
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Permission model
func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for Session model
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for AuditLog model
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// TableName returns the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// TableName returns the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}
