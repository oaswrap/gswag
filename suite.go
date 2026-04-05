package gswag

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// RegisterSuiteHandlers registers BeforeSuite and AfterSuite hooks that
// initialise gswag and write the spec on suite completion.
//
// Call this from your TestXxx function or at package init, passing the same
// Config you would pass to Init.
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
