package stdlib_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/gomega"
)

// Delete /api/users/{id} with Bearer JWT auth requirement — spec documentation.
var _ = Path("/api/users/{id}", func() {
	Delete("Delete user by ID", func() {
		Tag("users")
		OperationID("deleteUserById")
		BearerAuth()
		Parameter("id", PathParam, String)

		Response(200, "user deleted", func() {
			ResponseSchema(new(api.User))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
