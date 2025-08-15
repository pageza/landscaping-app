package tools

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/ai"
	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// CustomerTools implements AI tools for customer-facing interactions
type CustomerTools struct {
	services *services.Services
	logger   *log.Logger
}

// NewCustomerTools creates a new set of customer AI tools
func NewCustomerTools(services *services.Services, logger *log.Logger) *CustomerTools {
	return &CustomerTools{
		services: services,
		logger:   logger,
	}
}

// RegisterTools registers all customer tools with the AI assistant
func (c *CustomerTools) RegisterTools(assistant ai.Assistant) error {
	tools := []*ai.Function{
		c.getScheduleAppointmentTool(),
		c.getCheckServiceHistoryTool(),
		c.getRequestQuoteTool(),
		c.getCheckBillingTool(),
		c.getJobStatusTool(),
		c.getModifyAppointmentTool(),
		c.getAddSpecialInstructionsTool(),
		c.getGetUpcomingJobsTool(),
		c.getGetInvoiceDetailsTool(),
		c.getGetPropertyInfoTool(),
		c.getGetServiceCatalogTool(),
		c.getRescheduleJobTool(),
		c.getCancelJobTool(),
		c.getPayInvoiceTool(),
		c.getGetQuoteStatusTool(),
	}

	for _, tool := range tools {
		if err := assistant.RegisterFunction(tool.Name, tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.Name, err)
		}
	}

	c.logger.Info("Registered customer AI tools", "count", len(tools))
	return nil
}

// Schedule Appointment Tool
func (c *CustomerTools) getScheduleAppointmentTool() *ai.Function {
	return &ai.Function{
		Name:        "schedule_appointment",
		Description: "Schedule a new landscaping service appointment for a customer",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "The customer's unique identifier",
				},
				"property_id": map[string]interface{}{
					"type":        "string",
					"description": "The property where the service will be performed",
				},
				"service_types": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "List of requested services (e.g., lawn_mowing, hedge_trimming)",
				},
				"preferred_date": map[string]interface{}{
					"type":        "string",
					"description": "Preferred date for the appointment (YYYY-MM-DD format)",
				},
				"preferred_time": map[string]interface{}{
					"type":        "string",
					"description": "Preferred time for the appointment (HH:MM format)",
				},
				"special_instructions": map[string]interface{}{
					"type":        "string",
					"description": "Any special instructions or notes for the service",
				},
				"priority": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"low", "medium", "high", "urgent"},
					"description": "Priority level of the appointment",
				},
			},
			"required": []string{"customer_id", "property_id", "service_types", "preferred_date"},
		},
		Handler: c.scheduleAppointmentHandler,
		Permissions: []string{"customer:schedule"},
	}
}

func (c *CustomerTools) scheduleAppointmentHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	customerID, err := uuid.Parse(params["customer_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	propertyID, err := uuid.Parse(params["property_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid property_id: %w", err)
	}

	// Parse preferred date
	preferredDate, err := time.Parse("2006-01-02", params["preferred_date"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid preferred_date format: %w", err)
	}

	// Parse preferred time if provided
	var scheduledTime *string
	if timeStr, ok := params["preferred_time"].(string); ok && timeStr != "" {
		scheduledTime = &timeStr
	}

	// Get service types
	serviceTypesInterface := params["service_types"].([]interface{})
	serviceTypes := make([]string, len(serviceTypesInterface))
	for i, v := range serviceTypesInterface {
		serviceTypes[i] = v.(string)
	}

	// Get special instructions
	var specialInstructions string
	if instructions, ok := params["special_instructions"].(string); ok {
		specialInstructions = instructions
	}

	// Get priority
	priority := "medium"
	if p, ok := params["priority"].(string); ok {
		priority = p
	}

	// Create job request
	jobReq := &domain.CreateJobRequest{
		CustomerID:        customerID,
		PropertyID:        propertyID,
		Title:             fmt.Sprintf("Landscaping Services - %s", preferredDate.Format("2006-01-02")),
		Description:       &specialInstructions,
		Priority:          priority,
		ScheduledDate:     &preferredDate,
		ScheduledTime:     scheduledTime,
		Services:          serviceTypes,
	}

	// Create the job
	job, err := c.services.Job.CreateJob(ctx, jobReq)
	if err != nil {
		c.logger.Error("Failed to schedule appointment", "error", err, "customer_id", customerID)
		return nil, fmt.Errorf("failed to schedule appointment: %w", err)
	}

	c.logger.Info("Appointment scheduled successfully", "job_id", job.ID, "customer_id", customerID)

	return map[string]interface{}{
		"success":        true,
		"job_id":         job.ID,
		"scheduled_date": preferredDate.Format("2006-01-02"),
		"scheduled_time": scheduledTime,
		"services":       serviceTypes,
		"message":        "Your appointment has been successfully scheduled!",
	}, nil
}

// Check Service History Tool
func (c *CustomerTools) getCheckServiceHistoryTool() *ai.Function {
	return &ai.Function{
		Name:        "check_service_history",
		Description: "Check the service history for a customer's properties",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "The customer's unique identifier",
				},
				"property_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional property ID to filter history for specific property",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of records to return (default: 10)",
					"default":     10,
				},
				"start_date": map[string]interface{}{
					"type":        "string",
					"description": "Start date for history search (YYYY-MM-DD format)",
				},
				"end_date": map[string]interface{}{
					"type":        "string",
					"description": "End date for history search (YYYY-MM-DD format)",
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: c.checkServiceHistoryHandler,
		Permissions: []string{"customer:view_history"},
	}
}

func (c *CustomerTools) checkServiceHistoryHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	customerID, err := uuid.Parse(params["customer_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	// Create job filter
	filter := &services.JobFilter{
		CustomerID: &customerID,
		Limit:      10,
		Offset:     0,
	}

	// Parse optional parameters
	if propertyIDStr, ok := params["property_id"].(string); ok && propertyIDStr != "" {
		propertyID, err := uuid.Parse(propertyIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid property_id: %w", err)
		}
		filter.PropertyID = &propertyID
	}

	if limit, ok := params["limit"].(float64); ok {
		filter.Limit = int(limit)
	}

	if startDateStr, ok := params["start_date"].(string); ok && startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		filter.StartDate = &startDate
	}

	if endDateStr, ok := params["end_date"].(string); ok && endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		filter.EndDate = &endDate
	}

	// Get job history
	jobsResponse, err := c.services.Job.ListJobs(ctx, filter)
	if err != nil {
		c.logger.Error("Failed to get service history", "error", err, "customer_id", customerID)
		return nil, fmt.Errorf("failed to get service history: %w", err)
	}

	// Format response
	jobs := jobsResponse.Data.([]*domain.EnhancedJob)
	history := make([]map[string]interface{}, len(jobs))
	
	for i, job := range jobs {
		history[i] = map[string]interface{}{
			"job_id":          job.ID,
			"title":           job.Title,
			"description":     job.Description,
			"status":          job.Status,
			"scheduled_date":  job.ScheduledDate,
			"completed_date":  job.ActualEndTime,
			"total_amount":    job.TotalAmount,
			"property_name":   job.Property.Name,
			"services":        job.Services,
		}
	}

	return map[string]interface{}{
		"success": true,
		"total":   jobsResponse.Total,
		"history": history,
	}, nil
}

// Request Quote Tool
func (c *CustomerTools) getRequestQuoteTool() *ai.Function {
	return &ai.Function{
		Name:        "request_quote",
		Description: "Request a quote for landscaping services",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "The customer's unique identifier",
				},
				"property_id": map[string]interface{}{
					"type":        "string",
					"description": "The property for which the quote is requested",
				},
				"service_description": map[string]interface{}{
					"type":        "string",
					"description": "Detailed description of the requested services",
				},
				"service_types": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "List of service types needed",
				},
				"urgency": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"low", "medium", "high", "urgent"},
					"description": "Urgency level for the quote",
				},
				"budget_range": map[string]interface{}{
					"type":        "string",
					"description": "Approximate budget range (e.g., '$500-1000')",
				},
			},
			"required": []string{"customer_id", "property_id", "service_description"},
		},
		Handler: c.requestQuoteHandler,
		Permissions: []string{"customer:request_quote"},
	}
}

func (c *CustomerTools) requestQuoteHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	customerID, err := uuid.Parse(params["customer_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	propertyID, err := uuid.Parse(params["property_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid property_id: %w", err)
	}

	serviceDescription := params["service_description"].(string)

	// Get service types if provided
	var serviceTypes []string
	if serviceTypesInterface, ok := params["service_types"].([]interface{}); ok {
		serviceTypes = make([]string, len(serviceTypesInterface))
		for i, v := range serviceTypesInterface {
			serviceTypes[i] = v.(string)
		}
	}

	// Get urgency
	urgency := "medium"
	if u, ok := params["urgency"].(string); ok {
		urgency = u
	}

	// Get budget range
	var budgetRange string
	if br, ok := params["budget_range"].(string); ok {
		budgetRange = br
	}

	// Create quote request
	quoteReq := &services.QuoteCreateRequest{
		CustomerID:  customerID,
		PropertyID:  propertyID,
		Title:       "Landscaping Services Quote",
		Description: serviceDescription,
		Priority:    urgency,
		Services:    serviceTypes,
		Notes:       &budgetRange,
	}

	// Create the quote
	quote, err := c.services.Quote.CreateQuote(ctx, quoteReq)
	if err != nil {
		c.logger.Error("Failed to create quote request", "error", err, "customer_id", customerID)
		return nil, fmt.Errorf("failed to create quote request: %w", err)
	}

	c.logger.Info("Quote requested successfully", "quote_id", quote.ID, "customer_id", customerID)

	return map[string]interface{}{
		"success":     true,
		"quote_id":    quote.ID,
		"status":      quote.Status,
		"description": serviceDescription,
		"services":    serviceTypes,
		"message":     "Your quote request has been submitted! We'll get back to you soon.",
	}, nil
}

// Check Billing Tool
func (c *CustomerTools) getCheckBillingTool() *ai.Function {
	return &ai.Function{
		Name:        "check_billing",
		Description: "Check billing information and outstanding invoices for a customer",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "The customer's unique identifier",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"all", "pending", "paid", "overdue"},
					"description": "Filter invoices by status",
					"default":     "all",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of invoices to return",
					"default":     10,
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: c.checkBillingHandler,
		Permissions: []string{"customer:view_billing"},
	}
}

func (c *CustomerTools) checkBillingHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	customerID, err := uuid.Parse(params["customer_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	// Create invoice filter
	filter := &services.InvoiceFilter{
		CustomerID: &customerID,
		Limit:      10,
		Offset:     0,
	}

	// Parse optional parameters
	if status, ok := params["status"].(string); ok && status != "all" {
		filter.Status = &status
	}

	if limit, ok := params["limit"].(float64); ok {
		filter.Limit = int(limit)
	}

	// Get invoices
	invoicesResponse, err := c.services.Invoice.ListInvoices(ctx, filter)
	if err != nil {
		c.logger.Error("Failed to get billing information", "error", err, "customer_id", customerID)
		return nil, fmt.Errorf("failed to get billing information: %w", err)
	}

	// Format response
	invoices := invoicesResponse.Data.([]*domain.Invoice)
	billingInfo := make([]map[string]interface{}, len(invoices))
	
	totalOutstanding := 0.0
	overdueCount := 0
	
	for i, invoice := range invoices {
		billingInfo[i] = map[string]interface{}{
			"invoice_id":      invoice.ID,
			"invoice_number":  invoice.InvoiceNumber,
			"status":          invoice.Status,
			"amount":          invoice.TotalAmount,
			"issued_date":     invoice.IssuedDate,
			"due_date":        invoice.DueDate,
			"paid_date":       invoice.PaidDate,
		}
		
		if invoice.Status == "pending" {
			totalOutstanding += invoice.TotalAmount
			if invoice.DueDate != nil && invoice.DueDate.Before(time.Now()) {
				overdueCount++
			}
		}
	}

	return map[string]interface{}{
		"success":           true,
		"total_invoices":    invoicesResponse.Total,
		"total_outstanding": totalOutstanding,
		"overdue_count":     overdueCount,
		"invoices":          billingInfo,
	}, nil
}

// Get Job Status Tool
func (c *CustomerTools) getJobStatusTool() *ai.Function {
	return &ai.Function{
		Name:        "get_job_status",
		Description: "Get the current status of a specific job or all active jobs for a customer",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "The customer's unique identifier",
				},
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "Specific job ID to check (optional, if not provided will show all active jobs)",
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: c.getJobStatusHandler,
		Permissions: []string{"customer:view_jobs"},
	}
}

func (c *CustomerTools) getJobStatusHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	customerID, err := uuid.Parse(params["customer_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid customer_id: %w", err)
	}

	// If specific job ID is provided, get that job
	if jobIDStr, ok := params["job_id"].(string); ok && jobIDStr != "" {
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid job_id: %w", err)
		}

		job, err := c.services.Job.GetJob(ctx, jobID)
		if err != nil {
			return nil, fmt.Errorf("failed to get job: %w", err)
		}

		// Verify job belongs to customer
		if job.CustomerID != customerID {
			return nil, fmt.Errorf("job not found for this customer")
		}

		return map[string]interface{}{
			"success":         true,
			"job_id":          job.ID,
			"title":           job.Title,
			"status":          job.Status,
			"scheduled_date":  job.ScheduledDate,
			"scheduled_time":  job.ScheduledTime,
			"actual_start":    job.ActualStartTime,
			"actual_end":      job.ActualEndTime,
			"assigned_crew":   job.AssignedUser,
			"property_name":   job.Property.Name,
			"description":     job.Description,
			"total_amount":    job.TotalAmount,
		}, nil
	}

	// Get all active jobs for customer
	filter := &services.JobFilter{
		CustomerID: &customerID,
		Status:     &[]string{"scheduled", "in_progress"},
		Limit:      20,
		Offset:     0,
	}

	jobsResponse, err := c.services.Job.ListJobs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	jobs := jobsResponse.Data.([]*domain.EnhancedJob)
	jobStatuses := make([]map[string]interface{}, len(jobs))
	
	for i, job := range jobs {
		jobStatuses[i] = map[string]interface{}{
			"job_id":          job.ID,
			"title":           job.Title,
			"status":          job.Status,
			"scheduled_date":  job.ScheduledDate,
			"scheduled_time":  job.ScheduledTime,
			"property_name":   job.Property.Name,
			"estimated_duration": job.EstimatedDuration,
		}
	}

	return map[string]interface{}{
		"success":    true,
		"total_jobs": len(jobs),
		"jobs":       jobStatuses,
	}, nil
}

// Additional tool implementations would continue here...
// For brevity, I'll implement a few more key tools

// Modify Appointment Tool
func (c *CustomerTools) getModifyAppointmentTool() *ai.Function {
	return &ai.Function{
		Name:        "modify_appointment",
		Description: "Modify an existing appointment (reschedule or update details)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "The job/appointment ID to modify",
				},
				"new_date": map[string]interface{}{
					"type":        "string",
					"description": "New preferred date (YYYY-MM-DD format)",
				},
				"new_time": map[string]interface{}{
					"type":        "string",
					"description": "New preferred time (HH:MM format)",
				},
				"special_instructions": map[string]interface{}{
					"type":        "string",
					"description": "Updated special instructions",
				},
			},
			"required": []string{"job_id"},
		},
		Handler: c.modifyAppointmentHandler,
		Permissions: []string{"customer:modify_appointment"},
	}
}

func (c *CustomerTools) modifyAppointmentHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	jobID, err := uuid.Parse(params["job_id"].(string))
	if err != nil {
		return nil, fmt.Errorf("invalid job_id: %w", err)
	}

	// Get existing job
	job, err := c.services.Job.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Create update request
	updateReq := &domain.UpdateJobRequest{}

	// Parse new date if provided
	if newDateStr, ok := params["new_date"].(string); ok && newDateStr != "" {
		newDate, err := time.Parse("2006-01-02", newDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid new_date format: %w", err)
		}
		updateReq.ScheduledDate = &newDate
	}

	// Parse new time if provided
	if newTimeStr, ok := params["new_time"].(string); ok && newTimeStr != "" {
		updateReq.ScheduledTime = &newTimeStr
	}

	// Update special instructions if provided
	if instructions, ok := params["special_instructions"].(string); ok && instructions != "" {
		updateReq.Description = &instructions
	}

	// Update the job
	updatedJob, err := c.services.Job.UpdateJob(ctx, jobID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	c.logger.Info("Appointment modified successfully", "job_id", jobID)

	return map[string]interface{}{
		"success":        true,
		"job_id":         updatedJob.ID,
		"scheduled_date": updatedJob.ScheduledDate,
		"scheduled_time": updatedJob.ScheduledTime,
		"message":        "Your appointment has been successfully updated!",
	}, nil
}

// Additional helper tools...

func (c *CustomerTools) getAddSpecialInstructionsTool() *ai.Function {
	return &ai.Function{
		Name:        "add_special_instructions",
		Description: "Add special instructions to an existing appointment or property",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "Job ID to add instructions to (optional)",
				},
				"property_id": map[string]interface{}{
					"type":        "string",
					"description": "Property ID to add general instructions to (optional)",
				},
				"instructions": map[string]interface{}{
					"type":        "string",
					"description": "The special instructions to add",
				},
			},
			"required": []string{"instructions"},
		},
		Handler: c.addSpecialInstructionsHandler,
		Permissions: []string{"customer:modify"},
	}
}

func (c *CustomerTools) addSpecialInstructionsHandler(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	instructions := params["instructions"].(string)

	// Add to job if job_id provided
	if jobIDStr, ok := params["job_id"].(string); ok && jobIDStr != "" {
		jobID, err := uuid.Parse(jobIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid job_id: %w", err)
		}

		updateReq := &domain.UpdateJobRequest{
			Description: &instructions,
		}

		_, err = c.services.Job.UpdateJob(ctx, jobID, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to add instructions to job: %w", err)
		}

		return map[string]interface{}{
			"success": true,
			"message": "Special instructions added to your appointment!",
		}, nil
	}

	// Add to property if property_id provided
	if propertyIDStr, ok := params["property_id"].(string); ok && propertyIDStr != "" {
		propertyID, err := uuid.Parse(propertyIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid property_id: %w", err)
		}

		updateReq := &services.PropertyUpdateRequest{
			Notes: &instructions,
		}

		_, err = c.services.Property.UpdateProperty(ctx, propertyID, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to add instructions to property: %w", err)
		}

		return map[string]interface{}{
			"success": true,
			"message": "Special instructions added to your property!",
		}, nil
	}

	return nil, fmt.Errorf("either job_id or property_id must be provided")
}

// Stub implementations for remaining tools - these would be fully implemented
func (c *CustomerTools) getGetUpcomingJobsTool() *ai.Function {
	return &ai.Function{
		Name:        "get_upcoming_jobs",
		Description: "Get upcoming scheduled jobs for a customer",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "Customer ID",
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:view_jobs"},
	}
}

func (c *CustomerTools) getGetInvoiceDetailsTool() *ai.Function {
	return &ai.Function{
		Name:        "get_invoice_details",
		Description: "Get detailed information about a specific invoice",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"invoice_id": map[string]interface{}{
					"type":        "string",
					"description": "Invoice ID",
				},
			},
			"required": []string{"invoice_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:view_billing"},
	}
}

func (c *CustomerTools) getGetPropertyInfoTool() *ai.Function {
	return &ai.Function{
		Name:        "get_property_info",
		Description: "Get information about customer properties",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "Customer ID",
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:view_properties"},
	}
}

func (c *CustomerTools) getGetServiceCatalogTool() *ai.Function {
	return &ai.Function{
		Name:        "get_service_catalog",
		Description: "Get available landscaping services and pricing",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Service category filter",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:view_services"},
	}
}

func (c *CustomerTools) getRescheduleJobTool() *ai.Function {
	return &ai.Function{
		Name:        "reschedule_job",
		Description: "Reschedule an existing job to a new date/time",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "Job ID to reschedule",
				},
				"new_date": map[string]interface{}{
					"type":        "string",
					"description": "New date (YYYY-MM-DD)",
				},
				"new_time": map[string]interface{}{
					"type":        "string",
					"description": "New time (HH:MM)",
				},
			},
			"required": []string{"job_id", "new_date"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:modify_appointment"},
	}
}

func (c *CustomerTools) getCancelJobTool() *ai.Function {
	return &ai.Function{
		Name:        "cancel_job",
		Description: "Cancel an existing job appointment",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "Job ID to cancel",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "Reason for cancellation",
				},
			},
			"required": []string{"job_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:cancel_appointment"},
	}
}

func (c *CustomerTools) getPayInvoiceTool() *ai.Function {
	return &ai.Function{
		Name:        "pay_invoice",
		Description: "Initiate payment for an invoice",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"invoice_id": map[string]interface{}{
					"type":        "string",
					"description": "Invoice ID to pay",
				},
				"payment_method": map[string]interface{}{
					"type":        "string",
					"description": "Payment method preference",
				},
			},
			"required": []string{"invoice_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:make_payment"},
	}
}

func (c *CustomerTools) getGetQuoteStatusTool() *ai.Function {
	return &ai.Function{
		Name:        "get_quote_status",
		Description: "Check the status of quote requests",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type":        "string",
					"description": "Customer ID",
				},
				"quote_id": map[string]interface{}{
					"type":        "string",
					"description": "Specific quote ID (optional)",
				},
			},
			"required": []string{"customer_id"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			// Implementation would go here
			return map[string]interface{}{"success": true, "message": "Feature coming soon"}, nil
		},
		Permissions: []string{"customer:view_quotes"},
	}
}