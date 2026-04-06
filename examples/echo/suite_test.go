package echo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag"
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
	gswag.Init(&gswag.Config{
		Title:      "Products API (Echo)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
	})
	testServer = httptest.NewServer(api.NewRouter())
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(gswag.WriteSpec()).To(Succeed())
})
