package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// PropertyServiceImpl implements the PropertyService interface
type PropertyServiceImpl struct {
	propertyRepo   PropertyRepositoryExtended
	customerRepo   CustomerRepository
	jobRepo        JobRepositoryExtended
	quoteRepo      QuoteRepositoryExtended
	auditService   AuditService
	logger         *log.Logger
}

// PropertyRepositoryExtended defines the interface for property data access
type PropertyRepositoryExtended interface {
	// CRUD operations
	Create(ctx context.Context, property *domain.EnhancedProperty) error
	GetByID(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.EnhancedProperty, error)
	Update(ctx context.Context, property *domain.EnhancedProperty) error
	Delete(ctx context.Context, tenantID, propertyID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *PropertyFilter) ([]*domain.EnhancedProperty, int64, error)
	
	// Geographic operations
	GetNearby(ctx context.Context, tenantID uuid.UUID, lat, lng, radiusMiles float64) ([]*domain.EnhancedProperty, error)
	UpdateGeocoding(ctx context.Context, propertyID uuid.UUID, lat, lng float64) error
	GetPropertiesWithinBounds(ctx context.Context, tenantID uuid.UUID, northLat, southLat, eastLng, westLng float64) ([]*domain.EnhancedProperty, error)
	
	// Search operations
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter *PropertyFilter) ([]*domain.EnhancedProperty, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.EnhancedProperty, error)
	
	// Route optimization
	GetPropertiesForRouteOptimization(ctx context.Context, tenantID uuid.UUID, propertyIDs []uuid.UUID) ([]*PropertyRouteInfo, error)
	
	// Analytics
	GetPropertyValueAnalytics(ctx context.Context, tenantID uuid.UUID, city, state string) (*PropertyValueAnalytics, error)
	
	// Geocoding
	GetPropertiesNeedingGeocoding(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.EnhancedProperty, error)
	
	// Validation
	CheckAddressExists(ctx context.Context, tenantID uuid.UUID, address string, excludeID *uuid.UUID) (bool, error)
}

// NewPropertyService creates a new property service instance
func NewPropertyService(
	propertyRepo PropertyRepositoryExtended,
	customerRepo CustomerRepository,
	jobRepo JobRepository,
	quoteRepo QuoteRepository,
	auditService AuditService,
	logger *log.Logger,
) PropertyService {
	return &PropertyServiceImpl{
		propertyRepo: propertyRepo,
		customerRepo: customerRepo,
		jobRepo:      jobRepo.(JobRepositoryExtended),
		quoteRepo:    quoteRepo.(QuoteRepositoryExtended),
		auditService: auditService,
		logger:       logger,
	}
}

// CreateProperty creates a new property
func (s *PropertyServiceImpl) CreateProperty(ctx context.Context, req *domain.CreatePropertyRequest) (*domain.EnhancedProperty, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreatePropertyRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Create full address for validation
	fullAddress := s.buildFullAddress(req.AddressLine1, req.AddressLine2, req.City, req.State, req.ZipCode)
	
	// Check for duplicate address
	exists, err := s.propertyRepo.CheckAddressExists(ctx, tenantID, fullAddress, nil)
	if err != nil {
		s.logger.Printf("Failed to check for duplicate address", "error", err)
	} else if exists {
		return nil, fmt.Errorf("property with this address already exists")
	}

	// Create property entity
	property := &domain.EnhancedProperty{
		Property: domain.Property{
			ID:           uuid.New(),
			TenantID:     tenantID,
			CustomerID:   req.CustomerID,
			Name:         req.Name,
			AddressLine1: req.AddressLine1,
			AddressLine2: req.AddressLine2,
			City:         req.City,
			State:        req.State,
			ZipCode:      req.ZipCode,
			Country:      "US", // Default to US
			PropertyType: req.PropertyType,
			LotSize:      req.LotSize,
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		SquareFootage:       req.SquareFootage,
		AccessInstructions:  req.AccessInstructions,
		GateCode:           req.GateCode,
		SpecialInstructions: req.SpecialInstructions,
	}

	// Geocode the address
	lat, lng, err := s.geocodeAddress(fullAddress)
	if err != nil {
		s.logger.Printf("Failed to geocode address", "error", err, "address", fullAddress)
	} else {
		property.Latitude = &lat
		property.Longitude = &lng
	}

	// Save to database
	if err := s.propertyRepo.Create(ctx, property); err != nil {
		s.logger.Printf("Failed to create property", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create property: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "property.create",
		ResourceType: "property",
		ResourceID:   &property.ID,
		NewValues: map[string]interface{}{
			"name":          property.Name,
			"address":       fullAddress,
			"property_type": property.PropertyType,
			"customer_id":   property.CustomerID,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Property created successfully", "property_id", property.ID, "tenant_id", tenantID)
	return property, nil
}

// GetProperty retrieves a property by ID
func (s *PropertyServiceImpl) GetProperty(ctx context.Context, propertyID uuid.UUID) (*domain.EnhancedProperty, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		s.logger.Printf("Failed to get property", "error", err, "property_id", propertyID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	return property, nil
}

// UpdateProperty updates an existing property
func (s *PropertyServiceImpl) UpdateProperty(ctx context.Context, propertyID uuid.UUID, req *PropertyUpdateRequest) (*domain.EnhancedProperty, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing property
	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	// Store old values for audit
	oldAddress := s.buildFullAddress(property.AddressLine1, property.AddressLine2, property.City, property.State, property.ZipCode)
	oldValues := map[string]interface{}{
		"name":    property.Name,
		"address": oldAddress,
	}

	// Track if address changed for geocoding
	addressChanged := false

	// Update fields
	if req.Name != nil {
		property.Name = *req.Name
	}
	if req.AddressLine1 != nil {
		if *req.AddressLine1 != property.AddressLine1 {
			addressChanged = true
		}
		property.AddressLine1 = *req.AddressLine1
	}
	if req.AddressLine2 != nil {
		if (req.AddressLine2 == nil && property.AddressLine2 != nil) ||
			(req.AddressLine2 != nil && property.AddressLine2 == nil) ||
			(req.AddressLine2 != nil && property.AddressLine2 != nil && *req.AddressLine2 != *property.AddressLine2) {
			addressChanged = true
		}
		property.AddressLine2 = req.AddressLine2
	}
	if req.City != nil {
		if *req.City != property.City {
			addressChanged = true
		}
		property.City = *req.City
	}
	if req.State != nil {
		if *req.State != property.State {
			addressChanged = true
		}
		property.State = *req.State
	}
	if req.ZipCode != nil {
		if *req.ZipCode != property.ZipCode {
			addressChanged = true
		}
		property.ZipCode = *req.ZipCode
	}
	if req.PropertyType != nil {
		property.PropertyType = *req.PropertyType
	}
	if req.LotSize != nil {
		property.LotSize = req.LotSize
	}
	if req.SquareFootage != nil {
		property.SquareFootage = req.SquareFootage
	}
	if req.AccessInstructions != nil {
		property.AccessInstructions = req.AccessInstructions
	}
	if req.GateCode != nil {
		property.GateCode = req.GateCode
	}
	if req.SpecialInstructions != nil {
		property.SpecialInstructions = req.SpecialInstructions
	}
	if req.PropertyValue != nil {
		property.PropertyValue = req.PropertyValue
	}
	if req.Notes != nil {
		property.Notes = req.Notes
	}

	property.UpdatedAt = time.Now()

	// If address changed, check for duplicates and re-geocode
	if addressChanged {
		newAddress := s.buildFullAddress(property.AddressLine1, property.AddressLine2, property.City, property.State, property.ZipCode)
		
		// Check for duplicate address
		exists, err := s.propertyRepo.CheckAddressExists(ctx, tenantID, newAddress, &propertyID)
		if err != nil {
			s.logger.Printf("Failed to check for duplicate address", "error", err)
		} else if exists {
			return nil, fmt.Errorf("property with this address already exists")
		}

		// Re-geocode the address
		lat, lng, err := s.geocodeAddress(newAddress)
		if err != nil {
			s.logger.Printf("Failed to geocode updated address", "error", err, "address", newAddress)
		} else {
			property.Latitude = &lat
			property.Longitude = &lng
		}
	}

	// Validate the updated property
	if err := s.validateProperty(property); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.propertyRepo.Update(ctx, property); err != nil {
		s.logger.Printf("Failed to update property", "error", err, "property_id", propertyID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update property: %w", err)
	}

	// Log audit event
	newAddress := s.buildFullAddress(property.AddressLine1, property.AddressLine2, property.City, property.State, property.ZipCode)
	newValues := map[string]interface{}{
		"name":    property.Name,
		"address": newAddress,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "property.update",
		ResourceType: "property",
		ResourceID:   &property.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Property updated successfully", "property_id", propertyID, "tenant_id", tenantID)
	return property, nil
}

// DeleteProperty deletes a property
func (s *PropertyServiceImpl) DeleteProperty(ctx context.Context, propertyID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get property before deletion for audit log
	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		return fmt.Errorf("failed to get property: %w", err)
	}
	if property == nil {
		return fmt.Errorf("property not found")
	}

	// Check if property has active jobs
	jobs, _, err := s.jobRepo.GetByPropertyID(ctx, tenantID, propertyID, &JobFilter{
		Status: "pending,scheduled,in_progress",
	})
	if err != nil {
		s.logger.Printf("Failed to check for active jobs", "error", err)
	} else if len(jobs) > 0 {
		return fmt.Errorf("cannot delete property with active jobs")
	}

	// Soft delete by updating status
	property.Status = "deleted"
	property.UpdatedAt = time.Now()

	if err := s.propertyRepo.Update(ctx, property); err != nil {
		s.logger.Printf("Failed to delete property", "error", err, "property_id", propertyID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete property: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "property.delete",
		ResourceType: "property",
		ResourceID:   &property.ID,
		OldValues: map[string]interface{}{
			"status": "active",
		},
		NewValues: map[string]interface{}{
			"status": "deleted",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Property deleted successfully", "property_id", propertyID, "tenant_id", tenantID)
	return nil
}

// ListProperties lists properties with filtering and pagination
func (s *PropertyServiceImpl) ListProperties(ctx context.Context, filter *PropertyFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &PropertyFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	properties, total, err := s.propertyRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list properties", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list properties: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       properties,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetNearbyProperties finds properties within a radius
func (s *PropertyServiceImpl) GetNearbyProperties(ctx context.Context, lat, lng float64, radiusMiles float64) ([]*domain.EnhancedProperty, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	if radiusMiles <= 0 || radiusMiles > 100 {
		return nil, fmt.Errorf("radius must be between 0 and 100 miles")
	}

	properties, err := s.propertyRepo.GetNearby(ctx, tenantID, lat, lng, radiusMiles)
	if err != nil {
		s.logger.Printf("Failed to get nearby properties", "error", err, "lat", lat, "lng", lng, "radius", radiusMiles)
		return nil, fmt.Errorf("failed to get nearby properties: %w", err)
	}

	return properties, nil
}

// SearchProperties searches properties by query string
func (s *PropertyServiceImpl) SearchProperties(ctx context.Context, query string, filter *PropertyFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &PropertyFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	properties, total, err := s.propertyRepo.Search(ctx, tenantID, query, filter)
	if err != nil {
		s.logger.Printf("Failed to search properties", "error", err, "query", query, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to search properties: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       properties,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetPropertyJobs retrieves jobs for a property
func (s *PropertyServiceImpl) GetPropertyJobs(ctx context.Context, propertyID uuid.UUID, filter *JobFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify property exists
	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	// Set defaults
	if filter == nil {
		filter = &JobFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}

	jobs, total, err := s.jobRepo.GetByPropertyID(ctx, tenantID, propertyID, filter)
	if err != nil {
		s.logger.Printf("Failed to get property jobs", "error", err, "property_id", propertyID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get property jobs: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       jobs,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetPropertyQuotes retrieves quotes for a property
func (s *PropertyServiceImpl) GetPropertyQuotes(ctx context.Context, propertyID uuid.UUID, filter *QuoteFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify property exists
	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	// Set defaults
	if filter == nil {
		filter = &QuoteFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}

	quotes, total, err := s.quoteRepo.GetByPropertyID(ctx, tenantID, propertyID, filter)
	if err != nil {
		s.logger.Printf("Failed to get property quotes", "error", err, "property_id", propertyID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get property quotes: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       quotes,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetPropertyValue estimates property value
func (s *PropertyServiceImpl) GetPropertyValue(ctx context.Context, propertyID uuid.UUID) (*PropertyValuation, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify property exists
	property, err := s.propertyRepo.GetByID(ctx, tenantID, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	// If we have a stored property value, return it
	if property.PropertyValue != nil {
		return &PropertyValuation{
			EstimatedValue: *property.PropertyValue,
			LastUpdated:    property.UpdatedAt,
			Confidence:     0.8, // Medium confidence for manually entered values
			Source:         "manual",
		}, nil
	}

	// Otherwise, estimate based on lot size and property type
	estimated := s.estimatePropertyValue(property)
	
	return &PropertyValuation{
		EstimatedValue: estimated,
		LastUpdated:    time.Now(),
		Confidence:     0.5, // Lower confidence for estimates
		Source:         "estimated",
	}, nil
}

// Helper methods

func (s *PropertyServiceImpl) validateCreatePropertyRequest(req *domain.CreatePropertyRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("property name is required")
	}
	if strings.TrimSpace(req.AddressLine1) == "" {
		return fmt.Errorf("address line 1 is required")
	}
	if strings.TrimSpace(req.City) == "" {
		return fmt.Errorf("city is required")
	}
	if strings.TrimSpace(req.State) == "" {
		return fmt.Errorf("state is required")
	}
	if strings.TrimSpace(req.ZipCode) == "" {
		return fmt.Errorf("zip code is required")
	}
	if req.PropertyType != "residential" && req.PropertyType != "commercial" {
		return fmt.Errorf("property type must be 'residential' or 'commercial'")
	}
	if req.LotSize != nil && *req.LotSize <= 0 {
		return fmt.Errorf("lot size must be greater than 0")
	}
	if req.SquareFootage != nil && *req.SquareFootage <= 0 {
		return fmt.Errorf("square footage must be greater than 0")
	}

	return nil
}

func (s *PropertyServiceImpl) validateProperty(property *domain.EnhancedProperty) error {
	if strings.TrimSpace(property.Name) == "" {
		return fmt.Errorf("property name is required")
	}
	if strings.TrimSpace(property.AddressLine1) == "" {
		return fmt.Errorf("address line 1 is required")
	}
	if strings.TrimSpace(property.City) == "" {
		return fmt.Errorf("city is required")
	}
	if strings.TrimSpace(property.State) == "" {
		return fmt.Errorf("state is required")
	}
	if strings.TrimSpace(property.ZipCode) == "" {
		return fmt.Errorf("zip code is required")
	}
	return nil
}

func (s *PropertyServiceImpl) buildFullAddress(line1 string, line2 *string, city, state, zipCode string) string {
	address := line1
	if line2 != nil && *line2 != "" {
		address += ", " + *line2
	}
	address += fmt.Sprintf(", %s, %s %s", city, state, zipCode)
	return address
}

func (s *PropertyServiceImpl) geocodeAddress(address string) (float64, float64, error) {
	// This is a placeholder for actual geocoding service integration
	// In a real implementation, you would integrate with Google Maps, MapBox, or similar service
	
	// For now, return a mock coordinate based on address hash
	// This ensures consistent coordinates for the same address
	hash := s.simpleHash(address)
	lat := 40.7128 + (float64(hash%1000)-500)/10000.0  // Around New York area
	lng := -74.0060 + (float64(hash%1000)-500)/10000.0
	
	s.logger.Printf("Mock geocoding - address: %s, lat: %f, lng: %f", address, lat, lng)
	return lat, lng, nil
}

func (s *PropertyServiceImpl) simpleHash(str string) int {
	hash := 0
	for _, c := range str {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func (s *PropertyServiceImpl) estimatePropertyValue(property *domain.EnhancedProperty) float64 {
	// Simple property value estimation based on type and size
	baseValue := 200000.0 // Base value
	
	if property.PropertyType == "commercial" {
		baseValue = 500000.0
	}
	
	// Adjust based on lot size
	if property.LotSize != nil {
		baseValue += *property.LotSize * 1000.0 // $1000 per unit of lot size
	}
	
	// Adjust based on square footage
	if property.SquareFootage != nil {
		baseValue += float64(*property.SquareFootage) * 100.0 // $100 per square foot
	}
	
	return baseValue
}

// Additional interfaces needed by property service
type JobRepositoryExtended interface {
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
}

type QuoteRepositoryExtended interface {
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
}

// GetPropertiesWithinBounds gets properties within geographic bounds
func (s *PropertyServiceImpl) GetPropertiesWithinBounds(ctx context.Context, northLat, southLat, eastLng, westLng float64) ([]*domain.EnhancedProperty, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate bounds
	if northLat <= southLat || eastLng <= westLng {
		return nil, fmt.Errorf("invalid geographic bounds")
	}

	properties, err := s.propertyRepo.GetPropertiesWithinBounds(ctx, tenantID, northLat, southLat, eastLng, westLng)
	if err != nil {
		s.logger.Printf("Failed to get properties within bounds", "error", err, "bounds", map[string]float64{
			"north": northLat, "south": southLat, "east": eastLng, "west": westLng,
		})
		return nil, fmt.Errorf("failed to get properties within bounds: %w", err)
	}

	return properties, nil
}

// GetPropertiesForRouteOptimization gets property route information for optimization
func (s *PropertyServiceImpl) GetPropertiesForRouteOptimization(ctx context.Context, propertyIDs []uuid.UUID) ([]*PropertyRouteInfo, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	if len(propertyIDs) == 0 {
		return []*PropertyRouteInfo{}, nil
	}

	routeInfos, err := s.propertyRepo.GetPropertiesForRouteOptimization(ctx, tenantID, propertyIDs)
	if err != nil {
		s.logger.Printf("Failed to get properties for route optimization", "error", err, "property_count", len(propertyIDs))
		return nil, fmt.Errorf("failed to get properties for route optimization: %w", err)
	}

	return routeInfos, nil
}

// OptimizePropertyRoute optimizes the order of visiting properties
func (s *PropertyServiceImpl) OptimizePropertyRoute(ctx context.Context, propertyIDs []uuid.UUID, startLocation *Location) (*RouteOptimization, error) {
	if len(propertyIDs) <= 1 {
		return &RouteOptimization{
			OptimizedRoute: []RouteStop{},
			TotalDistance:  0,
			TotalDuration:  0,
			Savings:        0,
		}, nil
	}

	// Get property route information
	routeInfos, err := s.GetPropertiesForRouteOptimization(ctx, propertyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get property route info: %w", err)
	}

	// Filter out properties without coordinates
	validProperties := make([]*PropertyRouteInfo, 0)
	for _, info := range routeInfos {
		if info.Latitude != nil && info.Longitude != nil {
			validProperties = append(validProperties, info)
		}
	}

	if len(validProperties) <= 1 {
		return &RouteOptimization{
			OptimizedRoute: []RouteStop{},
			TotalDistance:  0,
			TotalDuration:  0,
			Savings:        0,
		}, nil
	}

	// Use nearest neighbor algorithm for route optimization
	optimizedRoute, totalDistance := s.optimizeRouteNearestNeighbor(validProperties, startLocation)

	// Calculate estimated duration (assuming 30 mph average speed + 30 min per stop)
	estimatedDriveTime := int(totalDistance / 30.0 * 60) // minutes
	estimatedServiceTime := len(optimizedRoute) * 30     // 30 minutes per property
	totalDuration := estimatedDriveTime + estimatedServiceTime

	// Calculate original distance for savings comparison
	originalDistance := s.calculateOriginalRouteDistance(validProperties, startLocation)
	savings := 0.0
	if originalDistance > 0 {
		savings = ((originalDistance - totalDistance) / originalDistance) * 100
	}

	return &RouteOptimization{
		OptimizedRoute: optimizedRoute,
		TotalDistance:  totalDistance,
		TotalDuration:  totalDuration,
		Savings:        savings,
	}, nil
}

// BatchGeocodeProperties geocodes multiple properties
func (s *PropertyServiceImpl) BatchGeocodeProperties(ctx context.Context, limit int) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	if limit <= 0 || limit > 100 {
		limit = 20 // Default batch size
	}

	properties, err := s.propertyRepo.GetPropertiesNeedingGeocoding(ctx, tenantID, limit)
	if err != nil {
		return fmt.Errorf("failed to get properties needing geocoding: %w", err)
	}

	if len(properties) == 0 {
		s.logger.Printf("No properties need geocoding", "tenant_id", tenantID)
		return nil
	}

	successCount := 0
	errorCount := 0

	for _, property := range properties {
		fullAddress := s.buildFullAddress(property.AddressLine1, property.AddressLine2, property.City, property.State, property.ZipCode)
		
		lat, lng, err := s.geocodeAddress(fullAddress)
		if err != nil {
			s.logger.Printf("Failed to geocode property", "error", err, "property_id", property.ID, "address", fullAddress)
			errorCount++
			continue
		}

		err = s.propertyRepo.UpdateGeocoding(ctx, property.ID, lat, lng)
		if err != nil {
			s.logger.Printf("Failed to update property coordinates", "error", err, "property_id", property.ID)
			errorCount++
			continue
		}

		successCount++
	}

	s.logger.Printf("Batch geocoding completed", 
		"tenant_id", tenantID, 
		"success_count", successCount, 
		"error_count", errorCount,
		"total_processed", len(properties))

	return nil
}

// GetPropertyValueAnalytics gets property value analytics for an area
func (s *PropertyServiceImpl) GetPropertyValueAnalytics(ctx context.Context, city, state string) (*PropertyValueAnalytics, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	analytics, err := s.propertyRepo.GetPropertyValueAnalytics(ctx, tenantID, city, state)
	if err != nil {
		s.logger.Printf("Failed to get property value analytics", "error", err, "city", city, "state", state)
		return nil, fmt.Errorf("failed to get property value analytics: %w", err)
	}

	return analytics, nil
}

// Helper methods for route optimization

func (s *PropertyServiceImpl) optimizeRouteNearestNeighbor(properties []*PropertyRouteInfo, startLocation *Location) ([]RouteStop, float64) {
	if len(properties) == 0 {
		return []RouteStop{}, 0
	}

	visited := make(map[int]bool)
	route := make([]RouteStop, 0, len(properties))
	totalDistance := 0.0

	// Starting location
	currentLat := startLocation.Latitude
	currentLng := startLocation.Longitude
	currentTime := time.Now().Add(time.Hour) // Assume route starts in 1 hour

	for len(route) < len(properties) {
		nearestIndex := -1
		nearestDistance := math.MaxFloat64

		// Find nearest unvisited property
		for i, property := range properties {
			if visited[i] || property.Latitude == nil || property.Longitude == nil {
				continue
			}

			distance := haversineDistance(currentLat, currentLng, *property.Latitude, *property.Longitude)
			if distance < nearestDistance {
				nearestDistance = distance
				nearestIndex = i
			}
		}

		if nearestIndex == -1 {
			break
		}

		// Add to route
		property := properties[nearestIndex]
		visited[nearestIndex] = true

		// Estimate travel time (30 mph average)
		travelTimeMinutes := int(nearestDistance / 30.0 * 60)
		arrivalTime := currentTime.Add(time.Duration(travelTimeMinutes) * time.Minute)

		route = append(route, RouteStop{
			JobID:       property.ID, // Using property ID as job ID for now
			Address:     s.buildPropertyAddress(property),
			Sequence:    len(route) + 1,
			ArrivalTime: arrivalTime,
			Duration:    30, // Assume 30 minutes per property
			Distance:    nearestDistance,
		})

		totalDistance += nearestDistance
		currentLat = *property.Latitude
		currentLng = *property.Longitude
		currentTime = arrivalTime.Add(30 * time.Minute) // Add service time
	}

	return route, totalDistance
}

func (s *PropertyServiceImpl) calculateOriginalRouteDistance(properties []*PropertyRouteInfo, startLocation *Location) float64 {
	if len(properties) <= 1 {
		return 0
	}

	totalDistance := 0.0
	currentLat := startLocation.Latitude
	currentLng := startLocation.Longitude

	// Calculate distance if visiting properties in original order
	for _, property := range properties {
		if property.Latitude != nil && property.Longitude != nil {
			distance := haversineDistance(currentLat, currentLng, *property.Latitude, *property.Longitude)
			totalDistance += distance
			currentLat = *property.Latitude
			currentLng = *property.Longitude
		}
	}

	return totalDistance
}

func (s *PropertyServiceImpl) buildPropertyAddress(property *PropertyRouteInfo) string {
	return fmt.Sprintf("%s, %s, %s %s", property.AddressLine1, property.City, property.State, property.ZipCode)
}

// Haversine formula to calculate distance between two points
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 3959 // miles

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

