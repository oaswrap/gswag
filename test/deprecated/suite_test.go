package deprecated_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	deprecated "github.com/oaswrap/gswag/test/deprecated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deprecated Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:                       "Versioned API",
		Version:                     "2.0.0",
		Description:                 "API demonstrating deprecated operations, descriptions, and operationIds.",
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"Deprecated"},
		Tags: []TagConfig{
			{Name: "items-v1", Description: "Legacy item endpoints (deprecated)"},
			{Name: "items-v2", Description: "Current item endpoints"},
		},
	})
	testServer = httptest.NewServer(deprecated.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "deprecated.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "deprecated.json", jsonData)
})

// v1 endpoints — deprecated, with Description and OperationID documented.
var _ = Path("/v1/items", func() {
	Get("List items (v1)", func() {
		Tag("items-v1")
		OperationID("listItemsV1")
		Description("Returns all items in the legacy format. Deprecated: use GET /v2/items instead.")
		Deprecated()

		Response(200, "list of legacy items", func() {
			ResponseSchema(new([]deprecated.LegacyItem))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})

var _ = Path("/v1/items/{id}", func() {
	Get("Get item by ID (v1)", func() {
		Tag("items-v1")
		OperationID("getItemV1")
		Description("Returns a single item by numeric ID. Deprecated: use GET /v2/items/{id} instead.")
		Deprecated()
		Parameter("id", PathParam, Integer)

		Response(200, "legacy item", func() {
			ResponseSchema(new(deprecated.LegacyItem))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&deprecated.LegacyItem{}))
			})
		})
	})
})

// v2 endpoints — current, not deprecated.
var _ = Path("/v2/items", func() {
	Get("List items (v2)", func() {
		Tag("items-v2")
		OperationID("listItemsV2")
		Description("Returns all items in the current format.")

		Response(200, "list of items", func() {
			ResponseSchema(new([]deprecated.ItemV2))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})
})
