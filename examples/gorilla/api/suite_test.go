package api_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/gorilla/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gorilla example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Tasks API (Gorilla)",
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
