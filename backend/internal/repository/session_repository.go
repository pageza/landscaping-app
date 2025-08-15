package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/pageza/landscaping-app/backend/internal/auth"
	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// sessionRepository implements auth.SessionRepository using PostgreSQL and Redis
type sessionRepository struct {
	db          *sql.DB
	redisClient *redis.Client
	keyPrefix   string
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB, redisClient *redis.Client) auth.SessionRepository {
	return &sessionRepository{
		db:          db,
		redisClient: redisClient,
		keyPrefix:   "session:",
	}
}

// CreateSession creates a new user session in both database and Redis
func (r *sessionRepository) CreateSession(ctx context.Context, session *domain.UserSession) error {
	// Insert into database first
	query := `
		INSERT INTO user_sessions (
			id, user_id, session_token, refresh_token, device_info, 
			ip_address, user_agent, expires_at, last_activity, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.SessionToken,
		session.RefreshToken,
		deviceInfoJSON,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.LastActivity,
		session.Status,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create session in database: %w", err)
	}

	// Cache in Redis for faster lookups
	return r.cacheSession(ctx, session)
}

// GetSession retrieves a session by ID, checking Redis first, then database
func (r *sessionRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	// Try Redis first
	if session, err := r.getSessionFromCache(ctx, sessionID); err == nil {
		return session, nil
	}

	// Fall back to database
	return r.getSessionFromDB(ctx, sessionID)
}

// UpdateSession updates session last activity time
func (r *sessionRepository) UpdateSession(ctx context.Context, sessionID uuid.UUID, lastActivity time.Time) error {
	// Update database
	query := `UPDATE user_sessions SET last_activity = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, lastActivity, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	// Update Redis cache
	redisKey := r.keyPrefix + sessionID.String()
	return r.redisClient.HSet(ctx, redisKey, "last_activity", lastActivity.Format(time.RFC3339)).Err()
}

// RevokeSession marks a session as revoked
func (r *sessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	// Update database
	query := `UPDATE user_sessions SET status = 'revoked' WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Remove from Redis cache
	redisKey := r.keyPrefix + sessionID.String()
	return r.redisClient.Del(ctx, redisKey).Err()
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *sessionRepository) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all active sessions for the user
	query := `SELECT id FROM user_sessions WHERE user_id = $1 AND status = 'active'`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to query user sessions: %w", err)
	}
	defer rows.Close()

	var sessionIDs []uuid.UUID
	for rows.Next() {
		var sessionID uuid.UUID
		if err := rows.Scan(&sessionID); err != nil {
			return fmt.Errorf("failed to scan session ID: %w", err)
		}
		sessionIDs = append(sessionIDs, sessionID)
	}

	// Update database
	updateQuery := `UPDATE user_sessions SET status = 'revoked' WHERE user_id = $1 AND status = 'active'`
	_, err = r.db.ExecContext(ctx, updateQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	// Remove from Redis cache
	if len(sessionIDs) > 0 {
		redisKeys := make([]string, len(sessionIDs))
		for i, sessionID := range sessionIDs {
			redisKeys[i] = r.keyPrefix + sessionID.String()
		}
		r.redisClient.Del(ctx, redisKeys...)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from database and cache
func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()

	// Get expired session IDs
	query := `SELECT id FROM user_sessions WHERE expires_at < $1`
	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return fmt.Errorf("failed to query expired sessions: %w", err)
	}
	defer rows.Close()

	var expiredSessions []uuid.UUID
	for rows.Next() {
		var sessionID uuid.UUID
		if err := rows.Scan(&sessionID); err != nil {
			return fmt.Errorf("failed to scan expired session ID: %w", err)
		}
		expiredSessions = append(expiredSessions, sessionID)
	}

	if len(expiredSessions) == 0 {
		return nil
	}

	// Delete from database
	deleteQuery := `DELETE FROM user_sessions WHERE expires_at < $1`
	_, err = r.db.ExecContext(ctx, deleteQuery, now)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	// Remove from Redis cache
	redisKeys := make([]string, len(expiredSessions))
	for i, sessionID := range expiredSessions {
		redisKeys[i] = r.keyPrefix + sessionID.String()
	}
	r.redisClient.Del(ctx, redisKeys...)

	return nil
}

// Helper methods

func (r *sessionRepository) cacheSession(ctx context.Context, session *domain.UserSession) error {
	redisKey := r.keyPrefix + session.ID.String()
	
	// Calculate TTL based on session expiration
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return nil // Don't cache expired sessions
	}

	// Store session data as hash
	sessionData := map[string]interface{}{
		"id":            session.ID.String(),
		"user_id":       session.UserID.String(),
		"session_token": session.SessionToken,
		"refresh_token": session.RefreshToken,
		"ip_address":    session.IPAddress,
		"user_agent":    session.UserAgent,
		"expires_at":    session.ExpiresAt.Format(time.RFC3339),
		"last_activity": session.LastActivity.Format(time.RFC3339),
		"status":        session.Status,
		"created_at":    session.CreatedAt.Format(time.RFC3339),
	}

	if session.DeviceInfo != nil {
		deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
		if err != nil {
			return fmt.Errorf("failed to marshal device info for cache: %w", err)
		}
		sessionData["device_info"] = string(deviceInfoJSON)
	}

	pipe := r.redisClient.Pipeline()
	pipe.HMSet(ctx, redisKey, sessionData)
	pipe.Expire(ctx, redisKey, ttl)
	_, err := pipe.Exec(ctx)

	return err
}

func (r *sessionRepository) getSessionFromCache(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	redisKey := r.keyPrefix + sessionID.String()
	
	result, err := r.redisClient.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session from cache: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("session not found in cache")
	}

	session, err := r.parseSessionFromCache(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cached session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		r.redisClient.Del(ctx, redisKey) // Clean up expired session
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

func (r *sessionRepository) getSessionFromDB(ctx context.Context, sessionID uuid.UUID) (*domain.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, refresh_token, device_info, 
		       ip_address, user_agent, expires_at, last_activity, status, created_at
		FROM user_sessions 
		WHERE id = $1 AND status = 'active'
	`

	row := r.db.QueryRowContext(ctx, query, sessionID)

	var session domain.UserSession
	var deviceInfoJSON []byte

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.SessionToken,
		&session.RefreshToken,
		&deviceInfoJSON,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.Status,
		&session.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query session: %w", err)
	}

	// Parse device info
	if len(deviceInfoJSON) > 0 {
		if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}
	}

	// Cache the session for future lookups
	go r.cacheSession(context.Background(), &session)

	return &session, nil
}

func (r *sessionRepository) parseSessionFromCache(data map[string]string) (*domain.UserSession, error) {
	session := &domain.UserSession{}

	// Parse required fields
	var err error
	
	session.ID, err = uuid.Parse(data["id"])
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %w", err)
	}

	session.UserID, err = uuid.Parse(data["user_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	session.SessionToken = data["session_token"]
	session.RefreshToken = data["refresh_token"]
	session.Status = data["status"]

	// Parse optional string fields
	if ipAddr := data["ip_address"]; ipAddr != "" {
		session.IPAddress = &ipAddr
	}
	if userAgent := data["user_agent"]; userAgent != "" {
		session.UserAgent = &userAgent
	}

	// Parse timestamps
	session.ExpiresAt, err = time.Parse(time.RFC3339, data["expires_at"])
	if err != nil {
		return nil, fmt.Errorf("invalid expires_at: %w", err)
	}

	session.LastActivity, err = time.Parse(time.RFC3339, data["last_activity"])
	if err != nil {
		return nil, fmt.Errorf("invalid last_activity: %w", err)
	}

	session.CreatedAt, err = time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	// Parse device info
	if deviceInfoStr := data["device_info"]; deviceInfoStr != "" {
		if err := json.Unmarshal([]byte(deviceInfoStr), &session.DeviceInfo); err != nil {
			return nil, fmt.Errorf("invalid device_info: %w", err)
		}
	}

	return session, nil
}