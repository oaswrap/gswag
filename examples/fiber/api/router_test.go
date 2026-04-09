package api_test

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
				Expect(resp).To(HaveStatus(http.StatusOK))
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
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
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
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&api.Review{}))
			})
		})

		// Negative: unknown review id → 404.
		Response(404, "review not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})
