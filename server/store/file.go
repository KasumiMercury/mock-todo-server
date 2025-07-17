package store

import (
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/goccy/go-json"
	"log"
	"os"
	"sync"
	"time"
)

type TaskFileStore struct {
	filePath string
	nextID   int
	mu       sync.RWMutex
}

func NewTaskFileStore(filePath string) *TaskFileStore {
	store := &TaskFileStore{
		filePath: filePath,
		nextID:   1,
	}

	// Initialize nextID based on existing tasks
	store.initializeNextID()

	return store
}

func (ts *TaskFileStore) initializeNextID() {
	tasks := ts.loadTasksFromFile()
	maxID := 0
	for _, task := range tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	ts.nextID = maxID + 1
}

func (ts *TaskFileStore) GetAll() []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return ts.loadTasksFromFile()
}

func (ts *TaskFileStore) loadTasksFromFile() []*domain.Task {
	// Create empty file if it doesn't exist
	if _, err := os.Stat(ts.filePath); os.IsNotExist(err) {
		if err := ts.createEmptyFile(); err != nil {
			log.Println("Error creating empty tasks file:", err)
			return []*domain.Task{}
		}
	}

	// Read json file and unmarshal into tasks
	file, err := os.ReadFile(ts.filePath)
	if err != nil {
		log.Println("Error reading tasks file:", err)
		return []*domain.Task{}
	}

	// Handle empty file
	if len(file) == 0 {
		return []*domain.Task{}
	}

	var tasks []*domain.Task
	if err := json.Unmarshal(file, &tasks); err != nil {
		log.Println("Error unmarshalling tasks:", err)
		return []*domain.Task{}
	}

	return tasks
}

func (ts *TaskFileStore) createEmptyFile() error {
	emptyJSON := []byte("[]")
	return os.WriteFile(ts.filePath, emptyJSON, 0644)
}

func (ts *TaskFileStore) GetByID(id int) (*domain.Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := ts.loadTasksFromFile()
	for _, task := range tasks {
		if task.ID == id {
			return task, true
		}
	}
	return nil, false
}

func (ts *TaskFileStore) Create(task *domain.Task) *domain.Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task.ID = ts.nextID
	ts.nextID++
	task.CreatedAt = time.Now().Format(time.RFC3339)

	tasks := ts.loadTasksFromFile()
	tasks = append(tasks, task)

	// Marshal tasks to json and write to file
	data, err := json.Marshal(tasks)
	if err != nil {
		log.Println("Error marshalling tasks:", err)
		return nil
	}

	if err := os.WriteFile(ts.filePath, data, 0644); err != nil {
		log.Println("Error writing tasks file:", err)
		return nil
	}

	return task
}

func (ts *TaskFileStore) Update(id int, updatedTask *domain.Task) (*domain.Task, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tasks := ts.loadTasksFromFile()
	for i, task := range tasks {
		if task.ID == id {
			updatedTask.ID = id
			updatedTask.CreatedAt = task.CreatedAt // Preserve the original creation time
			tasks[i] = updatedTask

			// Marshal tasks to json and write to file
			data, err := json.Marshal(tasks)
			if err != nil {
				log.Println("Error marshalling tasks:", err)
				return nil, false
			}

			if err := os.WriteFile(ts.filePath, data, 0644); err != nil {
				log.Println("Error writing tasks file:", err)
				return nil, false
			}

			return updatedTask, true
		}
	}
	return nil, false
}

func (ts *TaskFileStore) Delete(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	tasks := ts.loadTasksFromFile()
	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...) // Remove the task
			// Marshal tasks to json and write to file
			data, err := json.Marshal(tasks)
			if err != nil {
				log.Println("Error marshalling tasks:", err)
				return false
			}

			if err := os.WriteFile(ts.filePath, data, 0644); err != nil {
				log.Println("Error writing tasks file:", err)
				return false
			}

			return true
		}
	}
	return false
}
