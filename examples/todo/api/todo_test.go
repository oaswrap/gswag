package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/todo/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Path("/todos", func() {
	Get("List all todos", func() {
		Tag("todos")

		BeforeEach(func() {
			DeferCleanup(newTestEnv())
		})

		Response(200, "list of todos", func() {
			ResponseSchema(new(model.Response[[]model.Todo]))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("data"))
				Expect(resp).To(ContainJSONKey("message"))
			})
		})
	})

	Post("Create a todo", func() {
		Tag("todos")
		RequestBody(new(model.CreateTodoRequest))

		BeforeEach(func() {
			DeferCleanup(newTestEnv())
		})

		Response(201, "todo created", func() {
			ResponseSchema(new(model.Response[model.Todo]))
			SetBody(&model.CreateTodoRequest{
				Title:       "Buy groceries",
				Description: "Milk, eggs, bread",
			})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("data"))
				Expect(resp).To(ContainJSONKey("message"))
				Expect(resp).To(MatchJSONSchema(&model.Response[model.Todo]{}))
			})
		})

		Response(400, "missing or invalid body", func() {
			SetRawBody([]byte("not-json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/todos/{id}", func() {
	Get("Get a todo by ID", func() {
		Tag("todos")
		Parameter("id", PathParam, Integer)

		BeforeEach(func() {
			DeferCleanup(newTestEnv(model.CreateTodoRequest{
				Title:       "Buy groceries",
				Description: "Milk and eggs",
			}))
		})

		Response(200, "todo found", func() {
			ResponseSchema(new(model.Response[model.Todo]))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("data"))
				Expect(resp).To(MatchJSONSchema(&model.Response[model.Todo]{}))
			})
		})

		Response(404, "todo not found", func() {
			SetParam("id", "99999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Put("Update a todo", func() {
		Tag("todos")
		Parameter("id", PathParam, Integer)
		RequestBody(new(model.UpdateTodoRequest))

		BeforeEach(func() {
			DeferCleanup(newTestEnv(model.CreateTodoRequest{
				Title:       "Write tests",
				Description: "Cover all endpoints",
			}))
		})

		Response(200, "todo updated", func() {
			ResponseSchema(new(model.Response[model.Todo]))
			SetParam("id", "1")
			SetBody(&model.UpdateTodoRequest{
				Title:       "Write tests (revised)",
				Description: "Cover all endpoints and edge cases",
				Done:        false,
			})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("data"))
				Expect(resp).To(MatchJSONSchema(&model.Response[model.Todo]{}))
			})
		})

		Response(400, "missing or invalid body", func() {
			SetParam("id", "1")
			SetRawBody([]byte("not-json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})

		Response(404, "todo not found", func() {
			SetParam("id", "99999")
			SetBody(&model.UpdateTodoRequest{Title: "ghost"})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Delete("Delete a todo", func() {
		Tag("todos")
		Parameter("id", PathParam, Integer)

		BeforeEach(func() {
			DeferCleanup(newTestEnv(model.CreateTodoRequest{
				Title:       "Review PR",
				Description: "Check pending pull requests",
			}))
		})

		Response(204, "todo deleted", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNoContent))
			})
		})

		Response(404, "todo not found", func() {
			SetParam("id", "99999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/todos/{id}/done", func() {
	Patch("Mark a todo as done", func() {
		Tag("todos")
		Parameter("id", PathParam, Integer)

		BeforeEach(func() {
			DeferCleanup(newTestEnv(model.CreateTodoRequest{
				Title:       "Deploy service",
				Description: "Push to production",
			}))
		})

		Response(200, "todo marked as done", func() {
			ResponseSchema(new(model.Response[model.Todo]))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("data"))
				Expect(resp).To(MatchJSONSchema(&model.Response[model.Todo]{}))
			})
		})

		Response(404, "todo not found", func() {
			SetParam("id", "99999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})
