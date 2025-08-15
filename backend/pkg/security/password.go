package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// PasswordValidator validates password strength
type PasswordValidator struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumbers   bool
	RequireSpecial   bool
	ForbiddenWords   []string
}

// DefaultPasswordValidator returns a validator with sensible defaults
func DefaultPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
		ForbiddenWords: []string{
			"password", "123456", "qwerty", "admin", "letmein",
			"welcome", "monkey", "dragon", "pass", "master",
		},
	}
}

// ValidationError represents a password validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate checks if a password meets the requirements
func (v *PasswordValidator) Validate(password string) []ValidationError {
	var errors []ValidationError

	if len(password) < v.MinLength {
		errors = append(errors, ValidationError{
			Field:   "length",
			Message: fmt.Sprintf("Password must be at least %d characters long", v.MinLength),
		})
	}

	if v.RequireUppercase && !hasUppercase(password) {
		errors = append(errors, ValidationError{
			Field:   "uppercase",
			Message: "Password must contain at least one uppercase letter",
		})
	}

	if v.RequireLowercase && !hasLowercase(password) {
		errors = append(errors, ValidationError{
			Field:   "lowercase",
			Message: "Password must contain at least one lowercase letter",
		})
	}

	if v.RequireNumbers && !hasNumbers(password) {
		errors = append(errors, ValidationError{
			Field:   "numbers",
			Message: "Password must contain at least one number",
		})
	}

	if v.RequireSpecial && !hasSpecialChars(password) {
		errors = append(errors, ValidationError{
			Field:   "special",
			Message: "Password must contain at least one special character",
		})
	}

	if v.containsForbiddenWords(password) {
		errors = append(errors, ValidationError{
			Field:   "forbidden",
			Message: "Password contains forbidden words or common patterns",
		})
	}

	return errors
}

// IsValid returns true if the password is valid
func (v *PasswordValidator) IsValid(password string) bool {
	return len(v.Validate(password)) == 0
}

// Helper functions

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasNumbers(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecialChars(s string) bool {
	specialPattern := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
	return specialPattern.MatchString(s)
}

func (v *PasswordValidator) containsForbiddenWords(password string) bool {
	lowerPassword := strings.ToLower(password)
	for _, word := range v.ForbiddenWords {
		if strings.Contains(lowerPassword, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// PasswordHasher handles password hashing operations
type PasswordHasher struct {
	Cost int
}

// NewPasswordHasher creates a new password hasher with the given cost
func NewPasswordHasher(cost int) *PasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordHasher{Cost: cost}
}

// Hash hashes a password using bcrypt
func (h *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// Verify verifies a password against its hash
func (h *PasswordHasher) Verify(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// NeedsRehash checks if the password needs to be rehashed (e.g., cost changed)
func (h *PasswordHasher) NeedsRehash(hashedPassword string) bool {
	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		return true
	}
	return cost != h.Cost
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSecureBytes generates cryptographically secure random bytes
func GenerateSecureBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure bytes: %w", err)
	}
	return bytes, nil
}

// ConstantTimeCompare performs a constant-time comparison of two strings
func ConstantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}