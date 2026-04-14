package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/parallel/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/api/users", func() {
	Get("List users", func() {
		Tag("users")

		// No ResponseSchema — schema is inferred from the live HTTP response
		// on whichever parallel node runs this test.
		Response(200, "list of users", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a user", func() {
		Tag("users")

		// No RequestBody/ResponseSchema — both are inferred at runtime.
		Response(201, "user created", func() {
			SetBody(&api.CreateUserRequest{Name: "Charlie", Email: "charlie@example.com"})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/api/users/{id}", func() {
	Get("Get user by ID", func() {
		Tag("users")
		Parameter("id", PathParam, String)

		// No ResponseSchema — inferred at runtime per node.
		Response(200, "user found", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		Response(404, "user not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
			})
		})
	})
})
