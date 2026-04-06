package echo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/echo/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Echo example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Products API (Echo)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
	})
	testServer = httptest.NewServer(api.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(WriteSpec()).To(Succeed())
})

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
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
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
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
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
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(MatchJSONSchema(&api.Product{}))
			})
		})
	})
})
