// Package api wires together the HTTP router for the Todo application.
package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oaswrap/gswag/examples/todo/internal/handler"
	"github.com/oaswrap/gswag/examples/todo/internal/repository"
	"github.com/oaswrap/gswag/examples/todo/internal/service"
)

// APIKey is the shared secret used by the API key middleware.
const APIKey = "secret"

// requireAPIKey rejects requests that do not carry the correct X-API-Key header.
func requireAPIKey(c *gin.Context) {
	if c.GetHeader("X-API-Key") != APIKey {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Next()
}

// NewRouter wires repository → service → handler and returns a configured Gin engine.
func NewRouter(db *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo := repository.NewTodoRepository(db)
	svc := service.NewTodoService(repo)
	h := handler.NewTodoHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery())

	todos := r.Group("/todos", requireAPIKey)
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
