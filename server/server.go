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
	"github.com/gin-gonic/gin"
)

type Server struct {
	engine  *gin.Engine
	server  *http.Server
	store   *TaskStore
	handler *TaskHandler
	ctx     context.Context
	cancel  context.CancelFunc
}

var serverInstance *Server

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	store := NewTaskStore()
	handler := NewTaskHandler(store)

	return &Server{
		engine:  engine,
		store:   store,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/")
	{
		api.GET("/tasks", s.handler.GetTasks)
		api.POST("/tasks", s.handler.CreateTask)
		api.GET("/tasks/:id", s.handler.GetTask)
		api.PUT("/tasks/:id", s.handler.UpdateTask)
		api.DELETE("/tasks/:id", s.handler.DeleteTask)
	}
}

func Run(port int) error {
	if pid.CheckRunning() {
		return fmt.Errorf("server is already running")
	}

	serverInstance = NewServer()
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
