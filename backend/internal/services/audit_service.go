package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// AuditServiceImpl implements comprehensive audit logging
type AuditServiceImpl struct {
	db     *sql.DB
	logger *log.Logger
}

// AuditEvent represents an audit log entry
type AuditEvent struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UserID       *uuid.UUID             `json:"user_id"`
	SessionID    *string                `json:"session_id"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id"`
	OldValues    map[string]interface{} `json:"old_values"`
	NewValues    map[string]interface{} `json:"new_values"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	RequestID    *string                `json:"request_id"`
	Success      bool                   `json:"success"`
	ErrorMessage *string                `json:"error_message"`
	Duration     *int64                 `json:"duration_ms"`
	CreatedAt    time.Time              `json:"created_at"`
}

// AuditFilter represents filtering options for audit logs
type AuditFilter struct {
	UserID       *uuid.UUID `json:"user_id"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Success      *bool      `json:"success"`
	IPAddress    string     `json:"ip_address"`
	Page         int        `json:"page"`
	PerPage      int        `json:"per_page"`
	SortBy       string     `json:"sort_by"`
	SortDesc     bool       `json:"sort_desc"`
}

// AuditStatistics represents audit statistics
type AuditStatistics struct {
	Period              TimeRange                `json:"period"`
	TotalEvents         int64                    `json:"total_events"`
	SuccessfulEvents    int64                    `json:"successful_events"`
	FailedEvents        int64                    `json:"failed_events"`
	UniqueUsers         int64                    `json:"unique_users"`
	TopActions          []ActionCount            `json:"top_actions"`
	TopResourceTypes    []ResourceTypeCount      `json:"top_resource_types"`
	EventsByDay         []DayEventCount          `json:"events_by_day"`
	UserActivity        []UserActivitySummary    `json:"user_activity"`
	SecurityEvents      []SecurityEvent          `json:"security_events"`
}

// NewAuditService creates a new audit service instance
func NewAuditService(db *sql.DB, logger *log.Logger) AuditService {
	return &AuditServiceImpl{
		db:     db,
		logger: logger,
	}
}

// LogAction logs an audit event
func (s *AuditServiceImpl) LogAction(ctx context.Context, req *AuditLogRequest) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		// For system-level actions, allow logging without tenant
		tenantID = uuid.Nil
	}

	// Get additional context from request
	sessionID := GetSessionIDFromContext(ctx)
	requestID := GetRequestIDFromContext(ctx)
	ipAddress := GetIPAddressFromContext(ctx)
	userAgent := GetUserAgentFromContext(ctx)

	// Create audit event
	event := &AuditEvent{
		ID:           uuid.New(),
		TenantID:     tenantID,
		UserID:       req.UserID,
		SessionID:    &sessionID,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		OldValues:    req.OldValues,
		NewValues:    req.NewValues,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		RequestID:    &requestID,
		Success:      req.ErrorMessage == nil,
		ErrorMessage: req.ErrorMessage,
		Duration:     convertDurationToMs(req.Duration),
		CreatedAt:    time.Now(),
	}

	// Insert audit event into database
	if err := s.insertAuditEvent(ctx, event); err != nil {
		s.logger.Printf("Failed to insert audit event: %v, action: %s", err, req.Action)
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	// Check for security-relevant events
	if s.isSecurityEvent(req.Action) {
		if err := s.handleSecurityEvent(ctx, event); err != nil {
			s.logger.Printf("Failed to handle security event: %v, action: %s", err, req.Action)
		}
	}

	s.logger.Printf("Audit event logged: event_id=%s, action=%s", event.ID, req.Action)
	return nil
}

// GetAuditLog retrieves audit logs with filtering and pagination
func (s *AuditServiceImpl) GetAuditLogs(ctx context.Context, filter *AuditFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &AuditFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 1000 {
		filter.PerPage = 1000
	}

	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIndex := 2

	if filter.UserID != nil {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Action != "" {
		whereClause += fmt.Sprintf(" AND action ILIKE $%d", argIndex)
		args = append(args, "%"+filter.Action+"%")
		argIndex++
	}

	if filter.ResourceType != "" {
		whereClause += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, filter.ResourceType)
		argIndex++
	}

	if filter.ResourceID != nil {
		whereClause += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *filter.ResourceID)
		argIndex++
	}

	if filter.StartDate != nil {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}

	if filter.Success != nil {
		whereClause += fmt.Sprintf(" AND success = $%d", argIndex)
		args = append(args, *filter.Success)
		argIndex++
	}

	if filter.IPAddress != "" {
		whereClause += fmt.Sprintf(" AND ip_address = $%d", argIndex)
		args = append(args, filter.IPAddress)
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM audit_events " + whereClause
	var total int64
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortDesc {
			direction = "DESC"
		}
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]bool{
			"created_at":     true,
			"action":         true,
			"resource_type":  true,
			"user_id":        true,
			"success":        true,
		}
		if validSortFields[filter.SortBy] {
			orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
		}
	}

	// Add pagination
	limit := filter.PerPage
	offset := (filter.Page - 1) * filter.PerPage
	paginationClause := fmt.Sprintf(" %s LIMIT $%d OFFSET $%d", orderBy, argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute main query
	query := `
		SELECT 
			id, tenant_id, user_id, session_id, action, resource_type, resource_id,
			old_values, new_values, ip_address, user_agent, request_id, success,
			error_message, duration_ms, created_at
		FROM audit_events ` + whereClause + paginationClause

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	events := make([]*AuditEvent, 0)
	for rows.Next() {
		event := &AuditEvent{}
		var oldValuesJSON, newValuesJSON sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&event.UserID,
			&event.SessionID,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&oldValuesJSON,
			&newValuesJSON,
			&event.IPAddress,
			&event.UserAgent,
			&event.RequestID,
			&event.Success,
			&event.ErrorMessage,
			&event.Duration,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Parse JSON fields
		if oldValuesJSON.Valid && oldValuesJSON.String != "" {
			if err := json.Unmarshal([]byte(oldValuesJSON.String), &event.OldValues); err != nil {
				s.logger.Printf("Failed to parse old_values JSON: %v, event_id: %s", err, event.ID)
			}
		}

		if newValuesJSON.Valid && newValuesJSON.String != "" {
			if err := json.Unmarshal([]byte(newValuesJSON.String), &event.NewValues); err != nil {
				s.logger.Printf("Failed to parse new_values JSON: %v, event_id: %s", err, event.ID)
			}
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating audit events: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	response := &domain.PaginatedResponse{
		Data:       events,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}

	return response, nil
}

// GetAuditStatistics retrieves audit statistics for a time period
func (s *AuditServiceImpl) GetAuditStatistics(ctx context.Context, startDate, endDate time.Time) (*AuditStatistics, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	stats := &AuditStatistics{
		Period: TimeRange{Start: startDate, End: endDate},
	}

	// Get basic statistics
	basicStatsQuery := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN success = true THEN 1 END) as successful_events,
			COUNT(CASE WHEN success = false THEN 1 END) as failed_events,
			COUNT(DISTINCT user_id) as unique_users
		FROM audit_events
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3`

	err := s.db.QueryRowContext(ctx, basicStatsQuery, tenantID, startDate, endDate).Scan(
		&stats.TotalEvents,
		&stats.SuccessfulEvents,
		&stats.FailedEvents,
		&stats.UniqueUsers,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic audit statistics: %w", err)
	}

	// Get top actions
	topActionsQuery := `
		SELECT action, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY action
		ORDER BY count DESC
		LIMIT 10`

	rows, err := s.db.QueryContext(ctx, topActionsQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top actions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var actionCount ActionCount
		err := rows.Scan(&actionCount.Action, &actionCount.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action count: %w", err)
		}
		stats.TopActions = append(stats.TopActions, actionCount)
	}

	// Get top resource types
	topResourceTypesQuery := `
		SELECT resource_type, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY resource_type
		ORDER BY count DESC
		LIMIT 10`

	rows, err = s.db.QueryContext(ctx, topResourceTypesQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top resource types: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var resourceTypeCount ResourceTypeCount
		err := rows.Scan(&resourceTypeCount.ResourceType, &resourceTypeCount.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource type count: %w", err)
		}
		stats.TopResourceTypes = append(stats.TopResourceTypes, resourceTypeCount)
	}

	// Get events by day
	eventsByDayQuery := `
		SELECT DATE(created_at) as event_date, COUNT(*) as count
		FROM audit_events
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY event_date`

	rows, err = s.db.QueryContext(ctx, eventsByDayQuery, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events by day: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dayEventCount DayEventCount
		err := rows.Scan(&dayEventCount.Date, &dayEventCount.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan day event count: %w", err)
		}
		stats.EventsByDay = append(stats.EventsByDay, dayEventCount)
	}

	// Get security events
	securityEvents, err := s.getSecurityEvents(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Printf("Failed to get security events: %v", err)
	} else {
		stats.SecurityEvents = securityEvents
	}

	return stats, nil
}

// ExportAuditLog exports audit logs to various formats
func (s *AuditServiceImpl) ExportAuditLogs(ctx context.Context, filter *AuditFilter, format string) ([]byte, error) {
	// Get all matching events (with a reasonable limit)
	if filter.PerPage > 10000 {
		filter.PerPage = 10000
	}

	auditLog, err := s.GetAuditLogs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log for export: %w", err)
	}

	// Cast the Data field to []*AuditEvent
	events, ok := auditLog.Data.([]*AuditEvent)
	if !ok {
		return nil, fmt.Errorf("invalid data type in audit log response")
	}

	switch strings.ToLower(format) {
	case "json":
		return s.exportToJSON(events)
	case "csv":
		return s.exportToCSV(events)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// CleanupOldAuditLogs removes old audit logs based on retention policy
func (s *AuditServiceImpl) CleanupOldAuditLogs(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		return fmt.Errorf("retention days must be greater than 0")
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM audit_events WHERE created_at < $1`
	result, err := s.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	s.logger.Printf("Audit log cleanup completed: cutoff_date=%s, rows_deleted=%d", cutoffDate, rowsAffected)

	return nil
}

// Private helper methods

func (s *AuditServiceImpl) insertAuditEvent(ctx context.Context, event *AuditEvent) error {
	// Convert JSON fields to strings
	oldValuesJSON, err := json.Marshal(event.OldValues)
	if err != nil {
		return fmt.Errorf("failed to marshal old_values: %w", err)
	}

	newValuesJSON, err := json.Marshal(event.NewValues)
	if err != nil {
		return fmt.Errorf("failed to marshal new_values: %w", err)
	}

	query := `
		INSERT INTO audit_events (
			id, tenant_id, user_id, session_id, action, resource_type, resource_id,
			old_values, new_values, ip_address, user_agent, request_id, success,
			error_message, duration_ms, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err = s.db.ExecContext(ctx, query,
		event.ID,
		event.TenantID,
		event.UserID,
		event.SessionID,
		event.Action,
		event.ResourceType,
		event.ResourceID,
		string(oldValuesJSON),
		string(newValuesJSON),
		event.IPAddress,
		event.UserAgent,
		event.RequestID,
		event.Success,
		event.ErrorMessage,
		event.Duration,
		event.CreatedAt,
	)

	return err
}

func (s *AuditServiceImpl) isSecurityEvent(action string) bool {
	securityActions := map[string]bool{
		"auth.login":         true,
		"auth.logout":        true,
		"auth.failed_login":  true,
		"auth.password_change": true,
		"user.create":        true,
		"user.delete":        true,
		"user.role_change":   true,
		"permission.grant":   true,
		"permission.revoke":  true,
		"data.export":        true,
		"data.delete":        true,
	}

	return securityActions[action]
}

func (s *AuditServiceImpl) handleSecurityEvent(ctx context.Context, event *AuditEvent) error {
	// Check for suspicious patterns
	if err := s.checkSuspiciousActivity(ctx, event); err != nil {
		s.logger.Printf("Suspicious activity detected: %v, event_id: %s", err, event.ID)
	}

	// Log security event separately if needed
	s.logger.Printf("Security event recorded: action=%s, user_id=%v, ip_address=%s, success=%t", 
		event.Action, event.UserID, event.IPAddress, event.Success)

	return nil
}

func (s *AuditServiceImpl) checkSuspiciousActivity(ctx context.Context, event *AuditEvent) error {
	// Check for multiple failed logins from same IP
	if event.Action == "auth.failed_login" && !event.Success {
		count, err := s.countRecentFailedLogins(ctx, event.IPAddress, 15*time.Minute)
		if err != nil {
			return err
		}

		if count >= 5 {
			s.logger.Printf("Multiple failed login attempts detected: ip_address=%s, count=%d", 
				event.IPAddress, count)
			
			// Could trigger additional security measures here
			return fmt.Errorf("multiple failed login attempts from IP: %s", event.IPAddress)
		}
	}

	// Check for unusual access patterns
	if event.UserID != nil {
		// Check for access from unusual locations
		if err := s.checkUnusualLocation(ctx, *event.UserID, event.IPAddress); err != nil {
			s.logger.Printf("Unusual location access detected: user_id=%s, ip_address=%s", 
				*event.UserID, event.IPAddress)
		}
	}

	return nil
}

func (s *AuditServiceImpl) countRecentFailedLogins(ctx context.Context, ipAddress string, duration time.Duration) (int, error) {
	since := time.Now().Add(-duration)
	
	query := `
		SELECT COUNT(*)
		FROM audit_events
		WHERE action = 'auth.failed_login' 
		AND success = false 
		AND ip_address = $1 
		AND created_at >= $2`

	var count int
	err := s.db.QueryRowContext(ctx, query, ipAddress, since).Scan(&count)
	return count, err
}

func (s *AuditServiceImpl) checkUnusualLocation(ctx context.Context, userID uuid.UUID, ipAddress string) error {
	// Simple implementation - check if this IP is significantly different from usual IPs
	// In production, you might use GeoIP databases
	
	query := `
		SELECT DISTINCT ip_address
		FROM audit_events
		WHERE user_id = $1 
		AND created_at >= $2
		AND success = true
		LIMIT 10`

	since := time.Now().AddDate(0, 0, -30) // Last 30 days
	rows, err := s.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return err
	}
	defer rows.Close()

	knownIPs := make([]string, 0)
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			continue
		}
		knownIPs = append(knownIPs, ip)
	}

	// Check if current IP is in known IPs or same subnet
	for _, knownIP := range knownIPs {
		if ipAddress == knownIP || s.sameSubnet(ipAddress, knownIP) {
			return nil // Not unusual
		}
	}

	if len(knownIPs) > 0 {
		return fmt.Errorf("access from unusual IP address")
	}

	return nil
}

func (s *AuditServiceImpl) sameSubnet(ip1, ip2 string) bool {
	// Simple subnet check - compare first 3 octets for IPv4
	addr1 := net.ParseIP(ip1)
	addr2 := net.ParseIP(ip2)
	
	if addr1 == nil || addr2 == nil {
		return false
	}

	// For IPv4, check if in same /24 subnet
	if addr1.To4() != nil && addr2.To4() != nil {
		return addr1.To4()[0] == addr2.To4()[0] &&
			   addr1.To4()[1] == addr2.To4()[1] &&
			   addr1.To4()[2] == addr2.To4()[2]
	}

	return false
}

func (s *AuditServiceImpl) getSecurityEvents(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]SecurityEvent, error) {
	query := `
		SELECT action, COUNT(*) as count, COUNT(CASE WHEN success = false THEN 1 END) as failed_count
		FROM audit_events
		WHERE tenant_id = $1 
		AND created_at BETWEEN $2 AND $3
		AND action IN ('auth.login', 'auth.failed_login', 'auth.logout', 'user.create', 'user.delete', 'permission.grant', 'permission.revoke')
		GROUP BY action
		ORDER BY count DESC`

	rows, err := s.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]SecurityEvent, 0)
	for rows.Next() {
		var event SecurityEvent
		err := rows.Scan(&event.EventType, &event.Count, &event.FailedCount)
		if err != nil {
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (s *AuditServiceImpl) exportToJSON(events []*AuditEvent) ([]byte, error) {
	return json.MarshalIndent(events, "", "  ")
}

func (s *AuditServiceImpl) exportToCSV(events []*AuditEvent) ([]byte, error) {
	var csv strings.Builder
	
	// Write header
	csv.WriteString("ID,Tenant ID,User ID,Action,Resource Type,Resource ID,IP Address,Success,Created At\n")
	
	// Write data
	for _, event := range events {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%t,%s\n",
			event.ID,
			event.TenantID,
			formatUUID(event.UserID),
			event.Action,
			event.ResourceType,
			formatUUID(event.ResourceID),
			event.IPAddress,
			event.Success,
			event.CreatedAt.Format(time.RFC3339),
		)
		csv.WriteString(line)
	}
	
	return []byte(csv.String()), nil
}

func formatUUID(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// Data structures for audit logging

type AuditLogResponse struct {
	Events     []*AuditEvent `json:"events"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
	TotalPages int           `json:"total_pages"`
}

type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

type ResourceTypeCount struct {
	ResourceType string `json:"resource_type"`
	Count        int64  `json:"count"`
}

type DayEventCount struct {
	Date  time.Time `json:"date"`
	Count int64     `json:"count"`
}

type UserActivitySummary struct {
	UserID      uuid.UUID `json:"user_id"`
	TotalEvents int64     `json:"total_events"`
	LastActive  time.Time `json:"last_active"`
}

type SecurityEvent struct {
	EventType   string `json:"event_type"`
	Count       int64  `json:"count"`
	FailedCount int64  `json:"failed_count"`
}

// Context helper functions - these would ideally be in a shared middleware/context package

func GetSessionIDFromContext(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func GetIPAddressFromContext(ctx context.Context) string {
	if ipAddress, ok := ctx.Value("ip_address").(string); ok {
		return ipAddress
	}
	return ""
}

func GetUserAgentFromContext(ctx context.Context) string {
	if userAgent, ok := ctx.Value("user_agent").(string); ok {
		return userAgent
	}
	return ""
}

func convertDurationToMs(duration *time.Duration) *int64 {
	if duration == nil {
		return nil
	}
	ms := int64(*duration / time.Millisecond)
	return &ms
}

// GetComplianceReport generates a compliance report for the specified period
func (s *AuditServiceImpl) GetComplianceReport(ctx context.Context, startDate, endDate time.Time) (*ComplianceReport, error) {
	// TODO: Implement compliance report generation
	report := &ComplianceReport{
		Period: TimeRange{
			Start: startDate,
			End:   endDate,
		},
		OverallScore:    95, // Placeholder
		Violations:      []ComplianceViolation{},
		Recommendations: []string{"Implement regular security training"},
		GeneratedAt:     time.Now(),
	}
	return report, nil
}

// GetUserActivity retrieves user activity logs with filtering and pagination
func (s *AuditServiceImpl) GetUserActivity(ctx context.Context, userID uuid.UUID, filter *ActivityFilter) (*domain.PaginatedResponse, error) {
	// TODO: Implement user activity retrieval
	// For now, return empty response
	return &domain.PaginatedResponse{
		Data:       []*AuditEvent{},
		Total:      0,
		Page:       1,
		PerPage:    filter.PerPage,
		TotalPages: 0,
	}, nil
}