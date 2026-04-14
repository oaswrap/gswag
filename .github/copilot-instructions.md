# Copilot Workspace Instructions for gswag

## Overview
This workspace is for `gswag`, a tool to generate OpenAPI 3.0 specifications from Ginkgo integration tests in Go. It works by intercepting HTTP requests/responses in tests and building a live OpenAPI spec, with no code annotations required.

`gswag` uses an rswag-style nested DSL:
- `Path(..., func() { ... })`
- `Get/Post/Put/Patch/Delete(..., func() { ... })`
- `Response(..., func() { ... })`
- `RunTest(func(resp *http.Response) { ... })`

## Build & Test Commands
- **Run all tests:** `make test` or `go test ./...`
- **Test with coverage:** `make cover` or `make cover-html` (see `coverage.html`)
- **Lint:** `make lint` (requires `golangci-lint`)
- **Test all examples:** `make examples`

## Key Conventions
- **No code annotations:** Specs are generated from test behavior, not code comments.
- **Ginkgo/Gomega:** All tests use Ginkgo v2 and Gomega for assertions.
- **Spec output:** By default, specs are written to `docs/openapi.yaml` in each example.
- **Test server setup:** Use `SetTestServer(...)` in `BeforeSuite` after starting the test server.
- **Parallel test support:** See README for merging partial specs.
- **Framework-agnostic:** Works with any Go HTTP framework (see `examples/`).

## Project Structure
- `examples/`: Example projects for different frameworks
- `internal/`: Internal utilities (e.g., schema inference)
- Top-level Go files: Core library and DSL

## Documentation
- See [README.md](../README.md) for usage, configuration, and advanced features.
- Example test suites in `examples/*/suite_test.go`.

## Current DSL Pattern
Use this shape in examples and generated snippets:

```go
var _ = Path("/users/{id}", func() {
	Get("Get user", func() {
		Tag("users")
		Parameter("id", PathParam, String)

		Response(200, "ok", func() {
			ResponseSchema(new(User))
			SetParam("id", "1")
			RunTest(func(resp *http.Response) {
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})
```

## Common Pitfalls
- **Go 1.24+ required**
- **`golangci-lint` must be installed** for `make lint`
- **Config field names:** Use `OutputPath`/`OutputFormat` (not `Output`/`Format`).
- **Server target:** If `SetTestServer` is not called, `RunTest` will fail.
- **Matchers input type:** Matchers now expect `*http.Response` in `RunTest` callbacks.
- **Parallel Ginkgo:** Use the documented pattern for merging specs

## Example Prompts
- "How do I generate and validate an OpenAPI spec from my Ginkgo tests?"
- "Show me how to use the gswag DSL for a POST endpoint."
- "How do I run all example suites and validate their specs?"
- "What Makefile targets are available?"
- "How do I set up `SetTestServer` in `BeforeSuite` for Chi/Echo/Gin/Fiber?"
- "How do I model query/path params with `Parameter` and `QueryParamStruct`?"

## Link, Don't Embed
For details on the DSL, configuration, and advanced usage, refer to [README.md](../README.md) and example projects in `examples/`.

## Assistant Guidance
- **Primary sources:** Prefer linking to [README.md](../README.md) and [ARCHITECTURE.md](../ARCHITECTURE.md) rather than embedding long docs.
- **Run commands only when requested:** Use `make test`, `make examples`, or `go test ./...` only after confirming with the developer.
- **Small, focused edits:** When changing code, make minimal, well-scoped patches and run unit tests for the affected package if possible.
- **Respect golden files:** When tests reference golden fixtures, ask before regenerating them; use `UPDATE_GOLDEN=true` when the developer approves.
- **Parallel-aware:** Be careful modifying merge/parallel logic (see [parallel.go](../parallel.go)) — this affects merged spec outputs across Ginkgo nodes.

## Suggested Agent Customizations
- **Create an `AGENTS.md`** that defines applyTo patterns for `examples/*`, and `test/*` work so assistants can apply different behaviors per area.
- **Add a `prompts/init.prompt.md`** with recommended starter prompts for common tasks (run tests, scaffold example, validate spec).
- **Add a `skills/run-make-targets.md`** to document standard `make` targets and expected effects for safe automation.

## Next Steps (pick one)
- I can create the suggested agent files (`AGENTS.md`, a prompt template, or a small `skills` doc). Tell me which to add first.

