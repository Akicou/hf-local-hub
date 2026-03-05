package middleware

import (
	"net"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipInfo struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	ips       map[string]*ipInfo
	mu        sync.RWMutex
	r         rate.Limit
	b         int
	cleanupCh chan struct{}
}

func NewRateLimiter(requestsPerMin, burst int) *RateLimiter {
	r := rate.Every(time.Minute / time.Duration(requestsPerMin))
	rl := &RateLimiter{
		ips:       make(map[string]*ipInfo),
		r:         r,
		b:         burst,
		cleanupCh: make(chan struct{}),
	}

	// Start cleanup goroutine that runs every minute
	go rl.cleanupLoop()

	return rl
}

// cleanupLoop removes stale IP entries every minute
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.cleanupCh:
			return
		}
	}
}

// cleanup removes IP entries that haven't been seen in the last 10 minutes
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	staleThreshold := 10 * time.Minute

	for ip, info := range rl.ips {
		if now.Sub(info.lastSeen) > staleThreshold {
			delete(rl.ips, ip)
		}
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	info, exists := rl.ips[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.r, rl.b)
		rl.ips[ip] = &ipInfo{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	info.lastSeen = time.Now()
	return info.limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			ip = c.Request.RemoteAddr
		}

		limiter := rl.getLimiter(ip)
		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Stop stops the cleanup goroutine (call on server shutdown)
func (rl *RateLimiter) Stop() {
	close(rl.cleanupCh)
}
