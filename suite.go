package gswag

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// RegisterSuiteHandlers registers BeforeSuite and AfterSuite hooks that
// initialise gswag and write the spec on suite completion.
//
// Call this from your TestXxx function or at package init, passing the same
// Config you would pass to Init. For parallel test runs use
// RegisterParallelSuiteHandlers instead.
//
//	func TestAPI(t *testing.T) {
//	    gswag.RegisterSuiteHandlers(&gswag.Config{...})
//	    gomega.RegisterFailHandler(gomega.Fail)
//	    ginkgo.RunSpecs(t, "API Suite")
//	}
func RegisterSuiteHandlers(cfg *Config) {
	ginkgo.BeforeSuite(func() {
		Init(cfg)
	})
	ginkgo.AfterSuite(func() {
		gomega.Expect(WriteSpec()).To(gomega.Succeed())
	})
}

// RegisterParallelSuiteHandlers registers suite hooks suitable for parallel
// Ginkgo runs (`ginkgo -p`). Each node writes a partial spec; node 1 then
// merges them all into the final output.
//
// partialDir is a temporary directory used to store per-node partial specs.
// It must be accessible by all parallel nodes (i.e. on a shared filesystem).
//
//	func TestAPI(t *testing.T) {
//	    gswag.RegisterParallelSuiteHandlers(&gswag.Config{...}, "./tmp/gswag")
//	    gomega.RegisterFailHandler(gomega.Fail)
//	    ginkgo.RunSpecs(t, "API Suite")
//	}
func RegisterParallelSuiteHandlers(cfg *Config, partialDir string) {
	ginkgo.BeforeSuite(func() {
		Init(cfg)
	})
	// SynchronizedAfterSuite guarantees:
	// 1. The first (all-nodes) function runs on every node.
	// 2. The second (node-1-only) function runs only on node 1, AFTER all
	//    other nodes have completed the first function.
	// This eliminates the race where MergeAndWriteSpec reads a file that
	// another node has not yet written.
	ginkgo.SynchronizedAfterSuite(func() {
		gomega.Expect(WritePartialSpec(ginkgo.GinkgoParallelProcess(), partialDir)).To(gomega.Succeed())
	}, func() {
		suiteConfig, _ := ginkgo.GinkgoConfiguration()
		gomega.Expect(MergeAndWriteSpec(suiteConfig.ParallelTotal, partialDir)).To(gomega.Succeed())
	})
}
