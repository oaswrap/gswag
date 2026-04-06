package gin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/gin/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gin example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Items API (Gin)",
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

// ------------------------- merged from items_test.go -------------------------

type ListQuery struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
}

var _ = Path("/items", func() {
	Get("List all items", func() {
		Tag("items")
		QueryParamStruct(new(ListQuery))

		Response(200, "list of items", func() {
			ResponseSchema(new([]api.Item))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create an item", func() {
		Tag("items")
		RequestBody(new(api.CreateItemRequest))

		Response(201, "item created", func() {
			ResponseSchema(new(api.Item))
			SetBody(&api.CreateItemRequest{Name: "Wrench", Price: 9.99})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/items/{id}", func() {
	Get("Get item by ID", func() {
		Tag("items")
		Parameter("id", PathParam, Integer)

		Response(200, "item found", func() {
			ResponseSchema(new(api.Item))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})

	Delete("Delete an item", func() {
		Tag("items")
		Parameter("id", PathParam, Integer)

		Response(204, "item deleted", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})
	})
})
