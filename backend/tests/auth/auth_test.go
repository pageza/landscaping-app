package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/auth"
	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// MockSessionRepository is a mock implementation of auth.SessionRepository
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSession), args.Error(1)
}

func (m *MockSessionRepository) UpdateSession(ctx context.Context, sessionID uuid.UUID, lastActivity time.Time) error {
	args := m.Called(ctx, sessionID, lastActivity)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockAPIKeyRepository is a mock implementation of auth.APIKeyRepository
type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) UpdateAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.APIKey), args.Error(1)
}

// Test data
func createTestUser() *domain.EnhancedUser {
	return &domain.EnhancedUser{
		User: domain.User{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Role:     domain.RoleUser,
		},
		Permissions: []string{domain.PermissionCustomerManage, domain.PermissionJobManage},
	}
}

func createTestSession(userID uuid.UUID) *domain.UserSession {
	return &domain.UserSession{
		ID:           uuid.New(),
		UserID:       userID,
		SessionToken: "test-session-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
		Status:       "active",
		CreatedAt:    time.Now(),
	}
}

func createTestAPIKey(tenantID uuid.UUID) *domain.APIKey {
	return &domain.APIKey{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "Test API Key",
		KeyHash:     "hashed-key",
		KeyPrefix:   "test-key",
		Permissions: []string{domain.PermissionCustomerManage},
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TestAuthService_HashPassword tests password hashing functionality
func TestAuthService_HashPassword(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
		{
			name:     "long password",
			password: string(make([]byte, 100)),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := authSvc.HashPassword(tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
				
				// Verify password
				err = authSvc.ComparePassword(hash, tt.password)
				assert.NoError(t, err)
			}
		})
	}
}

// TestAuthService_ComparePassword tests password verification
func TestAuthService_ComparePassword(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	password := "testpassword123"
	hash, err := authSvc.HashPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name         string
		hashedPassword string
		password     string
		wantErr      bool
	}{
		{
			name:         "correct password",
			hashedPassword: hash,
			password:     password,
			wantErr:      false,
		},
		{
			name:         "incorrect password",
			hashedPassword: hash,
			password:     "wrongpassword",
			wantErr:      true,
		},
		{
			name:         "empty password",
			hashedPassword: hash,
			password:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authSvc.ComparePassword(tt.hashedPassword, tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAuthService_GenerateTokens tests JWT token generation
func TestAuthService_GenerateTokens(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	user := createTestUser()
	sessionID := uuid.New()

	tokens, err := authSvc.GenerateTokens(user, sessionID)
	
	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.True(t, tokens.ExpiresIn > 0)
	assert.True(t, tokens.ExpiresAt.After(time.Now()))
}

// TestAuthService_ValidateToken tests JWT token validation
func TestAuthService_ValidateToken(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	user := createTestUser()
	sessionID := uuid.New()

	// Generate valid tokens
	tokens, err := authSvc.GenerateTokens(user, sessionID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		tokenType auth.TokenType
		wantErr   bool
	}{
		{
			name:      "valid access token",
			token:     tokens.AccessToken,
			tokenType: auth.AccessToken,
			wantErr:   false,
		},
		{
			name:      "valid refresh token",
			token:     tokens.RefreshToken,
			tokenType: auth.RefreshToken,
			wantErr:   false,
		},
		{
			name:      "wrong token type",
			token:     tokens.AccessToken,
			tokenType: auth.RefreshToken,
			wantErr:   true,
		},
		{
			name:      "invalid token",
			token:     "invalid.token.here",
			tokenType: auth.AccessToken,
			wantErr:   true,
		},
		{
			name:      "empty token",
			token:     "",
			tokenType: auth.AccessToken,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authSvc.ValidateToken(tt.token, tt.tokenType)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.ID, claims.UserID)
				assert.Equal(t, user.TenantID, claims.TenantID)
				assert.Equal(t, user.Role, claims.Role)
				assert.Equal(t, sessionID, claims.SessionID)
				assert.Equal(t, tt.tokenType, claims.TokenType)
			}
		})
	}
}

// TestAuthService_CreateSession tests session creation
func TestAuthService_CreateSession(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	userID := uuid.New()
	deviceInfo := map[string]interface{}{
		"platform": "web",
		"browser":  "Chrome",
	}
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0..."

	// Mock successful session creation
	mockSessionRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*domain.UserSession")).Return(nil)

	session, err := authSvc.CreateSession(userID, deviceInfo, ipAddress, userAgent)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, userID, session.UserID)
	assert.Equal(t, &ipAddress, session.IPAddress)
	assert.Equal(t, &userAgent, session.UserAgent)
	assert.Equal(t, "active", session.Status)
	assert.NotEmpty(t, session.SessionToken)
	assert.NotEmpty(t, session.RefreshToken)
	
	mockSessionRepo.AssertExpectations(t)
}

// TestAuthService_ValidateSession tests session validation
func TestAuthService_ValidateSession(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	sessionID := uuid.New()
	validSession := createTestSession(uuid.New())
	expiredSession := createTestSession(uuid.New())
	expiredSession.ExpiresAt = time.Now().Add(-time.Hour) // Expired

	tests := []struct {
		name      string
		sessionID uuid.UUID
		mockSetup func()
		wantErr   bool
	}{
		{
			name:      "valid session",
			sessionID: sessionID,
			mockSetup: func() {
				mockSessionRepo.On("GetSession", mock.Anything, sessionID).Return(validSession, nil)
			},
			wantErr: false,
		},
		{
			name:      "expired session",
			sessionID: sessionID,
			mockSetup: func() {
				mockSessionRepo.On("GetSession", mock.Anything, sessionID).Return(expiredSession, nil)
			},
			wantErr: true,
		},
		{
			name:      "session not found",
			sessionID: sessionID,
			mockSetup: func() {
				mockSessionRepo.On("GetSession", mock.Anything, sessionID).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockSessionRepo.ExpectedCalls = nil
			tt.mockSetup()

			session, err := authSvc.ValidateSession(tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, validSession.ID, session.ID)
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}

// TestAuthService_GenerateTOTPSecret tests TOTP secret generation
func TestAuthService_GenerateTOTPSecret(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	secret, err := authSvc.GenerateTOTPSecret()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.True(t, len(secret) > 20) // Base32 encoded secret should be longer than 20 chars
}

// TestAuthService_ValidateTOTP tests TOTP token validation
func TestAuthService_ValidateTOTP(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	// Generate a secret and get current token
	secret, err := authSvc.GenerateTOTPSecret()
	require.NoError(t, err)

	tests := []struct {
		name   string
		secret string
		token  string
		want   bool
	}{
		{
			name:   "empty token",
			secret: secret,
			token:  "",
			want:   false,
		},
		{
			name:   "invalid token",
			secret: secret,
			token:  "123456",
			want:   false,
		},
		{
			name:   "empty secret",
			secret: "",
			token:  "123456",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authSvc.ValidateTOTP(tt.secret, tt.token)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestAuthService_GenerateBackupCodes tests backup code generation
func TestAuthService_GenerateBackupCodes(t *testing.T) {
	mockSessionRepo := &MockSessionRepository{}
	mockAPIKeyRepo := &MockAPIKeyRepository{}
	
	authSvc := auth.NewAuthService(
		"test-secret",
		time.Hour,
		24*time.Hour,
		12,
		mockSessionRepo,
		mockAPIKeyRepo,
	)

	codes, err := authSvc.GenerateBackupCodes()
	
	assert.NoError(t, err)
	assert.Len(t, codes, 10) // Should generate 10 backup codes
	
	// Ensure all codes are unique
	codeSet := make(map[string]bool)
	for _, code := range codes {
		assert.NotEmpty(t, code)
		assert.False(t, codeSet[code], "Duplicate backup code generated")
		codeSet[code] = true
	}
}

// TestAuthService_Permission_Helpers tests permission checking utilities
func TestAuthService_PermissionHelpers(t *testing.T) {
	tests := []struct {
		name               string
		userPermissions    []string
		requiredPermission string
		want               bool
	}{
		{
			name:               "has specific permission",
			userPermissions:    []string{domain.PermissionCustomerManage, domain.PermissionJobManage},
			requiredPermission: domain.PermissionCustomerManage,
			want:               true,
		},
		{
			name:               "has wildcard permission",
			userPermissions:    []string{"*"},
			requiredPermission: domain.PermissionCustomerManage,
			want:               true,
		},
		{
			name:               "does not have permission",
			userPermissions:    []string{domain.PermissionJobManage},
			requiredPermission: domain.PermissionCustomerManage,
			want:               false,
		},
		{
			name:               "empty permissions",
			userPermissions:    []string{},
			requiredPermission: domain.PermissionCustomerManage,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.HasPermission(tt.userPermissions, tt.requiredPermission)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestAuthService_CanAccessTenant tests tenant access validation
func TestAuthService_CanAccessTenant(t *testing.T) {
	userTenantID := uuid.New()
	otherTenantID := uuid.New()

	tests := []struct {
		name              string
		userTenantID      uuid.UUID
		requestedTenantID uuid.UUID
		userRole          string
		want              bool
	}{
		{
			name:              "same tenant access",
			userTenantID:      userTenantID,
			requestedTenantID: userTenantID,
			userRole:          domain.RoleUser,
			want:              true,
		},
		{
			name:              "super admin cross-tenant access",
			userTenantID:      userTenantID,
			requestedTenantID: otherTenantID,
			userRole:          domain.RoleSuperAdmin,
			want:              true,
		},
		{
			name:              "regular user cross-tenant denied",
			userTenantID:      userTenantID,
			requestedTenantID: otherTenantID,
			userRole:          domain.RoleUser,
			want:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.CanAccessTenant(tt.userTenantID, tt.requestedTenantID, tt.userRole)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestAuthService_GetDefaultPermissions tests default permission assignment
func TestAuthService_GetDefaultPermissions(t *testing.T) {
	tests := []struct {
		name            string
		role            string
		expectedContains []string
		expectedLength   int
	}{
		{
			name:            "super admin permissions",
			role:            domain.RoleSuperAdmin,
			expectedContains: []string{"*"},
			expectedLength:   1,
		},
		{
			name:            "owner permissions",
			role:            domain.RoleOwner,
			expectedContains: []string{domain.PermissionTenantManage, domain.PermissionUserManage, domain.PermissionAuditView},
			expectedLength:   12,
		},
		{
			name:            "user permissions",
			role:            domain.RoleUser,
			expectedContains: []string{domain.PermissionCustomerManage, domain.PermissionJobManage},
			expectedLength:   5,
		},
		{
			name:            "customer permissions",
			role:            domain.RoleCustomer,
			expectedContains: []string{},
			expectedLength:   0,
		},
		{
			name:            "unknown role",
			role:            "unknown",
			expectedContains: []string{},
			expectedLength:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permissions := auth.GetDefaultPermissions(tt.role)
			
			assert.Len(t, permissions, tt.expectedLength)
			
			for _, expectedPerm := range tt.expectedContains {
				assert.Contains(t, permissions, expectedPerm)
			}
		})
	}
}