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

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mock TODO server",
	Long:  `Start the mock TODO server that provides REST API endpoints for managing tasks.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Run(port, jsonFilePath); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().StringVarP(&jsonFilePath, "jsonFilePath", "f", "", "File as a Data source for test")
}
