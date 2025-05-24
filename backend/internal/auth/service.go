package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/mynodecp/mynodecp/backend/internal/config"
	"github.com/mynodecp/mynodecp/backend/internal/models"
)

// Service handles authentication operations
type Service struct {
	db     *gorm.DB
	redis  *redis.Client
	config config.AuthConfig
}

// NewService creates a new authentication service
func NewService(db *gorm.DB, redis *redis.Client, config config.AuthConfig) *Service {
	return &Service{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	SessionID uuid.UUID `json:"session_id"`
	jwt.RegisteredClaims
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TwoFactorCode string `json:"two_factor_code,omitempty"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *models.User `json:"user"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Find user by username or email
	var user models.User
	if err := s.db.WithContext(ctx).
		Preload("Roles").
		Where("username = ? OR email = ?", req.Username, req.Username).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is disabled")
	}

	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, fmt.Errorf("account is locked until %v", user.LockedUntil)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed login count
		s.incrementFailedLogin(ctx, &user, req.IPAddress)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check two-factor authentication if enabled
	if user.IsTwoFactorEnabled {
		if req.TwoFactorCode == "" {
			return nil, fmt.Errorf("two-factor code required")
		}
		if !s.verifyTwoFactorCode(user.TwoFactorSecret, req.TwoFactorCode) {
			return nil, fmt.Errorf("invalid two-factor code")
		}
	}

	// Reset failed login count on successful login
	if err := s.db.WithContext(ctx).Model(&user).Updates(map[string]interface{}{
		"failed_login_count": 0,
		"locked_until":       nil,
		"last_login_at":      time.Now(),
		"last_login_ip":      req.IPAddress,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update user login info: %w", err)
	}

	// Create session
	session, err := s.createSession(ctx, &user, req.IPAddress, req.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(&user, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with tokens
	session.Token = accessToken
	session.RefreshToken = refreshToken
	if err := s.db.WithContext(ctx).Save(session).Error; err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Store session in Redis
	if err := s.storeSessionInRedis(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session in Redis: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		User:         &user,
	}, nil
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if username or email already exists
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ? OR email = ?", req.Username, req.Email).
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("username or email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		IsActive:     true,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default role
	if err := s.assignDefaultRole(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to assign default role: %w", err)
	}

	return user, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken refreshes an access token using a refresh token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Find session by refresh token
	var session models.Session
	if err := s.db.WithContext(ctx).
		Preload("User.Roles").
		Where("refresh_token = ? AND revoked_at IS NULL AND expires_at > ?", refreshToken, time.Now()).
		First(&session).Error; err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(&session.User, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update session
	session.Token = accessToken
	session.LastUsedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Update session in Redis
	if err := s.storeSessionInRedis(ctx, &session); err != nil {
		return nil, fmt.Errorf("failed to update session in Redis: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		User:         &session.User,
	}, nil
}

// Logout revokes a session
func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now()
	
	// Revoke session in database
	if err := s.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Remove session from Redis
	if err := s.redis.Del(ctx, fmt.Sprintf("session:%s", sessionID)).Err(); err != nil {
		return fmt.Errorf("failed to remove session from Redis: %w", err)
	}

	return nil
}

// Helper methods

func (s *Service) incrementFailedLogin(ctx context.Context, user *models.User, ipAddress string) {
	user.FailedLoginCount++
	
	// Lock account after 5 failed attempts
	if user.FailedLoginCount >= 5 {
		lockUntil := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &lockUntil
	}

	s.db.WithContext(ctx).Save(user)

	// Log security event
	securityEvent := &models.SecurityEvent{
		UserID:      &user.ID,
		Type:        "login_failed",
		Severity:    "medium",
		Source:      "web",
		IPAddress:   ipAddress,
		Description: fmt.Sprintf("Failed login attempt for user %s", user.Username),
	}
	s.db.WithContext(ctx).Create(securityEvent)
}

func (s *Service) createSession(ctx context.Context, user *models.User, ipAddress, userAgent string) (*models.Session, error) {
	session := &models.Session{
		UserID:     user.ID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  time.Now().Add(s.config.RefreshExpiration),
		LastUsedAt: time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) generateAccessToken(user *models.User, sessionID uuid.UUID) (string, error) {
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	claims := &Claims{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Roles:     roles,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mynodecp",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *Service) storeSessionInRedis(ctx context.Context, session *models.Session) error {
	key := fmt.Sprintf("session:%s", session.ID)
	return s.redis.Set(ctx, key, session.UserID.String(), s.config.SessionTimeout).Err()
}

func (s *Service) validatePassword(password string) error {
	if len(password) < s.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", s.config.PasswordMinLength)
	}
	// Add more password validation logic here
	return nil
}

func (s *Service) verifyTwoFactorCode(secret, code string) bool {
	// Implement TOTP verification here
	// This is a placeholder - you would use a library like github.com/pquerna/otp
	return true
}

func (s *Service) assignDefaultRole(ctx context.Context, user *models.User) error {
	// Find or create default user role
	var role models.Role
	if err := s.db.WithContext(ctx).Where("name = ?", "user").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default user role
			role = models.Role{
				Name:        "user",
				DisplayName: "User",
				Description: "Default user role",
				IsSystem:    true,
			}
			if err := s.db.WithContext(ctx).Create(&role).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Assign role to user
	userRole := &models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	}
	return s.db.WithContext(ctx).Create(userRole).Error
}
