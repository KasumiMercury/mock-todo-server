/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/KasumiMercury/mock-todo-server/server"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the mock TODO server",
	Long:  `Stop the running mock TODO server gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.Stop(); err != nil {
			log.Fatal("Failed to stop server:", err)
		}
		log.Println("Server stop request sent")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
