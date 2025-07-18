package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OIDCHandler handles OIDC endpoints
type OIDCHandler struct {
	oidcService *OIDCService
	authService *AuthService
}

// NewOIDCHandler creates a new OIDC handler
func NewOIDCHandler(oidcService *OIDCService, authService *AuthService) *OIDCHandler {
	return &OIDCHandler{
		oidcService: oidcService,
		authService: authService,
	}
}

// Authorize handles the authorization endpoint
func (h *OIDCHandler) Authorize(c *gin.Context) {
	// Parse query parameters
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")
	responseType := c.Query("response_type")
	scope := c.Query("scope")
	state := c.Query("state")

	// Validate required parameters
	if clientID == "" || redirectURI == "" || responseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Missing required parameters",
		})
		return
	}

	// Validate client ID and redirect URI
	if clientID != h.oidcService.config.ClientID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_client",
			"error_description": "Invalid client_id",
		})
		return
	}

	if !h.oidcService.config.ValidateRedirectURI(redirectURI) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "Invalid redirect_uri",
		})
		return
	}

	// Validate response type
	if responseType != "code" {
		redirectError(c, redirectURI, "unsupported_response_type", "Only authorization code flow is supported", state)
		return
	}

	// Parse and validate scopes
	scopes := h.oidcService.ParseScopes(scope)
	if err := h.oidcService.ValidateScopes(scopes); err != nil {
		redirectError(c, redirectURI, "invalid_scope", err.Error(), state)
		return
	}

	// Check if user is already authenticated (simple implementation)
	if c.Request.Method == "POST" {
		h.handleLogin(c, clientID, redirectURI, scopes, state)
		return
	}

	// Show login form
	h.showLoginForm(c, clientID, redirectURI, scope, state)
}

// handleLogin processes the login form submission
func (h *OIDCHandler) handleLogin(c *gin.Context, clientID, redirectURI string, scopes []string, state string) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if username == "" || password == "" {
		h.showLoginForm(c, clientID, redirectURI, strings.Join(scopes, " "), state)
		return
	}

	// Authenticate user
	user, _, err := h.authService.Login(username, password)
	if err != nil {
		h.showLoginForm(c, clientID, redirectURI, strings.Join(scopes, " "), state)
		return
	}

	// Generate authorization code
	code, err := h.oidcService.GenerateAuthCode(clientID, user.ID, redirectURI, scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to generate authorization code",
		})
		return
	}

	// Redirect back to client with code
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += fmt.Sprintf("&state=%s", url.QueryEscape(state))
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// showLoginForm displays the login form
func (h *OIDCHandler) showLoginForm(c *gin.Context, clientID, redirectURI, scope, state string) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"ClientID":    clientID,
		"RedirectURI": redirectURI,
		"Scope":       scope,
		"State":       state,
	})
}

// Token handles the token endpoint
func (h *OIDCHandler) Token(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	code := c.PostForm("code")
	redirectURI := c.PostForm("redirect_uri")
	clientID := c.PostForm("client_id")
	clientSecret := c.PostForm("client_secret")

	// Validate grant type
	if grantType != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "unsupported_grant_type",
			"error_description": "Only authorization_code grant type is supported",
		})
		return
	}

	// Validate client credentials
	if clientID != h.oidcService.config.ClientID || clientSecret != h.oidcService.config.ClientSecret {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_client",
			"error_description": "Invalid client credentials",
		})
		return
	}

	// Validate and consume authorization code
	authCode, err := h.oidcService.ValidateAuthCode(code, clientID, redirectURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_grant",
			"error_description": err.Error(),
		})
		return
	}

	// Get user
	user, exists := h.oidcService.userStore.GetByID(authCode.UserID)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "User not found",
		})
		return
	}

	// Generate access token
	accessToken, err := h.oidcService.GenerateAccessToken(user, authCode.Scopes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Failed to generate access token",
		})
		return
	}

	// Generate ID token if openid scope is requested
	var idToken string
	if h.oidcService.containsScope(authCode.Scopes, "openid") {
		idToken, err = h.oidcService.GenerateIDToken(user, authCode.Scopes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":             "server_error",
				"error_description": "Failed to generate ID token",
			})
			return
		}
	}

	// Return token response
	response := gin.H{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"scope":        strings.Join(authCode.Scopes, " "),
	}

	if idToken != "" {
		response["id_token"] = idToken
	}

	c.JSON(http.StatusOK, response)
}

// UserInfo handles the userinfo endpoint
func (h *OIDCHandler) UserInfo(c *gin.Context) {
	// Get access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Bearer token required",
		})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate access token
	token, err := h.oidcService.ValidateAccessToken(tokenString)
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":             "invalid_token",
			"error_description": "Invalid access token",
		})
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Invalid token claims",
		})
		return
	}

	// Get user ID and scopes
	userIDStr, ok := claims["sub"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Invalid user ID in token",
		})
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "Invalid user ID format",
		})
		return
	}

	scopeStr, _ := claims["scope"].(string)
	scopes := h.oidcService.ParseScopes(scopeStr)

	// Get user info
	userInfo, err := h.oidcService.GetUserInfo(userID, scopes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":             "invalid_token",
			"error_description": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// GetOpenIDConfiguration handles the discovery endpoint
func (h *OIDCHandler) GetOpenIDConfiguration(c *gin.Context) {
	config := h.oidcService.GetOpenIDConfiguration()
	c.JSON(http.StatusOK, config)
}

// redirectError redirects back to client with an error
func redirectError(c *gin.Context, redirectURI, errorCode, errorDescription, state string) {
	redirectURL := fmt.Sprintf("%s?error=%s&error_description=%s",
		redirectURI, url.QueryEscape(errorCode), url.QueryEscape(errorDescription))

	if state != "" {
		redirectURL += fmt.Sprintf("&state=%s", url.QueryEscape(state))
	}

	c.Redirect(http.StatusFound, redirectURL)
}
