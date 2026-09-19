package httpapi

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vignesh/livepoll/internal/auth"
	"github.com/vignesh/livepoll/internal/live"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RealIP extracts the client's real IP address.
func RealIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.GetHeader("X-Forwarded-For")
		if ip == "" {
			ip = c.GetHeader("X-Real-IP")
		}
		if ip == "" {
			ip, _, _ = net.SplitHostPort(c.Request.RemoteAddr)
		} else {
			ip = strings.Split(ip, ",")[0]
		}
		c.Set("realIP", strings.TrimSpace(ip))
		c.Next()
	}
}

// AuthRequired enforces that the request has a valid session token.
func AuthRequired(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(auth.SessionCookie)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}

		claims, err := tm.Parse(cookie)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "invalid session")
			c.Abort()
			return
		}

		uid, err := primitive.ObjectIDFromHex(claims.Subject)
		if err != nil {
			respondError(c, http.StatusUnauthorized, "invalid session")
			c.Abort()
			return
		}

		c.Set("userID", uid)
		c.Set("claims", claims)
		c.Next()
	}
}

// AuthOptional sets the user ID if a valid session exists, but allows the
// request to proceed regardless.
func AuthOptional(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(auth.SessionCookie)
		if err == nil && cookie != "" {
			if claims, err := tm.Parse(cookie); err == nil {
				if uid, err := primitive.ObjectIDFromHex(claims.Subject); err == nil {
					c.Set("userID", uid)
					c.Set("claims", claims)
				}
			}
		}
		c.Next()
	}
}

// RateLimit uses Redis to throttle requests by IP.
func RateLimit(l *live.Live, scope string, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.GetString("realIP")
		allowed, err := l.Allow(c.Request.Context(), scope, ip, limit)
		if err != nil {
			// On redis failure, fail open or fail closed? Fail closed is safer,
			// but fail open keeps the app partially alive. Given Redis is critical,
			// we can fail closed with 500.
			respondError(c, http.StatusInternalServerError, "rate limiter unavailable")
			c.Abort()
			return
		}
		if !allowed {
			respondError(c, http.StatusTooManyRequests, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

// getUserID is a helper to fetch the authenticated user ID from context.
func getUserID(c *gin.Context) (primitive.ObjectID, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return primitive.NilObjectID, false
	}
	uid, ok := val.(primitive.ObjectID)
	return uid, ok
}
