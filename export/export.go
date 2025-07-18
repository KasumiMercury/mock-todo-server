package export

import (
	"encoding/json"
	"fmt"
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileData struct {
	Tasks []*domain.Task `json:"tasks"`
	Users []*domain.User `json:"users"`
}

func Export(args []string, templateMode, memoryMode, oidcMode bool) error {
	if templateMode {
		filePath := GetOutputPath(args, "data-template.json")
		return Template(filePath)
	}

	if memoryMode {
		filePath := GetOutputPath(args, "memory-state.json")
		return MemoryState(filePath)
	}

	if oidcMode {
		filePath := GetOutputPath(args, "oidc-config.json")
		return OidcTemplate(filePath)
	}

	return fmt.Errorf("no valid export mode specified")
}

func GetOutputPath(args []string, defaultFilename string) string {
	if len(args) > 0 {
		return args[0]
	}

	return defaultFilename
}

func Template(filePath string) error {
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
				CreatedAt: now,
			},
			{
				ID:        2,
				Username:  "user2",
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

func MemoryState(filePath string) error {
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

// OidcTemplate exports an OIDC configuration template
func OidcTemplate(filePath string) error {
	oidcTemplate := map[string]interface{}{
		"client_id":     "mock-client-id-12345",
		"client_secret": "mock-client-secret-67890",
		"redirect_uris": []string{
			"http://localhost:3000/callback",
			"http://localhost:3000/auth/callback",
			"https://your-app.example.com/callback",
		},
		"issuer": "http://localhost:8080",
		"scopes": []string{
			"openid",
			"profile",
		},
	}

	data, err := json.MarshalIndent(oidcTemplate, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OIDC template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write OIDC template file: %w", err)
	}

	return nil
}
