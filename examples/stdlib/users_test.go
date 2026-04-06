package stdlib_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/api/users", func() {
	Get("List all users", func() {
		Tag("users")

		Response(200, "list of users", func() {
			ResponseSchema(new([]api.User))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a user", func() {
		Tag("users")
		RequestBody(new(api.User))

		Response(201, "user created", func() {
			ResponseSchema(new(api.User))
			SetBody(&api.User{Name: "Charlie", Email: "charlie@example.com"})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/api/users/{id}", func() {
	Get("Get user by ID", func() {
		Tag("users")
		Parameter("id", PathParam, String)

		Response(200, "user found", func() {
			ResponseSchema(new(api.User))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(ContainJSONKey("name"))
			})
		})
	})
})
