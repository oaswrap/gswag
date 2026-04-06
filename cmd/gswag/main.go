// Command gswag is a CLI companion for the gswag library.
//
// Usage:
//
//	gswag validate [flags] <spec-file>
//	gswag version
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/oaswrap/gswag"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		runValidate(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "init":
		runInit(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Println("gswag", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// runInit scaffolds a minimal Ginkgo `suite_test.go` and a GitHub Actions workflow
// for using gswag in CI. Usage: gswag init [--force] [path]
func runInit(args []string) {
	code := runInitNoExit(args)
	if code != 0 {
		os.Exit(code)
	}
}

// runInitNoExit is like runInit but returns an exit code instead of calling os.Exit.
func runInitNoExit(args []string) int {
	force := false
	target := "."
	for _, a := range args {
		if strings.EqualFold(a, "--force") {
			force = true
			continue
		}
		// last non-flag arg is target
		target = a
	}

	// Ensure target directory exists.
	if err := os.MkdirAll(target, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating target directory: %v\n", err)
		return 2
	}

	// Files to create.
	files := map[string]string{
		filepath.Join(target, "suite_test.go"):                     suiteTestTemplate,
		filepath.Join(target, ".github", "workflows", "gswag.yml"): ghActionsTemplate,
	}

	for path, content := range files {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating dir %s: %v\n", dir, err)
			return 2
		}
		if _, err := os.Stat(path); err == nil && !force {
			fmt.Fprintf(os.Stderr, "skipping existing file %s (use --force to overwrite)\n", path)
			continue
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", path, err)
			return 2
		}
		fmt.Println("created", path)
	}
	return 0
}

const suiteTestTemplate = `package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This scaffold creates a minimal Ginkgo suite for gswag.
// Replace the TODO section with your project's router import and startup.
var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Example Suite")
}

var _ = BeforeSuite(func() {
	gswag.Init(&gswag.Config{
		Title:      "Example API",
		Version:    "0.1.0",
		OutputPath: "./docs/openapi.yaml",
	})
	// TODO: start your server here, for example:
	//   import yourpkg "github.com/your/module/path"
	//   testServer = httptest.NewServer(yourpkg.NewRouter())
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(gswag.WriteSpec()).To(Succeed())
})
`

const ghActionsTemplate = `name: gswag

on:
	push:
		branches: [ main ]
	pull_request:
		branches: [ main ]

jobs:
	test:
		runs-on: ubuntu-latest
		steps:
			- uses: actions/checkout@v4
			- name: Set up Go
				uses: actions/setup-go@v4
				with:
					go-version: '1.24'
			- name: Install dependencies
				run: go mod download
			- name: Run tests
				run: go test ./... -v
			- name: Validate OpenAPI spec (if present)
				run: |
					if [ -f ./docs/openapi.yaml ]; then
						gswag validate ./docs/openapi.yaml || exit 1
					fi
`

func runValidate(args []string) {
	code := runValidateNoExit(args)
	if code != 0 {
		os.Exit(code)
	}
}

// runValidateNoExit is like runValidate but returns an exit code instead of calling os.Exit.
func runValidateNoExit(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gswag validate <spec-file>")
		return 1
	}

	path := args[len(args)-1]
	strict := hasFlag(args, "--strict")

	issues, err := gswag.ValidateSpecFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	if len(issues) == 0 {
		fmt.Println("✓ spec is valid")
		return 0
	}

	hasErrors := false
	for _, issue := range issues {
		fmt.Println(issue)
		if issue.Severity == "error" {
			hasErrors = true
		}
	}

	if hasErrors || (strict && len(issues) > 0) {
		return 1
	}
	return 0
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if strings.EqualFold(a, flag) {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Print(`gswag — OpenAPI spec tooling for Ginkgo test suites

Commands:
	validate [--strict] <spec-file>   Validate an OpenAPI spec file.
																		Exits 1 if errors are found.
																		With --strict, warnings also cause exit 1.
	diff <base-spec> <new-spec>       Diff two OpenAPI spec files.
																		Exits 1 if breaking changes are detected.
	init [--force] [path]             Scaffold a Ginkgo suite_test.go and CI workflow.
																		Writes files into the target path (default: current dir).
																		Use --force to overwrite existing files.
	version                           Print the gswag version.
	help                              Show this help.

Examples:
	gswag validate ./docs/openapi.yaml
	gswag validate --strict ./docs/openapi.json
	gswag diff ./docs/openapi-v1.yaml ./docs/openapi-v2.yaml
	gswag init .
`)
}
