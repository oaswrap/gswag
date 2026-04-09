package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/gorilla/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/tasks", func() {
	Get("List all tasks", func() {
		Tag("tasks")

		Response(200, "list of tasks", func() {
			ResponseSchema(new([]api.Task))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a task", func() {
		Tag("tasks")
		RequestBody(new(api.CreateTaskRequest))

		Response(201, "task created", func() {
			ResponseSchema(new(api.Task))
			SetBody(&api.CreateTaskRequest{Title: "Read documentation"})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "bad request", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/tasks/{id}", func() {
	Get("Get task by ID", func() {
		Tag("tasks")
		Parameter("id", PathParam, Integer)

		Response(200, "task found", func() {
			ResponseSchema(new(api.Task))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&api.Task{}))
			})
		})

		// Negative: unknown task id → 404.
		Response(404, "task not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Delete("Delete a task", func() {
		Tag("tasks")
		Parameter("id", PathParam, Integer)

		Response(204, "task deleted", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNoContent))
			})
		})
	})
})
