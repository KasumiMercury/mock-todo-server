package flagmanager

import (
	"testing"

	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

func TestNewServeFlagConfig(t *testing.T) {
	config := NewServeFlagConfig()
	if config == nil {
		t.Fatal("NewServeFlagConfig returned nil")
	}

	// Check default values
	if config.Port != 8080 {
		t.Errorf("Expected default port to be 8080, got %d", config.Port)
	}
	if config.JWTKeyModeStr != "secret" {
		t.Errorf("Expected default JWTKeyModeStr to be 'secret', got %s", config.JWTKeyModeStr)
	}
	if config.AuthModeStr != "jwt" {
		t.Errorf("Expected default AuthModeStr to be 'jwt', got %s", config.AuthModeStr)
	}
	if !config.AuthRequired {
		t.Error("Expected default AuthRequired to be true")
	}
}

func TestRegisterFlags(t *testing.T) {
	config := NewServeFlagConfig()
	cmd := &cobra.Command{Use: "test"}

	config.RegisterFlags(cmd)

	// Check that expected flags were registered
	expectedFlags := []string{"port", "json-file-path", "jwt-key-mode", "jwt-secret", "auth-required", "auth-mode", "oidc-config-path"}
	for _, flagName := range expectedFlags {
		if flag := cmd.Flags().Lookup(flagName); flag == nil {
			t.Errorf("Expected flag %s was not registered", flagName)
		}
	}

	// Check that short flags are properly registered
	expectedShortFlags := map[string]string{
		"p": "port",
		"f": "json-file-path",
		"a": "auth-required",
	}

	for shortFlag, longFlag := range expectedShortFlags {
		if flag := cmd.Flags().ShorthandLookup(shortFlag); flag == nil {
			t.Errorf("Expected short flag -%s for %s was not registered", shortFlag, longFlag)
		} else if flag.Name != longFlag {
			t.Errorf("Short flag -%s points to %s, expected %s", shortFlag, flag.Name, longFlag)
		}
	}
}

func TestToServerConfig(t *testing.T) {
	flagConfig := NewServeFlagConfig()
	flagConfig.Port = 9090
	flagConfig.AuthRequired = false
	flagConfig.JWTSecretKey = "test-secret"
	flagConfig.JWTKeyModeStr = "rsa"
	flagConfig.AuthModeStr = "session"

	serverConfig, err := flagConfig.ToServerConfig()
	if err != nil {
		t.Fatalf("ToServerConfig failed: %v", err)
	}

	if serverConfig.Port != 9090 {
		t.Errorf("Expected port to be 9090, got %d", serverConfig.Port)
	}
	if serverConfig.AuthRequired != false {
		t.Errorf("Expected auth-required to be false, got %t", serverConfig.AuthRequired)
	}
	if serverConfig.JWTSecretKey != "test-secret" {
		t.Errorf("Expected jwt-secret to be 'test-secret', got %s", serverConfig.JWTSecretKey)
	}
}

func TestFromServerConfig(t *testing.T) {
	serverConfig := server.NewServerConfig()
	serverConfig.Port = 8888
	serverConfig.AuthRequired = false
	serverConfig.JWTSecretKey = "custom-secret"

	flagConfig := NewServeFlagConfig()
	flagConfig.FromServerConfig(serverConfig)

	if flagConfig.Port != 8888 {
		t.Errorf("Expected port to be 8888, got %d", flagConfig.Port)
	}
	if flagConfig.AuthRequired != false {
		t.Errorf("Expected auth-required to be false, got %t", flagConfig.AuthRequired)
	}
	if flagConfig.JWTSecretKey != "custom-secret" {
		t.Errorf("Expected jwt-secret to be 'custom-secret', got %s", flagConfig.JWTSecretKey)
	}
}

func TestReconstructFlags(t *testing.T) {
	flagConfig := NewServeFlagConfig()
	flagConfig.Port = 9090
	flagConfig.AuthRequired = false
	flagConfig.JWTKeyModeStr = "rsa"
	flagConfig.AuthModeStr = "session"
	flagConfig.JsonFilePath = "test.json"

	flags := flagConfig.ReconstructFlags()

	// Check that non-default values are included in flags with short names preferred
	expectedContains := map[string]bool{
		"-p":             false, // port has short name
		"9090":           false,
		"-a=false":       false, // auth-required has short name
		"--jwt-key-mode": false, // no short name
		"rsa":            false,
		"--auth-mode":    false, // no short name
		"session":        false,
		"-f":             false, // json-file-path has short name
		"test.json":      false,
	}

	for _, flag := range flags {
		if _, exists := expectedContains[flag]; exists {
			expectedContains[flag] = true
		}
	}

	for expected, found := range expectedContains {
		if !found {
			t.Errorf("Expected flag %s not found in reconstructed flags: %v", expected, flags)
		}
	}
}

func TestBidirectionalConversion(t *testing.T) {
	// Create original flag config
	original := NewServeFlagConfig()
	original.Port = 9999
	original.AuthRequired = false
	original.JWTSecretKey = "bidirectional-test"
	original.JWTKeyModeStr = "rsa"
	original.AuthModeStr = "both"

	// Convert to server config
	serverConfig, err := original.ToServerConfig()
	if err != nil {
		t.Fatalf("ToServerConfig failed: %v", err)
	}

	// Convert back to flag config
	reconstructed := NewServeFlagConfig()
	reconstructed.FromServerConfig(serverConfig)

	// Compare
	if reconstructed.Port != original.Port {
		t.Errorf("Port mismatch: original=%d, reconstructed=%d", original.Port, reconstructed.Port)
	}
	if reconstructed.AuthRequired != original.AuthRequired {
		t.Errorf("AuthRequired mismatch: original=%t, reconstructed=%t", original.AuthRequired, reconstructed.AuthRequired)
	}
	if reconstructed.JWTSecretKey != original.JWTSecretKey {
		t.Errorf("JWTSecretKey mismatch: original=%s, reconstructed=%s", original.JWTSecretKey, reconstructed.JWTSecretKey)
	}
	if reconstructed.JWTKeyModeStr != original.JWTKeyModeStr {
		t.Errorf("JWTKeyModeStr mismatch: original=%s, reconstructed=%s", original.JWTKeyModeStr, reconstructed.JWTKeyModeStr)
	}
	if reconstructed.AuthModeStr != original.AuthModeStr {
		t.Errorf("AuthModeStr mismatch: original=%s, reconstructed=%s", original.AuthModeStr, reconstructed.AuthModeStr)
	}
}

func TestFlagDefinitionsConsistency(t *testing.T) {
	// Test that flag definitions are consistent
	if len(flagDefinitions) == 0 {
		t.Fatal("flagDefinitions is empty")
	}

	// Check that all flag definitions have required fields
	for i, flagDef := range flagDefinitions {
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

		// Validate FlagType matches DefaultVal type
		switch flagDef.FlagType {
		case FlagTypeInt:
			if _, ok := flagDef.DefaultVal.(int); !ok {
				t.Errorf("Flag definition %d (%s) has FlagTypeInt but DefaultVal is not int: %T", i, flagDef.Name, flagDef.DefaultVal)
			}
		case FlagTypeString:
			if _, ok := flagDef.DefaultVal.(string); !ok {
				t.Errorf("Flag definition %d (%s) has FlagTypeString but DefaultVal is not string: %T", i, flagDef.Name, flagDef.DefaultVal)
			}
		case FlagTypeBool:
			if _, ok := flagDef.DefaultVal.(bool); !ok {
				t.Errorf("Flag definition %d (%s) has FlagTypeBool but DefaultVal is not bool: %T", i, flagDef.Name, flagDef.DefaultVal)
			}
		default:
			t.Errorf("Flag definition %d (%s) has unknown FlagType: %v", i, flagDef.Name, flagDef.FlagType)
		}
	}

	// Test that NewServeFlagConfig uses flag definitions for defaults
	config := NewServeFlagConfig()
	for _, flagDef := range flagDefinitions {
		fieldPtr := flagDef.BindFunc(config)

		switch flagDef.FlagType {
		case FlagTypeInt:
			currentVal := *fieldPtr.(*int)
			expectedVal := flagDef.DefaultVal.(int)
			if currentVal != expectedVal {
				t.Errorf("Int flag %s default mismatch: config=%d, definition=%d", flagDef.Name, currentVal, expectedVal)
			}
		case FlagTypeString:
			currentVal := *fieldPtr.(*string)
			expectedVal := flagDef.DefaultVal.(string)
			if currentVal != expectedVal {
				t.Errorf("String flag %s default mismatch: config=%s, definition=%s", flagDef.Name, currentVal, expectedVal)
			}
		case FlagTypeBool:
			currentVal := *fieldPtr.(*bool)
			expectedVal := flagDef.DefaultVal.(bool)
			if currentVal != expectedVal {
				t.Errorf("Bool flag %s default mismatch: config=%t, definition=%t", flagDef.Name, currentVal, expectedVal)
			}
		}
	}
}

func TestFlagTypeEnum(t *testing.T) {
	// Test that FlagType constants are properly defined
	if FlagTypeInt != 0 {
		t.Errorf("Expected FlagTypeInt to be 0, got %d", FlagTypeInt)
	}
	if FlagTypeString != 1 {
		t.Errorf("Expected FlagTypeString to be 1, got %d", FlagTypeString)
	}
	if FlagTypeBool != 2 {
		t.Errorf("Expected FlagTypeBool to be 2, got %d", FlagTypeBool)
	}
}

func TestReconstructFlagsConsistency(t *testing.T) {
	// Test that ReconstructFlags only includes non-default values
	config := NewServeFlagConfig()

	// Default config should produce no flags
	flags := config.ReconstructFlags()
	if len(flags) != 0 {
		t.Errorf("Default config should produce no flags, got: %v", flags)
	}

	// Change one value and verify flag appears (should use short flag if available)
	config.Port = 9090
	flags = config.ReconstructFlags()

	foundPortFlag := false
	foundPortValue := false
	for _, flag := range flags {
		if flag == "-p" { // Port should use short flag -p
			foundPortFlag = true
		}
		if flag == "9090" {
			foundPortValue = true
		}
	}

	if !foundPortFlag {
		t.Error("Expected -p (short) flag not found in reconstructed flags")
	}
	if !foundPortValue {
		t.Error("Expected port value '9090' not found in reconstructed flags")
	}
}

func TestReconstructFlagsShortVsLong(t *testing.T) {
	config := NewServeFlagConfig()

	// Test flags with short names use short format
	config.Port = 9090                // has short name -p
	config.JsonFilePath = "test.json" // has short name -f
	config.AuthRequired = false       // has short name -a

	// Test flags without short names use long format
	config.JWTKeyModeStr = "rsa"   // no short name
	config.AuthModeStr = "session" // no short name

	flags := config.ReconstructFlags()

	// Verify short flags are used where available
	shortFlagsFound := map[string]bool{
		"-p":       false,
		"-f":       false,
		"-a=false": false,
	}

	// Verify long flags are used where no short flag exists
	longFlagsFound := map[string]bool{
		"--jwt-key-mode": false,
		"--auth-mode":    false,
	}

	for _, flag := range flags {
		if _, exists := shortFlagsFound[flag]; exists {
			shortFlagsFound[flag] = true
		}
		if _, exists := longFlagsFound[flag]; exists {
			longFlagsFound[flag] = true
		}
	}

	// Check all expected short flags were found
	for flag, found := range shortFlagsFound {
		if !found {
			t.Errorf("Expected short flag %s not found in: %v", flag, flags)
		}
	}

	// Check all expected long flags were found
	for flag, found := range longFlagsFound {
		if !found {
			t.Errorf("Expected long flag %s not found in: %v", flag, flags)
		}
	}

	// Verify no long versions of flags with short names are present
	unwantedLongFlags := []string{"--port", "--json-file-path", "--auth-required"}
	for _, flag := range flags {
		for _, unwanted := range unwantedLongFlags {
			if flag == unwanted {
				t.Errorf("Found unwanted long flag %s when short version should be used", unwanted)
			}
		}
	}
}
