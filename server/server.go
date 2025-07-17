package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/KasumiMercury/mock-todo-server/pid"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

type Server struct {
	listener net.Listener
	server   *http.Server
	tasks    map[int]*Task
	nextID   int
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

var serverInstance *Server

func NewServer() *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		tasks:  make(map[int]*Task),
		nextID: 1,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", s.handleTasks)
	mux.HandleFunc("/tasks/", s.handleTaskByID)

	return mux
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.getTasks(w, r)
	case http.MethodPost:
		s.createTask(w, r)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if path == "" {
		http.Error(w, `{"error":"Task ID required"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, `{"error":"Invalid task ID"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getTask(w, r, id)
	case http.MethodPut:
		s.updateTask(w, r, id)
	case http.MethodDelete:
		s.deleteTask(w, r, id)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) getTasks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	task.ID = s.nextID
	s.nextID++
	task.CreatedAt = time.Now().Format(time.RFC3339)
	s.tasks[task.ID] = &task
	s.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&task)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request, id int) {
	s.mu.RLock()
	task, exists := s.tasks[id]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingTask, exists := s.tasks[id]
	if !exists {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	var updatedTask Task
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	updatedTask.ID = id
	updatedTask.CreatedAt = existingTask.CreatedAt
	s.tasks[id] = &updatedTask

	json.NewEncoder(w).Encode(&updatedTask)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request, id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	}

	delete(s.tasks, id)
	w.WriteHeader(http.StatusNoContent)
}

func Run(port int) error {
	if pid.CheckRunning() {
		return fmt.Errorf("server is already running")
	}

	serverInstance = NewServer()

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	serverInstance.listener = listener
	serverInstance.server = &http.Server{
		Handler: serverInstance.setupRoutes(),
	}

	if err := pid.CreatePidFile(os.Getpid()); err != nil {
		listener.Close()
		return fmt.Errorf("failed to create PID file: %w", err)
	}

	log.Printf("Mock TODO server starting on %s", addr)

	go func() {
		if err := serverInstance.server.Serve(listener); err != nil && err != http.ErrServerClosed {
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
