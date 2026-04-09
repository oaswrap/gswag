package inlinerefs_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	inlinerefs "github.com/oaswrap/gswag/test/inline_refs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InlineRefs Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:      "InlineRefs API",
		Version:    "1.0.0",
		OutputPath: filepath.Join(rootOutDir, "openapi.yaml"),
		// InlineRefs: schemas are expanded in-place rather than hoisted to
		// #/components/schemas and referenced via $ref.
		InlineRefs: true,
	})
	testServer = httptest.NewServer(inlinerefs.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())

	// With InlineRefs the generated spec must not contain any $ref pointers.
	Expect(string(yamlData)).NotTo(ContainSubstring("$ref:"))
	// Nested field names must appear inline.
	Expect(string(yamlData)).To(ContainSubstring("street"))
	Expect(string(yamlData)).To(ContainSubstring("city"))

	golden.Check(GinkgoT(), "inline_refs.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "inline_refs.json", jsonData)
})

var _ = Path("/customers", func() {
	Get("List customers", func() {
		Tag("customers")
		OperationID("listCustomers")

		Response(200, "list of customers", func() {
			ResponseSchema(new([]inlinerefs.Customer))
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusOK))
				Expect(resp).To(HaveNonEmptyBody())
			})
		})
	})

	Post("Create customer", func() {
		Tag("customers")
		OperationID("createCustomer")
		RequestBody(new(inlinerefs.Customer))

		Response(201, "customer created", func() {
			ResponseSchema(new(inlinerefs.Customer))
			SetBody(&inlinerefs.Customer{
				Name:    "Bob",
				Address: inlinerefs.Address{Street: "456 Oak Ave", City: "Shelbyville", Country: "US"},
				Contact: inlinerefs.Contact{Email: "bob@example.com", Phone: "555-0200"},
			})
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusCreated))
				Expect(resp).To(ContainJSONKey("id"))
				Expect(resp).To(MatchJSONSchema(&inlinerefs.Customer{}))
			})
		})

		// Negative: malformed JSON body → 400.
		Response(400, "invalid input", func() {
			SetRawBody([]byte("not json"), "application/json")
			RunTest(func(resp *http.Response) {
				Expect(resp).To(HaveStatus(http.StatusBadRequest))
				Expect(resp).To(ContainJSONKey("error"))
			})
		})
	})
})
