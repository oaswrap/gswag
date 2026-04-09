package api_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/petstore/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Petstore example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:          "Swagger Petstore - OpenAPI 3.0",
		Description:    "This is a sample Pet Store Server based on the OpenAPI 3.0 specification.  You can find out more about\nSwagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach!\nYou can now help us improve the API whether it's by making changes to the definition itself or to the code.\nThat way, with time, we can improve the API in general, and expose some of the new features in OAS3.\n\nSome useful links:\n- [The Pet Store repository](https://github.com/swagger-api/swagger-petstore)\n- [The source API definition for the Pet Store](https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml)",
		TermsOfService: "https://swagger.io/terms/",
		Contact: &ContactConfig{
			Email: "apiteam@swagger.io",
		},
		License: &LicenseConfig{
			Name: "Apache 2.0",
			URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
		},
		ExternalDocs: &ExternalDocsConfig{
			Description: "Find out more about Swagger",
			URL:         "https://swagger.io",
		},
		Tags: []TagConfig{
			{
				Name:        "pet",
				Description: "Everything about your Pets",
				ExternalDocs: &ExternalDocsConfig{
					Description: "Find out more",
					URL:         "https://swagger.io",
				},
			},
			{
				Name:        "store",
				Description: "Access to Petstore orders",
				ExternalDocs: &ExternalDocsConfig{
					Description: "Find out more about our store",
					URL:         "https://swagger.io",
				},
			},
			{
				Name:        "user",
				Description: "Operations about user",
			},
		},
		Version:    "1.0.27",
		OutputPath: "./docs/openapi.yaml",
		Servers: []ServerConfig{{
			URL: "/api/v3",
		}},
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"petstore_auth": OAuth2Implicit("https://petstore3.swagger.io/oauth/authorize", map[string]string{
				"write:pets": "modify pets in your account",
				"read:pets":  "read your pets",
			}),
			"api_key": APIKeyHeader("api_key"),
		},
	})
	testServer = httptest.NewServer(api.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	testServer.Close()
	Expect(WriteSpecTo("../docs/openapi.yaml", YAML)).To(Succeed())
})
