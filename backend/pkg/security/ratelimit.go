package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RateLimiter interface defines rate limiting operations
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	AllowN(ctx context.Context, key string, n int) (bool, error)
	Reset(ctx context.Context, key string) error
	GetInfo(ctx context.Context, key string) (*RateLimitInfo, error)
}

// RateLimitInfo contains information about rate limiting status
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	Window    time.Duration `json:"window"`
}

// MemoryRateLimiter implements rate limiting using in-memory storage
type MemoryRateLimiter struct {
	limiters map[string]*rate.Limiter
	limit    rate.Limit
	burst    int
}

// NewMemoryRateLimiter creates a new memory-based rate limiter
func NewMemoryRateLimiter(requestsPerSecond float64, burst int) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(requestsPerSecond),
		burst:    burst,
	}
}

// Allow checks if a request is allowed for the given key
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	limiter := m.getLimiter(key)
	return limiter.Allow(), nil
}

// AllowN checks if N requests are allowed for the given key
func (m *MemoryRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	limiter := m.getLimiter(key)
	return limiter.AllowN(time.Now(), n), nil
}

// Reset resets the rate limiter for the given key
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	delete(m.limiters, key)
	return nil
}

// GetInfo returns rate limit information for the given key
func (m *MemoryRateLimiter) GetInfo(ctx context.Context, key string) (*RateLimitInfo, error) {
	limiter := m.getLimiter(key)
	
	// This is approximate since rate.Limiter doesn't expose internal state
	return &RateLimitInfo{
		Limit:     m.burst,
		Remaining: m.burst - int(limiter.Tokens()),
		Reset:     time.Now().Add(time.Second),
		Window:    time.Second,
	}, nil
}

func (m *MemoryRateLimiter) getLimiter(key string) *rate.Limiter {
	if limiter, exists := m.limiters[key]; exists {
		return limiter
	}
	
	limiter := rate.NewLimiter(m.limit, m.burst)
	m.limiters[key] = limiter
	return limiter
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client      *redis.Client
	limit       int
	window      time.Duration
	keyPrefix   string
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, limit int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		limit:     limit,
		window:    window,
		keyPrefix: "ratelimit:",
	}
}

// Allow checks if a request is allowed for the given key
func (r *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	return r.AllowN(ctx, key, 1)
}

// AllowN checks if N requests are allowed for the given key using sliding window
func (r *RedisRateLimiter) AllowN(ctx context.Context, key string, n int) (bool, error) {
	redisKey := r.keyPrefix + key
	now := time.Now()
	windowStart := now.Add(-r.window)
	
	pipe := r.client.Pipeline()
	
	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", strconv.FormatInt(windowStart.UnixNano(), 10))
	
	// Count current entries
	countCmd := pipe.ZCard(ctx, redisKey)
	
	// Add current request
	for i := 0; i < n; i++ {
		score := now.Add(time.Duration(i) * time.Nanosecond).UnixNano()
		pipe.ZAdd(ctx, redisKey, redis.Z{Score: float64(score), Member: score})
	}
	
	// Set expiration
	pipe.Expire(ctx, redisKey, r.window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit check: %w", err)
	}
	
	currentCount := countCmd.Val()
	return int(currentCount) <= r.limit, nil
}

// Reset resets the rate limiter for the given key
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	redisKey := r.keyPrefix + key
	return r.client.Del(ctx, redisKey).Err()
}

// GetInfo returns rate limit information for the given key
func (r *RedisRateLimiter) GetInfo(ctx context.Context, key string) (*RateLimitInfo, error) {
	redisKey := r.keyPrefix + key
	now := time.Now()
	windowStart := now.Add(-r.window)
	
	// Count current entries in the window
	count, err := r.client.ZCount(ctx, redisKey, strconv.FormatInt(windowStart.UnixNano(), 10), "+inf").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit info: %w", err)
	}
	
	remaining := r.limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	
	return &RateLimitInfo{
		Limit:     r.limit,
		Remaining: remaining,
		Reset:     now.Add(r.window),
		Window:    r.window,
	}, nil
}

// SlidingWindowRateLimiter implements sliding window rate limiting
type SlidingWindowRateLimiter struct {
	client    *redis.Client
	keyPrefix string
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter
func NewSlidingWindowRateLimiter(client *redis.Client) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		client:    client,
		keyPrefix: "sliding_window:",
	}
}

// SlidingWindowConfig represents configuration for sliding window rate limiting
type SlidingWindowConfig struct {
	Limit  int           `json:"limit"`
	Window time.Duration `json:"window"`
}

// AllowWithConfig checks if a request is allowed with specific configuration
func (s *SlidingWindowRateLimiter) AllowWithConfig(ctx context.Context, key string, config SlidingWindowConfig) (bool, *RateLimitInfo, error) {
	redisKey := s.keyPrefix + key
	now := time.Now()
	windowStart := now.Add(-config.Window)
	
	// Lua script for atomic sliding window check
	luaScript := `
		local key = KEYS[1]
		local window_start = ARGV[1]
		local current_time = ARGV[2]
		local limit = tonumber(ARGV[3])
		local expiry = tonumber(ARGV[4])
		
		-- Remove expired entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Count current entries
		local current_count = redis.call('ZCARD', key)
		
		-- Check if we can add a new entry
		if current_count < limit then
			-- Add current request
			redis.call('ZADD', key, current_time, current_time)
			redis.call('EXPIRE', key, expiry)
			return {1, current_count + 1}
		else
			return {0, current_count}
		end
	`
	
	result, err := s.client.Eval(ctx, luaScript, []string{redisKey},
		windowStart.UnixNano(),
		now.UnixNano(),
		config.Limit,
		int(config.Window.Seconds())+1,
	).Result()
	
	if err != nil {
		return false, nil, fmt.Errorf("failed to execute sliding window check: %w", err)
	}
	
	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1
	currentCount := int(resultSlice[1].(int64))
	
	remaining := config.Limit - currentCount
	if remaining < 0 {
		remaining = 0
	}
	
	info := &RateLimitInfo{
		Limit:     config.Limit,
		Remaining: remaining,
		Reset:     now.Add(config.Window),
		Window:    config.Window,
	}
	
	return allowed, info, nil
}

// HierarchicalRateLimiter implements multiple rate limits (e.g., per second, per minute, per hour)
type HierarchicalRateLimiter struct {
	limiters map[string]RateLimiter
}

// NewHierarchicalRateLimiter creates a new hierarchical rate limiter
func NewHierarchicalRateLimiter() *HierarchicalRateLimiter {
	return &HierarchicalRateLimiter{
		limiters: make(map[string]RateLimiter),
	}
}

// AddLimiter adds a rate limiter with a specific name
func (h *HierarchicalRateLimiter) AddLimiter(name string, limiter RateLimiter) {
	h.limiters[name] = limiter
}

// Allow checks if a request is allowed across all configured limiters
func (h *HierarchicalRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	for name, limiter := range h.limiters {
		allowed, err := limiter.Allow(ctx, fmt.Sprintf("%s:%s", name, key))
		if err != nil {
			return false, fmt.Errorf("limiter %s failed: %w", name, err)
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}

// GetAllInfo returns rate limit information from all limiters
func (h *HierarchicalRateLimiter) GetAllInfo(ctx context.Context, key string) (map[string]*RateLimitInfo, error) {
	infos := make(map[string]*RateLimitInfo)
	
	for name, limiter := range h.limiters {
		info, err := limiter.GetInfo(ctx, fmt.Sprintf("%s:%s", name, key))
		if err != nil {
			return nil, fmt.Errorf("failed to get info from limiter %s: %w", name, err)
		}
		infos[name] = info
	}
	
	return infos, nil
}

// AdaptiveRateLimiter adjusts rate limits based on system load
type AdaptiveRateLimiter struct {
	baseLimiter RateLimiter
	baseLimit   int
	maxLimit    int
	minLimit    int
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter
func NewAdaptiveRateLimiter(baseLimiter RateLimiter, baseLimit, minLimit, maxLimit int) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		baseLimiter: baseLimiter,
		baseLimit:   baseLimit,
		maxLimit:    maxLimit,
		minLimit:    minLimit,
	}
}

// Allow checks if a request is allowed with adaptive limits
func (a *AdaptiveRateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	// In a real implementation, you would check system metrics here
	// and adjust the limit based on CPU, memory, response times, etc.
	
	return a.baseLimiter.Allow(ctx, key)
}

// GetCurrentLimit returns the current adaptive limit (placeholder)
func (a *AdaptiveRateLimiter) GetCurrentLimit() int {
	// This would calculate the current limit based on system metrics
	return a.baseLimit
}