package server

import (
	"sync"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

type TaskStore struct {
	tasks  map[int]*Task
	nextID int
	mu     sync.RWMutex
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

func (ts *TaskStore) GetAll() []*Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func (ts *TaskStore) GetByID(id int) (*Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	task, exists := ts.tasks[id]
	return task, exists
}

func (ts *TaskStore) Create(task *Task) *Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task.ID = ts.nextID
	ts.nextID++
	task.CreatedAt = time.Now().Format(time.RFC3339)
	ts.tasks[task.ID] = task

	return task
}

func (ts *TaskStore) Update(id int, updatedTask *Task) (*Task, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	existingTask, exists := ts.tasks[id]
	if !exists {
		return nil, false
	}

	updatedTask.ID = id
	updatedTask.CreatedAt = existingTask.CreatedAt
	ts.tasks[id] = updatedTask

	return updatedTask, true
}

func (ts *TaskStore) Delete(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.tasks[id]; !exists {
		return false
	}

	delete(ts.tasks, id)
	return true
}
