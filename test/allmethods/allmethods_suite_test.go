package allmethods_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	"github.com/oaswrap/gswag/test/allmethods"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AllMethods Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:                       "All Methods API",
		Version:                     "1.0.0",
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"Allmethods"},
	})
	testServer = httptest.NewServer(allmethods.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	// Golden: compare the complete YAML spec
	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "allmethods.yaml", yamlData)

	// Golden: compare the complete JSON spec.
	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "allmethods.json", jsonData)
})

var _ = Path("/allmethods", func() {
	Get("get allmethods", func() {
		Tag("allmethods")
		OperationID("get_allmethods")

		Response(200, "successful operation", func() {
			ResponseSchema(new(map[string]string))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Post("post allmethods", func() {
		Tag("allmethods")
		OperationID("post_allmethods")
		RequestBody(new(allmethods.AllMethodsModel))

		Response(200, "successful operation", func() {
			ResponseSchema(new(allmethods.AllMethodsModel))
			SetBody(&allmethods.AllMethodsModel{ID: 1, Name: "one"})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Put("put allmethods", func() {
		Tag("allmethods")
		OperationID("put_allmethods")
		RequestBody(new(allmethods.AllMethodsModel))

		Response(200, "successful operation", func() {
			ResponseSchema(new(allmethods.AllMethodsModel))
			SetBody(&allmethods.AllMethodsModel{ID: 2, Name: "two"})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Patch("patch allmethods", func() {
		Tag("allmethods")
		OperationID("patch_allmethods")
		RequestBody(new(allmethods.AllMethodsModel))

		Response(200, "successful operation", func() {
			ResponseSchema(new(allmethods.AllMethodsModel))
			SetBody(&allmethods.AllMethodsModel{ID: 3, Name: "three"})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Delete("delete allmethods", func() {
		Tag("allmethods")
		OperationID("delete_allmethods")

		Response(200, "successful operation", func() {
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
