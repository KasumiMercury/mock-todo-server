package store

import (
	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/goccy/go-json"
	"log"
	"os"
	"sync"
	"time"
)

type FileData struct {
	Tasks []*domain.Task        `json:"tasks"`
	Users []*domain.UserStorage `json:"users"`
}

type TaskFileStore struct {
	filePath   string
	nextTaskID int
	mu         sync.RWMutex
}

type UserFileStore struct {
	filePath   string
	nextUserID int
	mu         sync.RWMutex
}

func NewTaskFileStore(filePath string) *TaskFileStore {
	store := &TaskFileStore{
		filePath:   filePath,
		nextTaskID: 1,
	}

	// Initialize nextTaskID based on existing tasks
	store.initializeNextTaskID()

	return store
}

func NewUserFileStore(filePath string) *UserFileStore {
	store := &UserFileStore{
		filePath:   filePath,
		nextUserID: 1,
	}

	// Initialize nextUserID based on existing users
	store.initializeNextUserID()

	return store
}

func (ts *TaskFileStore) initializeNextTaskID() {
	data := ts.loadDataFromFile()
	maxID := 0
	for _, task := range data.Tasks {
		if task.ID > maxID {
			maxID = task.ID
		}
	}
	ts.nextTaskID = maxID + 1
}

func (us *UserFileStore) initializeNextUserID() {
	data := us.loadDataFromFile()
	maxID := 0
	for _, user := range data.Users {
		if user.ID > maxID {
			maxID = user.ID
		}
	}
	us.nextUserID = maxID + 1
}

func (ts *TaskFileStore) GetAll() []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data := ts.loadDataFromFile()
	return data.Tasks
}

func (ts *TaskFileStore) GetAllByUserID(userID int) []*domain.Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data := ts.loadDataFromFile()
	tasks := make([]*domain.Task, 0)
	for _, task := range data.Tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (ts *TaskFileStore) loadDataFromFile() *FileData {
	// Create empty file if it doesn't exist
	if _, err := os.Stat(ts.filePath); os.IsNotExist(err) {
		if err := ts.createEmptyFile(); err != nil {
			log.Println("Error creating empty data file:", err)
			return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
		}
	}

	// Read json file and unmarshal into data
	file, err := os.ReadFile(ts.filePath)
	if err != nil {
		log.Println("Error reading data file:", err)
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	// Handle empty file
	if len(file) == 0 {
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	var data FileData
	if err := json.Unmarshal(file, &data); err != nil {
		log.Println("Error unmarshalling data:", err)
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	// Initialize empty arrays if nil
	if data.Tasks == nil {
		data.Tasks = []*domain.Task{}
	}
	if data.Users == nil {
		data.Users = []*domain.UserStorage{}
	}

	return &data
}

func (us *UserFileStore) loadDataFromFile() *FileData {
	// Create empty file if it doesn't exist
	if _, err := os.Stat(us.filePath); os.IsNotExist(err) {
		if err := us.createEmptyFile(); err != nil {
			log.Println("Error creating empty data file:", err)
			return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
		}
	}

	// Read json file and unmarshal into data
	file, err := os.ReadFile(us.filePath)
	if err != nil {
		log.Println("Error reading data file:", err)
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	// Handle empty file
	if len(file) == 0 {
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	var data FileData
	if err := json.Unmarshal(file, &data); err != nil {
		log.Println("Error unmarshalling data:", err)
		return &FileData{Tasks: []*domain.Task{}, Users: []*domain.UserStorage{}}
	}

	// Initialize empty arrays if nil
	if data.Tasks == nil {
		data.Tasks = []*domain.Task{}
	}
	if data.Users == nil {
		data.Users = []*domain.UserStorage{}
	}

	return &data
}

func (ts *TaskFileStore) createEmptyFile() error {
	emptyData := FileData{
		Tasks: []*domain.Task{},
		Users: []*domain.UserStorage{},
	}
	data, err := json.Marshal(emptyData)
	if err != nil {
		return err
	}
	return os.WriteFile(ts.filePath, data, 0644)
}

func (us *UserFileStore) createEmptyFile() error {
	emptyData := FileData{
		Tasks: []*domain.Task{},
		Users: []*domain.UserStorage{},
	}
	data, err := json.Marshal(emptyData)
	if err != nil {
		return err
	}
	return os.WriteFile(us.filePath, data, 0644)
}

func (ts *TaskFileStore) GetByID(id int) (*domain.Task, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	data := ts.loadDataFromFile()
	for _, task := range data.Tasks {
		if task.ID == id {
			return task, true
		}
	}
	return nil, false
}

func (ts *TaskFileStore) Create(task *domain.Task) *domain.Task {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	task.ID = ts.nextTaskID
	ts.nextTaskID++
	task.CreatedAt = time.Now().Format(time.RFC3339)

	data := ts.loadDataFromFile()
	data.Tasks = append(data.Tasks, task)

	// Marshal data to json and write to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return nil
	}

	if err := os.WriteFile(ts.filePath, jsonData, 0644); err != nil {
		log.Println("Error writing data file:", err)
		return nil
	}

	return task
}

func (ts *TaskFileStore) Update(id int, updatedTask *domain.Task) (*domain.Task, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	data := ts.loadDataFromFile()
	for i, task := range data.Tasks {
		if task.ID == id {
			updatedTask.ID = id
			updatedTask.CreatedAt = task.CreatedAt // Preserve the original creation time
			data.Tasks[i] = updatedTask

			// Marshal data to json and write to file
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshalling data:", err)
				return nil, false
			}

			if err := os.WriteFile(ts.filePath, jsonData, 0644); err != nil {
				log.Println("Error writing data file:", err)
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

	data := ts.loadDataFromFile()
	for i, task := range data.Tasks {
		if task.ID == id {
			data.Tasks = append(data.Tasks[:i], data.Tasks[i+1:]...) // Remove the task
			// Marshal data to json and write to file
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshalling data:", err)
				return false
			}

			if err := os.WriteFile(ts.filePath, jsonData, 0644); err != nil {
				log.Println("Error writing data file:", err)
				return false
			}

			return true
		}
	}
	return false
}

// UserFileStore methods
func (us *UserFileStore) GetAll() []*domain.User {
	us.mu.RLock()
	defer us.mu.RUnlock()

	data := us.loadDataFromFile()
	users := make([]*domain.User, 0, len(data.Users))
	for _, userStorage := range data.Users {
		users = append(users, userStorage.ToUser())
	}
	return users
}

func (us *UserFileStore) GetByID(id int) (*domain.User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	data := us.loadDataFromFile()
	for _, userStorage := range data.Users {
		if userStorage.ID == id {
			return userStorage.ToUser(), true
		}
	}
	return nil, false
}

func (us *UserFileStore) GetByUsername(username string) (*domain.User, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	data := us.loadDataFromFile()
	for _, userStorage := range data.Users {
		if userStorage.Username == username {
			return userStorage.ToUser(), true
		}
	}
	return nil, false
}

func (us *UserFileStore) Create(user *domain.User) *domain.User {
	us.mu.Lock()
	defer us.mu.Unlock()

	user.ID = us.nextUserID
	us.nextUserID++
	user.CreatedAt = time.Now()

	data := us.loadDataFromFile()
	// Convert User to UserStorage for JSON persistence
	userStorage := user.ToStorage(user.HashedPassword)
	data.Users = append(data.Users, userStorage)

	// Marshal data to json and write to file
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling data:", err)
		return nil
	}

	if err := os.WriteFile(us.filePath, jsonData, 0644); err != nil {
		log.Println("Error writing data file:", err)
		return nil
	}

	return user
}

func (us *UserFileStore) Update(id int, updatedUser *domain.User) (*domain.User, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()

	data := us.loadDataFromFile()
	for i, userStorage := range data.Users {
		if userStorage.ID == id {
			updatedUser.ID = id
			updatedUser.CreatedAt = userStorage.CreatedAt // Preserve the original creation time
			// Convert User to UserStorage for JSON persistence
			newUserStorage := updatedUser.ToStorage(updatedUser.HashedPassword)
			data.Users[i] = newUserStorage

			// Marshal data to json and write to file
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshalling data:", err)
				return nil, false
			}

			if err := os.WriteFile(us.filePath, jsonData, 0644); err != nil {
				log.Println("Error writing data file:", err)
				return nil, false
			}

			return updatedUser, true
		}
	}
	return nil, false
}

func (us *UserFileStore) Delete(id int) bool {
	us.mu.Lock()
	defer us.mu.Unlock()

	data := us.loadDataFromFile()
	for i, userStorage := range data.Users {
		if userStorage.ID == id {
			data.Users = append(data.Users[:i], data.Users[i+1:]...) // Remove the user
			// Marshal data to json and write to file
			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshalling data:", err)
				return false
			}

			if err := os.WriteFile(us.filePath, jsonData, 0644); err != nil {
				log.Println("Error writing data file:", err)
				return false
			}

			return true
		}
	}
	return false
}
