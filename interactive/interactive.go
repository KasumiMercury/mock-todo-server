package interactive

import (
	"errors"
	"fmt"
	exportHandler "github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/charmbracelet/huh"
	"log"
	"strconv"
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
		case stop:
			err := server.Stop()
			if err != nil {
				log.Fatal("Failed to stop server:", err)
			}
			log.Println("Server stop request sent")
		case export:
			// TODO: fix multiple export modes
			// TODO: fix export to file path

			templateMode, memoryMode, oidcMode := exportForm()
			if templateMode {
				err := exportHandler.Template("data-template.json")
				if err != nil {
					log.Fatal("Failed to export data template:", err)
				} else {
					log.Println("Data template exported successfully to data-template.json")
				}
			}

			if memoryMode {
				err := exportHandler.MemoryState("memory-state.json")
				if err != nil {
					log.Fatal("Failed to export memory state:", err)
				} else {
					log.Println("Memory state exported successfully to memory-state.json")
				}
			}

			if oidcMode {
				err := exportHandler.OidcTemplate("oidc-config.json")
				if err != nil {
					log.Fatal("Failed to export OIDC configuration template:", err)
				} else {
					log.Println("OIDC configuration template exported successfully to oidc-config.json")
				}
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

func exportForm() (bool, bool, bool) {
	var templateMode, memoryMode, oidcMode bool

	templateConfirm := huh.NewConfirm().
		Title("Export JSON template?").
		Affirmative("Yes").
		Negative("No").
		Value(&templateMode)
	if err := templateConfirm.Run(); err != nil {
		log.Fatal("Failed to confirm template export:", err)
	}

	memoryConfirm := huh.NewConfirm().
		Title("Export current memory store state?").
		Affirmative("Yes").
		Negative("No").
		Value(&memoryMode)
	if err := memoryConfirm.Run(); err != nil {
		log.Fatal("Failed to confirm memory export:", err)
	}

	oidcConfirm := huh.NewConfirm().
		Title("Export OIDC configuration template?").
		Affirmative("Yes").
		Negative("No").
		Value(&oidcMode)
	if err := oidcConfirm.Run(); err != nil {
		log.Fatal("Failed to confirm OIDC export:", err)
	}

	return templateMode, memoryMode, oidcMode
}
