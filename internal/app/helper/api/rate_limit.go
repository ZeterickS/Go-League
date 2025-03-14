package apiHelper

import (
	"os"
	"strconv"
	"sync"
	"time"

	"discord-bot/internal/logger"

	"go.uber.org/zap"
)

type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	interval  time.Duration
	requests  []time.Time
}

func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		interval:  interval,
		requests:  make([]time.Time, 0, maxTokens),
	}
}

// Check if a token is available without consuming it
func (rl *RateLimiter) Check() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.cleanup(now)

	if len(rl.requests) > 0 {
		oldestTokenAge := now.Sub(rl.requests[0])
		logger.Logger.Debug("Oldest token age", zap.Duration("age", oldestTokenAge))
	}

	return len(rl.requests) < rl.maxTokens
}

// Consume a token if available
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.cleanup(now)

	if len(rl.requests) < rl.maxTokens {
		rl.requests = append(rl.requests, now)
		logger.Logger.Debug("Token consumed", zap.Time("time", now))
		return true
	}

	logger.Logger.Warn("Rate limit exceeded", zap.Time("time", now))
	return false
}

// Cleanup removes tokens that are older than the interval
func (rl *RateLimiter) cleanup(now time.Time) {
	cutoff := now.Add(-rl.interval)
	i := 0
	for i < len(rl.requests) && rl.requests[i].Before(cutoff) {
		i++
	}
	rl.requests = rl.requests[i:]
	logger.Logger.Debug("Tokens cleaned up", zap.Int("remainingTokens", len(rl.requests)))
}

func (rl *RateLimiter) GetInterval() time.Duration {
	return rateLimiterPerSecond.interval
}

func (rl *RateLimiter) GetMaxTokens() int {
	return rateLimiterPerSecond.maxTokens
}

func getEnvAsInt(name string, defaultValue int) int {
	valueStr := os.Getenv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

var rateLimiterRequestPerTime = NewRateLimiter(1, time.Duration(1000/getEnvAsInt("API_RATE_LIMIT_SECOND", 1)) * time.Millisecond)
var rateLimiterPerSecond = NewRateLimiter(getEnvAsInt("API_RATE_LIMIT_SECOND", 1), time.Second)
//var rateLimiterPer2Minutes = NewRateLimiter(getEnvAsInt("API_RATE_LIMIT_2_MINUTE", 50), 2*time.Minute)
