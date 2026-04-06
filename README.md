# gswag

Generate OpenAPI 3.0 specifications from your [Ginkgo](https://github.com/onsi/ginkgo) integration tests — no annotations, no code generation, just tests.

Inspired by [rswag](https://github.com/rswag/rswag) for Rails.

## How it works

`gswag` intercepts the HTTP requests and responses your Ginkgo tests already make. As each test runs, it records the method, path, parameters, request body, response status, and response body, then builds a live OpenAPI spec in memory. When the suite finishes, it writes the spec to disk.

```
Ginkgo test → gswag DSL → HTTP request → real server → response
                                                           ↓
                                              spec.Register() accumulates
                                                           ↓
                                              AfterSuite → WriteSpec()
```

## Installation

```sh
go get github.com/oaswrap/gswag
```

Requires Go 1.24+.

## Quick start

### 1. Configure the suite

```go
// suite_test.go
package api_test

import (
    "testing"

    "github.com/oaswrap/gswag"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestAPI(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "API Suite")
}

var _ = BeforeSuite(func() {
    gswag.Init(&gswag.Config{
        Title:   "My API",
        Version: "1.0.0",
        Output:  "docs/openapi.yaml",
    })
})

var _ = AfterSuite(func() {
    gswag.WriteSpec()
})
```

Or use the convenience helper which does the same thing:

```go
var _ = gswag.RegisterSuiteHandlers(&gswag.Config{
    Title:   "My API",
    Version: "1.0.0",
    Output:  "docs/openapi.yaml",
})
```

### 2. Write tests with the DSL

```go
// orders_test.go
package api_test

import (
    "net/http/httptest"

    "github.com/oaswrap/gswag"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

var _ = Describe("Orders", func() {
    var server *httptest.Server

    BeforeEach(func() {
        server = httptest.NewServer(NewRouter())
    })

    AfterEach(func() {
        server.Close()
    })

    Describe("GET /orders", func() {
        It("lists all orders", func() {
            resp := gswag.GET("/orders").
                WithTag("orders").
                WithSummary("List all orders").
                WithQueryParam("status", "pending").
                WithQueryParam("limit", "20").
                WithBearerAuth().
                Do(server)

            Expect(resp).To(gswag.HaveStatus(200))
            Expect(resp).To(gswag.HaveJSONBody([]Order{}))
        })
    })

    Describe("POST /orders", func() {
        It("creates an order", func() {
            resp := gswag.POST("/orders").
                WithTag("orders").
                WithSummary("Create an order").
                WithJSONBody(CreateOrderRequest{Product: "Widget", Quantity: 2}).
                WithBearerAuth().
                Do(server)

            Expect(resp).To(gswag.HaveStatus(201))
            Expect(resp).To(gswag.HaveJSONBody(Order{}))
        })
    })
})
```

Run the tests and `docs/openapi.yaml` is generated automatically.

## Configuration

```go
gswag.Init(&gswag.Config{
    Title       string         // required — info.title
    Version     string         // required — info.version
    Description string         // info.description
    Output      string         // output file path (default: "openapi.yaml")
    Format      gswag.OutputFormat  // gswag.YAML (default) or gswag.JSON
    APIVersion  gswag.OpenAPIVersion // gswag.V30 (default) or gswag.V31

    SecuritySchemes map[string]gswag.SecuritySchemeConfig
})
```

### Security scheme helpers

```go
gswag.BearerJWT()                  // HTTP bearer with JWT format
gswag.APIKeyHeader("X-API-Key")    // API key in header
gswag.APIKeyQuery("api_key")       // API key in query string
gswag.APIKeyCookie("session")      // API key in cookie
```

Example with multiple schemes:

```go
gswag.Init(&gswag.Config{
    Title:   "My API",
    Version: "1.0.0",
    SecuritySchemes: map[string]gswag.SecuritySchemeConfig{
        "bearerAuth": gswag.BearerJWT(),
        "apiKey":     gswag.APIKeyHeader("X-API-Key"),
    },
})
```

## Request builder DSL

### HTTP methods

```go
gswag.GET(path)
gswag.POST(path)
gswag.PUT(path)
gswag.PATCH(path)
gswag.DELETE(path)
```

### Metadata

```go
.WithTag("orders")
.WithSummary("List all orders")
.WithDescription("Returns a paginated list of orders for the current user.")
.WithOperationID("listOrders")
.WithDeprecated()
```

### Parameters

```go
.WithPathParam("id", "123")                // path parameter — {id} in path
.WithQueryParam("status", "active")        // individual query parameter
.WithQueryParamStruct(FilterParams{})      // typed struct → full query schema
.WithHeader("X-Request-ID", "abc-123")    // request header parameter
```

`WithQueryParamStruct` accepts any struct; fields with `query:"name"` tags are used (falls back to field name). Types are reflected from the struct definition.

### Request body

```go
.WithJSONBody(CreateOrderRequest{})        // typed struct → request body schema
.WithRawBody([]byte(`{"x":1}`), "application/json")  // raw bytes + content type
```

### Response schemas

```go
.WithResponse(200, Order{})               // typed response body schema
.WithResponse(201, Order{})
.WithResponse(404, ErrorResponse{})
```

If no `WithResponse` is called, gswag infers the schema from the actual response body at test time.

### Security

```go
.WithBearerAuth()                          // adds bearerAuth security requirement
.WithSecurity("apiKey", []string{})        // named scheme with optional scopes
.WithSecurity("oauth2", []string{"read:orders", "write:orders"})
```

`WithBearerAuth()` auto-registers the `bearerAuth` scheme as HTTP Bearer JWT if it is not already declared in `Config.SecuritySchemes`.

### Sending the request

```go
resp := builder.Do(server)   // *httptest.Server or base URL string
```

`Do` returns a `*gswag.RecordedResponse` which carries `StatusCode`, `Headers`, `BodyBytes`, and `Duration`.

## Gomega matchers

Use these matchers on a `*RecordedResponse`:

| Matcher | Description |
|---|---|
| `HaveStatus(code int)` | Exact HTTP status code |
| `HaveStatusInRange(lo, hi int)` | Status code in inclusive range |
| `HaveHeader(key, value string)` | Response header value |
| `HaveJSONBody(expected interface{})` | JSON body matches structure of expected value |
| `ContainJSONKey(key string)` | JSON object contains the given key |
| `MatchJSONSchema(model interface{})` | Response keys are a subset of model's JSON fields |
| `HaveNonEmptyBody()` | Response body is not empty |

```go
Expect(resp).To(gswag.HaveStatus(200))
Expect(resp).To(gswag.HaveStatusInRange(200, 299))
Expect(resp).To(gswag.HaveHeader("Content-Type", "application/json"))
Expect(resp).To(gswag.HaveJSONBody(Order{}))
Expect(resp).To(gswag.ContainJSONKey("id"))
Expect(resp).To(gswag.MatchJSONSchema(Order{}))
Expect(resp).To(gswag.HaveNonEmptyBody())
```

## Spec validation

Validate the in-memory spec programmatically:

```go
issues := gswag.ValidateSpec()
for _, issue := range issues {
    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Path, issue.Message)
}
```

Validate a spec file on disk:

```go
issues, err := gswag.ValidateSpecFile("docs/openapi.yaml")
```

Write and validate in one step (returns `gswag.ErrSpecInvalid` if errors are present):

```go
if err := gswag.WriteAndValidateSpec(); err != nil {
    log.Fatal(err)
}
```

Validation checks:
- `info.title` and `info.version` are required (error)
- `paths` must not be empty (warning)
- Operations should have summary and tags (warning)
- Security scheme references must be declared in `components.securitySchemes` (error)

## Runtime response validation

`gswag` can optionally validate actual HTTP responses observed during tests against the declared or inferred response schema. Enable this in your test suite by setting `EnforceResponseValidation` in the `Config` passed to `gswag.Init`.

The `ValidationMode` setting controls what happens when a validation error occurs:

- `fail` (default): a validation mismatch causes the test to fail (panic).
- `warn`: validation issues are printed to stderr and tests continue.

Example:

```go
gswag.Init(&gswag.Config{
    Title:                    "My API",
    Version:                  "1.0.0",
    EnforceResponseValidation: true,
    ValidationMode:           "warn", // or "fail"
})
```

Use `warn` during development to see schema drift without breaking tests, and switch to `fail` in CI to enforce schema correctness.

## Capture request/response examples

`gswag` can optionally capture actual request and response bodies observed during tests and attach them to the generated OpenAPI spec as examples. This is useful for populating live example sections in your docs, but be careful to avoid storing sensitive data.

Settings (in `gswag.Config`):

- `CaptureExamples bool`: enable example capture when true.
- `MaxExampleBytes int`: cap the number of bytes stored for any single example. `0` means no cap.
- `Sanitizer func([]byte) []byte`: optional hook called with the raw example bytes before they are stored; use it to redact or transform sensitive fields.

When enabled, `gswag` stores request examples at `requestBody.content.<mediaType>.example` and response examples at `responses.<status>.content.<mediaType>.example` (prefers `application/json`).

Example usage:

```go
gswag.Init(&gswag.Config{
    Title:           "My API",
    Version:         "1.0.0",
    CaptureExamples: true,
    MaxExampleBytes: 16384,
    Sanitizer: func(b []byte) []byte {
        // Example: redact a JSON password field before storing
        // (implement your own sanitizer according to your needs).
        return b
    },
})
```

Notes:

- Example capture is opt-in. Enable it only where you are comfortable storing runtime examples.
- Use a `Sanitizer` to remove or redact PII before writing the spec.
- Large responses can bloat specs; use `MaxExampleBytes` to bound size.

## Parallel Ginkgo support

When running Ginkgo in parallel (`ginkgo -p`), each worker process calls `WritePartialSpec` and node 1 merges them all:

```go
var _ = AfterSuite(func() {
    nodeIndex := GinkgoParallelProcess()
    totalNodes := GinkgoT().(*testing.T) // or use GinkgoConfiguration()

    if err := gswag.WritePartialSpec(nodeIndex, "tmp/specs"); err != nil {
        Fail(err.Error())
    }
    if nodeIndex == 1 {
        if err := gswag.MergeAndWriteSpec(totalNodes, "tmp/specs"); err != nil {
            Fail(err.Error())
        }
    }
})
```

Partial specs are written as `tmp/specs/node-N.json`. The merge uses no-clobber semantics: the first definition of any path, schema, or security scheme wins.

## CLI tool

Install:

```sh
go install github.com/oaswrap/gswag/cmd/gswag@latest
```

### Validate

```sh
gswag validate docs/openapi.yaml
gswag validate --strict docs/openapi.yaml   # treat warnings as errors
```

Exits 0 if valid, 1 if errors (or warnings in strict mode).

### Diff

Detect breaking changes between two spec files:

```sh
gswag diff base.yaml head.yaml
```

The `diff` command also supports machine-readable output and CI-friendly flags:

```sh
gswag diff --json base.yaml head.yaml      # emit JSON describing added/removed/modified
gswag diff --json --no-fail base head      # don't exit non-zero on breaking changes
gswag diff --format=json base head         # alias for --json
```

This is useful in CI to post a structured report or to decide whether to fail a job on breaking changes.

Exits 0 if no breaking changes, 1 if breaking changes are found.

Breaking changes detected:
- Removed paths or HTTP methods
- Removed parameters
- New required parameters added
- Removed response status codes

Non-breaking changes reported:
- Added paths or HTTP methods
- New optional parameters
- Added response status codes

### Version

```sh
gswag version
```

## Framework support

`gswag` works with any framework that exposes an `http.Handler` or can be wrapped in `httptest.Server`.

| Framework | Example |
|---|---|
| `net/http` | [`examples/stdlib`](examples/stdlib) |
| [Gin](https://github.com/gin-gonic/gin) | [`examples/gin`](examples/gin) |
| [Echo](https://github.com/labstack/echo) | [`examples/echo`](examples/echo) |
| [Chi](https://github.com/go-chi/chi) | [`examples/chi`](examples/chi) |
| [Fiber](https://github.com/gofiber/fiber) | [`examples/fiber`](examples/fiber) |

Each example is a self-contained Go module with a complete test suite that generates a valid `docs/openapi.yaml`.

## CI integration

Example GitHub Actions workflow:

```yaml
- name: Run tests and generate spec
  run: go test ./...

- name: Validate spec
  run: gswag validate docs/openapi.yaml

- name: Diff spec on PR
  if: github.event_name == 'pull_request'
  run: gswag diff origin/main/docs/openapi.yaml docs/openapi.yaml
```

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/swaggest/openapi-go` | OpenAPI 3.0 reflector and spec builder |
| `github.com/onsi/ginkgo/v2` | BDD test framework |
| `github.com/onsi/gomega` | Assertion library |

## License

MIT
