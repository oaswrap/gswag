package chi_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/chi/api"
)

// OrderQuery is a typed query param struct for listing orders.
type OrderQuery struct {
	Status string `query:"status"`
	Limit  int    `query:"limit"`
}

var _ = Describe("/orders", func() {

	Context("GET /orders", func() {
		It("returns a list of orders", func() {
			res := gswag.GET("/orders").
				WithTag("orders").
				WithSummary("List all orders").
				WithQueryParamStruct(new(OrderQuery)).
				WithSecurity("apiKey").
				ExpectResponseBody(new([]api.Order)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.HaveNonEmptyBody())
		})
	})

	Context("POST /orders", func() {
		It("creates an order and returns 201", func() {
			res := gswag.POST("/orders").
				WithTag("orders").
				WithSummary("Place an order").
				WithRequestBody(&api.CreateOrderRequest{Product: "Widget", Quantity: 3}).
				ExpectResponseBodyFor(http.StatusCreated, new(api.Order)).
				WithSecurity("apiKey").
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusCreated))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("GET /orders/{id}", func() {
		It("returns a single order by string ID", func() {
			res := gswag.GET("/orders/{id}").
				WithPathParam("id", "ord-1").
				WithTag("orders").
				WithSummary("Get order by ID").
				WithSecurity("apiKey").
				ExpectResponseBody(new(api.Order)).
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusOK))
			Expect(res).To(gswag.ContainJSONKey("id"))
		})
	})

	Context("DELETE /orders/{id}", func() {
		It("deletes an order and returns 204", func() {
			res := gswag.DELETE("/orders/{id}").
				WithPathParam("id", "ord-2").
				WithTag("orders").
				WithSummary("Cancel an order").
				WithSecurity("apiKey").
				AsDeprecated().
				Do(testServer)

			Expect(res).To(gswag.HaveStatus(http.StatusNoContent))
		})
	})
})
