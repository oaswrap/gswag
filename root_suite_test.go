package gswag_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var rootSrv *httptest.Server
var rootOutDir string

func TestGswagSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gswag root suite")
}

var _ = BeforeSuite(func() {
	rootSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1,"name":"test"}`)) //nolint:errcheck
	}))

	rootOutDir = GinkgoT().TempDir()
	gswag.Init(&gswag.Config{
		Title:           "Root Suite API",
		Version:         "1.0.0",
		OutputPath:      filepath.Join(rootOutDir, "openapi.yaml"),
		CaptureExamples: true,
	})
	gswag.SetTestServer(rootSrv)
})

var _ = AfterSuite(func() {
	rootSrv.Close()
	Expect(gswag.WriteSpec()).To(Succeed())
})
