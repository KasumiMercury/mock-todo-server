package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *AuthService, authMode AuthMode) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID int
		var authenticated bool

		switch authMode {
		case AuthModeJWT:
			userID, authenticated = authenticateWithJWT(c, authService)
		case AuthModeSession:
			userID, authenticated = authenticateWithSession(c, authService)
		case AuthModeBoth:
			userID, authenticated = authenticateWithBoth(c, authService)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid auth mode"})
			c.Abort()
			return
		}

		if !authenticated {
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("userID", userID)
		c.Next()
	}
}

func authenticateWithJWT(c *gin.Context, authService *AuthService) (int, bool) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
		return 0, false
	}

	// Check if header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		return 0, false
	}

	// Extract token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return 0, false
	}

	// Validate token
	token, err := authService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return 0, false
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return 0, false
	}

	// Extract user ID from token
	userID, err := authService.GetUserIDFromToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return 0, false
	}

	return userID, true
}

func authenticateWithSession(c *gin.Context, authService *AuthService) (int, bool) {
	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session required"})
		return 0, false
	}

	// Validate session
	session, valid := authService.ValidateSession(sessionID)
	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return 0, false
	}

	return session.UserID, true
}

func authenticateWithBoth(c *gin.Context, authService *AuthService) (int, bool) {
	// Try JWT first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		userID, authenticated := authenticateWithJWT(c, authService)
		if authenticated {
			return userID, true
		}
	}

	// Try session authentication
	_, err := c.Cookie("session_id")
	if err == nil {
		userID, authenticated := authenticateWithSession(c, authService)
		if authenticated {
			return userID, true
		}
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
	return 0, false
}

func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}

	id, ok := userID.(int)
	return id, ok
}
