package stdlib_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/gomega"
)

// PATCH /api/users/{id} — documents a partial update operation.
var _ = Path("/api/users/{id}", func() {
	Patch("Update user", func() {
		Tag("users")
		Parameter("id", PathParam, String)
		RequestBody(new(api.User))

		Response(200, "user updated", func() {
			ResponseSchema(new(api.User))
			SetParam("id", "1")
			SetBody(&api.User{Name: "Updated"})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
