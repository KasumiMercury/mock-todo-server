package interactive

import (
	"errors"
	"log"
	"strconv"

	exportHandler "github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/charmbracelet/huh"
)

func serveForm() *server.Config {
	config := server.NewServerConfig()

	config.Port = createPortInput()
	config.AuthRequired = createAuthRequiredConfirm()
	config.JsonFilePath = createJSONFilePathInput()
	config.JWTKeyMode = createJWTModeSelector()

	if config.JWTKeyMode == auth.JWTKeyModeSecret {
		config.JWTSecretKey = createJWTSecretInput()
	}

	config.AuthMode = createAuthModeSelector()

	if config.AuthMode == auth.AuthModeOIDC {
		config.OIDCConfigPath = createOIDCConfigInput()
	}

	return config
}

func createPortInput() int {
	portString := "8080"
	portInput := huh.NewInput().
		Title("Port to run the server on").
		Prompt("Enter port:").
		Validate(
			func(s string) error {
				if s == "" {
					return errors.New("port cannot be empty")
				}
				port, err := strconv.Atoi(s)
				if err != nil {
					return errors.New("port must be an integer")
				}
				if port < 1 || port > 65535 {
					return errors.New("port must be between 1 and 65535")
				}
				return nil
			},
		).
		Value(&portString)

	if err := portInput.Run(); err != nil {
		log.Fatal("Failed to get port input:", err)
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatal("Invalid port number:", portString)
	}

	return port
}

func createAuthRequiredConfirm() bool {
	var authRequired bool
	authDisableConfirm := huh.NewConfirm().
		Title("Enable Authentication for the tasks endpoint").
		Affirmative("Yes").
		Negative("No").
		Value(&authRequired)

	if err := authDisableConfirm.Run(); err != nil {
		log.Fatal("Failed to confirm authentication disable:", err)
	}

	return authRequired
}

func createJSONFilePathInput() string {
	var jsonFilePath string
	jsonFilePathInput := huh.NewInput().
		Title("JSON File Path").
		Description("Enter the path to the JSON file as a data storage source. (leave empty for using memory state)").
		Prompt("Enter JSON file path:").
		Placeholder("data.json").
		Value(&jsonFilePath)

	if err := jsonFilePathInput.Run(); err != nil {
		log.Fatal("Failed to get JSON file path input:", err)
	}

	return jsonFilePath
}

func createJWTModeSelector() auth.JWTKeyMode {
	var jwtKeyMode auth.JWTKeyMode
	jwtModeSelector := huh.NewSelect[auth.JWTKeyMode]().
		Title("Select JWT Key Mode").
		Options(
			huh.NewOption("Secret Key", auth.JWTKeyModeSecret),
			huh.NewOption("RSA Key", auth.JWTKeyModeRSA),
		).
		Value(&jwtKeyMode)

	if err := jwtModeSelector.Run(); err != nil {
		log.Fatal("Failed to select JWT key mode:", err)
	}

	return jwtKeyMode
}

func createJWTSecretInput() string {
	var jwtSecretKey string
	jwtSecretInput := huh.NewInput().
		Title("JWT Secret Key").
		Prompt("Enter JWT secret key:").
		Validate(
			func(s string) error {
				if s == "" {
					return errors.New("JWT secret key cannot be empty")
				}
				return nil
			},
		).
		Value(&jwtSecretKey)

	if err := jwtSecretInput.Run(); err != nil {
		log.Fatal("Failed to get JWT secret key input:", err)
	}

	return jwtSecretKey
}

func createAuthModeSelector() auth.AuthMode {
	var authMode auth.AuthMode
	authModeSelector := huh.NewSelect[auth.AuthMode]().
		Title("Select Authentication Mode").
		Options(
			huh.NewOption("JWT", auth.AuthModeJWT),
			huh.NewOption("Session", auth.AuthModeSession),
			huh.NewOption("Both", auth.AuthModeBoth),
			huh.NewOption("OIDC", auth.AuthModeOIDC),
		).
		Value(&authMode)

	if err := authModeSelector.Run(); err != nil {
		log.Fatal("Failed to select authentication mode:", err)
	}

	return authMode
}

func createOIDCConfigInput() string {
	var oidcConfigPath string
	oidcConfigPathInput := huh.NewInput().
		Title("OIDC Configuration Path").
		Prompt("Enter path to OIDC configuration JSON file:").
		Validate(
			func(s string) error {
				if s == "" {
					return errors.New("OIDC configuration path cannot be empty")
				}
				return nil
			},
		).
		Value(&oidcConfigPath)

	if err := oidcConfigPathInput.Run(); err != nil {
		log.Fatal("Failed to get OIDC configuration path input:", err)
	}

	return oidcConfigPath
}

func exportForm() ExportConfig {
	var selectedMode exportHandler.ExportMode
	modeSelector := huh.NewSelect[exportHandler.ExportMode]().
		Title("Select export modes").
		Options(
			huh.NewOption("JSON Data Template", exportHandler.StoreMode),
			huh.NewOption("Memory State", exportHandler.MemoryExportMode),
			huh.NewOption("OIDC Configuration Template", exportHandler.OidcMode),
		).
		Value(&selectedMode)

	if err := modeSelector.Run(); err != nil {
		log.Fatal("Failed to select export modes:", err)
	}

	filePath := ""
	filePathInput := huh.NewInput().
		TitleFunc(
			func() string {
				switch selectedMode {
				case exportHandler.StoreMode:
					return "JSON Data Template File Path"
				case exportHandler.MemoryExportMode:
					return "Memory State File Path"
				case exportHandler.OidcMode:
					return "OIDC Configuration Template File Path"
				default:
					log.Fatal("Unknown mode:", selectedMode)
					return ""
				}
			}, &selectedMode).
		Description("leave empty for default file path").
		Prompt("Enter file path:").
		PlaceholderFunc(func() string {
			switch selectedMode {
			case exportHandler.StoreMode:
				return exportHandler.DefaultStoreFile
			case exportHandler.MemoryExportMode:
				return exportHandler.DefaultMemoryFile
			case exportHandler.OidcMode:
				return exportHandler.DefaultOidcFile
			default:
				log.Fatal("Unknown mode:", selectedMode)
				return ""
			}
		}, &selectedMode).
		Value(&filePath)

	if err := filePathInput.Run(); err != nil {
		log.Fatal("Failed to get file path input:", err)
	}

	return ExportConfig{
		Mode:     selectedMode,
		FilePath: filePath,
	}
}
