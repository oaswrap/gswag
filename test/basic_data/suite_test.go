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
	"github.com/oaswrap/gswag/test/util"
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
				Int:     util.Ptr(1),
				Int8:    util.Ptr(int8(2)),
				Int16:   util.Ptr(int16(3)),
				Int32:   util.Ptr(int32(4)),
				Int64:   util.Ptr(int64(5)),
				Uint:    util.Ptr(uint(6)),
				Uint8:   util.Ptr(uint8(7)),
				Uint16:  util.Ptr(uint16(8)),
				Uint32:  util.Ptr(uint32(9)),
				Uint64:  util.Ptr(uint64(10)),
				Float32: util.Ptr(float32(1.23)),
				Float64: util.Ptr(float64(4.56)),
				Byte:    util.Ptr(byte('a')),
				Rune:    util.Ptr(rune('b')),
				String:  util.Ptr("test"),
				Bool:    util.Ptr(true),
			})
			RunTest(func(r *http.Response) {
				Expect(r.StatusCode).To(Equal(200))
			})
		})
	})
})
