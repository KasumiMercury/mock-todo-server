package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KasumiMercury/mock-todo-server/pid"
	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/KasumiMercury/mock-todo-server/server/store"
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine       *gin.Engine
	server       *http.Server
	taskStore    store.TaskStore
	userStore    store.UserStore
	authService  *auth.AuthService
	taskHandler  *TaskHandler
	authHandler  *auth.AuthHandler
	authRequired bool
	ctx          context.Context
	cancel       context.CancelFunc
}

var serverInstance *Server

func NewServer(filePath string, keyMode auth.JWTKeyMode, secretKey string, authRequired bool) (*Server, error) {
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
	authHandler := auth.NewAuthHandler(authService)

	return &Server{
		engine:       engine,
		taskStore:    taskStore,
		userStore:    userStore,
		authService:  authService,
		taskHandler:  taskHandler,
		authHandler:  authHandler,
		authRequired: authRequired,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (s *Server) setupRoutes() {
	// Authentication routes (no auth required)
	authGroup := s.engine.Group("/auth")
	{
		authGroup.POST("/login", s.authHandler.Login)
		authGroup.POST("/register", s.authHandler.Register)
		authGroup.GET("/jwks", s.authHandler.GetJWKs)
		authGroup.GET("/public-key", s.authHandler.GetPublicKey)
	}

	// Standard well-known endpoints (no auth required)
	wellKnownGroup := s.engine.Group("/.well-known")
	{
		wellKnownGroup.GET("/jwks.json", s.authHandler.GetJWKs)
		wellKnownGroup.GET("/openid_configuration", s.authHandler.GetOpenIDConfiguration)
	}

	if s.authRequired {
		// Protected routes (auth required)
		api := s.engine.Group("/")
		api.Use(auth.AuthMiddleware(s.authService))
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
			authRequired.Use(auth.AuthMiddleware(s.authService))
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

func Run(port int, filePath string, keyMode string, secretKey string, authRequired bool) error {
	if pid.CheckRunning() {
		return fmt.Errorf("server is already running")
	}

	var jwtKeyMode auth.JWTKeyMode
	switch keyMode {
	case "secret":
		jwtKeyMode = auth.JWTKeyModeSecret
	case "rsa":
		jwtKeyMode = auth.JWTKeyModeRSA
	default:
		return fmt.Errorf("invalid jwt-key-mode: %s (must be 'secret' or 'rsa')", keyMode)
	}

	var err error
	serverInstance, err = NewServer(filePath, jwtKeyMode, secretKey, authRequired)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	serverInstance.setupRoutes()

	addr := fmt.Sprintf(":%d", port)
	serverInstance.server = &http.Server{
		Addr:    addr,
		Handler: serverInstance.engine,
	}

	if err := pid.CreatePidFile(os.Getpid()); err != nil {
		return fmt.Errorf("failed to create PID file: %w", err)
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
