package api_test

import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	"github.com/oaswrap/gswag/examples/parallel/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// partialDir is the shared directory where each parallel node writes its
// partial spec file (node-1.json, node-2.json, …).
const partialDir = "./tmp/gswag"

var testServer *httptest.Server

// TestAPI is the entry point for the test suite.
//
// To exercise the parallel merge mechanism, run with ginkgo:
//
//	ginkgo -p ./...
//
// Each Ginkgo node receives a subset of the specs in users_test.go and
// posts_test.go.  After execution:
//  1. Every node writes its own partial spec to ./tmp/gswag/node-N.json via
//     WritePartialSpec (first function of SynchronizedAfterSuite).
//  2. Node 1 waits for all partial files to appear and merges them into the
//     final docs/openapi.yaml via MergeAndWriteSpec (second function, node-1
//     only).
//
// Running with `go test ./...` (sequential, 1 node) also works — node 1 is
// both writer and merger so the spec is produced correctly.
func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parallel example suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "Blog API (parallel)",
		Version:    "1.0.0",
		OutputPath: "../docs/openapi.yaml",
	})
	testServer = httptest.NewServer(api.NewRouter())
	SetTestServer(testServer)
})

// SynchronizedAfterSuite implements the two-phase parallel teardown:
//   - All nodes: close the server and write their partial spec file.
//   - Node 1 only: wait for all partial files and merge into the final YAML.
var _ = SynchronizedAfterSuite(func() {
	testServer.Close()
	Expect(WritePartialSpec(GinkgoParallelProcess(), partialDir)).To(Succeed())
}, func() {
	suiteConfig, _ := GinkgoConfiguration()
	Expect(MergeAndWriteSpec(suiteConfig.ParallelTotal, partialDir)).To(Succeed())
})
