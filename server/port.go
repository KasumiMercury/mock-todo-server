package server

import (
	"net/http"
	"strconv"

	"github.com/KasumiMercury/mock-todo-server/server/auth"
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/KasumiMercury/mock-todo-server/server/store"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	store        store.TaskStore
	authRequired bool
}

func NewTaskHandler(store store.TaskStore, authRequired bool) *TaskHandler {
	return &TaskHandler{
		store:        store,
		authRequired: authRequired,
	}
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	if h.authRequired {
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		tasks := h.store.GetAllByUserID(userID)
		c.JSON(http.StatusOK, tasks)
	} else {
		// No authentication required, return all tasks
		tasks := h.store.GetAll()
		c.JSON(http.StatusOK, tasks)
	}
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task domain.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if h.authRequired {
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		// Set the user ID for the task
		task.UserID = userID
	} else {
		// No authentication required, set UserID to 0 (anonymous)
		task.UserID = 0
	}

	createdTask := h.store.Create(&task)
	c.JSON(http.StatusCreated, createdTask)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, exists := h.store.GetByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if h.authRequired {
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		// Check if the task belongs to the authenticated user
		if task.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if the task exists
	existingTask, exists := h.store.GetByID(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if h.authRequired {
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		// Check if the task belongs to the authenticated user
		if existingTask.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	var updatedTask domain.Task
	if err := c.ShouldBindJSON(&updatedTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if h.authRequired {
		userID, _ := auth.GetUserIDFromContext(c)
		// Ensure the task still belongs to the same user
		updatedTask.UserID = userID
	} else {
		// Preserve the original UserID when auth is not required
		updatedTask.UserID = existingTask.UserID
	}

	task, exists := h.store.Update(id, &updatedTask)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if h.authRequired {
		userID, exists := auth.GetUserIDFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		// Check if the task exists and belongs to the user
		existingTask, exists := h.store.GetByID(id)
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		if existingTask.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	if !h.store.Delete(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
