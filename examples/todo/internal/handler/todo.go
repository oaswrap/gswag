// Package handler provides Gin HTTP handlers for the Todo application.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oaswrap/gswag/examples/todo/internal/model"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
	"github.com/oaswrap/gswag/examples/todo/internal/service"
)

// TodoHandler holds the service dependency for todo HTTP handlers.
type TodoHandler struct {
	svc service.TodoService
}

// NewTodoHandler returns a new TodoHandler.
func NewTodoHandler(svc service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// List handles GET /todos.
func (h *TodoHandler) List(c *gin.Context) {
	todos, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.NewResponse(todos, "ok"))
}

// Create handles POST /todos.
func (h *TodoHandler) Create(c *gin.Context) {
	var req model.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "title is required"})
		return
	}
	todo, err := h.svc.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, model.NewResponse(*todo, "created"))
}

// Get handles GET /todos/:id.
func (h *TodoHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}
	todo, err := h.svc.Get(id)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "todo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.NewResponse(*todo, "ok"))
}

// Update handles PUT /todos/:id.
func (h *TodoHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}
	var req model.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "title is required"})
		return
	}
	todo, err := h.svc.Update(id, req)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "todo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.NewResponse(*todo, "ok"))
}

// MarkDone handles PATCH /todos/:id/done.
func (h *TodoHandler) MarkDone(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}
	todo, err := h.svc.MarkDone(id)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "todo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.NewResponse(*todo, "ok"))
}

// Delete handles DELETE /todos/:id.
func (h *TodoHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid id"})
		return
	}
	err = h.svc.Delete(id)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "todo not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
