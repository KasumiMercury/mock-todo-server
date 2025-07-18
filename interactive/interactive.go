package interactive

import (
	"errors"
	"fmt"
	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/charmbracelet/huh"
	"log"
	"strconv"
)

func Start() {
	//var service InteractiveService

	for {
		selectedCommand := commandSelector()

		switch selectedCommand {
		case serve:
			config := serveForm()
			if err := server.Run(config); err != nil {
				log.Fatal("Failed to start server:", err)
			}
		case stop:
			//service.stop()
		case export:
			//service.export()
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
