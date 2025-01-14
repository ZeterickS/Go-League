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

	log.Printf("RateLimiter Check: tokens=%d, maxTokens=%d, requests=%d", rl.tokens, rl.maxTokens, len(rl.requests))

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
		log.Printf("RateLimiter Allow: tokens=%d, maxTokens=%d, requests=%d", rl.tokens, rl.maxTokens, len(rl.requests))
		return true
	}

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
}

var rateLimiterPerSecond = NewRateLimiter(10, time.Second)      // 10 requests per second
var rateLimiterPer2Minutes = NewRateLimiter(100, 2*time.Minute) // 100 requests per 2 minutes
