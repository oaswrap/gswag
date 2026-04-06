package gswag_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/oaswrap/gswag"
)

var _ = Describe("WriteSpec / WriteSpecTo", func() {
	It("creates a YAML file containing the API title", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "out.yaml")

		Expect(gswag.WriteSpecTo(outPath, gswag.YAML)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(string(data), "Root Suite API")).To(BeTrue())
	})

	It("creates a JSON file containing the API title", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "out.json")

		Expect(gswag.WriteSpecTo(outPath, gswag.JSON)).To(Succeed())

		data, err := os.ReadFile(outPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Contains(string(data), `"Root Suite API"`)).To(BeTrue())
	})

	It("creates parent directories if they don't exist", func() {
		dir := GinkgoT().TempDir()
		outPath := filepath.Join(dir, "nested", "dir", "out.yaml")

		Expect(gswag.WriteSpecTo(outPath, gswag.YAML)).To(Succeed())

		_, err := os.Stat(outPath)
		Expect(err).NotTo(HaveOccurred())
	})
})
