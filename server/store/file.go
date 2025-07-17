package store

import (
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/goccy/go-json"
	"log"
	"os"
	"sync"
)

type TaskFileStore struct {
	filePath string
	nextID   int
	mu       sync.RWMutex
}

func NewTaskFileStore(filePath string) *TaskFileStore {
	return &TaskFileStore{
		filePath: filePath,
		nextID:   1,
	}
}

func (ts *TaskFileStore) GetAll() []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// Read json file and unmarshal into tasks
	file, err := os.ReadFile(ts.filePath)
	if err != nil {
		log.Println("Error reading tasks file:", err)
		return nil
	}

	var tasks []*domain.Task
	if err := json.Unmarshal(file, &tasks); err != nil {
		log.Println("Error unmarshalling tasks:", err)
		return nil
	}

	return tasks
}

func (ts *TaskFileStore) GetByID(id int) (*domain.Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := ts.GetAll()
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
	tasks := ts.GetAll()
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

	tasks := ts.GetAll()
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

	tasks := ts.GetAll()
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
