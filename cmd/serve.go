/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/KasumiMercury/mock-todo-server/flagmanager"
	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

var flagConfig = flagmanager.NewServeFlagConfig()

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mock TODO server",
	Long:  `Start the mock TODO server that provides REST API endpoints for managing tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Convert flag config to server config
		config, err := flagConfig.ToServerConfig()
		if err != nil {
			log.Fatal("Invalid configuration:", err)
		}

		// Run standard validation
		if err := config.Validate(); err != nil {
			log.Fatal("Invalid configuration:", err)
		}

		if err := server.Run(config); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Register flags using simplified flag manager
	flagConfig.RegisterFlags(serveCmd)
}
