package store

import "github.com/KasumiMercury/mock-todo-server/server/domain"

type TaskStore interface {
	GetAll() []*domain.Task
	GetByID(id int) (*domain.Task, bool)
	Create(task *domain.Task) *domain.Task
	Update(id int, updatedTask *domain.Task) (*domain.Task, bool)
	Delete(id int) bool
}
