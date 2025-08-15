package security

import (
	"fmt"
	"html"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// InputValidator provides comprehensive input validation and sanitization
type InputValidator struct {
	maxStringLength int
	allowedTags     []string
	blockedPatterns []*regexp.Regexp
}

// NewInputValidator creates a new input validator with security defaults
func NewInputValidator() *InputValidator {
	// Common SQL injection patterns
	sqlPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|update\s+.*\s+set)`),
		regexp.MustCompile(`(?i)(exec(\s|\+)+(s|x)p\w+|sp_\w+)`),
		regexp.MustCompile(`(?i)(--|#|/\*|\*/)`),
		regexp.MustCompile(`(?i)(\bor\b.*=.*\bor\b|\band\b.*=.*\band\b)`),
		regexp.MustCompile(`(?i)(drop\s+table|truncate\s+table|alter\s+table)`),
		regexp.MustCompile(`(?i)('(\s*|\+|%20)(or|and)(\s*|\+|%20)')`),
	}

	// XSS patterns
	xssPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)<iframe\b[^>]*>`),
		regexp.MustCompile(`(?i)<object\b[^>]*>`),
		regexp.MustCompile(`(?i)<embed\b[^>]*>`),
	}

	blockedPatterns := append(sqlPatterns, xssPatterns...)

	return &InputValidator{
		maxStringLength: 10000,
		allowedTags:     []string{"p", "br", "strong", "em"},
		blockedPatterns: blockedPatterns,
	}
}

// ValidateAndSanitizeString validates and sanitizes string input
func (v *InputValidator) ValidateAndSanitizeString(input string) (string, error) {
	if len(input) > v.maxStringLength {
		return "", fmt.Errorf("input exceeds maximum length of %d characters", v.maxStringLength)
	}

	// Check for malicious patterns
	for _, pattern := range v.blockedPatterns {
		if pattern.MatchString(input) {
			return "", fmt.Errorf("input contains potentially malicious content")
		}
	}

	// HTML encode the input to prevent XSS
	sanitized := html.EscapeString(input)
	
	// Normalize unicode characters
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1 // Remove control characters
		}
		return r
	}, sanitized)

	return strings.TrimSpace(sanitized), nil
}

// ValidateEmail validates and sanitizes email addresses
func (v *InputValidator) ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	
	// Basic email regex - not perfect but good enough for validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}

	if len(email) > 254 { // RFC 5321 limit
		return "", fmt.Errorf("email address too long")
	}

	// Check for suspicious patterns in email
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[<>'"()[\]]`), // HTML/script injection characters
		regexp.MustCompile(`\s`),          // No spaces allowed
		regexp.MustCompile(`\.{2,}`),      // No consecutive dots
	}

	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(email) {
			return "", fmt.Errorf("email contains invalid characters")
		}
	}

	return email, nil
}

// ValidatePhoneNumber validates and formats phone numbers
func (v *InputValidator) ValidatePhoneNumber(phone string) (string, error) {
	// Remove all non-digit characters except + for international numbers
	phoneRegex := regexp.MustCompile(`[^\d+]`)
	cleaned := phoneRegex.ReplaceAllString(phone, "")

	// Basic phone number validation
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return "", fmt.Errorf("invalid phone number length")
	}

	// Check for valid phone number patterns
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^\+?1?[2-9]\d{2}[2-9]\d{2}\d{4}$`), // US format
		regexp.MustCompile(`^\+\d{1,3}\d{7,12}$`),               // International format
	}

	isValid := false
	for _, pattern := range validPatterns {
		if pattern.MatchString(cleaned) {
			isValid = true
			break
		}
	}

	if !isValid {
		return "", fmt.Errorf("invalid phone number format")
	}

	return cleaned, nil
}

// ValidateURL validates and sanitizes URL input
func (v *InputValidator) ValidateURL(rawURL string) (string, error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow HTTP and HTTPS schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("only HTTP and HTTPS URLs are allowed")
	}

	// Validate hostname
	if parsedURL.Host == "" {
		return "", fmt.Errorf("URL must have a valid hostname")
	}

	// Check for potentially dangerous URLs
	dangerousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(javascript|vbscript|data):`),
		regexp.MustCompile(`(?i)localhost|127\.0\.0\.1|0\.0\.0\.0`),
		regexp.MustCompile(`(?i)\.local|\.internal`),
	}

	urlString := parsedURL.String()
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(urlString) {
			return "", fmt.Errorf("URL contains potentially dangerous content")
		}
	}

	return urlString, nil
}

// ValidateIPAddress validates IP addresses and blocks private ranges in production
func (v *InputValidator) ValidateIPAddress(ipStr string, allowPrivate bool) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address format")
	}

	// Check for private IP ranges if not allowed
	if !allowPrivate {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return "", fmt.Errorf("private IP addresses not allowed")
		}
	}

	return ip.String(), nil
}

// ValidateJSONField validates JSON input for specific fields
func (v *InputValidator) ValidateJSONField(jsonStr string, maxDepth, maxKeys int) error {
	if len(jsonStr) > 100000 { // 100KB limit for JSON
		return fmt.Errorf("JSON input too large")
	}

	// Basic JSON structure validation
	depth := 0
	maxDepthSeen := 0
	keyCount := 0

	for i, char := range jsonStr {
		switch char {
		case '{', '[':
			depth++
			if depth > maxDepthSeen {
				maxDepthSeen = depth
			}
			if maxDepthSeen > maxDepth {
				return fmt.Errorf("JSON nesting depth exceeds limit of %d", maxDepth)
			}
		case '}', ']':
			depth--
		case '"':
			// Count potential keys (simplified)
			if i > 0 && jsonStr[i-1] != '\\' {
				keyCount++
			}
		}
	}

	if keyCount > maxKeys*2 { // Rough estimate (keys + values)
		return fmt.Errorf("JSON contains too many keys (limit: %d)", maxKeys)
	}

	return nil
}

// SanitizeFilename sanitizes filenames for safe storage
func (v *InputValidator) SanitizeFilename(filename string) (string, error) {
	if strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")

	// Remove dangerous characters
	dangerousChars := regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)
	filename = dangerousChars.ReplaceAllString(filename, "")

	// Limit filename length
	if len(filename) > 255 {
		return "", fmt.Errorf("filename too long (max 255 characters)")
	}

	// Check for reserved Windows filenames
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperFilename := strings.ToUpper(filename)
	for _, reserved := range reservedNames {
		if upperFilename == reserved || strings.HasPrefix(upperFilename, reserved+".") {
			return "", fmt.Errorf("filename uses reserved name: %s", reserved)
		}
	}

	return filename, nil
}

// ValidateUUID validates UUID format
func (v *InputValidator) ValidateUUID(uuid string) error {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// ValidateAndSanitizeHTML sanitizes HTML content while preserving safe tags
func (v *InputValidator) ValidateAndSanitizeHTML(htmlContent string) (string, error) {
	if len(htmlContent) > v.maxStringLength {
		return "", fmt.Errorf("HTML content exceeds maximum length")
	}

	// Remove potentially dangerous HTML elements and attributes
	dangerousElements := regexp.MustCompile(`(?i)<(script|iframe|object|embed|form|input|button|select|textarea|link|meta|style)[^>]*>.*?</\1>`)
	htmlContent = dangerousElements.ReplaceAllString(htmlContent, "")

	// Remove dangerous attributes
	dangerousAttrs := regexp.MustCompile(`(?i)\s+(on\w+|javascript:|vbscript:|data:)\s*=\s*[^>\s]+`)
	htmlContent = dangerousAttrs.ReplaceAllString(htmlContent, "")

	// Basic HTML escaping for remaining content
	htmlContent = html.EscapeString(htmlContent)

	return htmlContent, nil
}

// ValidationResult represents the result of input validation
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	SanitizedInput string   `json:"sanitized_input,omitempty"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// ValidateStruct validates multiple fields in a structured way
func (v *InputValidator) ValidateStruct(fields map[string]interface{}) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []string{},
		Warnings: []string{},
	}

	for fieldName, value := range fields {
		switch val := value.(type) {
		case string:
			if sanitized, err := v.ValidateAndSanitizeString(val); err != nil {
				result.IsValid = false
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", fieldName, err))
			} else {
				result.SanitizedInput += fmt.Sprintf("%s: %s; ", fieldName, sanitized)
			}
		default:
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: unsupported validation type", fieldName))
		}
	}

	return result
}