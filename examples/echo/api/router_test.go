package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/echo/api"
	. "github.com/onsi/gomega"
)

type ProductQuery struct {
	Category string  `query:"category"`
	MaxPrice float64 `query:"max_price"`
}

var _ = Path("/products", func() {
	Get("List all products", func() {
		Tag("products")
		QueryParamStruct(new(ProductQuery))

		Response(200, "list of products", func() {
			ResponseSchema(new([]api.Product))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create a product", func() {
		Tag("products")
		RequestBody(new(api.CreateProductRequest))

		Response(201, "product created", func() {
			ResponseSchema(new(api.Product))
			SetBody(&api.CreateProductRequest{Title: "Headphones", Category: "Electronics", Price: 79.99})
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

var _ = Path("/products/{id}", func() {
	Get("Get product by ID", func() {
		Tag("products")
		Parameter("id", PathParam, Integer)

		Response(200, "product found", func() {
			ResponseSchema(new(api.Product))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(MatchJSONSchema(&api.Product{}))
			})
		})
	})
})
