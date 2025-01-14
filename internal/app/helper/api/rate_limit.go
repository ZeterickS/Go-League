package apiHelper

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu        sync.Mutex
	tokens    int
	maxTokens int
	interval  time.Duration
	lastCheck time.Time
}

func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		tokens:    maxTokens,
		maxTokens: maxTokens,
		interval:  interval,
		lastCheck: time.Now(),
	}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheck)
	rl.lastCheck = now

	// Add tokens based on the elapsed time
	rl.tokens += int(elapsed / rl.interval)
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

var rateLimiter = NewRateLimiter(50, time.Minute) // 50 requests per minute
