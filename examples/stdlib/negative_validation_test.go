package stdlib_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validation failure example", func() {
	It("panics when response doesn't match declared schema", func() {
		// Start a server that returns intentionally invalid data for declared schema.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// api.User expects numeric id and string fields — send wrong types.
			_, _ = w.Write([]byte(`{"id":"not-a-number","name":123}`))
		}))
		defer srv.Close()

		// The builder declares the response model as api.User; since the body does
		// not match, EnforceResponseValidation (enabled in the suite) should cause a panic.
		Expect(func() {
			gswag.GET("/bad").
				WithTag("bad").
				WithSummary("Bad response").
				ExpectResponseBody(new(api.User)).
				Do(srv)
		}).To(Panic())
	})
})
