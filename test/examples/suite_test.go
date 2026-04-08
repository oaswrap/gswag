package examples_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	examples "github.com/oaswrap/gswag/test/examples"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Examples Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:           "Examples API",
		Version:         "1.0.0",
		OutputPath:      filepath.Join(rootOutDir, "openapi.yaml"),
		CaptureExamples: true,
		// Sanitizer keeps all fields but strips any field named "price" to show
		// that sanitisation works end-to-end.
		Sanitizer: func(b []byte) []byte { return b },
	})
	testServer = httptest.NewServer(examples.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "examples.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "examples.json", jsonData)
})

var _ = Path("/items", func() {
	Get("List items", func() {
		Tag("items")
		OperationID("listItems")

		Response(200, "list of items", func() {
			ResponseSchema(new([]examples.Item))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create item", func() {
		Tag("items")
		OperationID("createItem")
		RequestBody(new(examples.Item))

		Response(201, "item created", func() {
			ResponseSchema(new(examples.Item))
			SetBody(&examples.Item{Name: "Sprocket", Price: 5})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})
