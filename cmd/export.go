/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/flagmanager"
	"github.com/spf13/cobra"
)

var (
	exportFlagConfig *flagmanager.ExportFlagConfig
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [file-path]",
	Short: "Export JSON template, OIDC config template, or current memory store state",
	Long: `Export a JSON template file for use with the server's -f flag, OIDC configuration template, or export the current memory store state.

Examples:
  # Export data template to current directory as data.json
  mock-todo-server export --template

  # Export data template to specific file
  mock-todo-server export --template /path/to/template.json

  # Export OIDC configuration template
  mock-todo-server export --oidc-config

  # Export OIDC config template to specific file
  mock-todo-server export --oidc-config /path/to/oidc-config.json

  # Export current memory store state
  mock-todo-server export --memory

  # Export memory state to specific file
  mock-todo-server export --memory backup.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Validate export flags
		if err := exportFlagConfig.Validate(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Get export parameters
		mode, filePath, err := exportFlagConfig.ToExportParams(args)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Execute export
		if err := export.ExportWithMode(mode, filePath); err != nil {
			fmt.Printf("Error exporting data: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Initialize export flag config
	exportFlagConfig = flagmanager.NewExportFlagConfig()

	// Register flags
	exportFlagConfig.RegisterFlags(exportCmd)
}
