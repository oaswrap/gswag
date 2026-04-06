package stdlib_test

import (
	"net/http"

	"github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("/api/users (auth)", func() {

	Context("GET /api/users with bearer auth", func() {
		It("returns 200 — operation is marked as requiring bearerAuth", func() {
			res := gswag.GET("/api/users").
				WithTag("users").
				WithSummary("List users (authenticated)").
				WithOperationID("listUsersAuth").
				WithBearerAuth().
				ExpectResponseBody(new([]struct {
					ID    string `json:"id"`
					Email string `json:"email"`
				})).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
		})
	})
})
