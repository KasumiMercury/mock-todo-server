package flagmanager

import (
	"fmt"

	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

type FlagType int

const (
	FlagTypeInt FlagType = iota
	FlagTypeString
	FlagTypeBool
)

// FlagDef defines metadata for a single flag
type FlagDef struct {
	FlagType    FlagType
	Name        string
	ShortName   string
	Description string
	DefaultVal  interface{}
	BindFunc    func(*ServeFlagConfig) interface{} // Returns pointer to the field
}

// ServeFlagConfig holds all flag configurations for the serve command
type ServeFlagConfig struct {
	Port           int
	JsonFilePath   string
	JWTKeyModeStr  string
	JWTSecretKey   string
	AuthRequired   bool
	AuthModeStr    string
	OIDCConfigPath string
}

// flagDefinitions holds all flag metadata for the serve command
var flagDefinitions = []FlagDef{
	{
		FlagType:    FlagTypeInt,
		Name:        "port",
		ShortName:   "p",
		Description: "Port to run the server on",
		DefaultVal:  8080,
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.Port },
	},
	{
		FlagType:    FlagTypeString,
		Name:        "json-file-path",
		ShortName:   "f",
		Description: "File as a Data source for test",
		DefaultVal:  "",
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.JsonFilePath },
	},
	{
		FlagType:    FlagTypeString,
		Name:        "jwt-key-mode",
		ShortName:   "",
		Description: "JWT key mode: 'secret' or 'rsa'",
		DefaultVal:  "secret",
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.JWTKeyModeStr },
	},
	{
		FlagType:    FlagTypeString,
		Name:        "jwt-secret",
		ShortName:   "",
		Description: "JWT secret key (used when jwt-key-mode is 'secret')",
		DefaultVal:  "test-secret-key",
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.JWTSecretKey },
	},
	{
		FlagType:    FlagTypeBool,
		Name:        "auth-required",
		ShortName:   "a",
		Description: "Require authentication for task API endpoints",
		DefaultVal:  true,
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.AuthRequired },
	},
	{
		FlagType:    FlagTypeString,
		Name:        "auth-mode",
		ShortName:   "",
		Description: "Authentication mode: 'jwt', 'session', 'both', or 'oidc'",
		DefaultVal:  "jwt",
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.AuthModeStr },
	},
	{
		FlagType:    FlagTypeString,
		Name:        "oidc-config-path",
		ShortName:   "",
		Description: "Path to OIDC configuration JSON file (required when auth-mode is 'oidc')",
		DefaultVal:  "",
		BindFunc:    func(c *ServeFlagConfig) interface{} { return &c.OIDCConfigPath },
	},
}

// NewServeFlagConfig creates a new ServeFlagConfig with default values
func NewServeFlagConfig() *ServeFlagConfig {
	config := &ServeFlagConfig{}

	// Set default values from flag definitions using BindFunc
	for _, flagDef := range flagDefinitions {
		fieldPtr := flagDef.BindFunc(config)

		switch flagDef.FlagType {
		case FlagTypeInt:
			*fieldPtr.(*int) = flagDef.DefaultVal.(int)
		case FlagTypeString:
			*fieldPtr.(*string) = flagDef.DefaultVal.(string)
		case FlagTypeBool:
			*fieldPtr.(*bool) = flagDef.DefaultVal.(bool)
		}
	}

	return config
}

// RegisterFlags registers all serve command flags
func (c *ServeFlagConfig) RegisterFlags(cmd *cobra.Command) {
	for _, flagDef := range flagDefinitions {
		fieldPtr := flagDef.BindFunc(c)

		switch flagDef.FlagType {
		case FlagTypeInt:
			defaultVal := flagDef.DefaultVal.(int)
			if flagDef.ShortName != "" {
				cmd.Flags().IntVarP(fieldPtr.(*int), flagDef.Name, flagDef.ShortName, defaultVal, flagDef.Description)
			} else {
				cmd.Flags().IntVar(fieldPtr.(*int), flagDef.Name, defaultVal, flagDef.Description)
			}
		case FlagTypeString:
			defaultVal := flagDef.DefaultVal.(string)
			if flagDef.ShortName != "" {
				cmd.Flags().StringVarP(fieldPtr.(*string), flagDef.Name, flagDef.ShortName, defaultVal, flagDef.Description)
			} else {
				cmd.Flags().StringVar(fieldPtr.(*string), flagDef.Name, defaultVal, flagDef.Description)
			}
		case FlagTypeBool:
			defaultVal := flagDef.DefaultVal.(bool)
			if flagDef.ShortName != "" {
				cmd.Flags().BoolVarP(fieldPtr.(*bool), flagDef.Name, flagDef.ShortName, defaultVal, flagDef.Description)
			} else {
				cmd.Flags().BoolVar(fieldPtr.(*bool), flagDef.Name, defaultVal, flagDef.Description)
			}
		}
	}
}

// ToServerConfig converts flag config to server config
func (c *ServeFlagConfig) ToServerConfig() (*server.Config, error) {
	config := server.NewServerConfig()
	config.Port = c.Port
	config.JsonFilePath = c.JsonFilePath
	config.JWTSecretKey = c.JWTSecretKey
	config.AuthRequired = c.AuthRequired
	config.OIDCConfigPath = c.OIDCConfigPath

	// Validate and convert enum fields
	if err := config.ValidateEnumFields(c.JWTKeyModeStr, c.AuthModeStr); err != nil {
		return nil, err
	}

	return config, nil
}

// FromServerConfig populates flag config from server config
func (c *ServeFlagConfig) FromServerConfig(config *server.Config) {
	c.Port = config.Port
	c.JsonFilePath = config.JsonFilePath
	c.JWTSecretKey = config.JWTSecretKey
	c.AuthRequired = config.AuthRequired
	c.OIDCConfigPath = config.OIDCConfigPath

	// Convert enum fields to strings
	enumStrings := config.ToFlagsString()
	c.JWTKeyModeStr = enumStrings["jwt-key-mode"]
	c.AuthModeStr = enumStrings["auth-mode"]
}

// ReconstructFlags returns the command line flags equivalent to this config
func (c *ServeFlagConfig) ReconstructFlags() []string {
	var flags []string

	for _, flagDef := range flagDefinitions {
		fieldPtr := flagDef.BindFunc(c)
		defaultValue := flagDef.DefaultVal

		var currentValue interface{}
		var isDefault bool

		switch flagDef.FlagType {
		case FlagTypeInt:
			currentValue = *fieldPtr.(*int)
			isDefault = (currentValue == defaultValue)
		case FlagTypeString:
			currentValue = *fieldPtr.(*string)
			isDefault = (currentValue == defaultValue)
		case FlagTypeBool:
			currentValue = *fieldPtr.(*bool)
			isDefault = (currentValue == defaultValue)
		}

		// Compare current value with default value
		if !isDefault {
			flagName := "--" + flagDef.Name
			if flagDef.ShortName != "" {
				flagName = "-" + flagDef.ShortName
			}

			switch v := currentValue.(type) {
			case bool:
				if !v {
					// For boolean false, use flag=false format
					flags = append(flags, flagName+"=false")
				}
				// For boolean true, we don't need to add the flag since true is usually the default
			case int:
				flags = append(flags, flagName, fmt.Sprintf("%d", v))
			case string:
				if v != "" || (v == "" && defaultValue != "") {
					flags = append(flags, flagName, v)
				}
			}
		}
	}

	return flags
}
