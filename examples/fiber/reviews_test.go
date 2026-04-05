package fiber_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/fiber/api"
)

var _ = Describe("/reviews", func() {

	Context("GET /reviews", func() {
		It("returns a list of reviews", func() {
			res := gswag.GET("/reviews").
				WithTag("reviews").
				WithSummary("List all reviews").
				WithBearerAuth().
				ExpectResponseBody(new([]api.Review)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.HaveNonEmptyBody())
		})
	})

	Context("POST /reviews", func() {
		It("creates a review and returns 201", func() {
			res := gswag.POST("/reviews").
				WithTag("reviews").
				WithSummary("Submit a review").
				WithBearerAuth().
				WithRequestBody(&api.CreateReviewRequest{Author: "Charlie", Rating: 5, Comment: "Perfect"}).
				ExpectResponseBodyFor(http.StatusCreated, new(api.Review)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusCreated))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("GET /reviews/{id}", func() {
		It("returns a review by integer ID", func() {
			res := gswag.GET("/reviews/{id}").
				WithPathParam("id", "1").
				WithTag("reviews").
				WithSummary("Get review by ID").
				WithBearerAuth().
				ExpectResponseBody(new(api.Review)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.MatchJSONSchema(&api.Review{}))
		})
	})
})
