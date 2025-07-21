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

// oidcCmd represents the oidc command
var oidcCmd = &cobra.Command{
	Use:   "oidc [file-path]",
	Short: "Export OIDC configuration template",
	Long: `Export an OIDC configuration template file for testing OpenID Connect integrations.
This template contains sample client credentials and configuration for use with the server's OIDC mode.

The template includes:
- Sample client ID and secret
- Common redirect URIs for local development
- Issuer URL pointing to localhost
- Standard OpenID scopes (openid, profile)

Examples:
  # Export OIDC config to current directory as oidc-config.json
  mock-todo-server export oidc

  # Export OIDC config to specific file
  mock-todo-server export oidc /path/to/oidc-config.json`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := export.GetOutputPath(args, export.DefaultOidcFile)

		if err := export.ExportWithMode(export.OidcMode, filePath); err != nil {
			fmt.Printf("Error exporting OIDC configuration: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	exportCmd.AddCommand(oidcCmd)
}
