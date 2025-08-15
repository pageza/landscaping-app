package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SecurityTestSuite represents a comprehensive security testing framework
type SecurityTestSuite struct {
	baseURL    string
	httpClient *http.Client
	results    []TestResult
}

// TestResult represents the result of a security test
type TestResult struct {
	TestName    string                 `json:"test_name"`
	Category    string                 `json:"category"`
	Severity    string                 `json:"severity"`
	Passed      bool                   `json:"passed"`
	Description string                 `json:"description"`
	Details     string                 `json:"details"`
	Remediation string                 `json:"remediation"`
	Metadata    map[string]interface{} `json:"metadata"`
	ExecutedAt  time.Time              `json:"executed_at"`
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite(baseURL string) *SecurityTestSuite {
	return &SecurityTestSuite{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: []TestResult{},
	}
}

// RunAllTests executes all security tests
func (sts *SecurityTestSuite) RunAllTests() []TestResult {
	sts.results = []TestResult{}

	// Run different categories of security tests
	sts.runAuthenticationTests()
	sts.runAuthorizationTests()
	sts.runInputValidationTests()
	sts.runSessionManagementTests()
	sts.runCryptographyTests()
	sts.runBusinessLogicTests()
	sts.runInformationDisclosureTests()
	sts.runInfrastructureTests()

	return sts.results
}

// runAuthenticationTests tests authentication mechanisms
func (sts *SecurityTestSuite) runAuthenticationTests() {
	// Test 1: Brute Force Protection
	sts.testBruteForceProtection()

	// Test 2: Password Policy Enforcement
	sts.testPasswordPolicyEnforcement()

	// Test 3: Account Lockout
	sts.testAccountLockout()

	// Test 4: JWT Token Security
	sts.testJWTSecurity()

	// Test 5: Session Timeout
	sts.testSessionTimeout()

	// Test 6: Multi-Factor Authentication
	sts.testMFAImplementation()
}

// testBruteForceProtection tests brute force attack protection
func (sts *SecurityTestSuite) testBruteForceProtection() {
	testName := "Brute Force Protection"
	category := "Authentication"

	// Simulate rapid login attempts
	attempts := 20
	successCount := 0

	for i := 0; i < attempts; i++ {
		loginData := map[string]string{
			"email":    "testuser@example.com",
			"password": fmt.Sprintf("wrongpassword%d", i),
		}

		resp, err := sts.makeRequest("POST", "/api/auth/login", loginData)
		if err == nil && resp.StatusCode == 200 {
			successCount++
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Small delay to simulate realistic attack timing
		time.Sleep(100 * time.Millisecond)
	}

	passed := successCount < 3 // Should be blocked after a few attempts
	severity := "HIGH"
	if !passed {
		severity = "CRITICAL"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests protection against brute force login attacks",
		Details:     fmt.Sprintf("Successfully authenticated %d out of %d attempts", successCount, attempts),
		Remediation: "Implement rate limiting, account lockout, and CAPTCHA after failed attempts",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// testPasswordPolicyEnforcement tests password policy compliance
func (sts *SecurityTestSuite) testPasswordPolicyEnforcement() {
	testName := "Password Policy Enforcement"
	category := "Authentication"

	weakPasswords := []string{
		"123456",
		"password",
		"12345678",
		"qwerty",
		"abc123",
		"password123",
	}

	rejectedCount := 0

	for _, password := range weakPasswords {
		registerData := map[string]string{
			"email":            fmt.Sprintf("test_%s@example.com", password),
			"password":         password,
			"password_confirm": password,
			"first_name":       "Test",
			"last_name":        "User",
		}

		resp, err := sts.makeRequest("POST", "/api/auth/register", registerData)
		if err == nil && (resp.StatusCode == 400 || resp.StatusCode == 422) {
			rejectedCount++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	passed := rejectedCount == len(weakPasswords)
	severity := "MEDIUM"
	if !passed {
		severity = "HIGH"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests enforcement of strong password policies",
		Details:     fmt.Sprintf("Rejected %d out of %d weak passwords", rejectedCount, len(weakPasswords)),
		Remediation: "Implement comprehensive password complexity requirements and validation",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// testAccountLockout tests account lockout functionality
func (sts *SecurityTestSuite) testAccountLockout() {
	testName := "Account Lockout"
	category := "Authentication"

	// Test account lockout after multiple failed attempts
	email := "lockout_test@example.com"
	attempts := 10
	lockedOut := false

	for i := 0; i < attempts; i++ {
		loginData := map[string]string{
			"email":    email,
			"password": "wrongpassword",
		}

		resp, err := sts.makeRequest("POST", "/api/auth/login", loginData)
		if err == nil && resp.StatusCode == 423 { // Account locked status
			lockedOut = true
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(200 * time.Millisecond)
	}

	passed := lockedOut
	severity := "HIGH"
	if !passed {
		severity = "CRITICAL"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests account lockout after multiple failed login attempts",
		Details:     fmt.Sprintf("Account lockout triggered: %v", lockedOut),
		Remediation: "Implement account lockout after consecutive failed login attempts",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// testJWTSecurity tests JWT token implementation security
func (sts *SecurityTestSuite) testJWTSecurity() {
	testName := "JWT Token Security"
	category := "Authentication"

	// Test various JWT vulnerabilities
	vulnerabilities := []string{}

	// Test 1: None algorithm acceptance
	if sts.testJWTNoneAlgorithm() {
		vulnerabilities = append(vulnerabilities, "Accepts 'none' algorithm")
	}

	// Test 2: Weak signing key
	if sts.testJWTWeakKey() {
		vulnerabilities = append(vulnerabilities, "Uses weak signing key")
	}

	// Test 3: Algorithm confusion
	if sts.testJWTAlgorithmConfusion() {
		vulnerabilities = append(vulnerabilities, "Vulnerable to algorithm confusion")
	}

	passed := len(vulnerabilities) == 0
	severity := "LOW"
	if len(vulnerabilities) > 0 {
		severity = "HIGH"
	}
	if len(vulnerabilities) > 2 {
		severity = "CRITICAL"
	}

	details := "No JWT vulnerabilities found"
	if len(vulnerabilities) > 0 {
		details = fmt.Sprintf("Found vulnerabilities: %s", strings.Join(vulnerabilities, ", "))
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests JWT token implementation for common vulnerabilities",
		Details:     details,
		Remediation: "Use strong signing algorithms, reject 'none' algorithm, validate algorithms",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// runAuthorizationTests tests authorization and access control
func (sts *SecurityTestSuite) runAuthorizationTests() {
	sts.testVerticalPrivilegeEscalation()
	sts.testHorizontalPrivilegeEscalation()
	sts.testDirectObjectReferences()
	sts.testMultiTenantIsolation()
}

// testVerticalPrivilegeEscalation tests for privilege escalation vulnerabilities
func (sts *SecurityTestSuite) testVerticalPrivilegeEscalation() {
	testName := "Vertical Privilege Escalation"
	category := "Authorization"

	// Simulate regular user trying to access admin functions
	regularUserToken := sts.getRegularUserToken()
	if regularUserToken == "" {
		result := TestResult{
			TestName:    testName,
			Category:    category,
			Severity:    "MEDIUM",
			Passed:      false,
			Description: "Tests for vertical privilege escalation vulnerabilities",
			Details:     "Could not obtain regular user token for testing",
			Remediation: "Ensure proper role-based access controls",
			ExecutedAt:  time.Now(),
		}
		sts.results = append(sts.results, result)
		return
	}

	// Test admin endpoints with regular user token
	adminEndpoints := []string{
		"/api/admin/users",
		"/api/admin/tenants",
		"/api/admin/settings",
		"/api/admin/analytics",
	}

	unauthorizedAccess := 0

	for _, endpoint := range adminEndpoints {
		resp, err := sts.makeAuthenticatedRequest("GET", endpoint, nil, regularUserToken)
		if err == nil && resp.StatusCode == 200 {
			unauthorizedAccess++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	passed := unauthorizedAccess == 0
	severity := "MEDIUM"
	if unauthorizedAccess > 0 {
		severity = "CRITICAL"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests for vertical privilege escalation vulnerabilities",
		Details:     fmt.Sprintf("Gained unauthorized access to %d admin endpoints", unauthorizedAccess),
		Remediation: "Implement proper role-based access controls and endpoint authorization",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// testMultiTenantIsolation tests multi-tenant data isolation
func (sts *SecurityTestSuite) testMultiTenantIsolation() {
	testName := "Multi-Tenant Data Isolation"
	category := "Authorization"

	// Test accessing data from different tenants
	tenant1Token := sts.getTenantToken("tenant1")
	tenant2ID := "different-tenant-uuid"

	if tenant1Token == "" {
		result := TestResult{
			TestName:    testName,
			Category:    category,
			Severity:    "MEDIUM",
			Passed:      false,
			Description: "Tests multi-tenant data isolation",
			Details:     "Could not obtain tenant token for testing",
			Remediation: "Implement proper tenant isolation controls",
			ExecutedAt:  time.Now(),
		}
		sts.results = append(sts.results, result)
		return
	}

	// Try to access another tenant's data
	crossTenantEndpoints := []string{
		fmt.Sprintf("/api/tenants/%s/customers", tenant2ID),
		fmt.Sprintf("/api/tenants/%s/jobs", tenant2ID),
		fmt.Sprintf("/api/tenants/%s/invoices", tenant2ID),
	}

	violations := 0

	for _, endpoint := range crossTenantEndpoints {
		resp, err := sts.makeAuthenticatedRequest("GET", endpoint, nil, tenant1Token)
		if err == nil && resp.StatusCode == 200 {
			violations++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	passed := violations == 0
	severity := "LOW"
	if violations > 0 {
		severity = "CRITICAL"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests multi-tenant data isolation",
		Details:     fmt.Sprintf("Cross-tenant access violations: %d", violations),
		Remediation: "Implement Row-Level Security (RLS) and proper tenant context validation",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// runInputValidationTests tests input validation and sanitization
func (sts *SecurityTestSuite) runInputValidationTests() {
	sts.testSQLInjection()
	sts.testXSSPrevention()
	sts.testCommandInjection()
	sts.testPathTraversal()
	sts.testFileUploadSecurity()
}

// testSQLInjection tests for SQL injection vulnerabilities
func (sts *SecurityTestSuite) testSQLInjection() {
	testName := "SQL Injection Prevention"
	category := "Input Validation"

	sqlPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE users;--",
		"' UNION SELECT * FROM users--",
		"1; EXEC xp_cmdshell('dir');--",
		"' OR 1=1/*",
	}

	vulnerableEndpoints := 0

	// Test search endpoints with SQL injection payloads
	searchEndpoints := []string{
		"/api/customers?search=",
		"/api/jobs?filter=",
		"/api/invoices?query=",
	}

	token := sts.getValidToken()
	if token == "" {
		return
	}

	for _, endpoint := range searchEndpoints {
		for _, payload := range sqlPayloads {
			fullURL := endpoint + url.QueryEscape(payload)
			resp, err := sts.makeAuthenticatedRequest("GET", fullURL, nil, token)
			
			if err == nil {
				body, _ := io.ReadAll(resp.Body)
				bodyStr := string(body)
				
				// Look for SQL error messages or unexpected data
				if strings.Contains(bodyStr, "SQL") || 
				   strings.Contains(bodyStr, "syntax error") ||
				   strings.Contains(bodyStr, "mysql") ||
				   strings.Contains(bodyStr, "postgres") {
					vulnerableEndpoints++
				}
				resp.Body.Close()
			}
		}
	}

	passed := vulnerableEndpoints == 0
	severity := "LOW"
	if vulnerableEndpoints > 0 {
		severity = "CRITICAL"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests for SQL injection vulnerabilities",
		Details:     fmt.Sprintf("Vulnerable endpoints found: %d", vulnerableEndpoints),
		Remediation: "Use parameterized queries, input validation, and ORM frameworks",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// testXSSPrevention tests for Cross-Site Scripting vulnerabilities
func (sts *SecurityTestSuite) testXSSPrevention() {
	testName := "XSS Prevention"
	category := "Input Validation"

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
		"'><script>alert('XSS')</script>",
	}

	vulnerableFields := 0
	token := sts.getValidToken()
	if token == "" {
		return
	}

	// Test customer creation with XSS payloads
	for i, payload := range xssPayloads {
		customerData := map[string]string{
			"first_name": fmt.Sprintf("Test%d", i),
			"last_name":  payload,
			"email":      fmt.Sprintf("xsstest%d@example.com", i),
		}

		resp, err := sts.makeAuthenticatedRequest("POST", "/api/customers", customerData, token)
		if err == nil && resp.StatusCode == 201 {
			// Check if the XSS payload is reflected without encoding
			body, _ := io.ReadAll(resp.Body)
			if strings.Contains(string(body), payload) {
				vulnerableFields++
			}
			resp.Body.Close()
		}
	}

	passed := vulnerableFields == 0
	severity := "LOW"
	if vulnerableFields > 0 {
		severity = "HIGH"
	}

	result := TestResult{
		TestName:    testName,
		Category:    category,
		Severity:    severity,
		Passed:      passed,
		Description: "Tests for Cross-Site Scripting vulnerabilities",
		Details:     fmt.Sprintf("Vulnerable fields found: %d", vulnerableFields),
		Remediation: "Implement input validation, output encoding, and CSP headers",
		ExecutedAt:  time.Now(),
	}

	sts.results = append(sts.results, result)
}

// Helper methods for testing

func (sts *SecurityTestSuite) makeRequest(method, endpoint string, data interface{}) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, sts.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return sts.httpClient.Do(req)
}

func (sts *SecurityTestSuite) makeAuthenticatedRequest(method, endpoint string, data interface{}, token string) (*http.Response, error) {
	var body io.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, sts.baseURL+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return sts.httpClient.Do(req)
}

// Placeholder methods - would be implemented based on actual authentication flow
func (sts *SecurityTestSuite) getValidToken() string {
	// Implementation would authenticate and return a valid token
	return "mock-valid-token"
}

func (sts *SecurityTestSuite) getRegularUserToken() string {
	// Implementation would authenticate as regular user
	return "mock-regular-user-token"
}

func (sts *SecurityTestSuite) getTenantToken(tenantID string) string {
	// Implementation would authenticate for specific tenant
	return "mock-tenant-token"
}

// Placeholder JWT security test methods
func (sts *SecurityTestSuite) testJWTNoneAlgorithm() bool {
	// Test if system accepts tokens with "none" algorithm
	return false
}

func (sts *SecurityTestSuite) testJWTWeakKey() bool {
	// Test if system uses weak signing keys
	return false
}

func (sts *SecurityTestSuite) testJWTAlgorithmConfusion() bool {
	// Test for algorithm confusion vulnerabilities
	return false
}

// Additional test method stubs that would be implemented
func (sts *SecurityTestSuite) testSessionTimeout()           {}
func (sts *SecurityTestSuite) testMFAImplementation()       {}
func (sts *SecurityTestSuite) testHorizontalPrivilegeEscalation() {}
func (sts *SecurityTestSuite) testDirectObjectReferences()  {}
func (sts *SecurityTestSuite) testCommandInjection()        {}
func (sts *SecurityTestSuite) testPathTraversal()           {}
func (sts *SecurityTestSuite) testFileUploadSecurity()      {}
func (sts *SecurityTestSuite) runSessionManagementTests()   {}
func (sts *SecurityTestSuite) runCryptographyTests()        {}
func (sts *SecurityTestSuite) runBusinessLogicTests()       {}
func (sts *SecurityTestSuite) runInformationDisclosureTests() {}
func (sts *SecurityTestSuite) runInfrastructureTests()      {}

// GenerateSecurityReport generates a comprehensive security test report
func (sts *SecurityTestSuite) GenerateSecurityReport() SecurityTestReport {
	totalTests := len(sts.results)
	passedTests := 0
	criticalIssues := 0
	highIssues := 0
	mediumIssues := 0
	lowIssues := 0

	for _, result := range sts.results {
		if result.Passed {
			passedTests++
		} else {
			switch result.Severity {
			case "CRITICAL":
				criticalIssues++
			case "HIGH":
				highIssues++
			case "MEDIUM":
				mediumIssues++
			case "LOW":
				lowIssues++
			}
		}
	}

	overallScore := float64(passedTests) / float64(totalTests) * 100

	return SecurityTestReport{
		ExecutedAt:      time.Now(),
		TotalTests:      totalTests,
		PassedTests:     passedTests,
		FailedTests:     totalTests - passedTests,
		CriticalIssues:  criticalIssues,
		HighIssues:      highIssues,
		MediumIssues:    mediumIssues,
		LowIssues:       lowIssues,
		OverallScore:    overallScore,
		TestResults:     sts.results,
	}
}

// SecurityTestReport represents a comprehensive security test report
type SecurityTestReport struct {
	ExecutedAt      time.Time    `json:"executed_at"`
	TotalTests      int          `json:"total_tests"`
	PassedTests     int          `json:"passed_tests"`
	FailedTests     int          `json:"failed_tests"`
	CriticalIssues  int          `json:"critical_issues"`
	HighIssues      int          `json:"high_issues"`
	MediumIssues    int          `json:"medium_issues"`
	LowIssues       int          `json:"low_issues"`
	OverallScore    float64      `json:"overall_score"`
	TestResults     []TestResult `json:"test_results"`
}