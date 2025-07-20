package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KasumiMercury/mock-todo-server/export"
	"github.com/KasumiMercury/mock-todo-server/pid"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/KasumiMercury/mock-todo-server/server/store"
	"github.com/gin-gonic/gin"
)

//go:embed templates
var templates embed.FS

type Server struct {
	engine       *gin.Engine
	server       *http.Server
	taskStore    store.TaskStore
	userStore    store.UserStore
	authService  *auth.AuthService
	taskHandler  *TaskHandler
	authHandler  *auth.AuthHandler
	oidcHandler  *auth.OIDCHandler
	authRequired bool
	authMode     auth.AuthMode
	ctx          context.Context
	cancel       context.CancelFunc
}

var serverInstance *Server

func NewServer(filePath string, keyMode auth.JWTKeyMode, secretKey string, authRequired bool, authMode auth.AuthMode, oidcConfigPath string) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	var taskStore store.TaskStore
	var userStore store.UserStore

	if filePath == "" {
		taskStore = store.NewTaskMemoryStore()
		userStore = store.NewUserMemoryStore()
	} else {
		taskStore = store.NewTaskFileStore(filePath)
		userStore = store.NewUserFileStore(filePath)
		log.Printf("Using file store at %s", filePath)
	}

	authService, err := auth.NewAuthService(userStore, keyMode, secretKey)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	taskHandler := NewTaskHandler(taskStore, authRequired)
	authHandler := auth.NewAuthHandler(authService, authMode)

	// Create OIDC handler if OIDC mode is enabled
	var oidcHandler *auth.OIDCHandler
	if authMode == auth.AuthModeOIDC {
		oidcConfig, err := auth.LoadOIDCConfig(oidcConfigPath)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to load OIDC config: %w", err)
		}

		oidcService := auth.NewOIDCService(oidcConfig, userStore, authService)
		oidcHandler = auth.NewOIDCHandler(oidcService, authService)

		// Load HTML templates for OIDC
		t := template.Must(template.ParseFS(templates, "templates/*.html"))
		engine.SetHTMLTemplate(t)
	}

	return &Server{
		engine:       engine,
		taskStore:    taskStore,
		userStore:    userStore,
		authService:  authService,
		taskHandler:  taskHandler,
		authHandler:  authHandler,
		oidcHandler:  oidcHandler,
		authRequired: authRequired,
		authMode:     authMode,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (s *Server) GetMemoryState() (*export.FileData, error) {
	tasks := s.taskStore.GetAll()
	users := s.userStore.GetAll()

	// Convert users to UserStorage for export
	userStorages := make([]*domain.UserStorage, 0, len(users))
	for _, user := range users {
		userStorage := user.ToStorage(user.HashedPassword)
		userStorages = append(userStorages, userStorage)
	}

	return &export.FileData{
		Tasks: tasks,
		Users: userStorages,
	}, nil
}

func (s *Server) getMemoryStateHandler(c *gin.Context) {
	data, err := s.GetMemoryState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (s *Server) setupRoutes() {
	// Authentication routes (no auth required)
	authGroup := s.engine.Group("/auth")
	{
		if s.authMode == auth.AuthModeOIDC {
			// OIDC specific routes
			authGroup.GET("/authorize", s.oidcHandler.Authorize)
			authGroup.POST("/authorize", s.oidcHandler.Authorize)
			authGroup.POST("/token", s.oidcHandler.Token)
			authGroup.GET("/userinfo", s.oidcHandler.UserInfo)
			authGroup.GET("/jwks", s.authHandler.GetJWKs)
			authGroup.GET("/register", s.oidcHandler.Register)
			authGroup.POST("/register", s.oidcHandler.Register)
		} else {
			// Standard auth routes
			authGroup.POST("/login", s.authHandler.Login)
			authGroup.POST("/register", s.authHandler.Register)
			authGroup.POST("/logout", s.authHandler.Logout)
			authGroup.GET("/jwks", s.authHandler.GetJWKs)
		}
	}

	// Internal API endpoints (no auth required)
	internalGroup := s.engine.Group("/internal")
	{
		internalGroup.GET("/memory-state", s.getMemoryStateHandler)
	}

	// Standard well-known endpoints (no auth required)
	wellKnownGroup := s.engine.Group("/.well-known")
	{
		wellKnownGroup.GET("/jwks.json", s.authHandler.GetJWKs)
		if s.authMode == auth.AuthModeOIDC {
			wellKnownGroup.GET("/openid_configuration", s.oidcHandler.GetOpenIDConfiguration)
		} else {
			wellKnownGroup.GET("/openid_configuration", s.authHandler.GetOpenIDConfiguration)
		}
	}

	if s.authRequired {
		// Protected routes (auth required)
		api := s.engine.Group("/")
		api.Use(auth.AuthMiddleware(s.authService, s.authMode))
		{
			api.GET("/auth/me", s.authHandler.Me)
			api.GET("/tasks", s.taskHandler.GetTasks)
			api.POST("/tasks", s.taskHandler.CreateTask)
			api.GET("/tasks/:id", s.taskHandler.GetTask)
			api.PUT("/tasks/:id", s.taskHandler.UpdateTask)
			api.DELETE("/tasks/:id", s.taskHandler.DeleteTask)
		}
	} else {
		// Unprotected routes (no auth required)
		api := s.engine.Group("/")
		{
			// Still provide /auth/me endpoint but with auth middleware
			authRequired := api.Group("/auth")
			authRequired.Use(auth.AuthMiddleware(s.authService, s.authMode))
			{
				authRequired.GET("/me", s.authHandler.Me)
			}

			// Task routes without auth middleware
			api.GET("/tasks", s.taskHandler.GetTasks)
			api.POST("/tasks", s.taskHandler.CreateTask)
			api.GET("/tasks/:id", s.taskHandler.GetTask)
			api.PUT("/tasks/:id", s.taskHandler.UpdateTask)
			api.DELETE("/tasks/:id", s.taskHandler.DeleteTask)
		}
	}
}

func Run(config *Config) error {
	if pid.CheckRunning() {
		return fmt.Errorf("server is already running")
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	var err error
	serverInstance, err = NewServer(config.JsonFilePath, config.JWTKeyMode, config.JWTSecretKey, config.AuthRequired, config.AuthMode, config.OIDCConfigPath)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Set server instance as export provider
	export.SetServerProvider(serverInstance)

	serverInstance.setupRoutes()

	addr := fmt.Sprintf(":%d", config.Port)
	serverInstance.server = &http.Server{
		Addr:    addr,
		Handler: serverInstance.engine,
	}

	if err := pid.CreatePidFile(os.Getpid()); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	if err := pid.CreateServerInfoFile(os.Getpid(), config.Port); err != nil {
		return fmt.Errorf("failed to create server info file: %w", err)
	}

	log.Printf("Mock TODO server starting on %s", addr)

	go func() {
		if err := serverInstance.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-serverInstance.ctx.Done():
		log.Println("Server shutdown initiated via Stop()")
	case sig := <-sigChan:
		log.Printf("Server shutdown initiated via signal: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serverInstance.server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if err := os.Remove(pid.PidFile); err != nil {
		log.Printf("Failed to remove PID file: %v", err)
	}

	if err := pid.RemoveServerInfoFile(); err != nil {
		log.Printf("Failed to remove server info file: %v", err)
	}

	log.Println("Server stopped")
	return nil
}

func Stop() error {
	if !pid.CheckRunning() {
		return fmt.Errorf("server is not running")
	}

	if serverInstance != nil {
		serverInstance.cancel()
		return nil
	}

	return pid.StopByPid()
}

func GetServerInstance() *Server {
	return serverInstance
}
