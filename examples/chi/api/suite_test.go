package api_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/chi/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chi example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Orders API (Chi)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"apiKey": APIKeyHeader("X-API-Key"),
		},
	})
	testServer = httptest.NewServer(api.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(WriteSpecTo("../docs/openapi.yaml", YAML)).To(Succeed())
})
