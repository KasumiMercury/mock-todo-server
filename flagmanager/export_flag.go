package flagmanager

import (
	"fmt"

	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/spf13/cobra"
)

// ExportFlagConfig holds all flag configurations for the export command
type ExportFlagConfig struct {
	TemplateMode bool
	MemoryMode   bool
	OidcMode     bool
}

// ExportFlagDef defines metadata for export command flags
type ExportFlagDef struct {
	FlagType    FlagType
	Name        string
	ShortName   string
	Description string
	DefaultVal  interface{}
	BindFunc    func(*ExportFlagConfig) interface{} // Returns pointer to the field
}

// exportFlagDefinitions holds all flag metadata for the export command
var exportFlagDefinitions = []ExportFlagDef{
	{
		FlagType:    FlagTypeBool,
		Name:        "template",
		ShortName:   "t",
		Description: "Export JSON data template file",
		DefaultVal:  false,
		BindFunc:    func(c *ExportFlagConfig) interface{} { return &c.TemplateMode },
	},
	{
		FlagType:    FlagTypeBool,
		Name:        "memory",
		ShortName:   "m",
		Description: "Export current memory store state",
		DefaultVal:  false,
		BindFunc:    func(c *ExportFlagConfig) interface{} { return &c.MemoryMode },
	},
	{
		FlagType:    FlagTypeBool,
		Name:        "oidc-config",
		ShortName:   "",
		Description: "Export OIDC configuration template",
		DefaultVal:  false,
		BindFunc:    func(c *ExportFlagConfig) interface{} { return &c.OidcMode },
	},
}

// NewExportFlagConfig creates a new ExportFlagConfig with default values
func NewExportFlagConfig() *ExportFlagConfig {
	config := &ExportFlagConfig{}

	// Set default values from flag definitions using BindFunc
	for _, flagDef := range exportFlagDefinitions {
		fieldPtr := flagDef.BindFunc(config)

		switch flagDef.FlagType {
		case FlagTypeBool:
			*fieldPtr.(*bool) = flagDef.DefaultVal.(bool)
		default:
			panic("unhandled default case")
		}
	}

	return config
}

// RegisterFlags registers all export command flags
func (c *ExportFlagConfig) RegisterFlags(cmd *cobra.Command) {
	for _, flagDef := range exportFlagDefinitions {
		fieldPtr := flagDef.BindFunc(c)

		switch flagDef.FlagType {
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

// ReconstructFlags returns the command line flags equivalent to this config
func (c *ExportFlagConfig) ReconstructFlags() []string {
	var flags []string

	for _, flagDef := range exportFlagDefinitions {
		fieldPtr := flagDef.BindFunc(c)
		defaultValue := flagDef.DefaultVal

		var currentValue interface{}
		var isDefault bool

		switch flagDef.FlagType {
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
				if v {
					// For boolean true, just add the flag
					flags = append(flags, flagName)
				}
				// For boolean false, we don't need to add anything since false is the default
			}
		}
	}

	return flags
}

// Validate checks that exactly one export mode is specified
func (c *ExportFlagConfig) Validate() error {
	activeFlags := 0
	if c.TemplateMode {
		activeFlags++
	}
	if c.MemoryMode {
		activeFlags++
	}
	if c.OidcMode {
		activeFlags++
	}

	if activeFlags == 0 {
		return fmt.Errorf("must specify one of --template, --memory, or --oidc-config flag")
	}

	if activeFlags > 1 {
		return fmt.Errorf("cannot specify multiple export flags simultaneously")
	}

	return nil
}

// GetActiveMode returns the active export mode
func (c *ExportFlagConfig) GetActiveMode() (export.ExportMode, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	if c.TemplateMode {
		return export.TemplateMode, nil
	}
	if c.MemoryMode {
		return export.MemoryExportMode, nil
	}
	if c.OidcMode {
		return export.OidcMode, nil
	}

	return "", fmt.Errorf("no active mode found")
}

// ToExportParams converts flag config and args to export parameters
func (c *ExportFlagConfig) ToExportParams(args []string) (mode export.ExportMode, filePath string, err error) {
	mode, err = c.GetActiveMode()
	if err != nil {
		return "", "", err
	}

	// Get file path from args or use default
	filePath = export.GetOutputPath(args, getDefaultFilename(mode))

	return mode, filePath, nil
}

// getDefaultFilename returns the default filename for the given export mode
func getDefaultFilename(mode export.ExportMode) string {
	switch mode {
	case export.TemplateMode:
		return export.DefaultTemplateFile
	case export.MemoryExportMode:
		return export.DefaultMemoryFile
	case export.OidcMode:
		return export.DefaultOidcFile
	default:
		return "export.json"
	}
}

// FromExportMode populates flag config from export mode
func (c *ExportFlagConfig) FromExportMode(mode export.ExportMode) {
	// Reset all flags to false first
	c.TemplateMode = false
	c.MemoryMode = false
	c.OidcMode = false

	// Set the appropriate flag based on the mode
	switch mode {
	case export.TemplateMode:
		c.TemplateMode = true
	case export.MemoryExportMode:
		c.MemoryMode = true
	case export.OidcMode:
		c.OidcMode = true
	}
}
