/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/spf13/cobra"
	"os"
)

var (
	templateMode bool
	memoryMode   bool
	oidcMode     bool
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
		// Count active flags
		activeFlags := 0
		if templateMode {
			activeFlags++
		}
		if memoryMode {
			activeFlags++
		}
		if oidcMode {
			activeFlags++
		}

		if activeFlags == 0 {
			fmt.Println("Error: Must specify one of --template, --memory, or --oidc-config flag")
			os.Exit(1)
		}

		if activeFlags > 1 {
			fmt.Println("Error: Cannot specify multiple export flags simultaneously")
			os.Exit(1)
		}

		if err := export.Export(
			args,
			templateMode,
			memoryMode,
			oidcMode,
		); err != nil {
			fmt.Printf("Error exporting data: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&templateMode, "template", "t", false, "Export JSON data template file")
	exportCmd.Flags().BoolVarP(&memoryMode, "memory", "m", false, "Export current memory store state")
	exportCmd.Flags().BoolVar(&oidcMode, "oidc-config", false, "Export OIDC configuration template")
}
