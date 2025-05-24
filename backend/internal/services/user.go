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

// UserService handles user-related operations
type UserService struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *UserService {
	return &UserService{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).
		Preload("Roles").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUsers retrieves all users with pagination
func (s *UserService) GetUsers(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	// Get total count
	if err := s.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	if err := s.db.WithContext(ctx).
		Preload("Roles").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
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

	if err := s.db.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Reload user with relationships
	if err := s.db.WithContext(ctx).Preload("Roles").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	return &user, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{}).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// AssignRole assigns a role to a user
func (s *UserService) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	// Check if user exists
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if role exists
	var role models.Role
	if err := s.db.WithContext(ctx).Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if assignment already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing assignment: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("user already has this role")
	}

	// Create assignment
	userRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := s.db.WithContext(ctx).Create(userRole).Error; err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RemoveRole removes a role from a user
func (s *UserService) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles for a user
func (s *UserService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.Role, error) {
	var roles []*models.Role
	if err := s.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *UserService) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*models.Permission, error) {
	var permissions []*models.Permission
	if err := s.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct().
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if a user has a specific permission
func (s *UserService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&models.Permission{}).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.resource = ? AND permissions.action = ?", userID, resource, action).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return count > 0, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := s.db.WithContext(ctx).Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// EnableTwoFactor enables two-factor authentication for a user
func (s *UserService) EnableTwoFactor(ctx context.Context, userID uuid.UUID, secret string) error {
	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_two_factor_enabled": true,
			"two_factor_secret":     secret,
		}).Error; err != nil {
		return fmt.Errorf("failed to enable two-factor authentication: %w", err)
	}

	return nil
}

// DisableTwoFactor disables two-factor authentication for a user
func (s *UserService) DisableTwoFactor(ctx context.Context, userID uuid.UUID) error {
	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_two_factor_enabled": false,
			"two_factor_secret":     "",
		}).Error; err != nil {
		return fmt.Errorf("failed to disable two-factor authentication: %w", err)
	}

	return nil
}
