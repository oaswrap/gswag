# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Does

`gswag` generates OpenAPI 3.0 specifications **as a side-effect of running Ginkgo v2 integration tests**. It intercepts real HTTP requests/responses during test execution and builds a live spec with zero code annotations. Inspired by rswag (Ruby on Rails).

## Commands

```bash
# Build
make build          # Build CLI binary to bin/gswag
make install        # go install

# Test
make test           # go test ./... (all unit + golden tests)
make test-verbose   # with -v
make test-race      # with race detector
make cover          # coverage report
make examples       # build and run all 7 example suites
make validate-examples  # run examples + validate generated specs

# Code quality
make vet
make fmt
make lint           # requires golangci-lint installed separately
make tidy           # go mod tidy on main module + all examples

# Golden files
make update-golden  # regenerate golden test fixtures (UPDATE_GOLDEN=true go test ./...)

# Run a single test suite
go test -run TestName ./...
cd test/basic_data && go test ./...   # single golden suite
cd examples/gin && go test ./...      # single example suite

# CLI usage
gswag init [dir]                       # scaffold suite file
gswag validate [--strict] <spec.yaml>
gswag diff [--json] [--no-fail] base.yaml head.yaml
```

## Architecture

### Three-Phase Model

```
Phase 1 — Tree Construction (synchronous, before tests)
  DSL calls (Path/Get/Response/...) build in-memory operation tree
  and enqueue metadata via dslPendingOps

Phase 2 — Test Execution (parallel-safe, inside It blocks)
  RunTest fires real HTTP requests, asserts responses
  Injects runtime-observed schemas/examples into spec

Phase 3 — Output (AfterSuite)
  WriteSpec/WritePartialSpec serialize to YAML or JSON
  MergeAndWriteSpec merges per-node partial specs for parallel runs
```

### Key Files

| File | Role |
|---|---|
| `dsl.go` | All DSL entry points: `Path`, `Get/Post/Patch/Put/Delete`, `Response`, `RunTest`, security/params/schema setters |
| `spec.go` | `SpecCollector` — thread-safe accumulator; registers operations, infers schemas, appends examples |
| `config.go` | Global `Config` struct and security scheme helpers (`BearerJWT`, `APIKeyHeader`, etc.) |
| `builder.go` | `requestBuilder` — constructs and fires HTTP requests, carries spec metadata + execution values |
| `recorder.go` | `recordedResponse` — captures status, headers, body from HTTP responses |
| `output.go` | `WriteSpec`/`WriteSpecTo` — serialize to YAML or JSON |
| `parallel.go` | `WritePartialSpec`, `MergeAndWriteSpec` — parallel Ginkgo node orchestration |
| `suite.go` | `RegisterSuiteHandlers`, `RegisterParallelSuiteHandlers` |
| `validate.go` | Structural checks + JSON Schema validation against OpenAPI 3.0 |
| `matchers.go` | Gomega matchers: `HaveStatus`, `HaveHeader`, `ContainJSONKey`, `MatchJSONSchema`, `HaveNonEmptyBody` |
| `internal/schemautil` | Infer JSON schemas from raw response bytes |
| `internal/golden` | Golden-file test helpers |
| `cmd/gswag/` | CLI: `init`, `validate`, `diff`, `version` |

### Global DSL State (single-threaded, tree-construction phase only)

- `dslPathStack` — accumulates nested path segments
- `dslOpStack` — current operation being described
- `dslRespExecStack` — current response block's execution values
- `dslPendingOps` — snapshot queue flushed on first `RunTest`

### Parallel Process Model

Each Ginkgo process has isolated global state (no cross-process races). Each writes `node-N.json` partial files. Node 1 uses `SynchronizedAfterSuite` to merge all partials via `MergeAndWriteSpec` (last-write-loses, no-clobber strategy for components).

### Content-Type Precedence (highest → lowest)

1. `Consumes()` + `RequestBody()` — injection skipped if schema exists
2. `SetRawBody(data, "Y")` — injected only when no schema exists
3. `SetBody(struct)` — injected only when no schema exists
4. Default `application/json`

## Test Organization

- **Unit tests** — `*_test.go` in root package (`dsl_test.go`, `spec_test.go`, `builder_test.go`, `matchers_test.go`, etc.)
- **Golden integration tests** — `test/*/` — full Ginkgo suite execution compared against golden files. Regenerate with `make update-golden`.
- **Example suites** — `examples/*/` — real `httptest.Server` instances for stdlib, gin, echo, chi, fiber, gorilla, petstore
- **Parallel tests** — `parallel_merge_test.go`, `parallel_components_test.go`, `parallel_test.go`

## Conventions

- Single-letter receivers: `(sc *SpecCollector)`, `(b *requestBuilder)`, `(r *recordedResponse)`
- Internal DSL types use snake_case: `dslOp`, `dslRespExec`, `dslParam`, `dslRespSpec`
- Public DSL API uses PascalCase: `Path()`, `Response()`, `ResponseSchema()`
- DSL entry points wrap Ginkgo `Describe`/`Context` nodes
- CLI commands use `*NoExit` variants returning exit codes for testability

## Key Design Notes

- **No annotations** — specs emerge from test behavior, not comments or struct tags
- **Two registration paths**: DSL path (recommended, metadata declared before `RunTest`) vs. legacy runtime path (`requestBuilder.Register` called directly)
- `buildPathParamsStruct` dynamically creates structs with `path:"name"` tags for path parameter reflection
- `dslSchemaTypeToReflect` maps `SchemaType` constants to Go `reflect.Type`

See `ARCHITECTURE.md` for deeper internal documentation and `README.md` for the full DSL reference.
