package auth

import (
	"encoding/json"
	"fmt"
	"os"
)

// OIDCConfig represents the OIDC provider configuration
type OIDCConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	Issuer       string   `json:"issuer"`
	Scopes       []string `json:"scopes"`
}

// LoadOIDCConfig loads OIDC configuration from a JSON file
func LoadOIDCConfig(configPath string) (*OIDCConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("OIDC config file path is required")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read OIDC config file: %w", err)
	}

	var config OIDCConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse OIDC config JSON: %w", err)
	}

	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("client_id is required in OIDC config")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret is required in OIDC config")
	}
	if len(config.RedirectURIs) == 0 {
		return nil, fmt.Errorf("redirect_uris is required in OIDC config")
	}
	if config.Issuer == "" {
		return nil, fmt.Errorf("issuer is required in OIDC config")
	}

	// Set default scopes if not provided
	if len(config.Scopes) == 0 {
		config.Scopes = []string{"openid", "profile"}
	}

	return &config, nil
}

// ValidateRedirectURI checks if the provided redirect URI is allowed
func (c *OIDCConfig) ValidateRedirectURI(uri string) bool {
	for _, allowedURI := range c.RedirectURIs {
		if uri == allowedURI {
			return true
		}
	}
	return false
}

// ValidateScope checks if the provided scope is supported
func (c *OIDCConfig) ValidateScope(scope string) bool {
	for _, supportedScope := range c.Scopes {
		if scope == supportedScope {
			return true
		}
	}
	return false
}
