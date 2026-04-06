package fiber_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/fiber/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fiber example suite")
}

var _ = BeforeSuite(func() {
	gswag.Init(&gswag.Config{
		Title:      "Reviews API (Fiber)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"bearerAuth": gswag.BearerJWT(),
		},
	})
	testServer = httptest.NewServer(adaptor.FiberApp(api.NewRouter()))
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(gswag.WriteSpec()).To(Succeed())
})
