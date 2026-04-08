package contenttype_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	contenttype "github.com/oaswrap/gswag/test/content_type"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ContentType Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:      "Content-Type API",
		Version:    "1.0.0",
		OutputPath: filepath.Join(rootOutDir, "openapi.yaml"),
	})
	testServer = httptest.NewServer(contenttype.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "content_type.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "content_type.json", jsonData)
})

// buildMultipart creates a minimal multipart/form-data body with a single file part.
func buildMultipart() ([]byte, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, _ := w.CreateFormFile("file", "hello.txt")
	fw.Write([]byte("hello world")) //nolint:errcheck
	w.Close()
	return buf.Bytes(), w.FormDataContentType()
}

var _ = Path("/upload", func() {
	Post("Upload a file", func() {
		Tag("upload")
		OperationID("uploadFile")
		// Declare that the request body is multipart, not JSON.
		Consumes("multipart/form-data")
		RequestBody(new(contenttype.UploadRequest))

		Response(200, "upload successful", func() {
			ResponseSchema(new(contenttype.UploadResponse))
			body, ct := buildMultipart()
			SetRawBody(body, ct)
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp).To(ContainJSONKey("id"))
			})
		})
	})
})

var _ = Path("/report", func() {
	Get("Get report", func() {
		Tag("report")
		OperationID("getReport")
		// Document that the endpoint can serve both JSON and CSV.
		Produces("application/json", "text/csv")

		Response(200, "report data", func() {
			ResponseSchema(new(contenttype.Report))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
