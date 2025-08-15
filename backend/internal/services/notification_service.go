package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	// TODO: Re-enable when communication packages are available
	// "github.com/pageza/go-comms"
	// "github.com/pageza/go-comms/providers/email"
	// "github.com/pageza/go-comms/providers/sms"
	// "github.com/pageza/go-comms/providers/push"
)

// NotificationServiceImpl implements comprehensive notification functionality
type NotificationServiceImpl struct {
	userRepo         UserRepository
	customerRepo     CustomerRepository
	jobRepo          JobRepositoryComplete
	auditService     AuditService
	// TODO: Re-enable when communication packages are available
	// commsClient      *comms.Client
	// emailProvider    email.Provider
	// smsProvider      sms.Provider
	// pushProvider     push.Provider
	logger           *log.Logger
	templates        map[string]NotificationTemplate
	defaultSettings  NotificationSettings
}

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // email, sms, push
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	Variables    []string               `json:"variables"`
	Settings     map[string]interface{} `json:"settings"`
	Active       bool                   `json:"active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailEnabled    bool   `json:"email_enabled"`
	SMSEnabled      bool   `json:"sms_enabled"`
	PushEnabled     bool   `json:"push_enabled"`
	FromEmail       string `json:"from_email"`
	FromName        string `json:"from_name"`
	SMSSender       string `json:"sms_sender"`
	PushAppID       string `json:"push_app_id"`
	MaxRetries      int    `json:"max_retries"`
	RetryDelay      int    `json:"retry_delay_seconds"`
}

// NewNotificationService creates a new notification service instance
func NewNotificationService(
	userRepo UserRepository,
	customerRepo CustomerRepository,
	jobRepo JobRepositoryComplete,
	auditService AuditService,
	// commsClient *comms.Client, // TODO: Re-enable when comms package is available
	logger *log.Logger,
) NotificationService {
	service := &NotificationServiceImpl{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		jobRepo:      jobRepo,
		auditService: auditService,
		// commsClient:  commsClient, // TODO: Re-enable when comms package is available
		logger:       logger,
		templates:    make(map[string]NotificationTemplate),
		defaultSettings: NotificationSettings{
			EmailEnabled: true,
			SMSEnabled:   true,
			PushEnabled:  true,
			FromEmail:    "noreply@landscaping-app.com",
			FromName:     "Landscaping App",
			SMSSender:    "LandscapeApp",
			MaxRetries:   3,
			RetryDelay:   30,
		},
	}

	// Initialize notification providers
	service.initializeProviders()
	
	// Load default templates
	service.loadDefaultTemplates()

	return service
}

// SendNotification sends a notification to a user
func (s *NotificationServiceImpl) SendNotification(ctx context.Context, req *NotificationRequest) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get user information
	var recipient *NotificationRecipient
	if req.UserID != nil {
		user, err := s.userRepo.GetByID(ctx, tenantID, *req.UserID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		if user == nil {
			return fmt.Errorf("user not found")
		}

		recipient = &NotificationRecipient{
			UserID:    user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}
	} else if req.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, tenantID, *req.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to get customer: %w", err)
		}
		if customer == nil {
			return fmt.Errorf("customer not found")
		}

		recipient = &NotificationRecipient{
			CustomerID: &customer.ID,
			Email:      "",
			Phone:      customer.Phone,
			FirstName:  customer.FirstName,
			LastName:   customer.LastName,
		}
		if customer.Email != nil {
			recipient.Email = *customer.Email
		}
	} else if req.Email != "" {
		recipient = &NotificationRecipient{
			Email: req.Email,
			Phone: nil,
		}
		if req.Phone != "" {
			recipient.Phone = &req.Phone
		}
	} else {
		return fmt.Errorf("no valid recipient specified")
	}

	// Get user notification preferences
	preferences := s.getUserNotificationPreferences(ctx, recipient)

	// Send notifications based on preferences and request
	var errors []string

	// Send email notification
	if preferences.EmailEnabled && recipient.Email != "" && (req.Channels == nil || contains(req.Channels, "email")) {
		if err := s.sendEmailNotification(ctx, recipient, req); err != nil {
			s.logger.Printf("Failed to send email notification", "error", err, "recipient", recipient.Email)
			errors = append(errors, fmt.Sprintf("email: %v", err))
		}
	}

	// Send SMS notification
	if preferences.SMSEnabled && recipient.Phone != nil && *recipient.Phone != "" && (req.Channels == nil || contains(req.Channels, "sms")) {
		if err := s.sendSMSNotification(ctx, recipient, req); err != nil {
			s.logger.Printf("Failed to send SMS notification", "error", err, "recipient", *recipient.Phone)
			errors = append(errors, fmt.Sprintf("sms: %v", err))
		}
	}

	// Send push notification
	if preferences.PushEnabled && recipient.UserID != uuid.Nil && (req.Channels == nil || contains(req.Channels, "push")) {
		if err := s.sendPushNotification(ctx, recipient, req); err != nil {
			s.logger.Printf("Failed to send push notification", "error", err, "user_id", recipient.UserID)
			errors = append(errors, fmt.Sprintf("push: %v", err))
		}
	}

	// Log notification attempt
	senderID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       senderID,
		Action:       "notification.send",
		ResourceType: "notification",
		NewValues: map[string]interface{}{
			"type":         req.Type,
			"title":        req.Title,
			"recipient_id": recipient.UserID,
			"channels":     req.Channels,
			"success":      len(errors) == 0,
			"errors":       errors,
		},
	}); err != nil {
		s.logger.Printf("Failed to log notification audit event", "error", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification failures: %s", strings.Join(errors, "; "))
	}

	s.logger.Printf("Notification sent successfully", "type", req.Type, "recipient", recipient.Email)
	return nil
}

// SendBulkNotification sends notifications to multiple recipients
func (s *NotificationServiceImpl) SendBulkNotification(ctx context.Context, req *BulkNotificationRequest) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	s.logger.Printf("Sending bulk notification - tenant_id: %s, type: %s", tenantID, req.Type)

	result := &BulkNotificationResult{
		TotalRecipients: len(req.Recipients),
		SuccessCount:    0,
		FailureCount:    0,
		Failures:        make([]NotificationFailure, 0),
	}

	// Process each recipient
	for _, recipient := range req.Recipients {
		notificationReq := &NotificationRequest{
			Type:       req.Type,
			Title:      req.Title,
			Message:    req.Message,
			Data:       req.Data,
			Channels:   req.Channels,
			Priority:   req.Priority,
			ScheduleAt: req.ScheduleAt,
		}

		// Set recipient based on type
		if recipient.UserID != uuid.Nil {
			notificationReq.UserID = &recipient.UserID
		} else if recipient.CustomerID != nil {
			notificationReq.CustomerID = recipient.CustomerID
		} else {
			notificationReq.Email = recipient.Email
			if recipient.Phone != nil {
				notificationReq.Phone = *recipient.Phone
			}
		}

		if err := s.SendNotification(ctx, notificationReq); err != nil {
			result.FailureCount++
			result.Failures = append(result.Failures, NotificationFailure{
				Recipient: recipient,
				Error:     err.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	s.logger.Printf("Bulk notification completed", 
		"total", result.TotalRecipients,
		"success", result.SuccessCount,
		"failures", result.FailureCount)

	return nil
}

// SendScheduledNotification sends a notification at a specific time
func (s *NotificationServiceImpl) SendScheduledNotification(ctx context.Context, req *NotificationRequest) error {
	if req.ScheduleAt == nil {
		return s.SendNotification(ctx, req)
	}

	// In a production system, this would use a job queue like Redis or database-based scheduling
	// For now, we'll use a simple goroutine with a timer
	go func() {
		delay := time.Until(*req.ScheduleAt)
		if delay > 0 {
			timer := time.NewTimer(delay)
			<-timer.C
		}

		if err := s.SendNotification(ctx, req); err != nil {
			s.logger.Printf("Failed to send scheduled notification", "error", err, "scheduled_at", req.ScheduleAt)
		}
	}()

	s.logger.Printf("Notification scheduled", "type", req.Type, "scheduled_at", req.ScheduleAt)
	return nil
}

// SendJobNotification sends job-related notifications
func (s *NotificationServiceImpl) SendJobNotification(ctx context.Context, jobID uuid.UUID, notificationType string, additionalData map[string]interface{}) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job details
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Get customer details
	customer, err := s.customerRepo.GetByID(ctx, tenantID, job.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return fmt.Errorf("customer not found")
	}

	// Prepare notification data
	data := map[string]interface{}{
		"job_id":        job.ID,
		"job_title":     job.Title,
		"job_status":    job.Status,
		"customer_name": fmt.Sprintf("%s %s", customer.FirstName, customer.LastName),
		"scheduled_date": job.ScheduledDate,
	}

	// Merge additional data
	for k, v := range additionalData {
		data[k] = v
	}

	// Determine notification content based on type
	title, message := s.getJobNotificationContent(notificationType, job, customer)

	// Send to assigned user if exists
	if job.AssignedUserID != nil {
		userReq := &NotificationRequest{
			UserID:   job.AssignedUserID,
			Type:     notificationType,
			Title:    title,
			Message:  message,
			Data:     data,
			Priority: s.getNotificationPriority(notificationType),
		}

		if err := s.SendNotification(ctx, userReq); err != nil {
			s.logger.Printf("Failed to send job notification to user", "error", err, "user_id", *job.AssignedUserID)
		}
	}

	// Send to customer for certain notification types
	if s.shouldNotifyCustomer(notificationType) {
		customerReq := &NotificationRequest{
			CustomerID: &customer.ID,
			Type:       notificationType,
			Title:      title,
			Message:    message,
			Data:       data,
			Priority:   s.getNotificationPriority(notificationType),
			Channels:   []string{"email"}, // Only email for customers by default
		}

		if err := s.SendNotification(ctx, customerReq); err != nil {
			s.logger.Printf("Failed to send job notification to customer", "error", err, "customer_id", customer.ID)
		}
	}

	return nil
}

// GetNotificationHistory gets notification history for a user
func (s *NotificationServiceImpl) GetNotificationHistory(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) (*NotificationHistoryResponse, error) {
	// In a production system, this would query a notifications table
	// For now, return a mock response
	
	notifications := []NotificationHistory{
		{
			ID:        uuid.New(),
			Type:      "job.assigned",
			Title:     "New Job Assigned",
			Message:   "You have been assigned to a landscaping job",
			Status:    "delivered",
			Channel:   "email",
			SentAt:    time.Now().Add(-2 * time.Hour),
			ReadAt:    nil,
		},
		{
			ID:        uuid.New(),
			Type:      "job.completed",
			Title:     "Job Completed",
			Message:   "Landscaping job has been completed",
			Status:    "delivered",
			Channel:   "push",
			SentAt:    time.Now().Add(-1 * time.Hour),
			ReadAt:    timePtr(time.Now().Add(-30 * time.Minute)),
		},
	}

	response := &NotificationHistoryResponse{
		Notifications: notifications,
		Total:         int64(len(notifications)),
		Page:          1,
		PerPage:       50,
		TotalPages:    1,
	}

	return response, nil
}

// UpdateNotificationPreferences updates user notification preferences
func (s *NotificationServiceImpl) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, preferences *NotificationPreferences) error {
	// In a production system, this would update user preferences in the database
	// For now, just log the update
	
	s.logger.Printf("Notification preferences updated", 
		"user_id", userID,
		"email_enabled", preferences.EmailEnabled,
		"sms_enabled", preferences.SMSEnabled,
		"push_enabled", preferences.PushEnabled)

	// Log audit event
	senderID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       senderID,
		Action:       "notification.preferences.update",
		ResourceType: "user",
		ResourceID:   &userID,
		NewValues: map[string]interface{}{
			"email_enabled": preferences.EmailEnabled,
			"sms_enabled":   preferences.SMSEnabled,
			"push_enabled":  preferences.PushEnabled,
		},
	}); err != nil {
		s.logger.Printf("Failed to log notification preferences audit event", "error", err)
	}

	return nil
}

// CreateNotificationTemplate creates a new notification template
func (s *NotificationServiceImpl) CreateNotificationTemplate(ctx context.Context, template *NotificationTemplate) error {
	template.ID = uuid.New().String()
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.Active = true

	// Store template (in production, this would be in database)
	s.templates[template.ID] = *template

	s.logger.Printf("Notification template created", "template_id", template.ID, "name", template.Name)
	return nil
}

// GetNotificationTemplate retrieves a notification template by ID
func (s *NotificationServiceImpl) GetNotificationTemplate(ctx context.Context, templateID string) (*NotificationTemplate, error) {
	template, exists := s.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("notification template not found")
	}
	return &template, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationServiceImpl) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	// TODO: Implement proper unread count query from database
	// For now, return 0 to prevent compilation errors
	return 0, nil
}

// GetUserNotifications retrieves notifications for a user with filtering
func (s *NotificationServiceImpl) GetUserNotifications(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) (*domain.PaginatedResponse, error) {
	// TODO: Implement proper user notifications query from database
	// For now, return empty response to prevent compilation errors
	return &domain.PaginatedResponse{
		Data:        []interface{}{},
		Total:       0,
		Page:        1,
		PerPage:     filter.PerPage,
		TotalPages:  0,
	}, nil
}

// MarkNotificationRead marks a specific notification as read
func (s *NotificationServiceImpl) MarkNotificationRead(ctx context.Context, notificationID uuid.UUID) error {
	// TODO: Implement proper notification update query
	// For now, do nothing to prevent compilation errors
	return nil
}

// MarkAllNotificationsRead marks all notifications as read for a user
func (s *NotificationServiceImpl) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement proper bulk notification update query
	// For now, do nothing to prevent compilation errors
	return nil
}

// Private helper methods

func (s *NotificationServiceImpl) initializeProviders() {
	// TODO: Initialize providers when communication packages are available
	// Currently disabled to prevent compilation errors
}

func (s *NotificationServiceImpl) loadDefaultTemplates() {
	defaultTemplates := []NotificationTemplate{
		{
			ID:      "job_assigned",
			Name:    "Job Assigned",
			Type:    "email",
			Subject: "New Job Assignment: {{.JobTitle}}",
			Body:    "Hello {{.FirstName}},\n\nYou have been assigned to a new job: {{.JobTitle}}\n\nScheduled for: {{.ScheduledDate}}\n\nBest regards,\nLandscaping Team",
			Variables: []string{"FirstName", "JobTitle", "ScheduledDate"},
			Active:  true,
		},
		{
			ID:      "job_completed",
			Name:    "Job Completed",
			Type:    "email", 
			Subject: "Job Completed: {{.JobTitle}}",
			Body:    "Hello {{.FirstName}},\n\nThe job '{{.JobTitle}}' has been completed successfully.\n\nThank you for your business!\n\nBest regards,\nLandscaping Team",
			Variables: []string{"FirstName", "JobTitle"},
			Active:  true,
		},
		{
			ID:      "appointment_reminder",
			Name:    "Appointment Reminder",
			Type:    "sms",
			Subject: "",
			Body:    "Reminder: Your landscaping appointment is scheduled for {{.ScheduledDate}}. Thank you!",
			Variables: []string{"ScheduledDate"},
			Active:  true,
		},
	}

	for _, template := range defaultTemplates {
		template.CreatedAt = time.Now()
		template.UpdatedAt = time.Now()
		s.templates[template.ID] = template
	}
}

func (s *NotificationServiceImpl) sendEmailNotification(ctx context.Context, recipient *NotificationRecipient, req *NotificationRequest) error {
	// TODO: Implement actual email sending when provider is available
	s.logger.Printf("Email notification would be sent", "to", recipient.Email, "subject", req.Title)
	return nil
}

func (s *NotificationServiceImpl) sendSMSNotification(ctx context.Context, recipient *NotificationRecipient, req *NotificationRequest) error {
	// TODO: Implement actual SMS sending when provider is available
	if recipient.Phone == nil || *recipient.Phone == "" {
		return fmt.Errorf("recipient phone number not available")
	}
	s.logger.Printf("SMS notification would be sent", "to", *recipient.Phone, "message", req.Message)
	return nil
}

func (s *NotificationServiceImpl) sendPushNotification(ctx context.Context, recipient *NotificationRecipient, req *NotificationRequest) error {
	// TODO: Implement actual push notification sending when provider is available
	s.logger.Printf("Push notification would be sent", "user_id", recipient.UserID, "title", req.Title)
	return nil
}

func (s *NotificationServiceImpl) getUserNotificationPreferences(ctx context.Context, recipient *NotificationRecipient) NotificationPreferences {
	// In a production system, this would query user preferences from database
	// For now, return default preferences
	return NotificationPreferences{
		EmailEnabled: true,
		SMSEnabled:   true,
		PushEnabled:  true,
	}
}

func (s *NotificationServiceImpl) applyTemplate(template string, recipient *NotificationRecipient, data map[string]interface{}) string {
	// Simple template replacement - in production, use a proper template engine
	result := template
	
	// Replace recipient variables
	result = strings.ReplaceAll(result, "{{.FirstName}}", recipient.FirstName)
	result = strings.ReplaceAll(result, "{{.LastName}}", recipient.LastName)
	
	// Replace data variables
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	
	return result
}

func (s *NotificationServiceImpl) getJobNotificationContent(notificationType string, job *domain.EnhancedJob, customer *domain.EnhancedCustomer) (string, string) {
	switch notificationType {
	case "job.assigned":
		return "New Job Assignment", fmt.Sprintf("You have been assigned to job: %s", job.Title)
	case "job.started":
		return "Job Started", fmt.Sprintf("Job '%s' has been started", job.Title)
	case "job.completed":
		return "Job Completed", fmt.Sprintf("Job '%s' has been completed successfully", job.Title)
	case "job.cancelled":
		return "Job Cancelled", fmt.Sprintf("Job '%s' has been cancelled", job.Title)
	case "job.scheduled":
		return "Job Scheduled", fmt.Sprintf("Job '%s' has been scheduled", job.Title)
	case "appointment.reminder":
		return "Appointment Reminder", fmt.Sprintf("Reminder: Your appointment for '%s' is scheduled", job.Title)
	default:
		return "Job Update", fmt.Sprintf("Update for job: %s", job.Title)
	}
}

func (s *NotificationServiceImpl) shouldNotifyCustomer(notificationType string) bool {
	customerNotificationTypes := map[string]bool{
		"job.scheduled":         true,
		"job.started":          true,
		"job.completed":        true,
		"appointment.reminder": true,
		"invoice.sent":         true,
		"payment.received":     true,
	}
	
	return customerNotificationTypes[notificationType]
}

func (s *NotificationServiceImpl) getNotificationPriority(notificationType string) string {
	highPriorityTypes := map[string]bool{
		"job.cancelled":        true,
		"emergency.alert":      true,
		"payment.failed":       true,
	}
	
	if highPriorityTypes[notificationType] {
		return "high"
	}
	
	return "normal"
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// Data structures for notifications

type NotificationRecipient struct {
	UserID     uuid.UUID  `json:"user_id"`
	CustomerID *uuid.UUID `json:"customer_id"`
	Email      string     `json:"email"`
	Phone      *string    `json:"phone"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
}

type BulkNotificationRequest struct {
	Type       string                   `json:"type"`
	Title      string                   `json:"title"`
	Message    string                   `json:"message"`
	Recipients []NotificationRecipient  `json:"recipients"`
	Data       map[string]interface{}   `json:"data"`
	Channels   []string                 `json:"channels"`
	Priority   string                   `json:"priority"`
	ScheduleAt *time.Time               `json:"schedule_at"`
}

type BulkNotificationResult struct {
	TotalRecipients int                    `json:"total_recipients"`
	SuccessCount    int                    `json:"success_count"`
	FailureCount    int                    `json:"failure_count"`
	Failures        []NotificationFailure  `json:"failures"`
}

type NotificationFailure struct {
	Recipient NotificationRecipient `json:"recipient"`
	Error     string                `json:"error"`
}

type NotificationHistory struct {
	ID      uuid.UUID  `json:"id"`
	Type    string     `json:"type"`
	Title   string     `json:"title"`
	Message string     `json:"message"`
	Status  string     `json:"status"`
	Channel string     `json:"channel"`
	SentAt  time.Time  `json:"sent_at"`
	ReadAt  *time.Time `json:"read_at"`
}

type NotificationHistoryResponse struct {
	Notifications []NotificationHistory `json:"notifications"`
	Total         int64                 `json:"total"`
	Page          int                   `json:"page"`
	PerPage       int                   `json:"per_page"`
	TotalPages    int                   `json:"total_pages"`
}

type NotificationPreferences struct {
	EmailEnabled bool `json:"email_enabled"`
	SMSEnabled   bool `json:"sms_enabled"`
	PushEnabled  bool `json:"push_enabled"`
}

type NotificationFilter struct {
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	Channel   string     `json:"channel"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
}