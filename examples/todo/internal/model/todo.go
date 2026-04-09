// Package model defines domain types for the Todo application.
package model

import "time"

// Todo is the domain model for a todo item.
type Todo struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodoRequest is the request body for creating a todo.
type CreateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTodoRequest is the request body for updating a todo.
type UpdateTodoRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// Response is a generic envelope for successful API responses.
type Response[T any] struct {
	Data    T      `json:"data"`
	Message string `json:"message"`
}

// NewResponse wraps data in a Response envelope.
func NewResponse[T any](data T, message string) Response[T] {
	return Response[T]{Data: data, Message: message}
}

// ErrorResponse is a generic error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}
