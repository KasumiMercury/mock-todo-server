/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/spf13/cobra"
)

var (
	templateMode bool
	memoryMode   bool
)

type FileData struct {
	Tasks []*domain.Task `json:"tasks"`
	Users []*domain.User `json:"users"`
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export [file-path]",
	Short: "Export JSON template or current memory store state",
	Long: `Export a JSON template file for use with the server's -f flag, or export the current memory store state.

Examples:
  # Export template to current directory as data.json
  mock-todo-server export --template

  # Export template to specific file
  mock-todo-server export --template /path/to/template.json

  # Export current memory store state
  mock-todo-server export --memory

  # Export memory state to specific file
  mock-todo-server export --memory backup.json`,
	Run: func(cmd *cobra.Command, args []string) {
		if !templateMode && !memoryMode {
			fmt.Println("Error: Must specify either --template or --memory flag")
			os.Exit(1)
		}

		if templateMode && memoryMode {
			fmt.Println("Error: Cannot specify both --template and --memory flags")
			os.Exit(1)
		}

		filePath := getOutputPath(args)

		if templateMode {
			if err := exportTemplate(filePath); err != nil {
				log.Fatalf("Failed to export template: %v", err)
			}
			fmt.Printf("Template exported to: %s\n", filePath)
		} else if memoryMode {
			if err := exportMemoryState(filePath); err != nil {
				log.Fatalf("Failed to export memory state: %v", err)
			}
			fmt.Printf("Memory state exported to: %s\n", filePath)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().BoolVarP(&templateMode, "template", "t", false, "Export JSON template file")
	exportCmd.Flags().BoolVarP(&memoryMode, "memory", "m", false, "Export current memory store state")
}

func getOutputPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}

	return "data.json"
}

func exportTemplate(filePath string) error {
	now := time.Now()

	sampleData := FileData{
		Tasks: []*domain.Task{
			{
				ID:        1,
				Title:     "Sample Task 1",
				UserID:    1,
				CreatedAt: now.Format(time.RFC3339),
			},
			{
				ID:        2,
				Title:     "Sample Task 2",
				UserID:    2,
				CreatedAt: now.Add(time.Minute).Format(time.RFC3339),
			},
		},
		Users: []*domain.User{
			{
				ID:        1,
				Username:  "user1",
				Email:     "user1@example.com",
				CreatedAt: now,
			},
			{
				ID:        2,
				Username:  "user2",
				Email:     "user2@example.com",
				CreatedAt: now.Add(time.Minute),
			},
		},
	}

	data, err := json.MarshalIndent(sampleData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

func exportMemoryState(filePath string) error {
	resp, err := http.Get("http://localhost:8080/tasks")
	if err != nil {
		return fmt.Errorf("failed to connect to server (is it running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var tasks []*domain.Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks response: %w", err)
	}

	data := FileData{
		Tasks: tasks,
		Users: []*domain.User{},
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal memory state data: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write memory state file: %w", err)
	}

	return nil
}
