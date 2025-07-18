/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

var (
	config        = server.NewServerConfig()
	jwtKeyModeStr string
	authModeStr   string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mock TODO server",
	Long:  `Start the mock TODO server that provides REST API endpoints for managing tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ValidateAndConvert(jwtKeyModeStr, authModeStr); err != nil {
			log.Fatal("Invalid configuration:", err)
		}
		if err := server.Run(config); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&config.Port, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringVarP(&config.JsonFilePath, "json-file-Path", "f", "", "File as a Data source for test")
	serveCmd.Flags().StringVar(&jwtKeyModeStr, "jwt-key-mode", "secret", "JWT key mode: 'secret' or 'rsa'")
	serveCmd.Flags().StringVar(&config.JWTSecretKey, "jwt-secret", "test-secret-key", "JWT secret key (used when jwt-key-mode is 'secret')")
	serveCmd.Flags().BoolVarP(&config.AuthRequired, "auth-required", "a", true, "Require authentication for task API endpoints")
	serveCmd.Flags().StringVar(&authModeStr, "auth-mode", "jwt", "Authentication mode: 'jwt', 'session', 'both', or 'oidc'")
	serveCmd.Flags().StringVar(&config.OIDCConfigPath, "oidc-config-path", "", "Path to OIDC configuration JSON file (required when auth-mode is 'oidc')")
}
