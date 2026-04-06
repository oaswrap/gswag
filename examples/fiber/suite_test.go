package fiber_test

import (
	"net"
	"net/http"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/fiber/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var fiberURL string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fiber example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Reviews API (Fiber)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"bearerAuth": BearerJWT(),
		},
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())
	fiberURL = "http://" + ln.Addr().String()
	SetTestServer(fiberURL)

	app := api.NewRouter()
	go func() { _ = app.Listener(ln) }() //nolint:errcheck
})

var _ = AfterSuite(func() {
	Expect(WriteSpec()).To(Succeed())
})

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
