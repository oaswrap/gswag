package gswag_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// captureExamplesPath is a registered DSL operation used for capture-examples testing.
var captureExamplesPath = "/capture-echo"

type captureBody struct {
	Name string `json:"name"`
}

var _ = Path(captureExamplesPath, func() {
	Post("Echo capture", func() {
		Tag("capture")
		RequestBody(new(captureBody))

		Response(200, "echoed", func() {
			SetBody(&captureBody{Name: "alice"})
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(200))
			})
		})
	})
})

var _ = Describe("CaptureExamples / Sanitizer", func() {
	It("writes a spec with examples present (CaptureExamples:true in BeforeSuite)", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "with-examples.yaml")

		Expect(WriteSpecTo(outPath, YAML)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("capture"))
	})

	It("WriteSpecTo honours the JSON format flag", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "with-examples.json")

		Expect(WriteSpecTo(outPath, JSON)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(string(data), "{")).To(BeTrue()) // valid JSON
	})
})
