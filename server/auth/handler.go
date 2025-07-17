package auth

import (
	"net/http"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/gin-gonic/gin"
)

type AuthMode string

const (
	AuthModeJWT     AuthMode = "jwt"
	AuthModeSession AuthMode = "session"
	AuthModeBoth    AuthMode = "both"
)

type AuthHandler struct {
	authService *AuthService
	authMode    AuthMode
}

func NewAuthHandler(authService *AuthService, authMode AuthMode) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		authMode:    authMode,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	switch h.authMode {
	case AuthModeJWT:
		h.loginWithJWT(c, req)
	case AuthModeSession:
		h.loginWithSession(c, req)
	case AuthModeBoth:
		h.loginWithBoth(c, req)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid auth mode"})
	}
}

func (h *AuthHandler) loginWithJWT(c *gin.Context, req domain.LoginRequest) {
	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Don't return password hash in response
	responseUser := *user
	responseUser.HashedPassword = ""

	response := domain.AuthResponse{
		Token: token,
		User:  responseUser,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) loginWithSession(c *gin.Context, req domain.LoginRequest) {
	user, session, err := h.authService.LoginWithSession(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("session_id", session.ID, 0, "/", "", false, true)

	// Don't return password hash in response
	responseUser := *user
	responseUser.HashedPassword = ""

	response := domain.AuthResponse{
		Token: "",
		User:  responseUser,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) loginWithBoth(c *gin.Context, req domain.LoginRequest) {
	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	session, err := h.authService.CreateSession(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("session_id", session.ID, 0, "/", "", false, true)

	// Don't return password hash in response
	responseUser := *user
	responseUser.HashedPassword = ""

	response := domain.AuthResponse{
		Token: token,
		User:  responseUser,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Register(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Don't return password hash in response
	responseUser := *user
	responseUser.HashedPassword = ""

	response := domain.AuthResponse{
		Token: token,
		User:  responseUser,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if h.authMode == AuthModeJWT {
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
		return
	}

	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No session found"})
		return
	}

	h.authService.DestroySession(sessionID)

	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *AuthHandler) GetJWKs(c *gin.Context) {
	jwkSet, err := h.authService.GetJWKSet()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, jwkSet)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, exists := h.authService.userStore.GetByID(userID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Don't return password hash in response
	responseUser := *user
	responseUser.HashedPassword = ""

	c.JSON(http.StatusOK, responseUser)
}

// OpenID Connect Discovery endpoint
func (h *AuthHandler) GetOpenIDConfiguration(c *gin.Context) {
	// Get the base URL from the request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + c.Request.Host

	config := map[string]interface{}{
		"issuer":                 baseURL,
		"authorization_endpoint": baseURL + "/auth/authorize",
		"token_endpoint":         baseURL + "/auth/token",
		"userinfo_endpoint":      baseURL + "/auth/me",
		"jwks_uri":               baseURL + "/.well-known/jwks.json",
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"token id_token",
			"code token id_token",
		},
		"subject_types_supported": []string{"public"},
		"id_token_signing_alg_values_supported": []string{
			"RS256",
			"HS256",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_post",
			"client_secret_basic",
		},
	}

	c.JSON(http.StatusOK, config)
}
