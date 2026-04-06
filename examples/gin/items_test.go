package gin_test

import (
	"net/http"

	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/gin/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ListQuery demonstrates typed query parameter schema generation.
type ListQuery struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Sort  string `query:"sort"`
}

var _ = Describe("/items", func() {

	Context("GET /items", func() {
		It("returns a list of items", func() {
			res := gswag.GET("/items").
				WithTag("items").
				WithSummary("List all items").
				WithQueryParamStruct(new(ListQuery)).
				ExpectResponseBody(new([]api.Item)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.HaveNonEmptyBody())
		})
	})

	Context("POST /items", func() {
		It("creates an item and returns 201", func() {
			res := gswag.POST("/items").
				WithTag("items").
				WithSummary("Create an item").
				WithRequestBody(&api.CreateItemRequest{Name: "Doohickey", Price: 4.99}).
				ExpectResponseBodyFor(http.StatusCreated, new(api.Item)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusCreated))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("GET /items/{id}", func() {
		It("returns a single item by ID (integer path param)", func() {
			res := gswag.GET("/items/{id}").
				WithPathParam("id", "1").
				WithTag("items").
				WithSummary("Get item by ID").
				ExpectResponseBody(new(api.Item)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("DELETE /items/{id}", func() {
		It("deletes an item and returns 204 (JSON inference fallback)", func() {
			res := gswag.DELETE("/items/{id}").
				WithPathParam("id", "2").
				WithTag("items").
				WithSummary("Delete an item").
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusNoContent))
		})
	})
})
