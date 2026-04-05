package stdlib_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
)

var _ = Describe("/api/users", func() {

	Context("GET /api/users", func() {
		It("returns a list of users", func() {
			res := gswag.GET("/api/users").
				WithTag("users").
				WithSummary("List all users").
				ExpectResponseBody(new([]api.User)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.HaveHeader("Content-Type", "application/json"))
			Expect(res).To(gswag.HaveNonEmptyBody())
		})
	})

	Context("POST /api/users", func() {
		It("creates a user and returns 201", func() {
			res := gswag.POST("/api/users").
				WithTag("users").
				WithSummary("Create a user").
				WithRequestBody(&api.User{Name: "Charlie", Email: "charlie@example.com"}).
				ExpectResponseBodyFor(http.StatusCreated, new(api.User)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusCreated))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("GET /api/users/{id}", func() {
		It("returns a single user by ID", func() {
			res := gswag.GET("/api/users/{id}").
				WithPathParam("id", "1").
				WithTag("users").
				WithSummary("Get user by ID").
				ExpectResponseBody(new(api.User)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.ContainJSONKey("id"))
			Expect(res).To(gswag.ContainJSONKey("name"))
		})
	})
})
