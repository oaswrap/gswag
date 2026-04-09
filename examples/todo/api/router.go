// Package api wires together the HTTP router for the Todo application.
package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/oaswrap/gswag/examples/todo/internal/handler"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
	"github.com/oaswrap/gswag/examples/todo/internal/service"
)

// NewRouter wires repository → service → handler and returns a configured Gin engine.
func NewRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo := repository.NewTodoRepository(db)
	svc := service.NewTodoService(repo)
	h := handler.NewTodoHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery())

	todos := r.Group("/todos")
	{
		todos.GET("", h.List)
		todos.POST("", h.Create)
		todos.GET("/:id", h.Get)
		todos.PUT("/:id", h.Update)
		todos.PATCH("/:id/done", h.MarkDone)
		todos.DELETE("/:id", h.Delete)
	}

	return r
}
