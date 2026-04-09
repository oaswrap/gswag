package api_test

import (
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
	Expect(WriteSpecTo("../docs/openapi.yaml", YAML)).To(Succeed())
})
