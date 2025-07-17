package store

import (
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"sync"
	"time"
)

type TaskMemoryStore struct {
	tasks  map[int]*domain.Task
	nextID int
	mu     sync.RWMutex
}

func NewTaskMemoryStore() *TaskMemoryStore {
	return &TaskMemoryStore{
		tasks:  make(map[int]*domain.Task),
		nextID: 1,
	}
}

func (ts *TaskMemoryStore) GetAll() []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*domain.Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

func (ts *TaskMemoryStore) GetAllByUserID(userID int) []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks := make([]*domain.Task, 0)
	for _, task := range ts.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (ts *TaskMemoryStore) GetByID(id int) (*domain.Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	task, exists := ts.tasks[id]
	return task, exists
}

func (ts *TaskMemoryStore) Create(task *domain.Task) *domain.Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task.ID = ts.nextID
	ts.nextID++
	task.CreatedAt = time.Now().Format(time.RFC3339)
	ts.tasks[task.ID] = task

	return task
}

func (ts *TaskMemoryStore) Update(id int, updatedTask *domain.Task) (*domain.Task, bool) {
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

func (ts *TaskMemoryStore) Delete(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if _, exists := ts.tasks[id]; !exists {
		return false
	}

	delete(ts.tasks, id)
	return true
}

type UserMemoryStore struct {
	users  map[int]*domain.User
	nextID int
	mu     sync.RWMutex
}

func NewUserMemoryStore() *UserMemoryStore {
	return &UserMemoryStore{
		users:  make(map[int]*domain.User),
		nextID: 1,
	}
}

func (us *UserMemoryStore) GetByID(id int) (*domain.User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	user, exists := us.users[id]
	return user, exists
}

func (us *UserMemoryStore) GetByUsername(username string) (*domain.User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	for _, user := range us.users {
		if user.Username == username {
			return user, true
		}
	}
	return nil, false
}

func (us *UserMemoryStore) Create(user *domain.User) *domain.User {
	us.mu.Lock()
	defer us.mu.Unlock()

	user.ID = us.nextID
	us.nextID++
	user.CreatedAt = time.Now()
	us.users[user.ID] = user

	return user
}

func (us *UserMemoryStore) Update(id int, updatedUser *domain.User) (*domain.User, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()

	existingUser, exists := us.users[id]
	if !exists {
		return nil, false
	}

	updatedUser.ID = id
	updatedUser.CreatedAt = existingUser.CreatedAt
	us.users[id] = updatedUser

	return updatedUser, true
}

func (us *UserMemoryStore) Delete(id int) bool {
	us.mu.Lock()
	defer us.mu.Unlock()

	if _, exists := us.users[id]; !exists {
		return false
	}

	delete(us.users, id)
	return true
}
