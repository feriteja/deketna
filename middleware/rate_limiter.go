package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Visitor tracks request limits per IP
type Visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

// Global and Specific Rate Limits
const (
	GlobalLimitRequests   = 15
	GlobalLimitWindow     = 1 * time.Minute
	SpecificLimitRequests = 5
	SpecificLimitWindow   = 1 * time.Minute
)

var (
	globalVisitors   = make(map[string]*Visitor)
	specificVisitors = make(map[string]*Visitor)
	muGlobal         sync.Mutex
	muSpecific       sync.Mutex
)

// getVisitor retrieves or creates a rate limiter for an IP
func getVisitor(ip string, visitors map[string]*Visitor, mu *sync.Mutex, limit int, window time.Duration) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Every(window/time.Duration(limit)), limit)
		visitors[ip] = &Visitor{Limiter: limiter, LastSeen: time.Now()}
		return limiter
	}

	v.LastSeen = time.Now()
	return v.Limiter
}

// cleanupVisitors removes stale visitors from the map
func cleanupVisitors(visitors map[string]*Visitor, mu *sync.Mutex, window time.Duration) {
	for {
		time.Sleep(5 * time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.LastSeen) > window {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// GlobalRateLimiter limits general API usage
func GlobalRateLimiter() gin.HandlerFunc {
	// Start cleanup routine
	go cleanupVisitors(globalVisitors, &muGlobal, GlobalLimitWindow)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip, globalVisitors, &muGlobal, GlobalLimitRequests, GlobalLimitWindow)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SpecificRateLimiter limits specific API usage
func SpecificRateLimiter() gin.HandlerFunc {
	// Start cleanup routine
	go cleanupVisitors(specificVisitors, &muSpecific, SpecificLimitWindow)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip, specificVisitors, &muSpecific, SpecificLimitRequests, SpecificLimitWindow)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests to this endpoint. Please try again later.",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
