package gswag_test

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These paths are registered at package-init time (tree-building phase).
// The BeforeAll inside each Get/Post block fires at test-execution time AFTER BeforeSuite.

var _ = gswag.Path("/spec-items", func() {
	gswag.Get("List spec items", func() {
		gswag.Tag("items")
		gswag.Security("bearerAuth")

		gswag.Response(200, "ok", func() {
			gswag.RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

var _ = gswag.Path("/spec-items/{id}", func() {
	gswag.Get("Get spec item", func() {
		gswag.Tag("items")
		gswag.Parameter("id", gswag.PathParam, gswag.Integer)

		gswag.Response(200, "ok", func() {
			gswag.SetParam("id", "42")
			gswag.RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

type specItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var _ = gswag.Path("/spec-typed", func() {
	gswag.Get("Get typed item", func() {
		gswag.Tag("typed")

		gswag.Response(200, "ok", func() {
			gswag.ResponseSchema(new(specItem))
			gswag.RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

var _ = Describe("SpecCollector", func() {
	It("has no error-level validation issues after operations are registered", func() {
		issues := gswag.ValidateSpec()
		for _, iss := range issues {
			if iss.Severity == "error" {
				Fail("unexpected spec error: " + iss.String())
			}
		}
	})

	It("can write spec to an alternate path and read it back", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "test.yaml")
		Expect(gswag.WriteSpecTo(outPath, gswag.YAML)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("Root Suite API"))
	})

	It("can write spec in JSON format", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "test.json")
		Expect(gswag.WriteSpecTo(outPath, gswag.JSON)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(`"Root Suite API"`))
	})
})
