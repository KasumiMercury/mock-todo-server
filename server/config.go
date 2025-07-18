package server

import (
	"fmt"

	"github.com/KasumiMercury/mock-todo-server/server/auth"
)

// Config holds all configuration options for the server
type Config struct {
	Port           int
	JsonFilePath   string
	JWTKeyMode     auth.JWTKeyMode
	JWTSecretKey   string
	AuthRequired   bool
	AuthMode       auth.AuthMode
	OIDCConfigPath string
}

// NewServerConfig creates a new ServerConfig with default values
func NewServerConfig() *Config {
	return &Config{
		Port:         8080,
		JWTKeyMode:   auth.JWTKeyModeSecret,
		JWTSecretKey: "test-secret-key",
		AuthRequired: true,
		AuthMode:     auth.AuthModeJWT,
	}
}

// ValidateAndConvert validates the configuration and converts string values to appropriate types
func (c *Config) ValidateAndConvert(keyModeStr, authModeStr string) error {
	// Validate and convert JWT key mode
	switch keyModeStr {
	case "secret":
		c.JWTKeyMode = auth.JWTKeyModeSecret
	case "rsa":
		c.JWTKeyMode = auth.JWTKeyModeRSA
	default:
		return fmt.Errorf("invalid jwt-key-mode: %s (must be 'secret' or 'rsa')", keyModeStr)
	}

	// Validate and convert auth mode
	switch authModeStr {
	case "jwt":
		c.AuthMode = auth.AuthModeJWT
	case "session":
		c.AuthMode = auth.AuthModeSession
	case "both":
		c.AuthMode = auth.AuthModeBoth
	case "oidc":
		c.AuthMode = auth.AuthModeOIDC
		// OIDC mode requires config file
		if c.OIDCConfigPath == "" {
			return fmt.Errorf("OIDC config file path is required when using auth-mode=oidc")
		}
	default:
		return fmt.Errorf("invalid auth-mode: %s (must be 'jwt', 'session', 'both', or 'oidc')", authModeStr)
	}

	return nil
}

// Validate performs additional validation on the configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	// Additional validation for OIDC mode
	if c.AuthMode == auth.AuthModeOIDC && c.OIDCConfigPath == "" {
		return fmt.Errorf("OIDC config file path is required when using OIDC auth mode")
	}

	return nil
}
