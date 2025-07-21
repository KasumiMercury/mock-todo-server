/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data templates, memory state, or OIDC configuration",
	Long: `Export various types of data and configuration templates.

Available subcommands:
  store   - Export JSON data template for use with server's -f flag
  memory  - Export current memory store state
  oidc    - Export OIDC configuration template

Examples:
  # Export data template to current directory as data.json
  mock-todo-server export store

  # Export data template to specific file
  mock-todo-server export store /path/to/template.json

  # Export OIDC configuration template
  mock-todo-server export oidc

  # Export OIDC config template to specific file
  mock-todo-server export oidc /path/to/oidc-config.json

  # Export current memory store state
  mock-todo-server export memory

  # Export memory state to specific file
  mock-todo-server export memory backup.json`,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
