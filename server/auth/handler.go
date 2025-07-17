package auth

import (
	"net/http"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, token, err := h.authService.Register(req.Username, req.Email, req.Password)
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
