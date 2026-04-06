package init_example_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "My API",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"bearerAuth": BearerJWT(),
		},
	})
	// TODO: start your server here, for example:
	//   import yourpkg "github.com/your/module/path"
	//   testServer = httptest.NewServer(yourpkg.NewRouter())
	//   SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())
})
