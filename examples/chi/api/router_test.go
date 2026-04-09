package api_test

import (
	"net/http"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/chi/api"
	. "github.com/onsi/gomega"
)

type OrderQuery struct {
	Status string `query:"status"`
	Limit  int    `query:"limit"`
}

var _ = Path("/orders", func() {
	Get("List all orders", func() {
		Tag("orders")
		QueryParamStruct(new(OrderQuery))
		Security("apiKey")

		Response(200, "list of orders", func() {
			ResponseSchema(new([]api.Order))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Place an order", func() {
		Tag("orders")
		RequestBody(new(api.CreateOrderRequest))
		Security("apiKey")

		Response(201, "order placed", func() {
			ResponseSchema(new(api.Order))
			SetBody(&api.CreateOrderRequest{Product: "Widget", Quantity: 3})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})

		// Negative: malformed JSON body → 400.
		Response(400, "bad request", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/orders/{id}", func() {
	Get("Get order by ID", func() {
		Tag("orders")
		Parameter("id", PathParam, String)
		Security("apiKey")

		Response(200, "order found", func() {
			ResponseSchema(new(api.Order))
			SetParam("id", "ord-1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&api.Order{}))
			})
		})

		// Negative: unknown order id → 404.
		Response(404, "order not found", func() {
			SetParam("id", "ord-999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})

	Delete("Cancel an order", func() {
		Tag("orders")
		Parameter("id", PathParam, String)
		Security("apiKey")
		Deprecated()

		Response(204, "order cancelled", func() {
			SetParam("id", "ord-2")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNoContent))
			})
		})
	})
})
