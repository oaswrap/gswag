// Package api is a minimal Fiber HTTP server used by the gswag Fiber example tests.
package api

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

// Review is the domain model for the reviews resource.
type Review struct {
	ID      int    `json:"id"`
	Author  string `json:"author"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// CreateReviewRequest is the request body for submitting a review.
type CreateReviewRequest struct {
	Author  string `json:"author"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

var reviews = []Review{
	{ID: 1, Author: "Alice", Rating: 5, Comment: "Excellent!"},
	{ID: 2, Author: "Bob", Rating: 4, Comment: "Pretty good"},
}

// NewRouter returns a Fiber app wrapped as a standard http.Handler.
func NewRouter() http.Handler {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/reviews", func(c *fiber.Ctx) error {
		return c.JSON(reviews)
	})

	app.Post("/reviews", func(c *fiber.Ctx) error {
		var req CreateReviewRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		review := Review{ID: len(reviews) + 1, Author: req.Author, Rating: req.Rating, Comment: req.Comment}
		reviews = append(reviews, review)
		return c.Status(fiber.StatusCreated).JSON(review)
	})

	app.Get("/reviews/:id", func(c *fiber.Ctx) error {
		for _, r := range reviews {
			if c.Params("id") == "1" && r.ID == 1 {
				return c.JSON(r)
			}
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	})

	return adaptor.FiberApp(app)
}
