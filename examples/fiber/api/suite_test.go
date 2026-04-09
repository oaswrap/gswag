package api_test

import (
	"net"
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
	Expect(WriteSpecTo("../docs/openapi.yaml", YAML)).To(Succeed())
})
