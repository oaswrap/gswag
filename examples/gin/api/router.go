// Package api is a minimal Gin HTTP server used by the gswag Gin example tests.
package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Item is the domain model for the items resource.
type Item struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CreateItemRequest is the request body for creating an item.
type CreateItemRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

var items = []Item{
	{ID: 1, Name: "Widget", Price: 9.99},
	{ID: 2, Name: "Gadget", Price: 24.99},
}

// NewRouter returns a configured Gin engine.
func NewRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/items", func(c *gin.Context) {
		c.JSON(http.StatusOK, items)
	})

	r.POST("/items", func(c *gin.Context) {
		var req CreateItemRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item := Item{ID: len(items) + 1, Name: req.Name, Price: req.Price}
		items = append(items, item)
		c.JSON(http.StatusCreated, item)
	})

	r.GET("/items/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		for _, item := range items {
			if item.ID == id {
				c.JSON(http.StatusOK, item)
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	r.DELETE("/items/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	return r
}
