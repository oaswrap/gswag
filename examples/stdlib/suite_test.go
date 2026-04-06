package stdlib_test

import (
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/stdlib/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "stdlib example suite")
}

var _ = BeforeSuite(func() {
	gswag.Init(&gswag.Config{
		Title:      "Users API",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		// Enable opt-in test-time response validation for examples.
		EnforceResponseValidation: true,
		SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
			"apiKey": gswag.APIKeyHeader("X-API-Key"),
		},
	})
	testServer = httptest.NewServer(api.NewRouter())
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(gswag.WriteSpec()).To(Succeed())
})
