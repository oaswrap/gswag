package gin_test

import (
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag"
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
	gswag.Init(&gswag.Config{
		Title:      "Items API (Gin)",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
	})
	testServer = httptest.NewServer(api.NewRouter())
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(gswag.WriteSpec()).To(Succeed())
})
