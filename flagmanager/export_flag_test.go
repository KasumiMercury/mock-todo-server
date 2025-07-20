package flagmanager

import (
	"strings"
	"testing"

	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/spf13/cobra"
)

func TestNewExportFlagConfig(t *testing.T) {
	config := NewExportFlagConfig()
	if config == nil {
		t.Fatal("NewExportFlagConfig returned nil")
	}

	// Check default values (all should be false)
	if config.TemplateMode {
		t.Error("Expected default TemplateMode to be false")
	}
	if config.MemoryMode {
		t.Error("Expected default MemoryMode to be false")
	}
	if config.OidcMode {
		t.Error("Expected default OidcMode to be false")
	}
}

func TestExportRegisterFlags(t *testing.T) {
	config := NewExportFlagConfig()
	cmd := &cobra.Command{Use: "test"}

	config.RegisterFlags(cmd)

	// Check that expected flags were registered
	expectedFlags := []string{"template", "memory", "oidc-config"}
	for _, flagName := range expectedFlags {
		if flag := cmd.Flags().Lookup(flagName); flag == nil {
			t.Errorf("Expected flag %s was not registered", flagName)
		}
	}

	// Check that short flags are properly registered
	expectedShortFlags := map[string]string{
		"t": "template",
		"m": "memory",
	}

	for shortFlag, longFlag := range expectedShortFlags {
		if flag := cmd.Flags().ShorthandLookup(shortFlag); flag == nil {
			t.Errorf("Expected short flag -%s for %s was not registered", shortFlag, longFlag)
		} else if flag.Name != longFlag {
			t.Errorf("Short flag -%s points to %s, expected %s", shortFlag, flag.Name, longFlag)
		}
	}

	// Check that oidc-config has no short flag
	if flag := cmd.Flags().ShorthandLookup("o"); flag != nil {
		t.Error("oidc-config should not have a short flag")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      ExportFlagConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no flags set",
			config:      ExportFlagConfig{},
			expectError: true,
			errorMsg:    "must specify one of",
		},
		{
			name:        "template mode only",
			config:      ExportFlagConfig{TemplateMode: true},
			expectError: false,
		},
		{
			name:        "memory mode only",
			config:      ExportFlagConfig{MemoryMode: true},
			expectError: false,
		},
		{
			name:        "oidc mode only",
			config:      ExportFlagConfig{OidcMode: true},
			expectError: false,
		},
		{
			name:        "multiple flags set - template and memory",
			config:      ExportFlagConfig{TemplateMode: true, MemoryMode: true},
			expectError: true,
			errorMsg:    "cannot specify multiple",
		},
		{
			name:        "multiple flags set - all three",
			config:      ExportFlagConfig{TemplateMode: true, MemoryMode: true, OidcMode: true},
			expectError: true,
			errorMsg:    "cannot specify multiple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestGetActiveMode(t *testing.T) {
	tests := []struct {
		name         string
		config       ExportFlagConfig
		expectedMode export.ExportMode
		expectError  bool
	}{
		{
			name:        "no mode set",
			config:      ExportFlagConfig{},
			expectError: true,
		},
		{
			name:         "template mode",
			config:       ExportFlagConfig{TemplateMode: true},
			expectedMode: export.TemplateMode,
			expectError:  false,
		},
		{
			name:         "memory mode",
			config:       ExportFlagConfig{MemoryMode: true},
			expectedMode: export.MemoryExportMode,
			expectError:  false,
		},
		{
			name:         "oidc mode",
			config:       ExportFlagConfig{OidcMode: true},
			expectedMode: export.OidcMode,
			expectError:  false,
		},
		{
			name:        "multiple modes",
			config:      ExportFlagConfig{TemplateMode: true, MemoryMode: true},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := tt.config.GetActiveMode()
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if mode != tt.expectedMode {
					t.Errorf("Expected mode %s but got %s", tt.expectedMode, mode)
				}
			}
		})
	}
}

func TestToExportParams(t *testing.T) {
	tests := []struct {
		name             string
		config           ExportFlagConfig
		args             []string
		expectedMode     export.ExportMode
		expectedFilePath string
		expectError      bool
	}{
		{
			name:             "template mode with no args",
			config:           ExportFlagConfig{TemplateMode: true},
			args:             []string{},
			expectedMode:     export.TemplateMode,
			expectedFilePath: export.DefaultTemplateFile,
			expectError:      false,
		},
		{
			name:             "template mode with custom file",
			config:           ExportFlagConfig{TemplateMode: true},
			args:             []string{"custom-template.json"},
			expectedMode:     export.TemplateMode,
			expectedFilePath: "custom-template.json",
			expectError:      false,
		},
		{
			name:             "memory mode with no args",
			config:           ExportFlagConfig{MemoryMode: true},
			args:             []string{},
			expectedMode:     export.MemoryExportMode,
			expectedFilePath: export.DefaultMemoryFile,
			expectError:      false,
		},
		{
			name:             "oidc mode with custom file",
			config:           ExportFlagConfig{OidcMode: true},
			args:             []string{"my-oidc-config.json"},
			expectedMode:     export.OidcMode,
			expectedFilePath: "my-oidc-config.json",
			expectError:      false,
		},
		{
			name:        "no mode set",
			config:      ExportFlagConfig{},
			args:        []string{},
			expectError: true,
		},
		{
			name:        "multiple modes set",
			config:      ExportFlagConfig{TemplateMode: true, MemoryMode: true},
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, filePath, err := tt.config.ToExportParams(tt.args)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if mode != tt.expectedMode {
					t.Errorf("Expected mode %s but got %s", tt.expectedMode, mode)
				}
				if filePath != tt.expectedFilePath {
					t.Errorf("Expected file path %s but got %s", tt.expectedFilePath, filePath)
				}
			}
		})
	}
}

func TestExportReconstructFlags(t *testing.T) {
	tests := []struct {
		name          string
		config        ExportFlagConfig
		expectedFlags []string
	}{
		{
			name:          "default config (no flags)",
			config:        ExportFlagConfig{},
			expectedFlags: []string{},
		},
		{
			name:          "template mode",
			config:        ExportFlagConfig{TemplateMode: true},
			expectedFlags: []string{"-t"}, // should use short flag
		},
		{
			name:          "memory mode",
			config:        ExportFlagConfig{MemoryMode: true},
			expectedFlags: []string{"-m"}, // should use short flag
		},
		{
			name:          "oidc mode",
			config:        ExportFlagConfig{OidcMode: true},
			expectedFlags: []string{"--oidc-config"}, // no short flag available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := tt.config.ReconstructFlags()
			if len(flags) != len(tt.expectedFlags) {
				t.Errorf("Expected %d flags, got %d: %v", len(tt.expectedFlags), len(flags), flags)
				return
			}

			for i, expected := range tt.expectedFlags {
				if i >= len(flags) || flags[i] != expected {
					t.Errorf("Expected flag %d to be %s, got %s", i, expected, flags[i])
				}
			}
		})
	}
}

func TestExportFlagDefinitionsConsistency(t *testing.T) {
	// Test that flag definitions are consistent
	if len(exportFlagDefinitions) == 0 {
		t.Fatal("exportFlagDefinitions is empty")
	}

	// Check that all flag definitions have required fields
	for i, flagDef := range exportFlagDefinitions {
		if flagDef.Name == "" {
			t.Errorf("Flag definition %d has empty Name", i)
		}
		if flagDef.BindFunc == nil {
			t.Errorf("Flag definition %d has nil BindFunc", i)
		}
		if flagDef.Description == "" {
			t.Errorf("Flag definition %d has empty Description", i)
		}
		if flagDef.DefaultVal == nil {
			t.Errorf("Flag definition %d has nil DefaultVal", i)
		}

		// All export flags should be boolean
		if flagDef.FlagType != FlagTypeBool {
			t.Errorf("Flag definition %d (%s) should be FlagTypeBool, got %v", i, flagDef.Name, flagDef.FlagType)
		}

		// Validate FlagType matches DefaultVal type
		if _, ok := flagDef.DefaultVal.(bool); !ok {
			t.Errorf("Flag definition %d (%s) has FlagTypeBool but DefaultVal is not bool: %T", i, flagDef.Name, flagDef.DefaultVal)
		}
	}

	// Test that NewExportFlagConfig uses flag definitions for defaults
	config := NewExportFlagConfig()
	for _, flagDef := range exportFlagDefinitions {
		fieldPtr := flagDef.BindFunc(config)

		currentVal := *fieldPtr.(*bool)
		expectedVal := flagDef.DefaultVal.(bool)
		if currentVal != expectedVal {
			t.Errorf("Bool flag %s default mismatch: config=%t, definition=%t", flagDef.Name, currentVal, expectedVal)
		}
	}
}

func TestExportReconstructFlagsShortVsLong(t *testing.T) {
	config := ExportFlagConfig{
		TemplateMode: true, // has short name -t
		OidcMode:     true, // no short name, should use --oidc-config
	}

	flags := config.ReconstructFlags()

	// Verify short and long flags are used appropriately
	expectedFlags := map[string]bool{
		"-t":            false,
		"--oidc-config": false,
	}

	for _, flag := range flags {
		if _, exists := expectedFlags[flag]; exists {
			expectedFlags[flag] = true
		}
	}

	// Check all expected flags were found
	for flag, found := range expectedFlags {
		if !found {
			t.Errorf("Expected flag %s not found in: %v", flag, flags)
		}
	}

	// Verify no long versions of flags with short names are present
	unwantedLongFlags := []string{"--template", "--memory"}
	for _, flag := range flags {
		for _, unwanted := range unwantedLongFlags {
			if flag == unwanted {
				t.Errorf("Found unwanted long flag %s when short version should be used", unwanted)
			}
		}
	}
}
