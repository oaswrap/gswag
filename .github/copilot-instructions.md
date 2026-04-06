# Copilot Workspace Instructions for gswag

## Overview
This workspace is for `gswag`, a tool to generate OpenAPI 3.0 specifications from Ginkgo integration tests in Go. It works by intercepting HTTP requests/responses in tests and building a live OpenAPI spec, with no code annotations required.

## Build & Test Commands
- **Build CLI:** `make build` (outputs to `bin/gswag`)
- **Install CLI:** `make install`
- **Run all tests:** `make test` or `go test ./...`
- **Test with coverage:** `make cover` (see `coverage.html`)
- **Lint:** `make lint` (requires `golangci-lint`)
- **Test all examples:** `make examples`
- **Validate example specs:** `make validate-examples`
- **Validate a spec:** `make validate SPEC=path/to/openapi.yaml`
- **Diff two specs:** `make diff BASE=base.yaml HEAD=head.yaml`
- **Clean build artifacts:** `make clean`

## Key Conventions
- **No code annotations:** Specs are generated from test behavior, not code comments.
- **Ginkgo/Gomega:** All tests use Ginkgo v2 and Gomega for assertions.
- **Spec output:** By default, specs are written to `docs/openapi.yaml` in each example.
- **Parallel test support:** See README for merging partial specs.
- **Framework-agnostic:** Works with any Go HTTP framework (see `examples/`).

## Project Structure
- `cmd/gswag/`: CLI entrypoint and commands
- `examples/`: Example projects for different frameworks
- `internal/`: Internal utilities (e.g., schema inference)
- Top-level Go files: Core library and DSL

## Documentation
- See [README.md](../README.md) for usage, configuration, and advanced features.
- Example test suites in `examples/*/suite_test.go`.

## Common Pitfalls
- **Go 1.24+ required**
- **`golangci-lint` must be installed** for `make lint`
- **Spec output path:** Ensure `Output` in config matches your desired location
- **Parallel Ginkgo:** Use the documented pattern for merging specs

## Example Prompts
- "How do I generate and validate an OpenAPI spec from my Ginkgo tests?"
- "Show me how to use the gswag DSL for a POST endpoint."
- "How do I run all example suites and validate their specs?"
- "What Makefile targets are available?"

## Link, Don't Embed
For details on the DSL, configuration, and advanced usage, refer to [README.md](../README.md) and example projects in `examples/`.
