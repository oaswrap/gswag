package basicdata_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	basicdata "github.com/oaswrap/gswag/test/basic_data"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:   "My API",
		Version: "1.0.0",
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"bearerAuth": BearerJWT(),
		},
		OutputPath:                  filepath.Join(rootOutDir, "openapi.yaml"),
		StripDefinitionNamePrefixes: []string{"BasicData"},
	})
	testServer = httptest.NewServer(basicdata.NewRouter())
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
	golden.Check(GinkgoT(), "basic_data.yaml", yamlData)

	// Golden: compare the complete JSON spec.
	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "basic_data.json", jsonData)
})

var _ = Path("/basicdata", func() {
	Get("get basic data types", func() {
		Tag("basicdata")
		OperationID("get_basicdata")

		Response(200, "successful operation", func() {
			ResponseSchema(new(basicdata.AllBasicDataTypes))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
			})
		})
	})
	Post("post basic data types", func() {
		Tag("basicdata")
		OperationID("post_basicdata")
		RequestBody(new(basicdata.AllBasicDataTypes))

		Response(200, "successful operation", func() {
			ResponseSchema(new(basicdata.AllBasicDataTypes))
			SetBody(&basicdata.AllBasicDataTypes{
				Int:     1,
				Int8:    2,
				Int16:   3,
				Int32:   4,
				Int64:   5,
				Uint:    6,
				Uint8:   7,
				Uint16:  8,
				Uint32:  9,
				Uint64:  10,
				Float32: 1.23,
				Float64: 4.56,
				Byte:    'a',
				Rune:    'b',
				String:  "test",
				Bool:    true,
			})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
			})
		})
		Response(400, "invalid input", func() {
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(400))
				Expect(r).To(ContainJSONKey("error"))
			})
		})
		Response(400, "invalid input", func() {
			SetBody(map[string]any{
				"int": "not an int",
			})
			RunTest(func(r *http.Response) {
				Expect(r).To(ContainJSONKey("error"))
			})
		})
	})
})

var _ = Path("/basicdata-pointers", func() {
	Get("get basic data types with pointers", func() {
		Tag("basicdata")
		OperationID("get_basicdata_pointers")

		Response(200, "successful operation", func() {
			ResponseSchema(new(basicdata.AllBasicDataTypesPointers))
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
			})
		})
	})
	Post("post basic data types with pointers", func() {
		Tag("basicdata")
		OperationID("post_basicdata_pointers")
		RequestBody(new(basicdata.AllBasicDataTypesPointers))

		Response(200, "successful operation", func() {
			ResponseSchema(new(basicdata.AllBasicDataTypesPointers))
			SetBody(&basicdata.AllBasicDataTypesPointers{
				Int:     ptr(1),
				Int8:    ptr(int8(2)),
				Int16:   ptr(int16(3)),
				Int32:   ptr(int32(4)),
				Int64:   ptr(int64(5)),
				Uint:    ptr(uint(6)),
				Uint8:   ptr(uint8(7)),
				Uint16:  ptr(uint16(8)),
				Uint32:  ptr(uint32(9)),
				Uint64:  ptr(uint64(10)),
				Float32: ptr(float32(1.23)),
				Float64: ptr(float64(4.56)),
				Byte:    ptr(byte('a')),
				Rune:    ptr(rune('b')),
				String:  ptr("test"),
				Bool:    ptr(true),
			})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
			})
		})
	})
})

func ptr[T any](v T) *T {
	return &v
}
