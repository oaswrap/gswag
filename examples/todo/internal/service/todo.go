// Package service implements business logic for the Todo application.
package service

import (
	"github.com/oaswrap/gswag/examples/todo/internal/model"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
)

// TodoService defines the use-case contract.
type TodoService interface {
	List() ([]model.Todo, error)
	Get(id int64) (*model.Todo, error)
	Create(req model.CreateTodoRequest) (*model.Todo, error)
	Update(id int64, req model.UpdateTodoRequest) (*model.Todo, error)
	MarkDone(id int64) (*model.Todo, error)
	Delete(id int64) error
}

type todoService struct {
	repo repository.TodoRepository
}

// NewTodoService returns a TodoService backed by the given repository.
func NewTodoService(repo repository.TodoRepository) TodoService {
	return &todoService{repo: repo}
}

func (s *todoService) List() ([]model.Todo, error) {
	return s.repo.FindAll()
}

func (s *todoService) Get(id int64) (*model.Todo, error) {
	return s.repo.FindByID(id)
}

func (s *todoService) Create(req model.CreateTodoRequest) (*model.Todo, error) {
	return s.repo.Create(req)
}

func (s *todoService) Update(id int64, req model.UpdateTodoRequest) (*model.Todo, error) {
	return s.repo.Update(id, req)
}

func (s *todoService) MarkDone(id int64) (*model.Todo, error) {
	todo, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(id, model.UpdateTodoRequest{
		Title:       todo.Title,
		Description: todo.Description,
		Done:        true,
	})
}

func (s *todoService) Delete(id int64) error {
	return s.repo.Delete(id)
}
