# gswag Architecture

This document describes how gswag is structured internally and how its key components fit together.

---

## Overview

gswag generates OpenAPI 3.0 specifications as a side-effect of running Ginkgo integration tests.
It works in two clearly separated phases:

```
Phase 1 — Tree construction (synchronous, before any test runs)
  DSL calls (Path/Get/Response/...) build an in-memory operation tree
  and enqueue operation metadata for later registration.

Phase 2 — Test execution (parallel-safe, inside It blocks)
  RunTest fires a real HTTP request, asserts the response, and
  injects runtime-observed schemas/examples into the spec.

Phase 3 — Output (AfterSuite)
  WriteSpec / WritePartialSpec serialise the collected spec to YAML or JSON.
```

---

## Directory Layout

```
gswag/
├── gswag.go                  # package doc
├── config.go                 # Config struct, Init(), global state
├── dsl.go                    # DSL entry points: Path, Get, Response, RunTest, ...
├── dsl_types.go              # ParamLocation / SchemaType constants
│
├── spec.go                   # SpecCollector struct, shared constants & vars
├── spec_collector.go         # newSpecCollector (+ applySchemaOptions/applySpecInfo helpers), Register
├── spec_dsl.go               # RegisterDSLOperation, isExcludedPath
├── spec_params.go            # param builder helpers: locationToParamIn, stringParam, dslSchemaParam,
│                             #   buildPathParamsStruct*, dslSchemaTypeToReflect, appendParams, appendDSLParams
├── spec_schemas.go           # schema injection: injectInferredRequestSchema, injectInferredSchema,
│                             #   injectRecordedResponseSchema, appendResponseHeaders, appendDSLResponseHeaders,
│                             #   appendExamples, content-type helpers
├── spec_security.go          # ensureSecurityScheme, buildSecuritySchemeOrRef
│
├── builder.go                # requestBuilder: constructs and fires HTTP requests
├── recorder.go               # recordedResponse: captures status, headers, body
├── output.go                 # WriteSpec / WriteSpecTo
├── parallel.go               # WritePartialSpec / MergeAndWriteSpec / mergeSpec
├── validate.go               # ValidateSpec / ValidateSpecFile / WriteAndValidateSpec
├── matchers.go               # Gomega matchers: HaveStatus, ContainJSONKey, ...
├── output_sanitize.go        # example body sanitisation helpers
│
├── internal/
│   ├── schemautil/           # JSON schema inference from raw bytes
│   └── golden/               # golden-file test helpers
│
├── examples/                 # working example suites (stdlib, gin, echo, chi, fiber, petstore)
└── test/                     # golden integration tests (basic_data, all_methods, petstore)
```

---

## Core Data Flow

### Phase 1 — Tree Construction

Ginkgo evaluates all `var _ = ...` initialisers synchronously at startup.
Each DSL call runs inside a Ginkgo container closure:

```
Path("/users", fn)
  └─ ginkgo.Describe("/users", func() {
        push "/users" onto dslPathStack
        fn()                              ← user code runs here
        copyDslOp(op) → enqueuePendingDSLOp  ← snapshot queued
        pop dslPathStack
     })
```

Key global stacks (all touched only during tree construction — effectively single-threaded):

| Variable | Purpose |
|----------|---------|
| `dslPathStack []string` | Accumulates nested path segments |
| `dslOpStack []*dslOp` | Current operation being described |
| `dslRespExecStack []*dslRespExec` | Current response block's execution values |
| `dslPendingOps []*dslOp` | Snapshot queue flushed on first RunTest |

`dslOp` holds spec-side metadata (method, path, params, schemas, security, `consumes`, `produces`).
`dslRespExec` holds test-side values (pathParams, queryParams, headers, body).

### Phase 2 — Test Execution

`RunTest` registers a Ginkgo `It` block. When Ginkgo runs it:

1. **`flushPendingDSLOps()`** — uses `sync.Once` to register all queued operations into the `SpecCollector` exactly once per Init cycle.
2. **`dslBuildRequest`** — constructs a `requestBuilder` from the `dslRespExec` snapshot captured at tree-construction time.
3. **`requestBuilder.do(target)`** — fires the HTTP request, records the response.
4. **`injectInferredRequestSchema`** (`spec_schemas.go`) — if no request schema was declared, infers one from the actual request body bytes.
5. **`injectRecordedResponseSchema`** (`spec_schemas.go`) — infers/attaches response schema from the actual response body.
6. **`appendExamples`** (`spec_schemas.go`) — if `CaptureExamples: true`, attaches sanitised request/response bodies as OpenAPI examples.

### Phase 3 — Output

`WriteSpec()` locks the `SpecCollector`, serialises the `openapi3.Spec` to YAML or JSON, and writes the file.

---

## Key Types

### `dslOp` (`dsl.go`)

Spec-side snapshot of one HTTP operation — everything known at tree-construction time.

```
dslOp
├── method, path, summary, description, operationID
├── tags []string
├── deprecated, hidden bool
├── security []map[string][]string
├── params []dslParam            ← Parameter() calls
├── reqBodyModel interface{}     ← RequestBody()
├── queryStruct  interface{}     ← QueryParamStruct()
├── responses map[int]*dslRespSpec
├── consumes string              ← Consumes()
└── produces []string            ← Produces()
```

### `dslRespExec` (`dsl.go`)

Test-side values for one `Response` block — what to send in the HTTP request.

```
dslRespExec
├── status int
├── pathParams  map[string]string   ← SetParam()
├── queryParams map[string]string   ← SetQueryParam()
├── headers     map[string]string   ← SetHeader()
├── body        interface{}         ← SetBody()
└── bodyRaw     []byte              ← SetRawBody()
    bodyContentType string
```

### `SpecCollector` (`spec.go`, `spec_collector.go`)

Thread-safe accumulator wrapping the `openapi3.Reflector`.

```
SpecCollector          (struct declared in spec.go)
├── mu sync.Mutex          ← protects all spec mutations
├── reflector              ← swaggest/openapi-go reflector
└── excludePaths []string
```

`newSpecCollector` (in `spec_collector.go`) is decomposed into:
- `applySchemaOptions` — wires JSON-schema reflector options (generic name shortening, inline refs, type maps)
- `shortenGenericName` — strips package paths from generic type-argument names
- `applySpecInfo` → `applySpecTags` / `applySpecServers` — populates the OpenAPI Info/Tags/Servers blocks

Four lock points: `Register`, `RegisterDSLOperation`, `injectRecordedResponseSchema`, `appendExamplesLocked`.

### `requestBuilder` (`builder.go`)

Carries both spec metadata (tags, summary, security) and execution values (body, headers, params).
Used in the non-DSL flow where tests call `Register` directly.

---

## Parallel Execution Model

Ginkgo `-p` spawns separate OS processes (not goroutines). Each process has its own global state, so there are no cross-process data races on the DSL stacks or the `SpecCollector`.

```
Process 1          Process 2          Process N
   │                   │                  │
   ▼                   ▼                  ▼
 Init()             Init()            Init()
   │                   │                  │
 [runs its share   [runs its share   [runs its share
  of specs]         of specs]         of specs]
   │                   │                  │
   ▼                   ▼                  ▼
WritePartialSpec   WritePartialSpec  WritePartialSpec
 node-1.json        node-2.json       node-N.json
   │
   ▼ (SynchronizedAfterSuite node-1 function)
MergeAndWriteSpec
 reads node-1..N.json
 mergeSpec() → openapi.yaml
```

### Tree construction on every node

Every Ginkgo process re-runs all `var _ = Path(...)` blocks during tree construction to build the full spec tree — this is how Ginkgo knows which specs exist before distributing them. As a result, `dslPendingOps` is populated with all declared operations on every node. `flushPendingDSLOps` (called by the first `RunTest` or by `WriteSpecTo`) registers these into the collector, so every partial spec contains all path skeletons.

The per-node difference is in runtime-enriched fields:
- **Inferred response schemas** — set by `injectRecordedResponseSchema` only on the node that ran that `It` block.
- **Inferred request body schemas** — set by `injectInferredRequestSchema` only on the node that ran the `It` block with `SetBody`.

### mergeSpec rules

- **Paths / operations**: if two nodes recorded the same path+method, their responses are merged at the status-code level. A response from a later node fills in a status code that is missing from the base, or replaces an empty (no-content) response with one that has a schema inferred from a live HTTP call. `RequestBody` is also copied from a later node when the base node never ran that test.
- **Components** (Schemas, SecuritySchemes, Responses, Parameters, RequestBodies, Headers, Examples, Links, Callbacks): first-seen wins — no-clobber.

`MergeAndWriteSpec` polls for each partial file with a configurable timeout (`Config.MergeTimeout`, default 30 s) to tolerate slow nodes.

**Recommended pattern** — use `SynchronizedAfterSuite` so Ginkgo guarantees all per-node `SynchronizedAfterSuite` bodies finish before node 1 runs the merge:

```go
SynchronizedAfterSuite(func() {
    // all nodes
    WritePartialSpec(GinkgoParallelProcess(), "./tmp/gswag")
}, func() {
    // node 1 only, after all nodes finish above
    MergeAndWriteSpec(suiteConfig.ParallelTotal, "./tmp/gswag")
})
```

See `examples/parallel` for a working end-to-end example using schema inference across parallel nodes.

---

## Schema Registration: Two Paths

### DSL path (recommended)

1. `Path` / `Get` / `Response` / `RequestBody` / `ResponseSchema` declare metadata during tree construction.
2. `flushPendingDSLOps` (first `RunTest`) calls `RegisterDSLOperation` which drives the `openapi3.Reflector` directly.
3. `injectInferredRequestSchema` is **skipped** if the reflector already placed a schema for any content-type key in `requestBody.content` — preventing conflict when `Consumes` and `SetRawBody` use different content-type strings.

### Runtime-only path (no DSL)

For tests that use `requestBuilder` directly (or through the legacy `Register` codepath):

1. `Register` is called with the builder and recorded response.
2. `AddReqStructure` / `AddRespStructure` drive the reflector from the builder's typed fields.
3. `injectInferredRequestSchema` fills any gaps from actual request bytes.
4. `injectInferredSchema` fills response schema from actual response bytes when none was declared.

---

## Content-Type Precedence

| Source | Controls | Priority |
|--------|----------|----------|
| `Consumes("X")` + `RequestBody(M)` | spec: `requestBody.content["X"]` schema | Highest — injection skipped if schema exists |
| `SetRawBody(data, "Y")` | runtime: `Content-Type: Y` header sent | Injected only when no schema exists yet |
| `SetBody(struct)` | runtime: `Content-Type: application/json` | Injected only when no schema exists yet |
| Default | `application/json` | Lowest |

---

## Validation

`validate.go` runs two layers of checks:

1. **Structural checks** — title/version present, paths non-empty, operations have summary/tags, security scheme references resolve.
2. **JSON Schema validation** — the generated spec is marshalled to JSON and validated against the OpenAPI 3.0 meta-schema using `gojsonschema`.

`ValidationIssue` carries a severity (`"error"` or `"warning"`). `WriteAndValidateSpec` returns `ErrSpecInvalid` when any error-level issue exists.

---

## Testing Strategy

| Layer | Location | Mechanism |
|-------|----------|-----------|
| Unit | `*_test.go` (same package) | Direct function calls, stack manipulation |
| Golden integration | `test/basic_data`, `test/all_methods`, `test/petstore` | Full Ginkgo suite → `UPDATE_GOLDEN=1` to regenerate |
| Example suites | `examples/stdlib`, `gin`, `echo`, `chi`, `fiber` | Real httptest.Server, DSL end-to-end |
| Parallel merge | `parallel_merge_test.go`, `parallel_components_test.go` | `mergeSpec` unit tests covering all component types |
