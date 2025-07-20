package interactive

import (
	"errors"
	"fmt"
	exportHandler "github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/flagmanager"
	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/charmbracelet/huh"
	"log"
	"strconv"
	"strings"
)

func Start() {
	for {
		selectedCommand := commandSelector()

		switch selectedCommand {
		case serve:
			config := serveForm()

			if err := server.Run(config); err != nil {
				log.Fatal("Failed to start server:", err)
			}

			flagConfig := flagmanager.NewServeFlagConfig()
			flagConfig.FromServerConfig(config)
			flags := flagConfig.ReconstructFlags()
			displayOneLiner("serve", flags, "")
		case stop:
			err := server.Stop()
			if err != nil {
				log.Fatal("Failed to stop server:", err)
			}
			log.Println("Server stop request sent")
		case export:
			exportConfigs := exportForm()
			for _, config := range exportConfigs {
				if err := exportHandler.ExportWithMode(config.Mode, config.FilePath); err != nil {
					log.Printf("Failed to export %s: %v", config.Mode, err)
				}
			}

			for _, config := range exportConfigs {
				flagConfig := flagmanager.NewExportFlagConfig()
				flagConfig.FromExportMode(config.Mode)
				flags := flagConfig.ReconstructFlags()
				displayOneLiner("export", flags, config.FilePath)
			}
		case exit:
			return
		default:
			panic("unknown command")
		}
	}
}

type command int

const (
	_ command = iota
	serve
	stop
	export
	exit
)

type ExportConfig struct {
	Mode     exportHandler.ExportMode
	FilePath string
}

func commandSelector() command {
	var selectedCommand command
	selector := huh.NewSelect[command]().
		Title("Select a command").
		Options(
			huh.NewOption("Start Server", serve),
			huh.NewOption("Stop Server", stop),
			huh.NewOption("Export JSON Data/Template", export),
			huh.NewOption("Exit", exit),
		).
		Value(&selectedCommand)

	if err := selector.Run(); err != nil {
		panic("failed to select command: " + err.Error())
	}

	return selectedCommand
}

func serveForm() *server.Config {
	config := server.NewServerConfig()

	portString := "8080" // Default port
	portInput := huh.NewInput().
		Title("Port to run the server on").
		Prompt("Enter port:").
		Validate(
			func(s string) error {
				if s == "" {
					return errors.New("port cannot be empty")
				}
				// Validate that the input is a valid port number
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
		panic(fmt.Sprintf("invalid port number: %s", portString))
	}
	config.Port = port

	var disabledAuth bool
	authDisableConfirm := huh.NewConfirm().
		Title("Do you want to disable authentication for task API endpoints?").
		Affirmative("Yes").
		Negative("No").
		Value(&disabledAuth)
	if err := authDisableConfirm.Run(); err != nil {
		log.Fatal("Failed to confirm authentication disable:", err)
	}

	config.AuthRequired = !disabledAuth

	jsonFilePathInput := huh.NewInput().
		Title("JSON File Path").
		Description("Enter the path to the JSON file as a data storage source. (leave empty for using memory state)").
		Prompt("Enter JSON file path:").
		Placeholder("data.json").
		Value(&config.JsonFilePath)
	if err := jsonFilePathInput.Run(); err != nil {
		log.Fatal("Failed to get JSON file path input:", err)
	}

	jwtModeSelector := huh.NewSelect[auth.JWTKeyMode]().
		Title("Select JWT Key Mode").
		Options(
			huh.NewOption("Secret Key", auth.JWTKeyModeSecret),
			huh.NewOption("RSA Key", auth.JWTKeyModeRSA),
		).
		Value(&config.JWTKeyMode)
	if err := jwtModeSelector.Run(); err != nil {
		log.Fatal("Failed to select JWT key mode:", err)
	}

	if config.JWTKeyMode == auth.JWTKeyModeSecret {
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
			Value(&config.JWTSecretKey)
		if err := jwtSecretInput.Run(); err != nil {
			log.Fatal("Failed to get JWT secret key input:", err)
		}
	}

	authModeSelector := huh.NewSelect[auth.AuthMode]().
		Title("Select Authentication Mode").
		Options(
			huh.NewOption("JWT", auth.AuthModeJWT),
			huh.NewOption("Session", auth.AuthModeSession),
			huh.NewOption("Both", auth.AuthModeBoth),
			huh.NewOption("OIDC", auth.AuthModeOIDC),
		).
		Value(&config.AuthMode)
	if err := authModeSelector.Run(); err != nil {
		log.Fatal("Failed to select authentication mode:", err)
	}

	if config.AuthMode == auth.AuthModeOIDC {
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
			Value(&config.OIDCConfigPath)
		if err := oidcConfigPathInput.Run(); err != nil {
			log.Fatal("Failed to get OIDC configuration path input:", err)
		}
	}

	return config
}

func exportForm() []ExportConfig {
	var configs []ExportConfig

	var selectedModes []exportHandler.ExportMode
	modeSelector := huh.NewMultiSelect[exportHandler.ExportMode]().
		Title("Select export modes").
		Options(
			huh.NewOption("JSON Data Template", exportHandler.TemplateMode),
			huh.NewOption("Memory State", exportHandler.MemoryExportMode),
			huh.NewOption("OIDC Configuration Template", exportHandler.OidcMode),
		).
		Value(&selectedModes)
	if err := modeSelector.Run(); err != nil {
		log.Fatal("Failed to select export modes:", err)
	}

	for _, mode := range selectedModes {
		filePath := ""

		filePathInput := huh.NewInput().
			TitleFunc(
				func() string {
					switch mode {
					case exportHandler.TemplateMode:
						return "JSON Data Template File Path"
					case exportHandler.MemoryExportMode:
						return "Memory State File Path"
					case exportHandler.OidcMode:
						return "OIDC Configuration Template File Path"
					default:
						log.Fatal("Unknown mode:", mode)
						return ""
					}
				}, &mode).
			Description("leave empty for default file path").
			Prompt("Enter file path:").
			PlaceholderFunc(func() string {
				switch mode {
				case exportHandler.TemplateMode:
					return exportHandler.DefaultTemplateFile
				case exportHandler.MemoryExportMode:
					return exportHandler.DefaultMemoryFile
				case exportHandler.OidcMode:
					return exportHandler.DefaultOidcFile
				default:
					log.Fatal("Unknown mode:", mode)
					return ""
				}
			}, &mode).
			Value(&filePath)
		if err := filePathInput.Run(); err != nil {
			log.Fatal("Failed to get file path input:", err)
		}

		configs = append(configs, ExportConfig{
			Mode:     mode,
			FilePath: filePath,
		})
	}

	return configs
}

// displayOneLiner prints the command-line equivalent of the configuration
func displayOneLiner(command string, flags []string, filePath string) {
	fmt.Printf("\n--- Command line equivalent ---\n")

	var cmdParts []string
	cmdParts = append(cmdParts, "./mock-todo-server", command)
	cmdParts = append(cmdParts, flags...)

	if filePath != "" {
		cmdParts = append(cmdParts, filePath)
	}

	fmt.Printf("%s\n\n", strings.Join(cmdParts, " "))
}
