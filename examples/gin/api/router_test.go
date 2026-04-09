package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/gin/api"
	. "github.com/onsi/gomega"
)

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
				Expect(resp).To(HaveStatus(http.StatusOK))
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

var _ = Path("/items/{id}", func() {
	Get("Get item by ID", func() {
		Tag("items")
		Parameter("id", PathParam, Integer)

		Response(200, "item found", func() {
			ResponseSchema(new(api.Item))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&api.Item{}))
			})
		})

		// Negative: item not found → 404.
		Response(404, "item not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Delete("Delete an item", func() {
		Tag("items")
		Parameter("id", PathParam, Integer)

		Response(204, "item deleted", func() {
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNoContent))
			})
		})
	})
})
