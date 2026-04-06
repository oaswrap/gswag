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

// runInit scaffolds a minimal Ginkgo-style suite file.
// Usage: gswag init [--force] [path]
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

	// Detect package name and generate a Ginkgo-style file name:
	// <package>_suite_test.go.
	pkg := detectPackageName(target)
	if pkg == "" {
		pkg = defaultPackageName(target)
	}
	var testPkg string
	if strings.HasSuffix(pkg, "_test") {
		testPkg = pkg
	} else {
		testPkg = pkg + "_test"
	}
	suiteContent := suiteTemplateFor(testPkg)
	suiteFilename := pkg + "_suite_test.go"

	files := map[string]string{
		filepath.Join(target, suiteFilename): suiteContent,
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

func detectPackageName(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			for _, line := range strings.Split(string(b), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "package ") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						return parts[1]
					}
				}
			}
		}
	}
	return ""
}

func defaultPackageName(target string) string {
	base := filepath.Base(filepath.Clean(target))
	if base == "." || base == string(filepath.Separator) {
		return "api"
	}
	base = strings.ReplaceAll(base, "-", "_")
	base = strings.ReplaceAll(base, " ", "_")
	if base == "" {
		return "api"
	}
	return base
}

func suiteTemplateFor(pkg string) string {
	return "package " + pkg + "\n\n" + `import (
	"net/http/httptest"
	"testing"

	. "github.com/oaswrap/gswag"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var testServer *httptest.Server

func TestAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
	Init(&Config{
		Title:      "My API",
		Version:    "1.0.0",
		OutputPath: "./docs/openapi.yaml",
		SecuritySchemes: map[string]SecuritySchemeConfig{
			"bearerAuth": BearerJWT(),
		},
	})
	// TODO: start your server here, for example:
	//   import yourpkg "github.com/your/module/path"
	//   testServer = httptest.NewServer(yourpkg.NewRouter())
	//   SetTestServer(testServer)
})

var _ = AfterSuite(func() {
	if testServer != nil {
		testServer.Close()
	}
	Expect(WriteSpec()).To(Succeed())
})
`
}

/* GitHub Actions workflow generation removed from gswag init.
Previously the CLI created a .github/workflows/gswag.yml file.
That behavior has been removed: init now only creates <package>_suite_test.go. */

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
	init [--force] [path]             Scaffold a minimal Ginkgo test suite file.
		Writes a ` + "`<package>_suite_test.go`" + ` file into the target path (default: current dir).
		Detects package name from existing .go files; if none, uses target directory name.
		The generated file uses package <pkg>_test.
		Use --force to overwrite an existing <package>_suite_test.go.
	version                           Print the gswag version.
	help                              Show this help.

Examples:
	gswag validate ./docs/openapi.yaml
	gswag validate --strict ./docs/openapi.json
	gswag diff ./docs/openapi-v1.yaml ./docs/openapi-v2.yaml
	gswag init .                 # create <package>_suite_test.go in current directory
	gswag init ./tests           # create tests_suite_test.go in ./tests (if no package found)
	gswag init --force ./tests   # overwrite existing generated suite file
`)
}
