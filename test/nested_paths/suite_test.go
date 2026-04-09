package nestedpaths_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	nestedpaths "github.com/oaswrap/gswag/test/nested_paths"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NestedPaths Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:                       "Nested Paths API",
		Version:                     "1.0.0",
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"Nestedpaths"},
	})
	testServer = httptest.NewServer(nestedpaths.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "nested_paths.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "nested_paths.json", jsonData)
})

// Deeply nested Path() calls — the path stack must concatenate correctly.
var _ = Path("/api", func() {
	Path("/v1", func() {
		Path("/users", func() {
			Get("List users", func() {
				Tag("users")
				OperationID("listUsers")

				Response(200, "list of users", func() {
					ResponseSchema(new([]nestedpaths.User))
					RunTest(func(resp *http.Response) {
						Expect(resp).To(HaveStatus(http.StatusOK))
						Expect(resp).To(HaveNonEmptyBody())
					})
				})
			})

			Path("/{id}", func() {
				Get("Get user by ID", func() {
					Tag("users")
					OperationID("getUser")
					Parameter("id", PathParam, String)

					Response(200, "user found", func() {
						ResponseSchema(new(nestedpaths.User))
						SetParam("id", "u1")
						RunTest(func(resp *http.Response) {
							Expect(resp).To(HaveStatus(http.StatusOK))
							Expect(resp).To(ContainJSONKey("id"))
							Expect(resp).To(MatchJSONSchema(&nestedpaths.User{}))
						})
					})
				})

				Path("/orders", func() {
					Get("List orders for user", func() {
						Tag("orders")
						OperationID("listUserOrders")
						Parameter("id", PathParam, String)

						Response(200, "list of orders", func() {
							ResponseSchema(new([]nestedpaths.Order))
							SetParam("id", "u1")
							RunTest(func(resp *http.Response) {
								Expect(resp).To(HaveStatus(http.StatusOK))
								Expect(resp).To(HaveNonEmptyBody())
							})
						})
					})

					Path("/{orderId}", func() {
						Get("Get order", func() {
							Tag("orders")
							OperationID("getOrder")
							Parameter("id", PathParam, String)
							Parameter("orderId", PathParam, String)

							Response(200, "order found", func() {
								ResponseSchema(new(nestedpaths.Order))
								SetParam("id", "u1")
								SetParam("orderId", "o1")
								RunTest(func(resp *http.Response) {
									Expect(resp).To(HaveStatus(http.StatusOK))
									Expect(resp).To(ContainJSONKey("id"))
								})
							})
						})

						Path("/items", func() {
							Get("List order items", func() {
								Tag("orders")
								OperationID("listOrderItems")
								Parameter("id", PathParam, String)
								Parameter("orderId", PathParam, String)

								Response(200, "order items", func() {
									ResponseSchema(new([]nestedpaths.OrderItem))
									SetParam("id", "u1")
									SetParam("orderId", "o1")
									RunTest(func(resp *http.Response) {
										Expect(resp).To(HaveStatus(http.StatusOK))
										Expect(resp).To(HaveNonEmptyBody())
									})
								})
							})
						})
					})
				})
			})
		})
	})
})
