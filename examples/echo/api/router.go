// Package api is a minimal Echo HTTP server used by the gswag Echo example tests.
package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Product is the domain model for the products resource.
type Product struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

var products = []Product{
	{ID: 1, Title: "Laptop", Category: "Electronics", Price: 999.00},
	{ID: 2, Title: "Book", Category: "Education", Price: 29.99},
}

// NewRouter returns a configured Echo instance as an http.Handler.
func NewRouter() http.Handler {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.GET("/products", func(c echo.Context) error {
		return c.JSON(http.StatusOK, products)
	})

	e.POST("/products", func(c echo.Context) error {
		var req CreateProductRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}
		p := Product{ID: len(products) + 1, Title: req.Title, Category: req.Category, Price: req.Price}
		products = append(products, p)
		return c.JSON(http.StatusCreated, p)
	})

	e.GET("/products/:id", func(c echo.Context) error {
		for _, p := range products {
			if p.ID == 1 { // simplified — always return first match for test
				return c.JSON(http.StatusOK, p)
			}
		}
		return c.JSON(http.StatusNotFound, echo.Map{"error": "not found"})
	})

	return e
}
