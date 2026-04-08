package excludehidden_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/internal/golden"
	excludehidden "github.com/oaswrap/gswag/test/exclude_hidden"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server
var rootOutDir string

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExcludeHidden Suite")
}

var _ = BeforeSuite(func() {
	rootOutDir = GinkgoT().TempDir()

	Init(&Config{
		Title:   "Exclude Hidden API",
		Version: "1.0.0",
		OutputPath: filepath.Join(rootOutDir, "openapi.yaml"),
		// /internal and /admin/* are excluded from the spec entirely.
		ExcludePaths: []string{"/internal", "/admin/*"},
	})
	testServer = httptest.NewServer(excludehidden.NewRouter())
	SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())

	yamlData, err := os.ReadFile(filepath.Join(rootOutDir, "openapi.yaml"))
	Expect(err).NotTo(HaveOccurred())

	// Verify excluded paths are absent from the spec.
	Expect(string(yamlData)).NotTo(ContainSubstring("/internal"))
	Expect(string(yamlData)).NotTo(ContainSubstring("/admin/"))
	// Verify hidden path is absent from the spec.
	Expect(string(yamlData)).NotTo(ContainSubstring("/secret"))
	// Verify public path is present.
	Expect(strings.Contains(string(yamlData), "/public")).To(BeTrue())

	golden.Check(GinkgoT(), "exclude_hidden.yaml", yamlData)

	jsonPath := filepath.Join(rootOutDir, "openapi.json")
	Expect(WriteSpecTo(jsonPath, JSON)).To(Succeed())
	jsonData, err := os.ReadFile(jsonPath)
	Expect(err).NotTo(HaveOccurred())
	golden.Check(GinkgoT(), "exclude_hidden.json", jsonData)
})

// /public — appears in the spec.
var _ = Path("/public", func() {
	Get("Get public resource", func() {
		Tag("public")
		OperationID("getPublic")

		Response(200, "public resource", func() {
			ResponseSchema(new(excludehidden.PublicResource))
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})

// /internal — excluded via ExcludePaths; test still executes.
var _ = Path("/internal", func() {
	Get("Get internal resource", func() {
		Tag("internal")
		OperationID("getInternal")

		Response(200, "internal resource", func() {
			ResponseSchema(new(excludehidden.InternalResource))
			RunTest(func(resp *http.Response) {
				// The HTTP request still runs even though the operation is excluded.
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})

// /secret — excluded via Hidden(); test still executes.
var _ = Path("/secret", func() {
	Get("Get secret resource", func() {
		Hidden()
		Tag("secret")

		Response(200, "secret resource", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})

// /admin/metrics — excluded via ExcludePaths prefix pattern.
var _ = Path("/admin/metrics", func() {
	Get("Get admin metrics", func() {
		Tag("admin")
		OperationID("getAdminMetrics")

		Response(200, "metrics", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})

// /admin/health — also excluded by the /admin/* prefix.
var _ = Path("/admin/health", func() {
	Get("Get admin health", func() {
		Tag("admin")
		OperationID("getAdminHealth")

		Response(200, "health", func() {
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
