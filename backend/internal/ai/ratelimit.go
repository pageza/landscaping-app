package ai

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RateLimiter interface defines rate limiting functionality for AI assistants
type RateLimiter interface {
	// Check if a request is allowed
	CheckLimit(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error
	
	// Record token usage
	RecordUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, tokens int) error
	
	// Get current usage statistics
	GetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*UsageStats, error)
	
	// Reset usage counters (typically for testing)
	ResetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error
}

// UsageStats represents current usage statistics
type UsageStats struct {
	RequestsPerMinute int     `json:"requests_per_minute"`
	RequestsPerHour   int     `json:"requests_per_hour"`
	RequestsPerDay    int     `json:"requests_per_day"`
	TokensPerMinute   int     `json:"tokens_per_minute"`
	TokensPerHour     int     `json:"tokens_per_hour"`
	TokensPerDay      int     `json:"tokens_per_day"`
	CostPerDay        float64 `json:"cost_per_day"`
	LastRequest       time.Time `json:"last_request"`
	IsBlocked         bool    `json:"is_blocked"`
	BlockedUntil      *time.Time `json:"blocked_until,omitempty"`
}

// RateLimitError represents a rate limit violation
type RateLimitError struct {
	Message    string    `json:"message"`
	RetryAfter time.Time `json:"retry_after"`
	LimitType  string    `json:"limit_type"`
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	redis  *redis.Client
	config *RateLimitConfig
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(redis *redis.Client, config *RateLimitConfig) *RedisRateLimiter {
	return &RedisRateLimiter{
		redis:  redis,
		config: config,
	}
}

// CheckLimit checks if a request is within rate limits
func (r *RedisRateLimiter) CheckLimit(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	if !r.config.Enabled {
		return nil
	}

	// Check if user is whitelisted
	if userID != nil && r.isWhitelisted(*userID) {
		return nil
	}

	// Check if user is currently blocked
	if blocked, until := r.isBlocked(ctx, tenantID, userID); blocked {
		return &RateLimitError{
			Message:    "Rate limit exceeded, please try again later",
			RetryAfter: until,
			LimitType:  "cooldown",
		}
	}

	// Check various rate limits
	if err := r.checkRequestLimits(ctx, tenantID, userID); err != nil {
		return err
	}

	if err := r.checkTokenLimits(ctx, tenantID, userID); err != nil {
		return err
	}

	if err := r.checkCostLimits(ctx, tenantID, userID); err != nil {
		return err
	}

	// Record the request
	return r.recordRequest(ctx, tenantID, userID)
}

// RecordUsage records token usage for billing and limits
func (r *RedisRateLimiter) RecordUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, tokens int) error {
	if !r.config.Enabled {
		return nil
	}

	now := time.Now()
	pipe := r.redis.Pipeline()

	// Record token usage by time window
	r.incrementCounter(pipe, r.getTokenKey(tenantID, userID, "minute"), tokens, time.Minute, now)
	r.incrementCounter(pipe, r.getTokenKey(tenantID, userID, "hour"), tokens, time.Hour, now)
	r.incrementCounter(pipe, r.getTokenKey(tenantID, userID, "day"), tokens, 24*time.Hour, now)

	// Estimate cost (simplified pricing model)
	cost := float64(tokens) * 0.0001 // $0.0001 per token (example pricing)
	r.incrementCounterFloat(pipe, r.getCostKey(tenantID, userID, "day"), cost, 24*time.Hour, now)

	_, err := pipe.Exec(ctx)
	return err
}

// GetUsage retrieves current usage statistics
func (r *RedisRateLimiter) GetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*UsageStats, error) {
	pipe := r.redis.Pipeline()

	// Get request counters
	reqMinute := pipe.Get(ctx, r.getRequestKey(tenantID, userID, "minute"))
	reqHour := pipe.Get(ctx, r.getRequestKey(tenantID, userID, "hour"))
	reqDay := pipe.Get(ctx, r.getRequestKey(tenantID, userID, "day"))

	// Get token counters
	tokenMinute := pipe.Get(ctx, r.getTokenKey(tenantID, userID, "minute"))
	tokenHour := pipe.Get(ctx, r.getTokenKey(tenantID, userID, "hour"))
	tokenDay := pipe.Get(ctx, r.getTokenKey(tenantID, userID, "day"))

	// Get cost counter
	costDay := pipe.Get(ctx, r.getCostKey(tenantID, userID, "day"))

	// Get last request time
	lastReq := pipe.Get(ctx, r.getLastRequestKey(tenantID, userID))

	// Check if blocked
	blockedUntil := pipe.Get(ctx, r.getBlockKey(tenantID, userID))

	results, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	stats := &UsageStats{}

	// Parse results (ignoring redis.Nil errors)
	if val := reqMinute.Val(); val != "" {
		stats.RequestsPerMinute, _ = strconv.Atoi(val)
	}
	if val := reqHour.Val(); val != "" {
		stats.RequestsPerHour, _ = strconv.Atoi(val)
	}
	if val := reqDay.Val(); val != "" {
		stats.RequestsPerDay, _ = strconv.Atoi(val)
	}
	if val := tokenMinute.Val(); val != "" {
		stats.TokensPerMinute, _ = strconv.Atoi(val)
	}
	if val := tokenHour.Val(); val != "" {
		stats.TokensPerHour, _ = strconv.Atoi(val)
	}
	if val := tokenDay.Val(); val != "" {
		stats.TokensPerDay, _ = strconv.Atoi(val)
	}
	if val := costDay.Val(); val != "" {
		stats.CostPerDay, _ = strconv.ParseFloat(val, 64)
	}
	if val := lastReq.Val(); val != "" {
		if timestamp, err := strconv.ParseInt(val, 10, 64); err == nil {
			stats.LastRequest = time.Unix(timestamp, 0)
		}
	}
	if val := blockedUntil.Val(); val != "" {
		if timestamp, err := strconv.ParseInt(val, 10, 64); err == nil {
			until := time.Unix(timestamp, 0)
			stats.IsBlocked = time.Now().Before(until)
			if stats.IsBlocked {
				stats.BlockedUntil = &until
			}
		}
	}

	return stats, nil
}

// ResetUsage resets all usage counters
func (r *RedisRateLimiter) ResetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	keys := []string{
		r.getRequestKey(tenantID, userID, "minute"),
		r.getRequestKey(tenantID, userID, "hour"),
		r.getRequestKey(tenantID, userID, "day"),
		r.getTokenKey(tenantID, userID, "minute"),
		r.getTokenKey(tenantID, userID, "hour"),
		r.getTokenKey(tenantID, userID, "day"),
		r.getCostKey(tenantID, userID, "day"),
		r.getLastRequestKey(tenantID, userID),
		r.getBlockKey(tenantID, userID),
	}

	return r.redis.Del(ctx, keys...).Err()
}

// Private helper methods

func (r *RedisRateLimiter) isWhitelisted(userID uuid.UUID) bool {
	for _, id := range r.config.WhitelistedUsers {
		if id == userID {
			return true
		}
	}
	return false
}

func (r *RedisRateLimiter) isBlocked(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (bool, time.Time) {
	val := r.redis.Get(ctx, r.getBlockKey(tenantID, userID)).Val()
	if val == "" {
		return false, time.Time{}
	}

	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return false, time.Time{}
	}

	blockedUntil := time.Unix(timestamp, 0)
	return time.Now().Before(blockedUntil), blockedUntil
}

func (r *RedisRateLimiter) checkRequestLimits(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	if r.config.RequestsPerMinute > 0 {
		count := r.getCount(ctx, r.getRequestKey(tenantID, userID, "minute"))
		if count >= r.config.RequestsPerMinute {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Request rate limit exceeded: %d requests per minute", r.config.RequestsPerMinute),
				RetryAfter: time.Now().Add(time.Minute),
				LimitType:  "requests_per_minute",
			}
		}
	}

	if r.config.RequestsPerHour > 0 {
		count := r.getCount(ctx, r.getRequestKey(tenantID, userID, "hour"))
		if count >= r.config.RequestsPerHour {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Request rate limit exceeded: %d requests per hour", r.config.RequestsPerHour),
				RetryAfter: time.Now().Add(time.Hour),
				LimitType:  "requests_per_hour",
			}
		}
	}

	if r.config.RequestsPerDay > 0 {
		count := r.getCount(ctx, r.getRequestKey(tenantID, userID, "day"))
		if count >= r.config.RequestsPerDay {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Request rate limit exceeded: %d requests per day", r.config.RequestsPerDay),
				RetryAfter: time.Now().Add(24 * time.Hour),
				LimitType:  "requests_per_day",
			}
		}
	}

	return nil
}

func (r *RedisRateLimiter) checkTokenLimits(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	if r.config.TokensPerMinute > 0 {
		count := r.getCount(ctx, r.getTokenKey(tenantID, userID, "minute"))
		if count >= r.config.TokensPerMinute {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Token rate limit exceeded: %d tokens per minute", r.config.TokensPerMinute),
				RetryAfter: time.Now().Add(time.Minute),
				LimitType:  "tokens_per_minute",
			}
		}
	}

	if r.config.TokensPerHour > 0 {
		count := r.getCount(ctx, r.getTokenKey(tenantID, userID, "hour"))
		if count >= r.config.TokensPerHour {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Token rate limit exceeded: %d tokens per hour", r.config.TokensPerHour),
				RetryAfter: time.Now().Add(time.Hour),
				LimitType:  "tokens_per_hour",
			}
		}
	}

	if r.config.TokensPerDay > 0 {
		count := r.getCount(ctx, r.getTokenKey(tenantID, userID, "day"))
		if count >= r.config.TokensPerDay {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Token rate limit exceeded: %d tokens per day", r.config.TokensPerDay),
				RetryAfter: time.Now().Add(24 * time.Hour),
				LimitType:  "tokens_per_day",
			}
		}
	}

	return nil
}

func (r *RedisRateLimiter) checkCostLimits(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	if r.config.CostLimitPerDay > 0 {
		cost := r.getCountFloat(ctx, r.getCostKey(tenantID, userID, "day"))
		if cost >= r.config.CostLimitPerDay {
			r.blockUser(ctx, tenantID, userID)
			return &RateLimitError{
				Message:    fmt.Sprintf("Daily cost limit exceeded: $%.2f", r.config.CostLimitPerDay),
				RetryAfter: time.Now().Add(24 * time.Hour),
				LimitType:  "cost_per_day",
			}
		}
	}

	return nil
}

func (r *RedisRateLimiter) recordRequest(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	now := time.Now()
	pipe := r.redis.Pipeline()

	// Increment request counters
	r.incrementCounter(pipe, r.getRequestKey(tenantID, userID, "minute"), 1, time.Minute, now)
	r.incrementCounter(pipe, r.getRequestKey(tenantID, userID, "hour"), 1, time.Hour, now)
	r.incrementCounter(pipe, r.getRequestKey(tenantID, userID, "day"), 1, 24*time.Hour, now)

	// Record last request time
	pipe.Set(ctx, r.getLastRequestKey(tenantID, userID), now.Unix(), 24*time.Hour)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisRateLimiter) blockUser(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) {
	if r.config.CooldownPeriod <= 0 {
		return
	}

	blockedUntil := time.Now().Add(r.config.CooldownPeriod)
	r.redis.Set(ctx, r.getBlockKey(tenantID, userID), blockedUntil.Unix(), r.config.CooldownPeriod)
}

func (r *RedisRateLimiter) incrementCounter(pipe redis.Pipeliner, key string, value int, expiry time.Duration, now time.Time) {
	pipe.Incr(context.Background(), key)
	pipe.Expire(context.Background(), key, expiry)
}

func (r *RedisRateLimiter) incrementCounterFloat(pipe redis.Pipeliner, key string, value float64, expiry time.Duration, now time.Time) {
	pipe.IncrByFloat(context.Background(), key, value)
	pipe.Expire(context.Background(), key, expiry)
}

func (r *RedisRateLimiter) getCount(ctx context.Context, key string) int {
	val := r.redis.Get(ctx, key).Val()
	if val == "" {
		return 0
	}
	count, _ := strconv.Atoi(val)
	return count
}

func (r *RedisRateLimiter) getCountFloat(ctx context.Context, key string) float64 {
	val := r.redis.Get(ctx, key).Val()
	if val == "" {
		return 0
	}
	count, _ := strconv.ParseFloat(val, 64)
	return count
}

// Key generation methods
func (r *RedisRateLimiter) getRequestKey(tenantID uuid.UUID, userID *uuid.UUID, window string) string {
	if userID != nil {
		return fmt.Sprintf("ai:ratelimit:requests:%s:%s:%s", tenantID.String(), userID.String(), window)
	}
	return fmt.Sprintf("ai:ratelimit:requests:%s:anonymous:%s", tenantID.String(), window)
}

func (r *RedisRateLimiter) getTokenKey(tenantID uuid.UUID, userID *uuid.UUID, window string) string {
	if userID != nil {
		return fmt.Sprintf("ai:ratelimit:tokens:%s:%s:%s", tenantID.String(), userID.String(), window)
	}
	return fmt.Sprintf("ai:ratelimit:tokens:%s:anonymous:%s", tenantID.String(), window)
}

func (r *RedisRateLimiter) getCostKey(tenantID uuid.UUID, userID *uuid.UUID, window string) string {
	if userID != nil {
		return fmt.Sprintf("ai:ratelimit:cost:%s:%s:%s", tenantID.String(), userID.String(), window)
	}
	return fmt.Sprintf("ai:ratelimit:cost:%s:anonymous:%s", tenantID.String(), window)
}

func (r *RedisRateLimiter) getLastRequestKey(tenantID uuid.UUID, userID *uuid.UUID) string {
	if userID != nil {
		return fmt.Sprintf("ai:ratelimit:last:%s:%s", tenantID.String(), userID.String())
	}
	return fmt.Sprintf("ai:ratelimit:last:%s:anonymous", tenantID.String())
}

func (r *RedisRateLimiter) getBlockKey(tenantID uuid.UUID, userID *uuid.UUID) string {
	if userID != nil {
		return fmt.Sprintf("ai:ratelimit:blocked:%s:%s", tenantID.String(), userID.String())
	}
	return fmt.Sprintf("ai:ratelimit:blocked:%s:anonymous", tenantID.String())
}

// NoOpRateLimiter is a rate limiter that doesn't actually limit anything (for testing/development)
type NoOpRateLimiter struct{}

func NewNoOpRateLimiter() *NoOpRateLimiter {
	return &NoOpRateLimiter{}
}

func (n *NoOpRateLimiter) CheckLimit(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	return nil
}

func (n *NoOpRateLimiter) RecordUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, tokens int) error {
	return nil
}

func (n *NoOpRateLimiter) GetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*UsageStats, error) {
	return &UsageStats{}, nil
}

func (n *NoOpRateLimiter) ResetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	return nil
}