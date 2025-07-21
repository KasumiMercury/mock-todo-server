/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/spf13/cobra"
)

// storeCmd represents the store command
var storeCmd = &cobra.Command{
	Use:   "store [file-path]",
	Short: "Export JSON data template for use with server's -f flag",
	Long: `Export a JSON data template file containing sample tasks and users.
This template can be used with the server's -f flag to populate initial data.

The template includes:
- 2 sample tasks
- 2 sample users with hashed passwords (user1/password1, user2/password2)

Examples:
  # Export template to current directory as data.json
  mock-todo-server export store

  # Export template to specific file
  mock-todo-server export store /path/to/template.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := export.GetOutputPath(args, export.DefaultStoreFile)

		if err := export.ExportWithMode(export.StoreMode, filePath); err != nil {
			fmt.Printf("Error exporting store template: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	exportCmd.AddCommand(storeCmd)
}
