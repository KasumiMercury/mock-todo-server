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

// memoryCmd represents the memory command
var memoryCmd = &cobra.Command{
	Use:   "memory [file-path]",
	Short: "Export current memory store state",
	Long: `Export the current memory store state including all tasks and users.
This captures the live data from a running server instance, or empty data if the server is stopped.

The export includes:
- All current tasks in memory
- All current users with their hashed passwords
- Timestamps and other metadata

Examples:
  # Export memory state to current directory as memory-state.json
  mock-todo-server export memory

  # Export memory state to specific file
  mock-todo-server export memory backup.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := export.GetOutputPath(args, export.DefaultMemoryFile)

		if err := export.ExportWithMode(export.MemoryExportMode, filePath); err != nil {
			fmt.Printf("Error exporting memory state: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	exportCmd.AddCommand(memoryCmd)
}
