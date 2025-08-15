package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// PropertyRepositoryImpl implements the PropertyRepositoryExtended interface
type PropertyRepositoryImpl struct {
	db *Database
}

// NewPropertyRepositoryImpl creates a new property repository instance
func NewPropertyRepositoryImpl(db *Database) services.PropertyRepositoryExtended {
	return &PropertyRepositoryImpl{db: db}
}

// Create creates a new property
func (r *PropertyRepositoryImpl) Create(ctx context.Context, property *domain.EnhancedProperty) error {
	query := `
		INSERT INTO properties (
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23
		)`

	_, err := r.db.ExecContext(ctx, query,
		property.ID,
		property.TenantID,
		property.CustomerID,
		property.Name,
		property.AddressLine1,
		property.AddressLine2,
		property.City,
		property.State,
		property.ZipCode,
		property.Country,
		property.PropertyType,
		property.LotSize,
		property.SquareFootage,
		property.Latitude,
		property.Longitude,
		property.AccessInstructions,
		property.GateCode,
		property.SpecialInstructions,
		property.PropertyValue,
		property.Notes,
		property.Status,
		property.CreatedAt,
		property.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create property: %w", err)
	}

	return nil
}

// GetByID retrieves a property by ID
func (r *PropertyRepositoryImpl) GetByID(ctx context.Context, tenantID, propertyID uuid.UUID) (*domain.EnhancedProperty, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		FROM properties
		WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`

	row := r.db.QueryRowContext(ctx, query, propertyID, tenantID)

	property := &domain.EnhancedProperty{}
	err := row.Scan(
		&property.ID,
		&property.TenantID,
		&property.CustomerID,
		&property.Name,
		&property.AddressLine1,
		&property.AddressLine2,
		&property.City,
		&property.State,
		&property.ZipCode,
		&property.Country,
		&property.PropertyType,
		&property.LotSize,
		&property.SquareFootage,
		&property.Latitude,
		&property.Longitude,
		&property.AccessInstructions,
		&property.GateCode,
		&property.SpecialInstructions,
		&property.PropertyValue,
		&property.Notes,
		&property.Status,
		&property.CreatedAt,
		&property.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	return property, nil
}

// Update updates an existing property
func (r *PropertyRepositoryImpl) Update(ctx context.Context, property *domain.EnhancedProperty) error {
	query := `
		UPDATE properties SET
			customer_id = $3,
			name = $4,
			address_line1 = $5,
			address_line2 = $6,
			city = $7,
			state = $8,
			zip_code = $9,
			country = $10,
			property_type = $11,
			lot_size = $12,
			square_footage = $13,
			latitude = $14,
			longitude = $15,
			access_instructions = $16,
			gate_code = $17,
			special_instructions = $18,
			property_value = $19,
			notes = $20,
			status = $21,
			updated_at = $22
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		property.ID,
		property.TenantID,
		property.CustomerID,
		property.Name,
		property.AddressLine1,
		property.AddressLine2,
		property.City,
		property.State,
		property.ZipCode,
		property.Country,
		property.PropertyType,
		property.LotSize,
		property.SquareFootage,
		property.Latitude,
		property.Longitude,
		property.AccessInstructions,
		property.GateCode,
		property.SpecialInstructions,
		property.PropertyValue,
		property.Notes,
		property.Status,
		property.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update property: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("property not found or not authorized")
	}

	return nil
}

// Delete deletes a property (soft delete)
func (r *PropertyRepositoryImpl) Delete(ctx context.Context, tenantID, propertyID uuid.UUID) error {
	query := `
		UPDATE properties 
		SET status = 'deleted', updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, propertyID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete property: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("property not found or not authorized")
	}

	return nil
}

// List retrieves properties with filtering and pagination
func (r *PropertyRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.PropertyFilter) ([]*domain.EnhancedProperty, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1 AND status != 'deleted'"
	args := []interface{}{tenantID}
	argIndex := 2

	if filter.PropertyType != "" {
		whereClause += fmt.Sprintf(" AND property_type = $%d", argIndex)
		args = append(args, filter.PropertyType)
		argIndex++
	}

	if filter.CustomerID != nil {
		whereClause += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.City != "" {
		whereClause += fmt.Sprintf(" AND city ILIKE $%d", argIndex)
		args = append(args, "%"+filter.City+"%")
		argIndex++
	}

	if filter.State != "" {
		whereClause += fmt.Sprintf(" AND state = $%d", argIndex)
		args = append(args, filter.State)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR address_line1 ILIKE $%d OR city ILIKE $%d)", 
			argIndex, argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM properties " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count properties: %w", err)
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
			"name":          true,
			"address_line1": true,
			"city":          true,
			"state":         true,
			"property_type": true,
			"created_at":    true,
			"updated_at":    true,
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
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		FROM properties ` + whereClause + paginationClause

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list properties: %w", err)
	}
	defer rows.Close()

	properties := make([]*domain.EnhancedProperty, 0)
	for rows.Next() {
		property := &domain.EnhancedProperty{}
		err := rows.Scan(
			&property.ID,
			&property.TenantID,
			&property.CustomerID,
			&property.Name,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.City,
			&property.State,
			&property.ZipCode,
			&property.Country,
			&property.PropertyType,
			&property.LotSize,
			&property.SquareFootage,
			&property.Latitude,
			&property.Longitude,
			&property.AccessInstructions,
			&property.GateCode,
			&property.SpecialInstructions,
			&property.PropertyValue,
			&property.Notes,
			&property.Status,
			&property.CreatedAt,
			&property.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan property: %w", err)
		}
		properties = append(properties, property)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating properties: %w", err)
	}

	return properties, total, nil
}

// GetNearby finds properties within a radius using PostGIS
func (r *PropertyRepositoryImpl) GetNearby(ctx context.Context, tenantID uuid.UUID, lat, lng, radiusMiles float64) ([]*domain.EnhancedProperty, error) {
	// Using Haversine formula for distance calculation
	// In a production system with PostGIS, you would use ST_DWithin for better performance
	query := `
		SELECT 
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at,
			(
				3959 * acos(
					cos(radians($2)) * cos(radians(latitude)) * 
					cos(radians(longitude) - radians($3)) + 
					sin(radians($2)) * sin(radians(latitude))
				)
			) AS distance
		FROM properties
		WHERE tenant_id = $1 
			AND status != 'deleted' 
			AND latitude IS NOT NULL 
			AND longitude IS NOT NULL
		HAVING distance <= $4
		ORDER BY distance`

	rows, err := r.db.QueryContext(ctx, query, tenantID, lat, lng, radiusMiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby properties: %w", err)
	}
	defer rows.Close()

	properties := make([]*domain.EnhancedProperty, 0)
	for rows.Next() {
		property := &domain.EnhancedProperty{}
		var distance float64
		err := rows.Scan(
			&property.ID,
			&property.TenantID,
			&property.CustomerID,
			&property.Name,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.City,
			&property.State,
			&property.ZipCode,
			&property.Country,
			&property.PropertyType,
			&property.LotSize,
			&property.SquareFootage,
			&property.Latitude,
			&property.Longitude,
			&property.AccessInstructions,
			&property.GateCode,
			&property.SpecialInstructions,
			&property.PropertyValue,
			&property.Notes,
			&property.Status,
			&property.CreatedAt,
			&property.UpdatedAt,
			&distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan nearby property: %w", err)
		}
		properties = append(properties, property)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating nearby properties: %w", err)
	}

	return properties, nil
}

// UpdateGeocoding updates the latitude and longitude for a property
func (r *PropertyRepositoryImpl) UpdateGeocoding(ctx context.Context, propertyID uuid.UUID, lat, lng float64) error {
	query := `
		UPDATE properties 
		SET latitude = $2, longitude = $3, updated_at = $4
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, propertyID, lat, lng, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update property geocoding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("property not found")
	}

	return nil
}

// Search searches properties by query string
func (r *PropertyRepositoryImpl) Search(ctx context.Context, tenantID uuid.UUID, query string, filter *services.PropertyFilter) ([]*domain.EnhancedProperty, int64, error) {
	// Update filter to include search
	if filter == nil {
		filter = &services.PropertyFilter{}
	}
	filter.Search = query

	return r.List(ctx, tenantID, filter)
}

// GetByCustomerID retrieves properties for a specific customer
func (r *PropertyRepositoryImpl) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.EnhancedProperty, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		FROM properties
		WHERE tenant_id = $1 AND customer_id = $2 AND status != 'deleted'
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties by customer: %w", err)
	}
	defer rows.Close()

	properties := make([]*domain.EnhancedProperty, 0)
	for rows.Next() {
		property := &domain.EnhancedProperty{}
		err := rows.Scan(
			&property.ID,
			&property.TenantID,
			&property.CustomerID,
			&property.Name,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.City,
			&property.State,
			&property.ZipCode,
			&property.Country,
			&property.PropertyType,
			&property.LotSize,
			&property.SquareFootage,
			&property.Latitude,
			&property.Longitude,
			&property.AccessInstructions,
			&property.GateCode,
			&property.SpecialInstructions,
			&property.PropertyValue,
			&property.Notes,
			&property.Status,
			&property.CreatedAt,
			&property.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer property: %w", err)
		}
		properties = append(properties, property)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customer properties: %w", err)
	}

	return properties, nil
}

// CheckAddressExists checks if an address already exists for the tenant
func (r *PropertyRepositoryImpl) CheckAddressExists(ctx context.Context, tenantID uuid.UUID, address string, excludeID *uuid.UUID) (bool, error) {
	whereClause := "WHERE tenant_id = $1 AND status != 'deleted' AND CONCAT(address_line1, ', ', city, ', ', state, ' ', zip_code) = $2"
	args := []interface{}{tenantID, address}

	if excludeID != nil {
		whereClause += " AND id != $3"
		args = append(args, *excludeID)
	}

	query := "SELECT COUNT(*) FROM properties " + whereClause

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check address existence: %w", err)
	}

	return count > 0, nil
}

// GetPropertyStatistics retrieves statistics for properties
func (r *PropertyRepositoryImpl) GetPropertyStatistics(ctx context.Context, tenantID uuid.UUID) (*PropertyStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_properties,
			COUNT(CASE WHEN property_type = 'residential' THEN 1 END) as residential_count,
			COUNT(CASE WHEN property_type = 'commercial' THEN 1 END) as commercial_count,
			AVG(CASE WHEN lot_size IS NOT NULL THEN lot_size END) as avg_lot_size,
			AVG(CASE WHEN square_footage IS NOT NULL THEN square_footage END) as avg_square_footage,
			AVG(CASE WHEN property_value IS NOT NULL THEN property_value END) as avg_property_value
		FROM properties
		WHERE tenant_id = $1 AND status != 'deleted'`

	var stats PropertyStatistics
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&stats.TotalProperties,
		&stats.ResidentialCount,
		&stats.CommercialCount,
		&stats.AvgLotSize,
		&stats.AvgSquareFootage,
		&stats.AvgPropertyValue,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get property statistics: %w", err)
	}

	return &stats, nil
}

// GetPropertiesByZipCode retrieves properties grouped by zip code
func (r *PropertyRepositoryImpl) GetPropertiesByZipCode(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT zip_code, COUNT(*) as count
		FROM properties
		WHERE tenant_id = $1 AND status != 'deleted'
		GROUP BY zip_code
		ORDER BY count DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties by zip code: %w", err)
	}
	defer rows.Close()

	zipCodeCounts := make(map[string]int)
	for rows.Next() {
		var zipCode string
		var count int
		err := rows.Scan(&zipCode, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zip code count: %w", err)
		}
		zipCodeCounts[zipCode] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating zip code counts: %w", err)
	}

	return zipCodeCounts, nil
}

// BulkUpdateGeocoding updates geocoding for multiple properties
func (r *PropertyRepositoryImpl) BulkUpdateGeocoding(ctx context.Context, updates []PropertyGeocodingUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for bulk update
	stmt, err := tx.PrepareContext(ctx, `
		UPDATE properties 
		SET latitude = $2, longitude = $3, updated_at = $4
		WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute for each update
	for _, update := range updates {
		_, err := stmt.ExecContext(ctx, update.PropertyID, update.Latitude, update.Longitude, time.Now())
		if err != nil {
			return fmt.Errorf("failed to bulk update geocoding for property %s: %w", update.PropertyID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk geocoding update transaction: %w", err)
	}

	return nil
}

// GetPropertiesWithinBounds gets properties within geographic bounds
func (r *PropertyRepositoryImpl) GetPropertiesWithinBounds(ctx context.Context, tenantID uuid.UUID, northLat, southLat, eastLng, westLng float64) ([]*domain.EnhancedProperty, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		FROM properties
		WHERE tenant_id = $1 
		AND status != 'deleted'
		AND latitude IS NOT NULL 
		AND longitude IS NOT NULL
		AND latitude BETWEEN $2 AND $3
		AND longitude BETWEEN $4 AND $5
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, southLat, northLat, westLng, eastLng)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties within bounds: %w", err)
	}
	defer rows.Close()

	properties := make([]*domain.EnhancedProperty, 0)
	for rows.Next() {
		property := &domain.EnhancedProperty{}
		err := rows.Scan(
			&property.ID,
			&property.TenantID,
			&property.CustomerID,
			&property.Name,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.City,
			&property.State,
			&property.ZipCode,
			&property.Country,
			&property.PropertyType,
			&property.LotSize,
			&property.SquareFootage,
			&property.Latitude,
			&property.Longitude,
			&property.AccessInstructions,
			&property.GateCode,
			&property.SpecialInstructions,
			&property.PropertyValue,
			&property.Notes,
			&property.Status,
			&property.CreatedAt,
			&property.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan property within bounds: %w", err)
		}
		properties = append(properties, property)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating properties within bounds: %w", err)
	}

	return properties, nil
}

// GetPropertiesForRouteOptimization gets properties with coordinates for route planning
func (r *PropertyRepositoryImpl) GetPropertiesForRouteOptimization(ctx context.Context, tenantID uuid.UUID, propertyIDs []uuid.UUID) ([]*services.PropertyRouteInfo, error) {
	if len(propertyIDs) == 0 {
		return []*services.PropertyRouteInfo{}, nil
	}

	// Build IN clause for property IDs
	placeholders := make([]string, len(propertyIDs))
	args := []interface{}{tenantID}
	for i, id := range propertyIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT 
			id, name, address_line1, city, state, zip_code,
			latitude, longitude, access_instructions, gate_code
		FROM properties
		WHERE tenant_id = $1 
		AND id IN (%s)
		AND status != 'deleted'
		AND latitude IS NOT NULL 
		AND longitude IS NOT NULL
		ORDER BY name ASC`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for route optimization: %w", err)
	}
	defer rows.Close()

	routeInfos := make([]*services.PropertyRouteInfo, 0)
	for rows.Next() {
		info := &services.PropertyRouteInfo{}
		err := rows.Scan(
			&info.ID,
			&info.Name,
			&info.AddressLine1,
			&info.City,
			&info.State,
			&info.ZipCode,
			&info.Latitude,
			&info.Longitude,
			&info.AccessInstructions,
			&info.GateCode,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan property route info: %w", err)
		}
		routeInfos = append(routeInfos, info)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating property route infos: %w", err)
	}

	return routeInfos, nil
}

// GetPropertyValueAnalytics gets analytics for property values in an area
func (r *PropertyRepositoryImpl) GetPropertyValueAnalytics(ctx context.Context, tenantID uuid.UUID, city, state string) (*services.PropertyValueAnalytics, error) {
	query := `
		SELECT 
			COUNT(*) as total_properties,
			COUNT(CASE WHEN property_value IS NOT NULL THEN 1 END) as properties_with_value,
			AVG(CASE WHEN property_value IS NOT NULL THEN property_value END) as avg_value,
			MIN(CASE WHEN property_value IS NOT NULL THEN property_value END) as min_value,
			MAX(CASE WHEN property_value IS NOT NULL THEN property_value END) as max_value,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY property_value) as median_value
		FROM properties
		WHERE tenant_id = $1 
		AND status != 'deleted'
		AND city ILIKE $2
		AND state ILIKE $3`

	var analytics services.PropertyValueAnalytics
	err := r.db.QueryRowContext(ctx, query, tenantID, city, state).Scan(
		&analytics.TotalProperties,
		&analytics.PropertiesWithValue,
		&analytics.AvgValue,
		&analytics.MinValue,
		&analytics.MaxValue,
		&analytics.MedianValue,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get property value analytics: %w", err)
	}

	return &analytics, nil
}

// GetPropertiesNeedingGeocoding gets properties without coordinates
func (r *PropertyRepositoryImpl) GetPropertiesNeedingGeocoding(ctx context.Context, tenantID uuid.UUID, limit int) ([]*domain.EnhancedProperty, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, name, address_line1, address_line2,
			city, state, zip_code, country, property_type, lot_size, square_footage,
			latitude, longitude, access_instructions, gate_code, special_instructions,
			property_value, notes, status, created_at, updated_at
		FROM properties
		WHERE tenant_id = $1 
		AND status != 'deleted'
		AND (latitude IS NULL OR longitude IS NULL)
		AND address_line1 IS NOT NULL
		AND city IS NOT NULL
		AND state IS NOT NULL
		ORDER BY created_at ASC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties needing geocoding: %w", err)
	}
	defer rows.Close()

	properties := make([]*domain.EnhancedProperty, 0)
	for rows.Next() {
		property := &domain.EnhancedProperty{}
		err := rows.Scan(
			&property.ID,
			&property.TenantID,
			&property.CustomerID,
			&property.Name,
			&property.AddressLine1,
			&property.AddressLine2,
			&property.City,
			&property.State,
			&property.ZipCode,
			&property.Country,
			&property.PropertyType,
			&property.LotSize,
			&property.SquareFootage,
			&property.Latitude,
			&property.Longitude,
			&property.AccessInstructions,
			&property.GateCode,
			&property.SpecialInstructions,
			&property.PropertyValue,
			&property.Notes,
			&property.Status,
			&property.CreatedAt,
			&property.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan property needing geocoding: %w", err)
		}
		properties = append(properties, property)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating properties needing geocoding: %w", err)
	}

	return properties, nil
}

// Helper types for statistics and bulk operations
type PropertyStatistics struct {
	TotalProperties   int      `json:"total_properties"`
	ResidentialCount  int      `json:"residential_count"`
	CommercialCount   int      `json:"commercial_count"`
	AvgLotSize        *float64 `json:"avg_lot_size"`
	AvgSquareFootage  *float64 `json:"avg_square_footage"`
	AvgPropertyValue  *float64 `json:"avg_property_value"`
}

type PropertyGeocodingUpdate struct {
	PropertyID uuid.UUID `json:"property_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
}

