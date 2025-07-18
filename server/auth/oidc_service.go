package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/golang-jwt/jwt/v5"
)

// AuthCode represents an authorization code
type AuthCode struct {
	Code        string
	ClientID    string
	UserID      int
	RedirectURI string
	Scopes      []string
	ExpiresAt   time.Time
}

// OIDCService handles OIDC provider functionality
type OIDCService struct {
	config      *OIDCConfig
	authCodes   map[string]*AuthCode // In-memory storage for auth codes
	userStore   UserStore
	keyMode     JWTKeyMode
	secretKey   []byte
	authService *AuthService
}

type UserStore interface {
	GetByUsername(username string) (*domain.User, bool)
	GetByID(id int) (*domain.User, bool)
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(config *OIDCConfig, userStore UserStore, authService *AuthService) *OIDCService {
	return &OIDCService{
		config:      config,
		authCodes:   make(map[string]*AuthCode),
		userStore:   userStore,
		authService: authService,
	}
}

// GenerateAuthCode generates a new authorization code
func (s *OIDCService) GenerateAuthCode(clientID string, userID int, redirectURI string, scopes []string) (string, error) {
	// Generate random code
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate auth code: %w", err)
	}

	code := base64.URLEncoding.EncodeToString(bytes)

	// Store auth code
	s.authCodes[code] = &AuthCode{
		Code:        code,
		ClientID:    clientID,
		UserID:      userID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minutes expiry
	}

	return code, nil
}

// ValidateAuthCode validates and consumes an authorization code
func (s *OIDCService) ValidateAuthCode(code, clientID, redirectURI string) (*AuthCode, error) {
	authCode, exists := s.authCodes[code]
	if !exists {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Check expiry
	if time.Now().After(authCode.ExpiresAt) {
		delete(s.authCodes, code)
		return nil, fmt.Errorf("authorization code expired")
	}

	// Validate client and redirect URI
	if authCode.ClientID != clientID {
		return nil, fmt.Errorf("client ID mismatch")
	}

	if authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("redirect URI mismatch")
	}

	// Consume the code (one-time use)
	delete(s.authCodes, code)

	return authCode, nil
}

// GenerateIDToken generates an OpenID Connect ID token
func (s *OIDCService) GenerateIDToken(user *domain.User, scopes []string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"iss": s.config.Issuer,
		"sub": fmt.Sprintf("%d", user.ID),
		"aud": s.config.ClientID,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Hour).Unix(),
	}

	// Add profile information based on requested scopes
	if s.containsScope(scopes, "profile") {
		claims["name"] = user.Username
		claims["preferred_username"] = user.Username
	}

	// Use existing auth service to generate token
	return s.authService.generateJWTWithClaims(claims)
}

// GenerateAccessToken generates an access token
func (s *OIDCService) GenerateAccessToken(user *domain.User, scopes []string) (string, error) {
	now := time.Now()

	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID),
		"iat":   now.Unix(),
		"exp":   now.Add(1 * time.Hour).Unix(),
		"scope": strings.Join(scopes, " "),
		"iss":   s.config.Issuer,
		"aud":   s.config.ClientID,
	}

	return s.authService.generateJWTWithClaims(claims)
}

// ValidateAccessToken validates an access token
func (s *OIDCService) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return s.authService.ValidateToken(tokenString)
}

// GetUserInfo returns user information for the userinfo endpoint
func (s *OIDCService) GetUserInfo(userID int, scopes []string) (map[string]interface{}, error) {
	user, exists := s.userStore.GetByID(userID)
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	userInfo := map[string]interface{}{
		"sub": fmt.Sprintf("%d", user.ID),
	}

	if s.containsScope(scopes, "profile") {
		userInfo["name"] = user.Username
		userInfo["preferred_username"] = user.Username
	}

	return userInfo, nil
}

// GetOpenIDConfiguration returns the OpenID Connect discovery document
func (s *OIDCService) GetOpenIDConfiguration() map[string]interface{} {
	return map[string]interface{}{
		"issuer":                                s.config.Issuer,
		"authorization_endpoint":                fmt.Sprintf("%s/auth/authorize", s.config.Issuer),
		"token_endpoint":                        fmt.Sprintf("%s/auth/token", s.config.Issuer),
		"userinfo_endpoint":                     fmt.Sprintf("%s/auth/userinfo", s.config.Issuer),
		"jwks_uri":                              fmt.Sprintf("%s/.well-known/jwks.json", s.config.Issuer),
		"scopes_supported":                      s.config.Scopes,
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256", "HS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic"},
		"claims_supported":                      []string{"sub", "name", "preferred_username"},
	}
}

// containsScope checks if a specific scope is in the scopes list
func (s *OIDCService) containsScope(scopes []string, scope string) bool {
	for _, s := range scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// ParseScopes parses a space-separated scope string into a slice
func (s *OIDCService) ParseScopes(scopeString string) []string {
	if scopeString == "" {
		return []string{}
	}
	return strings.Fields(scopeString)
}

// ValidateScopes validates that all requested scopes are supported
func (s *OIDCService) ValidateScopes(scopes []string) error {
	for _, scope := range scopes {
		if !s.config.ValidateScope(scope) {
			return fmt.Errorf("unsupported scope: %s", scope)
		}
	}
	return nil
}
