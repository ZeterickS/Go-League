package apiHelper

import (
	"log"
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

// Check if a token is available without consuming it
func (rl *RateLimiter) Check() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastCheck)

	// Calculate how many tokens to add based on the elapsed time
	tokensToAdd := int(elapsed / rl.interval)
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		rl.lastCheck = rl.lastCheck.Add(time.Duration(tokensToAdd) * rl.interval)
	}

	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}

	log.Printf("RateLimiter Check: tokens=%d, maxTokens=%d, elapsed=%v, tokensToAdd=%d", rl.tokens, rl.maxTokens, elapsed, tokensToAdd)

	return rl.tokens > 0
}

// Consume a token if available
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.tokens > 0 {
		rl.tokens--
		log.Printf("RateLimiter Allow: tokens=%d", rl.tokens)
		return true
	}

	return false
}

var rateLimiterPerSecond = NewRateLimiter(10, time.Second)      // 10 requests per second
var rateLimiterPer2Minutes = NewRateLimiter(100, 2*time.Minute) // 100 requests per 2 minutes
