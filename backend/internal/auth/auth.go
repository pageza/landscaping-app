package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/pkg/security"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT claims
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Role       string    `json:"role"`
	TokenType  TokenType `json:"token_type"`
	SessionID  uuid.UUID `json:"session_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService interface {
	// Password operations
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	
	// Token operations
	GenerateTokens(user *domain.EnhancedUser, sessionID uuid.UUID) (*TokenPair, error)
	ValidateToken(tokenString string, tokenType TokenType) (*Claims, error)
	RefreshTokens(refreshToken string) (*TokenPair, error)
	RevokeToken(sessionID uuid.UUID) error
	
	// API Key operations
	GenerateAPIKey(name string, permissions []string) (*APIKeyPair, error)
	ValidateAPIKey(keyString string) (*APIKeyClaims, error)
	
	// Session operations
	CreateSession(userID uuid.UUID, deviceInfo map[string]interface{}, ipAddress, userAgent string) (*domain.UserSession, error)
	ValidateSession(sessionID uuid.UUID) (*domain.UserSession, error)
	RevokeSession(sessionID uuid.UUID) error
	RevokeAllUserSessions(userID uuid.UUID) error
	
	// Two-factor authentication
	GenerateTOTPSecret() (string, error)
	ValidateTOTP(secret, token string) bool
	GenerateBackupCodes() ([]string, error)
}

// TokenPair represents access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// APIKeyPair represents API key and its metadata
type APIKeyPair struct {
	Key       string    `json:"key"`
	KeyPrefix string    `json:"key_prefix"`
	KeyHash   string    `json:"-"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyClaims represents API key claims
type APIKeyClaims struct {
	TenantID    uuid.UUID `json:"tenant_id"`
	KeyID       uuid.UUID `json:"key_id"`
	Permissions []string  `json:"permissions"`
}

// authService implements AuthService
type authService struct {
	jwtSecret        []byte
	accessExpiry     time.Duration
	refreshExpiry    time.Duration
	bcryptCost       int
	sessionRepo      SessionRepository
	apiKeyRepo       APIKeyRepository
	totpManager      *security.TOTPManager
	backupCodeMgr    *security.BackupCodeManager
}

// SessionRepository defines session storage operations
type SessionRepository interface {
	CreateSession(ctx context.Context, session *domain.UserSession) error
	GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error)
	UpdateSession(ctx context.Context, sessionID uuid.UUID, lastActivity time.Time) error
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

// APIKeyRepository defines API key storage operations
type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	UpdateAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error
	RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error
	ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error)
}

// NewAuthService creates a new auth service
func NewAuthService(
	jwtSecret string,
	accessExpiry, refreshExpiry time.Duration,
	bcryptCost int,
	sessionRepo SessionRepository,
	apiKeyRepo APIKeyRepository,
) AuthService {
	return &authService{
		jwtSecret:     []byte(jwtSecret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		bcryptCost:    bcryptCost,
		sessionRepo:   sessionRepo,
		apiKeyRepo:    apiKeyRepo,
		totpManager:   security.NewTOTPManager("Landscaping App"),
		backupCodeMgr: security.NewBackupCodeManager(),
	}
}

// HashPassword hashes a password using bcrypt
func (s *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// ComparePassword compares a password with its hash
func (s *authService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateTokens generates access and refresh token pair
func (s *authService) GenerateTokens(user *domain.EnhancedUser, sessionID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(s.accessExpiry)
	refreshExpiresAt := now.Add(s.refreshExpiry)

	// Generate access token
	accessClaims := &Claims{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Role:        user.Role,
		TokenType:   AccessToken,
		SessionID:   sessionID,
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "landscaping-app",
			Audience:  []string{"landscaping-app"},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Role:      user.Role,
		TokenType: RefreshToken,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(refreshExpiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "landscaping-app",
			Audience:  []string{"landscaping-app"},
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(s.accessExpiry.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}, nil
}

// ValidateToken validates and parses a JWT token
func (s *authService) ValidateToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

// RefreshTokens generates new token pair using refresh token
func (s *authService) RefreshTokens(refreshToken string) (*TokenPair, error) {
	claims, err := s.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Validate session
	session, err := s.sessionRepo.GetSession(context.Background(), claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session is not active")
	}

	// Create new user object for token generation
	user := &domain.EnhancedUser{
		User: domain.User{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Role:     claims.Role,
		},
		Permissions: claims.Permissions,
	}

	// Generate new tokens
	tokens, err := s.GenerateTokens(user, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Update session activity
	err = s.sessionRepo.UpdateSession(context.Background(), claims.SessionID, time.Now())
	if err != nil {
		// Log error but don't fail the refresh
		fmt.Printf("Failed to update session activity: %v\n", err)
	}

	return tokens, nil
}

// RevokeToken revokes a token by revoking its session
func (s *authService) RevokeToken(sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeSession(context.Background(), sessionID)
}

// GenerateAPIKey generates a new API key
func (s *authService) GenerateAPIKey(name string, permissions []string) (*APIKeyPair, error) {
	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	key := base64.URLEncoding.EncodeToString(keyBytes)
	keyPrefix := key[:8]

	// Hash the key for storage
	keyHash, err := s.HashPassword(key)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	return &APIKeyPair{
		Key:       key,
		KeyPrefix: keyPrefix,
		KeyHash:   keyHash,
	}, nil
}

// ValidateAPIKey validates an API key
func (s *authService) ValidateAPIKey(keyString string) (*APIKeyClaims, error) {
	if len(keyString) < 8 {
		return nil, fmt.Errorf("invalid API key format")
	}

	// Hash the provided key to match against stored hash
	keyHash, err := s.HashPassword(keyString)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}

	// Look up API key by hash
	apiKey, err := s.apiKeyRepo.GetAPIKeyByHash(context.Background(), keyHash)
	if err != nil {
		return nil, fmt.Errorf("API key not found or invalid: %w", err)
	}

	// Update last used timestamp
	go func() {
		_ = s.apiKeyRepo.UpdateAPIKeyLastUsed(context.Background(), apiKey.ID)
	}()

	return &APIKeyClaims{
		TenantID:    apiKey.TenantID,
		KeyID:       apiKey.ID,
		Permissions: apiKey.Permissions,
	}, nil
}

// CreateSession creates a new user session
func (s *authService) CreateSession(userID uuid.UUID, deviceInfo map[string]interface{}, ipAddress, userAgent string) (*domain.UserSession, error) {
	session := &domain.UserSession{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: generateSessionToken(),
		RefreshToken: generateSessionToken(),
		DeviceInfo:   deviceInfo,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		ExpiresAt:    time.Now().Add(s.refreshExpiry),
		LastActivity: time.Now(),
		Status:       "active",
		CreatedAt:    time.Now(),
	}

	err := s.sessionRepo.CreateSession(context.Background(), session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session
func (s *authService) ValidateSession(sessionID uuid.UUID) (*domain.UserSession, error) {
	session, err := s.sessionRepo.GetSession(context.Background(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session is not active")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}

// RevokeSession revokes a specific session
func (s *authService) RevokeSession(sessionID uuid.UUID) error {
	return s.sessionRepo.RevokeSession(context.Background(), sessionID)
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *authService) RevokeAllUserSessions(userID uuid.UUID) error {
	return s.sessionRepo.RevokeAllUserSessions(context.Background(), userID)
}

// GenerateTOTPSecret generates a TOTP secret for 2FA
func (s *authService) GenerateTOTPSecret() (string, error) {
	// Generate a TOTP secret with QR code URL for a generic account
	// In practice, you'd pass the actual user's email or identifier
	totpSecret, err := s.totpManager.GenerateSecret("user@example.com")
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}
	return totpSecret.Secret, nil
}

// ValidateTOTP validates a TOTP token
func (s *authService) ValidateTOTP(secret, token string) bool {
	return s.totpManager.ValidateTokenWithSkew(secret, token)
}

// GenerateBackupCodes generates backup codes for 2FA
func (s *authService) GenerateBackupCodes() ([]string, error) {
	return s.backupCodeMgr.GenerateBackupCodes()
}

// Helper function to generate session tokens
func generateSessionToken() string {
	token := make([]byte, 32)
	rand.Read(token)
	return base64.URLEncoding.EncodeToString(token)
}

// Permission checking utilities

// HasPermission checks if a user has a specific permission
func HasPermission(userPermissions []string, requiredPermission string) bool {
	for _, permission := range userPermissions {
		if permission == requiredPermission || permission == "*" {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func HasAnyPermission(userPermissions []string, requiredPermissions []string) bool {
	for _, required := range requiredPermissions {
		if HasPermission(userPermissions, required) {
			return true
		}
	}
	return false
}

// HasAllPermissions checks if a user has all specified permissions
func HasAllPermissions(userPermissions []string, requiredPermissions []string) bool {
	for _, required := range requiredPermissions {
		if !HasPermission(userPermissions, required) {
			return false
		}
	}
	return true
}

// CanAccessTenant checks if a user can access a specific tenant
func CanAccessTenant(userTenantID, requestedTenantID uuid.UUID, userRole string) bool {
	// Super admins can access any tenant
	if userRole == domain.RoleSuperAdmin {
		return true
	}
	// Other users can only access their own tenant
	return userTenantID == requestedTenantID
}

// GetDefaultPermissions returns default permissions for a role
func GetDefaultPermissions(role string) []string {
	switch role {
	case domain.RoleSuperAdmin:
		return []string{"*"} // All permissions
	case domain.RoleOwner:
		return []string{
			domain.PermissionTenantManage,
			domain.PermissionUserManage,
			domain.PermissionCustomerManage,
			domain.PermissionPropertyManage,
			domain.PermissionJobManage,
			domain.PermissionJobAssign,
			domain.PermissionInvoiceManage,
			domain.PermissionPaymentManage,
			domain.PermissionEquipmentManage,
			domain.PermissionReportView,
			domain.PermissionWebhookManage,
			domain.PermissionAuditView,
		}
	case domain.RoleAdmin:
		return []string{
			domain.PermissionUserManage,
			domain.PermissionCustomerManage,
			domain.PermissionPropertyManage,
			domain.PermissionJobManage,
			domain.PermissionJobAssign,
			domain.PermissionInvoiceManage,
			domain.PermissionPaymentManage,
			domain.PermissionEquipmentManage,
			domain.PermissionReportView,
		}
	case domain.RoleUser:
		return []string{
			domain.PermissionCustomerManage,
			domain.PermissionPropertyManage,
			domain.PermissionJobManage,
			domain.PermissionInvoiceManage,
			domain.PermissionReportView,
		}
	case domain.RoleCrew:
		return []string{
			domain.PermissionJobManage, // Limited to assigned jobs
		}
	case domain.RoleCustomer:
		return []string{} // Very limited access
	default:
		return []string{}
	}
}