package store

import "github.com/KasumiMercury/mock-todo-server/server/domain"

type TaskStore interface {
	GetAll() []*domain.Task
	GetAllByUserID(userID int) []*domain.Task
	GetByID(id int) (*domain.Task, bool)
	Create(task *domain.Task) *domain.Task
	Update(id int, updatedTask *domain.Task) (*domain.Task, bool)
	Delete(id int) bool
}

type UserStore interface {
	GetAll() []*domain.User
	GetByID(id int) (*domain.User, bool)
	GetByUsername(username string) (*domain.User, bool)
	Create(user *domain.User) *domain.User
	Update(id int, updatedUser *domain.User) (*domain.User, bool)
	Delete(id int) bool
}
