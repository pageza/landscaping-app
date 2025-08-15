package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/scrypt"
)

// SecureAPIKeyManager handles secure API key generation and validation
type SecureAPIKeyManager struct {
	secretKey []byte
	keyLength int
}

// NewSecureAPIKeyManager creates a new secure API key manager
func NewSecureAPIKeyManager(secretKey string) *SecureAPIKeyManager {
	return &SecureAPIKeyManager{
		secretKey: []byte(secretKey),
		keyLength: 32, // 256 bits
	}
}

// SecureAPIKey represents a secure API key with metadata
type SecureAPIKey struct {
	ID          string    `json:"id"`
	KeyHash     string    `json:"key_hash"`
	KeyPreview  string    `json:"key_preview"`
	HMACKey     string    `json:"hmac_key"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Permissions []string  `json:"permissions"`
	Name        string    `json:"name"`
}

// GenerateSecureAPIKey generates a cryptographically secure API key using HMAC
func (m *SecureAPIKeyManager) GenerateSecureAPIKey(name string, permissions []string, expiresIn *time.Duration) (*SecureAPIKey, string, error) {
	// Generate random key material
	keyBytes := make([]byte, m.keyLength)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate random key: %w", err)
	}

	// Create key ID
	keyID := make([]byte, 16)
	if _, err := rand.Read(keyID); err != nil {
		return nil, "", fmt.Errorf("failed to generate key ID: %w", err)
	}

	keyIDStr := hex.EncodeToString(keyID)
	
	// Create the actual API key with structure: prefix.keyid.hmac
	keyPrefix := "lsk" // landscaping secret key
	baseKey := fmt.Sprintf("%s.%s", keyPrefix, keyIDStr)
	
	// Generate HMAC for the key
	h := hmac.New(sha256.New, m.secretKey)
	h.Write([]byte(baseKey))
	hmacBytes := h.Sum(nil)
	hmacStr := base64.RawURLEncoding.EncodeToString(hmacBytes)
	
	// Final API key format: lsk.keyid.hmac
	fullAPIKey := fmt.Sprintf("%s.%s", baseKey, hmacStr)
	
	// Create hash for storage (using scrypt for better security than bcrypt for API keys)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, "", fmt.Errorf("failed to generate salt: %w", err)
	}
	
	keyHash, err := scrypt.Key([]byte(fullAPIKey), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash API key: %w", err)
	}
	
	keyHashStr := fmt.Sprintf("scrypt:%s:%s", 
		base64.RawURLEncoding.EncodeToString(salt),
		base64.RawURLEncoding.EncodeToString(keyHash))

	var expiresAt *time.Time
	if expiresIn != nil {
		expiry := time.Now().Add(*expiresIn)
		expiresAt = &expiry
	}

	apiKey := &SecureAPIKey{
		ID:          keyIDStr,
		KeyHash:     keyHashStr,
		KeyPreview:  fmt.Sprintf("%s.%s.***", keyPrefix, keyIDStr[:8]),
		HMACKey:     hmacStr[:8], // Store preview of HMAC for identification
		CreatedAt:   time.Now(),
		ExpiresAt:   expiresAt,
		Permissions: permissions,
		Name:        name,
	}

	return apiKey, fullAPIKey, nil
}

// ValidateAPIKey validates an API key using constant-time comparison
func (m *SecureAPIKeyManager) ValidateAPIKey(providedKey, storedHash string) (bool, error) {
	// Parse stored hash
	parts := strings.Split(storedHash, ":")
	if len(parts) != 3 || parts[0] != "scrypt" {
		return false, fmt.Errorf("invalid stored hash format")
	}

	salt, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("invalid salt in stored hash: %w", err)
	}

	expectedHash, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("invalid hash in stored hash: %w", err)
	}

	// Hash the provided key with the same parameters
	providedHash, err := scrypt.Key([]byte(providedKey), salt, 32768, 8, 1, 32)
	if err != nil {
		return false, fmt.Errorf("failed to hash provided key: %w", err)
	}

	// Constant-time comparison
	return subtle.ConstantTimeCompare(expectedHash, providedHash) == 1, nil
}

// ValidateAPIKeyStructure validates the structure of an API key
func (m *SecureAPIKeyManager) ValidateAPIKeyStructure(apiKey string) error {
	parts := strings.Split(apiKey, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid API key format: expected 3 parts")
	}

	// Validate prefix
	if parts[0] != "lsk" {
		return fmt.Errorf("invalid API key prefix")
	}

	// Validate key ID (should be 32 hex characters)
	if len(parts[1]) != 32 {
		return fmt.Errorf("invalid key ID length")
	}
	
	if _, err := hex.DecodeString(parts[1]); err != nil {
		return fmt.Errorf("invalid key ID format: %w", err)
	}

	// Validate HMAC part
	if _, err := base64.RawURLEncoding.DecodeString(parts[2]); err != nil {
		return fmt.Errorf("invalid HMAC format: %w", err)
	}

	// Verify HMAC
	baseKey := fmt.Sprintf("lsk.%s", parts[1])
	h := hmac.New(sha256.New, m.secretKey)
	h.Write([]byte(baseKey))
	expectedHMAC := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedHMAC)) != 1 {
		return fmt.Errorf("invalid API key HMAC")
	}

	return nil
}

// JWTSecurityManager handles enhanced JWT security
type JWTSecurityManager struct {
	secretKey       []byte
	issuer          string
	allowedAudiences []string
	clockSkew       time.Duration
}

// NewJWTSecurityManager creates a new JWT security manager
func NewJWTSecurityManager(secretKey, issuer string, audiences []string) *JWTSecurityManager {
	return &JWTSecurityManager{
		secretKey:       []byte(secretKey),
		issuer:          issuer,
		allowedAudiences: audiences,
		clockSkew:       5 * time.Minute, // Allow 5 minutes clock skew
	}
}

// SecureJWTClaims extends standard JWT claims with security features
type SecureJWTClaims struct {
	UserID      string   `json:"uid"`
	TenantID    string   `json:"tid"`
	SessionID   string   `json:"sid"`
	Permissions []string `json:"perms"`
	TokenType   string   `json:"typ"`
	IPAddress   string   `json:"ip,omitempty"`
	UserAgent   string   `json:"ua,omitempty"`
	jwt.RegisteredClaims
}

// GenerateSecureJWT generates a JWT with enhanced security claims
func (j *JWTSecurityManager) GenerateSecureJWT(claims *SecureJWTClaims, expiry time.Duration) (string, error) {
	now := time.Now()
	
	// Set standard claims with security best practices
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    j.issuer,
		Subject:   claims.UserID,
		Audience:  j.allowedAudiences,
		ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        generateJTI(), // Unique token ID for revocation
	}

	// Create token with security headers
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Add security headers
	token.Header["alg"] = "HS256"
	token.Header["typ"] = "JWT"
	token.Header["kid"] = "primary" // Key ID for key rotation

	return token.SignedString(j.secretKey)
}

// ValidateSecureJWT validates a JWT with enhanced security checks
func (j *JWTSecurityManager) ValidateSecureJWT(tokenString string, expectedType string, clientIP string) (*SecureJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SecureJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Validate algorithm specifically
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("invalid algorithm: %s", token.Method.Alg())
		}

		// Validate key ID if present
		if kid, ok := token.Header["kid"].(string); ok && kid != "primary" {
			return nil, fmt.Errorf("invalid key ID: %s", kid)
		}

		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*SecureJWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Enhanced validation checks
	if err := j.validateClaims(claims, expectedType, clientIP); err != nil {
		return nil, err
	}

	return claims, nil
}

// validateClaims performs enhanced claim validation
func (j *JWTSecurityManager) validateClaims(claims *SecureJWTClaims, expectedType string, clientIP string) error {
	now := time.Now()

	// Validate token type
	if claims.TokenType != expectedType {
		return fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	// Validate audience
	validAudience := false
	for _, aud := range claims.Audience {
		for _, allowed := range j.allowedAudiences {
			if aud == allowed {
				validAudience = true
				break
			}
		}
		if validAudience {
			break
		}
	}
	if !validAudience {
		return fmt.Errorf("invalid audience")
	}

	// Enhanced time validation with clock skew
	if claims.ExpiresAt != nil && now.After(claims.ExpiresAt.Time.Add(j.clockSkew)) {
		return fmt.Errorf("token is expired")
	}

	if claims.NotBefore != nil && now.Before(claims.NotBefore.Time.Add(-j.clockSkew)) {
		return fmt.Errorf("token used before valid")
	}

	if claims.IssuedAt != nil && now.Before(claims.IssuedAt.Time.Add(-j.clockSkew)) {
		return fmt.Errorf("token used before issued")
	}

	// Validate IP address if present (optional security feature)
	if claims.IPAddress != "" && clientIP != "" && claims.IPAddress != clientIP {
		return fmt.Errorf("token IP mismatch")
	}

	// Validate required fields
	if claims.UserID == "" || claims.TenantID == "" || claims.SessionID == "" {
		return fmt.Errorf("missing required claims")
	}

	return nil
}

// generateJTI generates a unique JWT ID for token tracking and revocation
func generateJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.RawURLEncoding.EncodeToString(bytes)
}

// TokenRevocationManager manages revoked tokens
type TokenRevocationManager struct {
	revokedTokens map[string]time.Time // JTI -> expiry time
}

// NewTokenRevocationManager creates a new token revocation manager
func NewTokenRevocationManager() *TokenRevocationManager {
	return &TokenRevocationManager{
		revokedTokens: make(map[string]time.Time),
	}
}

// RevokeToken adds a token to the revocation list
func (trm *TokenRevocationManager) RevokeToken(jti string, expiry time.Time) {
	trm.revokedTokens[jti] = expiry
}

// IsTokenRevoked checks if a token is revoked
func (trm *TokenRevocationManager) IsTokenRevoked(jti string) bool {
	expiry, exists := trm.revokedTokens[jti]
	if !exists {
		return false
	}

	// Clean up expired revocation entries
	if time.Now().After(expiry) {
		delete(trm.revokedTokens, jti)
		return false
	}

	return true
}

// CleanupExpiredRevocations removes expired revocation entries
func (trm *TokenRevocationManager) CleanupExpiredRevocations() {
	now := time.Now()
	for jti, expiry := range trm.revokedTokens {
		if now.After(expiry) {
			delete(trm.revokedTokens, jti)
		}
	}
}