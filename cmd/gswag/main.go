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

func runValidate(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: gswag validate <spec-file>")
		os.Exit(1)
	}

	path := args[len(args)-1]
	strict := hasFlag(args, "--strict")

	issues, err := gswag.ValidateSpecFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if len(issues) == 0 {
		fmt.Println("✓ spec is valid")
		return
	}

	hasErrors := false
	for _, issue := range issues {
		fmt.Println(issue)
		if issue.Severity == "error" {
			hasErrors = true
		}
	}

	if hasErrors || (strict && len(issues) > 0) {
		os.Exit(1)
	}
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
  version                           Print the gswag version.
  help                              Show this help.

Examples:
  gswag validate ./docs/openapi.yaml
  gswag validate --strict ./docs/openapi.json
  gswag diff ./docs/openapi-v1.yaml ./docs/openapi-v2.yaml
`)
}
