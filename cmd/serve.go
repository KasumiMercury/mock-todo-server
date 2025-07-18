/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

var port int
var jsonFilePath string
var jwtKeyMode string
var jwtSecretKey string
var authRequired bool
var authMode string
var oidcConfigPath string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mock TODO server",
	Long:  `Start the mock TODO server that provides REST API endpoints for managing tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Run(port, jsonFilePath, jwtKeyMode, jwtSecretKey, authRequired, authMode, oidcConfigPath); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringVarP(&jsonFilePath, "json-file-Path", "f", "", "File as a Data source for test")
	serveCmd.Flags().StringVar(&jwtKeyMode, "jwt-key-mode", "secret", "JWT key mode: 'secret' or 'rsa'")
	serveCmd.Flags().StringVar(&jwtSecretKey, "jwt-secret", "test-secret-key", "JWT secret key (used when jwt-key-mode is 'secret')")
	serveCmd.Flags().BoolVarP(&authRequired, "auth-required", "a", true, "Require authentication for task API endpoints")
	serveCmd.Flags().StringVar(&authMode, "auth-mode", "jwt", "Authentication mode: 'jwt', 'session', 'both', or 'oidc'")
	serveCmd.Flags().StringVar(&oidcConfigPath, "oidc-config-path", "", "Path to OIDC configuration JSON file (required when auth-mode is 'oidc')")
}
