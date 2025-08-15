package security

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	UserID      string                 `json:"user_id,omitempty"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// SecurityMonitor handles security event monitoring and alerting
type SecurityMonitor struct {
	redisClient    *redis.Client
	eventQueue     chan SecurityEvent
	alertThresholds map[string]AlertThreshold
	suspiciousIPs   *sync.Map
	blockedIPs      *sync.Map
	mu             sync.RWMutex
}

// AlertThreshold defines when to trigger alerts for specific event types
type AlertThreshold struct {
	Count        int           `json:"count"`
	TimeWindow   time.Duration `json:"time_window"`
	Severity     string        `json:"severity"`
	AutoBlock    bool          `json:"auto_block"`
	BlockDuration time.Duration `json:"block_duration"`
}

// NewSecurityMonitor creates a new security monitoring system
func NewSecurityMonitor(redisClient *redis.Client) *SecurityMonitor {
	monitor := &SecurityMonitor{
		redisClient:   redisClient,
		eventQueue:    make(chan SecurityEvent, 1000),
		alertThresholds: map[string]AlertThreshold{
			"failed_login": {
				Count:         5,
				TimeWindow:    5 * time.Minute,
				Severity:      "HIGH",
				AutoBlock:     true,
				BlockDuration: 30 * time.Minute,
			},
			"suspicious_request": {
				Count:         10,
				TimeWindow:    10 * time.Minute,
				Severity:      "MEDIUM",
				AutoBlock:     false,
				BlockDuration: 0,
			},
			"rate_limit_exceeded": {
				Count:         3,
				TimeWindow:    5 * time.Minute,
				Severity:      "HIGH",
				AutoBlock:     true,
				BlockDuration: 15 * time.Minute,
			},
			"invalid_token": {
				Count:         10,
				TimeWindow:    10 * time.Minute,
				Severity:      "HIGH",
				AutoBlock:     true,
				BlockDuration: 1 * time.Hour,
			},
			"privilege_escalation": {
				Count:         1,
				TimeWindow:    1 * time.Minute,
				Severity:      "CRITICAL",
				AutoBlock:     true,
				BlockDuration: 24 * time.Hour,
			},
			"data_access_violation": {
				Count:         3,
				TimeWindow:    5 * time.Minute,
				Severity:      "CRITICAL",
				AutoBlock:     true,
				BlockDuration: 2 * time.Hour,
			},
		},
		suspiciousIPs: &sync.Map{},
		blockedIPs:    &sync.Map{},
	}

	// Start event processor
	go monitor.processEvents()

	return monitor
}

// LogSecurityEvent logs a security event for monitoring
func (sm *SecurityMonitor) LogSecurityEvent(event SecurityEvent) {
	event.Timestamp = time.Now()
	event.ID = generateSecureID()

	select {
	case sm.eventQueue <- event:
		// Event queued successfully
	default:
		// Queue is full, handle gracefully
		fmt.Printf("Security event queue full, dropping event: %s\n", event.Type)
	}
}

// processEvents processes security events and triggers alerts
func (sm *SecurityMonitor) processEvents() {
	for event := range sm.eventQueue {
		go sm.handleSecurityEvent(event)
	}
}

// handleSecurityEvent processes individual security events
func (sm *SecurityMonitor) handleSecurityEvent(event SecurityEvent) {
	// Store event in Redis for analysis
	if err := sm.storeEvent(event); err != nil {
		fmt.Printf("Failed to store security event: %v\n", err)
	}

	// Check if this event triggers an alert
	if threshold, exists := sm.alertThresholds[event.Type]; exists {
		if sm.shouldTriggerAlert(event, threshold) {
			sm.triggerAlert(event, threshold)
		}
	}

	// Update suspicious IP tracking
	sm.updateSuspiciousIPTracking(event)
}

// storeEvent stores a security event in Redis
func (sm *SecurityMonitor) storeEvent(event SecurityEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal security event: %w", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("security:events:%s:%s", event.Type, event.ID)

	// Store event with TTL
	if err := sm.redisClient.Set(ctx, key, eventJSON, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store event in Redis: %w", err)
	}

	// Add to time-series for analysis
	timeSeriesKey := fmt.Sprintf("security:timeseries:%s", event.Type)
	score := float64(event.Timestamp.Unix())
	if err := sm.redisClient.ZAdd(ctx, timeSeriesKey, redis.Z{
		Score:  score,
		Member: event.ID,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to time series: %w", err)
	}

	// Set TTL on time series
	sm.redisClient.Expire(ctx, timeSeriesKey, 7*24*time.Hour)

	return nil
}

// shouldTriggerAlert checks if an event should trigger an alert
func (sm *SecurityMonitor) shouldTriggerAlert(event SecurityEvent, threshold AlertThreshold) bool {
	ctx := context.Background()
	timeSeriesKey := fmt.Sprintf("security:timeseries:%s", event.Type)

	// Count events in the time window
	now := time.Now()
	windowStart := now.Add(-threshold.TimeWindow).Unix()

	count, err := sm.redisClient.ZCount(ctx, timeSeriesKey, 
		fmt.Sprintf("%d", windowStart), "+inf").Result()
	if err != nil {
		fmt.Printf("Failed to count events for alert: %v\n", err)
		return false
	}

	return int(count) >= threshold.Count
}

// triggerAlert triggers an alert for a security event
func (sm *SecurityMonitor) triggerAlert(event SecurityEvent, threshold AlertThreshold) {
	alert := SecurityAlert{
		ID:          generateSecureID(),
		EventType:   event.Type,
		Severity:    threshold.Severity,
		IPAddress:   event.IPAddress,
		UserID:      event.UserID,
		TenantID:    event.TenantID,
		Count:       threshold.Count,
		TimeWindow:  threshold.TimeWindow,
		TriggeredAt: time.Now(),
		AutoBlocked: threshold.AutoBlock,
		Description: fmt.Sprintf("Alert triggered for %s: %d events in %v", 
			event.Type, threshold.Count, threshold.TimeWindow),
	}

	// Store alert
	if err := sm.storeAlert(alert); err != nil {
		fmt.Printf("Failed to store security alert: %v\n", err)
	}

	// Auto-block if configured
	if threshold.AutoBlock && event.IPAddress != "" {
		sm.blockIP(event.IPAddress, threshold.BlockDuration, 
			fmt.Sprintf("Auto-blocked due to %s threshold exceeded", event.Type))
	}

	// Send notifications (implement based on your notification system)
	sm.sendAlertNotification(alert)
}

// SecurityAlert represents a triggered security alert
type SecurityAlert struct {
	ID          string        `json:"id"`
	EventType   string        `json:"event_type"`
	Severity    string        `json:"severity"`
	IPAddress   string        `json:"ip_address"`
	UserID      string        `json:"user_id,omitempty"`
	TenantID    string        `json:"tenant_id,omitempty"`
	Count       int           `json:"count"`
	TimeWindow  time.Duration `json:"time_window"`
	TriggeredAt time.Time     `json:"triggered_at"`
	AutoBlocked bool          `json:"auto_blocked"`
	Description string        `json:"description"`
	Resolved    bool          `json:"resolved"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
}

// storeAlert stores a security alert
func (sm *SecurityMonitor) storeAlert(alert SecurityAlert) error {
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	ctx := context.Background()
	key := fmt.Sprintf("security:alerts:%s", alert.ID)

	return sm.redisClient.Set(ctx, key, alertJSON, 30*24*time.Hour).Err()
}

// blockIP blocks an IP address for a specific duration
func (sm *SecurityMonitor) blockIP(ipAddress string, duration time.Duration, reason string) {
	ctx := context.Background()
	blockKey := fmt.Sprintf("security:blocked:%s", ipAddress)

	blockInfo := map[string]interface{}{
		"blocked_at": time.Now().Unix(),
		"expires_at": time.Now().Add(duration).Unix(),
		"reason":     reason,
		"duration":   duration.Seconds(),
	}

	blockJSON, _ := json.Marshal(blockInfo)
	sm.redisClient.Set(ctx, blockKey, blockJSON, duration)

	// Add to local cache
	sm.blockedIPs.Store(ipAddress, time.Now().Add(duration))

	fmt.Printf("SECURITY: Blocked IP %s for %v - %s\n", ipAddress, duration, reason)
}

// IsIPBlocked checks if an IP address is blocked
func (sm *SecurityMonitor) IsIPBlocked(ipAddress string) (bool, string) {
	// Check local cache first
	if expiry, exists := sm.blockedIPs.Load(ipAddress); exists {
		if time.Now().Before(expiry.(time.Time)) {
			return true, "IP address is temporarily blocked due to security violation"
		}
		sm.blockedIPs.Delete(ipAddress)
	}

	// Check Redis
	ctx := context.Background()
	blockKey := fmt.Sprintf("security:blocked:%s", ipAddress)

	blockData, err := sm.redisClient.Get(ctx, blockKey).Result()
	if err != nil {
		return false, ""
	}

	var blockInfo map[string]interface{}
	if err := json.Unmarshal([]byte(blockData), &blockInfo); err != nil {
		return false, ""
	}

	// Check if block is still active
	expiresAt := int64(blockInfo["expires_at"].(float64))
	if time.Now().Unix() > expiresAt {
		sm.redisClient.Del(ctx, blockKey)
		return false, ""
	}

	reason := blockInfo["reason"].(string)
	return true, reason
}

// updateSuspiciousIPTracking updates suspicious IP tracking
func (sm *SecurityMonitor) updateSuspiciousIPTracking(event SecurityEvent) {
	if event.IPAddress == "" {
		return
	}

	// Calculate suspicion score based on event type and severity
	score := sm.calculateSuspicionScore(event)
	
	ctx := context.Background()
	suspicionKey := fmt.Sprintf("security:suspicion:%s", event.IPAddress)

	// Add to suspicious IP score with expiry
	sm.redisClient.ZIncrBy(ctx, suspicionKey, score, event.IPAddress)
	sm.redisClient.Expire(ctx, suspicionKey, 24*time.Hour)

	// Update local cache
	currentScore, _ := sm.redisClient.ZScore(ctx, suspicionKey, event.IPAddress).Result()
	if currentScore > 50 { // Threshold for marking as suspicious
		sm.suspiciousIPs.Store(event.IPAddress, time.Now().Add(24*time.Hour))
	}
}

// calculateSuspicionScore calculates a suspicion score for an event
func (sm *SecurityMonitor) calculateSuspicionScore(event SecurityEvent) float64 {
	baseScores := map[string]float64{
		"failed_login":           10,
		"invalid_token":          15,
		"rate_limit_exceeded":    5,
		"suspicious_request":     8,
		"privilege_escalation":   50,
		"data_access_violation":  30,
		"unusual_access_pattern": 12,
	}

	severityMultipliers := map[string]float64{
		"LOW":      1.0,
		"MEDIUM":   1.5,
		"HIGH":     2.0,
		"CRITICAL": 3.0,
	}

	baseScore := baseScores[event.Type]
	if baseScore == 0 {
		baseScore = 5 // Default score for unknown events
	}

	multiplier := severityMultipliers[event.Severity]
	if multiplier == 0 {
		multiplier = 1.0
	}

	return baseScore * multiplier
}

// sendAlertNotification sends alert notifications (implement based on your system)
func (sm *SecurityMonitor) sendAlertNotification(alert SecurityAlert) {
	// This would integrate with your notification system (email, Slack, PagerDuty, etc.)
	fmt.Printf("SECURITY ALERT [%s]: %s\n", alert.Severity, alert.Description)
	
	// Example webhook notification
	go sm.sendWebhookAlert(alert)
}

// sendWebhookAlert sends alert to configured webhook endpoints
func (sm *SecurityMonitor) sendWebhookAlert(alert SecurityAlert) {
	// Implementation would depend on your webhook configuration
	fmt.Printf("Webhook alert sent for: %s\n", alert.ID)
}

// GetSecurityMetrics returns security metrics for monitoring dashboards
func (sm *SecurityMonitor) GetSecurityMetrics(timeRange time.Duration) (SecurityMetrics, error) {
	ctx := context.Background()
	now := time.Now()
	startTime := now.Add(-timeRange)

	metrics := SecurityMetrics{
		TimeRange:   timeRange,
		GeneratedAt: now,
		EventCounts: make(map[string]int),
		AlertCounts: make(map[string]int),
	}

	// Count events by type
	for eventType := range sm.alertThresholds {
		timeSeriesKey := fmt.Sprintf("security:timeseries:%s", eventType)
		count, err := sm.redisClient.ZCount(ctx, timeSeriesKey,
			fmt.Sprintf("%d", startTime.Unix()), "+inf").Result()
		if err == nil {
			metrics.EventCounts[eventType] = int(count)
		}
	}

	// Count blocked IPs
	blockedKeys, err := sm.redisClient.Keys(ctx, "security:blocked:*").Result()
	if err == nil {
		metrics.BlockedIPs = len(blockedKeys)
	}

	// Count active alerts
	alertKeys, err := sm.redisClient.Keys(ctx, "security:alerts:*").Result()
	if err == nil {
		metrics.ActiveAlerts = len(alertKeys)
	}

	return metrics, nil
}

// SecurityMetrics represents security monitoring metrics
type SecurityMetrics struct {
	TimeRange    time.Duration      `json:"time_range"`
	GeneratedAt  time.Time          `json:"generated_at"`
	EventCounts  map[string]int     `json:"event_counts"`
	AlertCounts  map[string]int     `json:"alert_counts"`
	BlockedIPs   int                `json:"blocked_ips"`
	ActiveAlerts int                `json:"active_alerts"`
}

// IntrusionDetectionMiddleware detects and prevents intrusion attempts
func (sm *SecurityMonitor) IntrusionDetectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		
		// Check if IP is blocked
		if blocked, reason := sm.IsIPBlocked(clientIP); blocked {
			sm.LogSecurityEvent(SecurityEvent{
				Type:        "blocked_ip_access",
				Severity:    "HIGH",
				Source:      "intrusion_detection",
				IPAddress:   clientIP,
				UserAgent:   r.UserAgent(),
				Description: fmt.Sprintf("Blocked IP attempted access: %s", reason),
			})

			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Analyze request for suspicious patterns
		if sm.isRequestSuspicious(r) {
			sm.LogSecurityEvent(SecurityEvent{
				Type:        "suspicious_request",
				Severity:    "MEDIUM",
				Source:      "intrusion_detection",
				IPAddress:   clientIP,
				UserAgent:   r.UserAgent(),
				Description: fmt.Sprintf("Suspicious request pattern detected: %s %s", r.Method, r.URL.Path),
				Metadata: map[string]interface{}{
					"method":      r.Method,
					"path":        r.URL.Path,
					"query":       r.URL.RawQuery,
					"content_type": r.Header.Get("Content-Type"),
				},
			})
		}

		next.ServeHTTP(w, r)
	})
}

// isRequestSuspicious analyzes request patterns for suspicious activity
func (sm *SecurityMonitor) isRequestSuspicious(r *http.Request) bool {
	// Check for common attack patterns
	suspiciousPatterns := []string{
		"../", "..\\", // Path traversal
		"<script", "javascript:", // XSS attempts
		"union select", "drop table", // SQL injection
		"/etc/passwd", "/etc/shadow", // File disclosure attempts
		"cmd.exe", "/bin/sh", // Command injection
		"php://", "file://", // Protocol manipulation
	}

	// Check URL, query parameters, and headers
	checkStrings := []string{
		r.URL.Path,
		r.URL.RawQuery,
		r.Header.Get("User-Agent"),
		r.Header.Get("Referer"),
	}

	for _, checkStr := range checkStrings {
		checkLower := strings.ToLower(checkStr)
		for _, pattern := range suspiciousPatterns {
			if strings.Contains(checkLower, pattern) {
				return true
			}
		}
	}

	// Check for unusual request methods
	unusualMethods := []string{"TRACE", "CONNECT", "PATCH", "PROPFIND"}
	for _, method := range unusualMethods {
		if r.Method == method {
			return true
		}
	}

	// Check for suspicious user agents
	suspiciousUserAgents := []string{
		"sqlmap", "nmap", "nikto", "burpsuite", "owasp zap",
		"python-requests", "curl", "wget", // Tool indicators
	}
	
	userAgent := strings.ToLower(r.UserAgent())
	for _, suspicious := range suspiciousUserAgents {
		if strings.Contains(userAgent, suspicious) {
			return true
		}
	}

	return false
}

// Helper function to extract client IP
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// generateSecureID generates a secure random ID
func generateSecureID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("err_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", bytes)
}