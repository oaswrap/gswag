package querystruct_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	querystruct "github.com/oaswrap/gswag/test/query_struct"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "QueryStruct Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:                       "QueryStruct API",
		Version:                     "1.0.0",
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"Querystruct"},
	})
	testServer = httptest.NewServer(querystruct.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "query_struct.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "query_struct.json", jsonData)
})

var _ = Path("/products", func() {
	Get("List products", func() {
		Tag("products")
		OperationID("listProducts")
		// QueryParamStruct derives all query parameters from the struct tags.
		QueryParamStruct(new(querystruct.ProductQuery))

		Response(200, "list of products", func() {
			ResponseSchema(new([]querystruct.Product))
			// ResponseHeader documents response headers returned by the server.
			ResponseHeader("X-Total-Count", 0)
			ResponseHeader("X-Page", 0)
			ResponseHeader("X-Page-Size", 0)
			SetQueryParam("page", "1")
			SetQueryParam("page_size", "10")
			SetQueryParam("tag", "hardware")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp.Header.Get("X-Total-Count")).NotTo(BeEmpty())
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/products/{id}", func() {
	Get("Get product by ID", func() {
		Tag("products")
		OperationID("getProduct")
		Parameter("id", PathParam, Integer)

		Response(200, "product found", func() {
			ResponseSchema(new(querystruct.Product))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&querystruct.Product{}))
			})
		})

		// Negative: non-numeric id → 400.
		Response(400, "invalid id", func() {
			SetParam("id", "abc")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})

		// Negative: id not found → 404.
		Response(404, "product not found", func() {
			SetParam("id", "9999")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusNotFound))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})
