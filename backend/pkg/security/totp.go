package security

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPManager handles Time-based One-Time Password operations
type TOTPManager struct {
	issuer      string
	algorithm   otp.Algorithm
	digits      otp.Digits
	period      uint
	skew        uint
	secretSize  uint
}

// NewTOTPManager creates a new TOTP manager with default settings
func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		issuer:     issuer,
		algorithm:  otp.AlgorithmSHA1,
		digits:     otp.DigitsSix,
		period:     30,
		skew:       1,
		secretSize: 20,
	}
}

// TOTPSecret represents a TOTP secret and its metadata
type TOTPSecret struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
	ManualKey string `json:"manual_key"`
	Issuer    string `json:"issuer"`
	Account   string `json:"account"`
}

// GenerateSecret generates a new TOTP secret for a user
func (tm *TOTPManager) GenerateSecret(accountName string) (*TOTPSecret, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      tm.issuer,
		AccountName: accountName,
		SecretSize:  tm.secretSize,
		Algorithm:   tm.algorithm,
		Digits:      tm.digits,
		Period:      tm.period,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return &TOTPSecret{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
		ManualKey: formatManualKey(key.Secret()),
		Issuer:    tm.issuer,
		Account:   accountName,
	}, nil
}

// ValidateToken validates a TOTP token against a secret
func (tm *TOTPManager) ValidateToken(secret, token string) bool {
	return totp.Validate(token, secret)
}

// ValidateTokenWithSkew validates a TOTP token with time skew tolerance
func (tm *TOTPManager) ValidateTokenWithSkew(secret, token string) bool {
	// Validate current time window
	if totp.Validate(token, secret) {
		return true
	}

	// Check previous and next time windows based on skew
	now := time.Now()
	for i := uint(1); i <= tm.skew; i++ {
		// Check previous windows
		pastTime := now.Add(-time.Duration(i) * time.Duration(tm.period) * time.Second)
		if validateTokenAtTime(secret, token, pastTime) {
			return true
		}

		// Check future windows
		futureTime := now.Add(time.Duration(i) * time.Duration(tm.period) * time.Second)
		if validateTokenAtTime(secret, token, futureTime) {
			return true
		}
	}

	return false
}

// GenerateCurrentToken generates the current TOTP token for a secret
func (tm *TOTPManager) GenerateCurrentToken(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}

// GetTimeRemaining returns the seconds remaining until the next token
func (tm *TOTPManager) GetTimeRemaining() uint64 {
	now := time.Now()
	period := time.Duration(tm.period) * time.Second
	return uint64(period.Seconds()) - uint64(now.Unix())%uint64(period.Seconds())
}

// BackupCodeManager handles backup codes for 2FA
type BackupCodeManager struct {
	codeLength int
	codeCount  int
}

// NewBackupCodeManager creates a new backup code manager
func NewBackupCodeManager() *BackupCodeManager {
	return &BackupCodeManager{
		codeLength: 8,
		codeCount:  10,
	}
}

// GenerateBackupCodes generates a set of backup codes
func (bcm *BackupCodeManager) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, bcm.codeCount)
	
	for i := 0; i < bcm.codeCount; i++ {
		code, err := generateBackupCode(bcm.codeLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}
	
	return codes, nil
}

// HashBackupCodes hashes backup codes for secure storage
func (bcm *BackupCodeManager) HashBackupCodes(codes []string) ([]string, error) {
	hasher := NewPasswordHasher(12) // Higher cost for backup codes
	hashedCodes := make([]string, len(codes))
	
	for i, code := range codes {
		hashed, err := hasher.Hash(code)
		if err != nil {
			return nil, fmt.Errorf("failed to hash backup code: %w", err)
		}
		hashedCodes[i] = hashed
	}
	
	return hashedCodes, nil
}

// ValidateBackupCode validates a backup code against hashed codes
func (bcm *BackupCodeManager) ValidateBackupCode(code string, hashedCodes []string) (bool, int) {
	hasher := NewPasswordHasher(12)
	
	for i, hashedCode := range hashedCodes {
		if hashedCode == "" {
			continue // Skip already used codes
		}
		
		if hasher.Verify(hashedCode, code) == nil {
			return true, i
		}
	}
	
	return false, -1
}

// Helper functions

func validateTokenAtTime(secret, token string, t time.Time) bool {
	code, err := totp.GenerateCodeCustom(secret, t, totp.ValidateOpts{
		Period:    30,
		Skew:      0,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		return false
	}
	return code == token
}

func formatManualKey(secret string) string {
	// Format secret as groups of 4 characters for easier manual entry
	var formatted string
	for i, r := range secret {
		if i > 0 && i%4 == 0 {
			formatted += " "
		}
		formatted += string(r)
	}
	return formatted
}

func generateBackupCode(length int) (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := make([]byte, length)
	
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	
	// Insert hyphen in the middle for readability
	if length >= 6 {
		middle := length / 2
		result := string(bytes[:middle]) + "-" + string(bytes[middle:])
		return result, nil
	}
	
	return string(bytes), nil
}

// ParseTOTPURL parses a TOTP URL and extracts the secret
func ParseTOTPURL(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid TOTP URL: %w", err)
	}
	
	if u.Scheme != "otpauth" || u.Host != "totp" {
		return "", fmt.Errorf("invalid TOTP URL scheme or type")
	}
	
	secret := u.Query().Get("secret")
	if secret == "" {
		return "", fmt.Errorf("secret not found in TOTP URL")
	}
	
	// Validate base32 encoding
	_, err = base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}
	
	return secret, nil
}