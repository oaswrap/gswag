package fiber_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/fiber/api"
	. "github.com/onsi/gomega"
)

var _ = Path("/reviews", func() {
	Get("List all reviews", func() {
		Tag("reviews")
		BearerAuth()

		Response(200, "list of reviews", func() {
			ResponseSchema(new([]api.Review))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a review", func() {
		Tag("reviews")
		BearerAuth()
		RequestBody(new(api.CreateReviewRequest))

		Response(201, "review created", func() {
			ResponseSchema(new(api.Review))
			SetBody(&api.CreateReviewRequest{Author: "Test User", Rating: 5, Comment: "Excellent!"})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/reviews/{id}", func() {
	Get("Get review by ID", func() {
		Tag("reviews")
		BearerAuth()
		Parameter("id", PathParam, Integer)

		Response(200, "review found", func() {
			ResponseSchema(new(api.Review))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})
