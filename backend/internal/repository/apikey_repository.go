package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/pageza/landscaping-app/backend/internal/auth"
	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// apiKeyRepository implements auth.APIKeyRepository using PostgreSQL and Redis
type apiKeyRepository struct {
	db          *sql.DB
	redisClient *redis.Client
	keyPrefix   string
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB, redisClient *redis.Client) auth.APIKeyRepository {
	return &apiKeyRepository{
		db:          db,
		redisClient: redisClient,
		keyPrefix:   "apikey:",
	}
}

// CreateAPIKey creates a new API key in the database
func (r *apiKeyRepository) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (
			id, tenant_id, name, key_hash, key_prefix, permissions, 
			expires_at, status, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		apiKey.ID,
		apiKey.TenantID,
		apiKey.Name,
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		permissionsJSON,
		apiKey.ExpiresAt,
		apiKey.Status,
		apiKey.CreatedBy,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	// Cache the API key for faster lookups
	return r.cacheAPIKey(ctx, apiKey)
}

// GetAPIKeyByHash retrieves an API key by its hash, checking cache first
func (r *apiKeyRepository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	// Try Redis first using hash as lookup key
	if apiKey, err := r.getAPIKeyFromCache(ctx, keyHash); err == nil {
		return apiKey, nil
	}

	// Fall back to database
	return r.getAPIKeyFromDB(ctx, keyHash)
}

// UpdateAPIKeyLastUsed updates the last used timestamp for an API key
func (r *apiKeyRepository) UpdateAPIKeyLastUsed(ctx context.Context, keyID uuid.UUID) error {
	now := time.Now()
	
	// Update database
	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, now, keyID)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}

	// Update cache if it exists
	// First, we need to find the cache key (hash) for this API key ID
	cacheKey := r.keyPrefix + "id:" + keyID.String()
	keyHash, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Update the cached last_used_at
		apiKeyCacheKey := r.keyPrefix + keyHash
		r.redisClient.HSet(ctx, apiKeyCacheKey, "last_used_at", now.Format(time.RFC3339))
	}

	return nil
}

// RevokeAPIKey marks an API key as revoked
func (r *apiKeyRepository) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	// Update database
	query := `UPDATE api_keys SET status = 'revoked', updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), keyID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	// Remove from cache
	// First, find the hash for this key ID
	cacheKey := r.keyPrefix + "id:" + keyID.String()
	keyHash, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Remove both the hash->data mapping and the id->hash mapping
		apiKeyCacheKey := r.keyPrefix + keyHash
		pipe := r.redisClient.Pipeline()
		pipe.Del(ctx, apiKeyCacheKey)
		pipe.Del(ctx, cacheKey)
		pipe.Exec(ctx)
	}

	return nil
}

// ListAPIKeys returns all API keys for a tenant
func (r *apiKeyRepository) ListAPIKeys(ctx context.Context, tenantID uuid.UUID) ([]*domain.APIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, key_prefix, permissions, 
		       last_used_at, expires_at, status, created_by, created_at, updated_at
		FROM api_keys 
		WHERE tenant_id = $1 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*domain.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// Helper methods

func (r *apiKeyRepository) cacheAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	// Calculate TTL - if API key has expiration, use it; otherwise use default
	var ttl time.Duration
	if apiKey.ExpiresAt != nil {
		ttl = time.Until(*apiKey.ExpiresAt)
		if ttl <= 0 {
			return nil // Don't cache expired keys
		}
	} else {
		ttl = 24 * time.Hour // Default cache TTL for non-expiring keys
	}

	// Store API key data as hash using key_hash as the Redis key
	apiKeyCacheKey := r.keyPrefix + apiKey.KeyHash
	
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions for cache: %w", err)
	}

	apiKeyData := map[string]interface{}{
		"id":         apiKey.ID.String(),
		"tenant_id":  apiKey.TenantID.String(),
		"name":       apiKey.Name,
		"key_hash":   apiKey.KeyHash,
		"key_prefix": apiKey.KeyPrefix,
		"permissions": string(permissionsJSON),
		"status":     apiKey.Status,
		"created_at": apiKey.CreatedAt.Format(time.RFC3339),
		"updated_at": apiKey.UpdatedAt.Format(time.RFC3339),
	}

	if apiKey.LastUsedAt != nil {
		apiKeyData["last_used_at"] = apiKey.LastUsedAt.Format(time.RFC3339)
	}
	if apiKey.ExpiresAt != nil {
		apiKeyData["expires_at"] = apiKey.ExpiresAt.Format(time.RFC3339)
	}
	if apiKey.CreatedBy != nil {
		apiKeyData["created_by"] = apiKey.CreatedBy.String()
	}

	pipe := r.redisClient.Pipeline()
	
	// Store the main API key data
	pipe.HMSet(ctx, apiKeyCacheKey, apiKeyData)
	pipe.Expire(ctx, apiKeyCacheKey, ttl)
	
	// Store ID -> hash mapping for reverse lookups
	idCacheKey := r.keyPrefix + "id:" + apiKey.ID.String()
	pipe.Set(ctx, idCacheKey, apiKey.KeyHash, ttl)
	
	_, err = pipe.Exec(ctx)
	return err
}

func (r *apiKeyRepository) getAPIKeyFromCache(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	apiKeyCacheKey := r.keyPrefix + keyHash
	
	result, err := r.redisClient.HGetAll(ctx, apiKeyCacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from cache: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("API key not found in cache")
	}

	apiKey, err := r.parseAPIKeyFromCache(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cached API key: %w", err)
	}

	// Check if API key is expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		r.redisClient.Del(ctx, apiKeyCacheKey) // Clean up expired key
		return nil, fmt.Errorf("API key expired")
	}

	// Check if API key is active
	if apiKey.Status != "active" {
		return nil, fmt.Errorf("API key is not active")
	}

	return apiKey, nil
}

func (r *apiKeyRepository) getAPIKeyFromDB(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `
		SELECT id, tenant_id, name, key_hash, key_prefix, permissions, 
		       last_used_at, expires_at, status, created_by, created_at, updated_at
		FROM api_keys 
		WHERE key_hash = $1 AND status = 'active'
	`

	row := r.db.QueryRowContext(ctx, query, keyHash)
	apiKey, err := r.scanAPIKey(row)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query API key: %w", err)
	}

	// Check if API key is expired
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	// Cache the API key for future lookups
	go r.cacheAPIKey(context.Background(), apiKey)

	return apiKey, nil
}

func (r *apiKeyRepository) scanAPIKey(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.APIKey, error) {
	var apiKey domain.APIKey
	var permissionsJSON []byte

	err := scanner.Scan(
		&apiKey.ID,
		&apiKey.TenantID,
		&apiKey.Name,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&permissionsJSON,
		&apiKey.LastUsedAt,
		&apiKey.ExpiresAt,
		&apiKey.Status,
		&apiKey.CreatedBy,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse permissions
	if len(permissionsJSON) > 0 {
		if err := json.Unmarshal(permissionsJSON, &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
		}
	}

	return &apiKey, nil
}

func (r *apiKeyRepository) parseAPIKeyFromCache(data map[string]string) (*domain.APIKey, error) {
	apiKey := &domain.APIKey{}

	// Parse required fields
	var err error
	
	apiKey.ID, err = uuid.Parse(data["id"])
	if err != nil {
		return nil, fmt.Errorf("invalid API key ID: %w", err)
	}

	apiKey.TenantID, err = uuid.Parse(data["tenant_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	apiKey.Name = data["name"]
	apiKey.KeyHash = data["key_hash"]
	apiKey.KeyPrefix = data["key_prefix"]
	apiKey.Status = data["status"]

	// Parse permissions
	if permissionsStr := data["permissions"]; permissionsStr != "" {
		if err := json.Unmarshal([]byte(permissionsStr), &apiKey.Permissions); err != nil {
			return nil, fmt.Errorf("invalid permissions: %w", err)
		}
	}

	// Parse optional UUID field
	if createdByStr := data["created_by"]; createdByStr != "" {
		createdBy, err := uuid.Parse(createdByStr)
		if err != nil {
			return nil, fmt.Errorf("invalid created_by: %w", err)
		}
		apiKey.CreatedBy = &createdBy
	}

	// Parse timestamps
	apiKey.CreatedAt, err = time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	apiKey.UpdatedAt, err = time.Parse(time.RFC3339, data["updated_at"])
	if err != nil {
		return nil, fmt.Errorf("invalid updated_at: %w", err)
	}

	// Parse optional timestamps
	if lastUsedStr := data["last_used_at"]; lastUsedStr != "" {
		lastUsed, err := time.Parse(time.RFC3339, lastUsedStr)
		if err != nil {
			return nil, fmt.Errorf("invalid last_used_at: %w", err)
		}
		apiKey.LastUsedAt = &lastUsed
	}

	if expiresStr := data["expires_at"]; expiresStr != "" {
		expires, err := time.Parse(time.RFC3339, expiresStr)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at: %w", err)
		}
		apiKey.ExpiresAt = &expires
	}

	return apiKey, nil
}