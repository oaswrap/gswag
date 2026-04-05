package echo_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/echo/api"
)

// ProductQuery demonstrates typed query parameter schema generation.
type ProductQuery struct {
	Category string `query:"category"`
	MaxPrice float64 `query:"max_price"`
}

var _ = Describe("/products", func() {

	Context("GET /products", func() {
		It("returns a list of products", func() {
			res := gswag.GET("/products").
				WithTag("products").
				WithSummary("List all products").
				WithQueryParamStruct(new(ProductQuery)).
				ExpectResponseBody(new([]api.Product)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.HaveNonEmptyBody())
		})
	})

	Context("POST /products", func() {
		It("creates a product and returns 201", func() {
			res := gswag.POST("/products").
				WithTag("products").
				WithSummary("Create a product").
				WithRequestBody(&api.CreateProductRequest{Title: "Headphones", Category: "Electronics", Price: 79.99}).
				ExpectResponseBodyFor(http.StatusCreated, new(api.Product)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusCreated))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("GET /products/{id}", func() {
		It("returns a single product by ID (integer path param)", func() {
			res := gswag.GET("/products/{id}").
				WithPathParam("id", "1").
				WithTag("products").
				WithSummary("Get product by ID").
				ExpectResponseBody(new(api.Product)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.MatchJSONSchema(&api.Product{}))
		})
	})
})
