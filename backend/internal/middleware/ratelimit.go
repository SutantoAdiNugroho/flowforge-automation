package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type rateLimitEntry struct {
	tokens    float64
	lastCheck time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	rate    float64 // tokens per second
	burst   int     // max tokens
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		rate:    requestsPerSecond,
		burst:   burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		key := ctx.IP()
		if tenantID := ctx.Locals("tenant_id"); tenantID != nil {
			switch v := tenantID.(type) {
			case string:
				key = v
			default:
				key = fmt.Sprint(v)
			}
		}

		if !rl.allow(key) {
			return ctx.Status(429).JSON(fiber.Map{
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": "too many requests, please try again later",
			})
		}

		return ctx.Next()
	}
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]
	if !exists {
		rl.entries[key] = &rateLimitEntry{
			tokens:    float64(rl.burst) - 1,
			lastCheck: now,
		}
		return true
	}

	elapsed := now.Sub(entry.lastCheck).Seconds()
	entry.tokens += elapsed * rl.rate
	if entry.tokens > float64(rl.burst) {
		entry.tokens = float64(rl.burst)
	}
	entry.lastCheck = now

	if entry.tokens < 1 {
		return false
	}

	entry.tokens--
	return true
}

// cleanup stale entries every 5 minutes
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for key, entry := range rl.entries {
			if entry.lastCheck.Before(cutoff) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}
