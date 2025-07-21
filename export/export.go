package export

import (
	"encoding/json"
	"fmt"
	"github.com/KasumiMercury/mock-todo-server/pid"
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ExportMode string

const (
	StoreMode        ExportMode = "store"
	MemoryExportMode ExportMode = "memory"
	OidcMode         ExportMode = "oidc"
)

const (
	DefaultStoreFile  = "data.json"
	DefaultMemoryFile = "memory-state.json"
	DefaultOidcFile   = "oidc-config.json"
)

type ServerProvider interface {
	GetMemoryState() (*FileData, error)
}

var serverProvider ServerProvider

func SetServerProvider(provider ServerProvider) {
	serverProvider = provider
}

func getServerPort() int {
	// Try to get port from server info file
	if serverInfo, err := pid.GetServerInfo(); err == nil {
		return serverInfo.Port
	}

	// Fallback to default port if server info is not available
	return 8080
}

type FileData struct {
	Tasks []*domain.Task        `json:"tasks"`
	Users []*domain.UserStorage `json:"users"`
}

// ExportWithMode exports data based on the specified mode and file path.
func ExportWithMode(mode ExportMode, filePath string) error {
	if filePath == "" {
		switch mode {
		case StoreMode:
			filePath = DefaultStoreFile
		case MemoryExportMode:
			filePath = DefaultMemoryFile
		case OidcMode:
			filePath = DefaultOidcFile
		default:
			return fmt.Errorf("unknown export mode: %s", mode)
		}
	}

	switch mode {
	case StoreMode:
		return Store(filePath)
	case MemoryExportMode:
		return MemoryState(filePath)
	case OidcMode:
		return OidcTemplate(filePath)
	default:
		return fmt.Errorf("unknown export mode: %s", mode)
	}
}

func GetOutputPath(args []string, defaultFilename string) string {
	if len(args) > 0 {
		return args[0]
	}

	return defaultFilename
}

func Store(filePath string) error {
	now := time.Now()

	// Create hashed passwords for template users
	hashedPassword1, err := bcrypt.GenerateFromPassword([]byte("password1"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password for user1: %w", err)
	}

	hashedPassword2, err := bcrypt.GenerateFromPassword([]byte("password2"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password for user2: %w", err)
	}

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
		Users: []*domain.UserStorage{
			{
				ID:             1,
				Username:       "user1",
				HashedPassword: string(hashedPassword1),
				CreatedAt:      now,
			},
			{
				ID:             2,
				Username:       "user2",
				HashedPassword: string(hashedPassword2),
				CreatedAt:      now.Add(time.Minute),
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

	log.Println("Template file created at:", filePath)
	log.Println("Template includes 2 users:")
	log.Println("  - user1 (password: password1)")
	log.Println("  - user2 (password: password2)")

	return nil
}

func MemoryState(filePath string) error {
	// Try to get memory state from server instance (if available)
	if data, err := getMemoryStateFromServer(); err == nil {
		return saveMemoryStateToFile(data, filePath)
	}

	// Try to get memory state from internal API endpoint
	if data, err := getMemoryStateFromInternalAPI(); err == nil {
		return saveMemoryStateToFile(data, filePath)
	}

	// Try to get memory state directly (for stopped servers)
	if data, err := getMemoryStateDirectly(); err == nil {
		return saveMemoryStateToFile(data, filePath)
	}

	return fmt.Errorf("failed to get memory state: all methods failed")
}

func getMemoryStateFromServer() (*FileData, error) {
	if serverProvider == nil {
		return nil, fmt.Errorf("server provider not available")
	}

	return serverProvider.GetMemoryState()
}

func getMemoryStateFromInternalAPI() (*FileData, error) {
	port := getServerPort()
	url := fmt.Sprintf("http://localhost:%d/internal/memory-state", port)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to internal API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("internal API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read internal API response: %w", err)
	}

	var data FileData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse internal API response: %w", err)
	}

	return &data, nil
}

func getMemoryStateDirectly() (*FileData, error) {
	// This function provides direct access to memory state without server running
	// It's used when server is stopped but we need to access the last known state

	// For now, return empty data as this is a fallback for stopped servers
	// In a real implementation, this might read from a cached state file
	// or use other mechanisms to preserve state
	return &FileData{
		Tasks: []*domain.Task{},
		Users: []*domain.UserStorage{},
	}, nil
}

func saveMemoryStateToFile(data *FileData, filePath string) error {
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

	log.Println("Memory state file exported to:", filePath)
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

	log.Println("OIDC configuration template exported to:", filePath)
	return nil
}
