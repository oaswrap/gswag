# gswag

[![CI](https://github.com/oaswrap/gswag/actions/workflows/test.yml/badge.svg)](https://github.com/oaswrap/gswag/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/oaswrap/gswag)](https://goreportcard.com/report/github.com/oaswrap/gswag)
[![Go Reference](https://pkg.go.dev/badge/github.com/oaswrap/gswag.svg)](https://pkg.go.dev/github.com/oaswrap/gswag)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Generate OpenAPI 3.0 specs directly from your [Ginkgo](https://github.com/onsi/ginkgo) integration tests.

Inspired by [rswag](https://github.com/rswag/rswag): define API docs alongside executable tests using a nested DSL.

## How it works

`gswag` wraps Ginkgo containers (`Path`, `Get/Post/...`, `Response`, `RunTest`).

1. During test tree construction, it records operation metadata (path, method, params, schemas, security).
2. During test execution, `RunTest` makes a real HTTP request against your test server.
3. Responses are asserted with Gomega and used to infer/capture examples when configured.
4. `WriteSpec()` serializes the in-memory OpenAPI document.

## Installation

```sh
go get github.com/oaswrap/gswag
```

Requires Go 1.24+.

## Quick Start

### 1. Configure suite and server

```go
// suite_test.go
package api_test

import (
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

    testServer = httptest.NewServer(NewRouter())
    SetTestServer(testServer)
})

var _ = AfterSuite(func() {
    testServer.Close()
    Expect(WriteSpec()).To(Succeed())
})
```

Or use the `RegisterSuiteHandlers` helper (single-process suites):

```go
func TestAPI(t *testing.T) {
    gswag.RegisterSuiteHandlers(&gswag.Config{
        Title:      "My API",
        Version:    "1.0.0",
        OutputPath: "./docs/openapi.yaml",
    })
    RegisterFailHandler(Fail)
    RunSpecs(t, "API Suite")
}
```

### 2. Write API specs with the DSL

```go
// users_test.go
package api_test

import (
    "net/http"

    . "github.com/oaswrap/gswag"
    . "github.com/onsi/gomega"
)

type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var _ = Path("/users/{id}", func() {
    Get("Get user by ID", func() {
        Tag("users")
        OperationID("getUserByID")
        BearerAuth()
        Parameter("id", PathParam, String)

        Response(200, "user found", func() {
            ResponseSchema(new(User))
            SetParam("id", "1")

            RunTest(func(resp *http.Response) {
                Expect(resp).To(HaveStatus(http.StatusOK))
                Expect(resp).To(ContainJSONKey("id"))
            })
        })
    })
})
```

## DSL Reference

### Path and operations

```go
Path("/users", func() {
    Get("List users", func() { ... })
    Post("Create user", func() { ... })
})

Path("/users/{id}", func() {
    Get("Get user", func() { ... })
    Put("Replace user", func() { ... })
    Patch("Update user", func() { ... })
    Delete("Delete user", func() { ... })
})
```

### Operation metadata

```go
Tag("users", "admin")
Description("Returns one user")
OperationID("getUser")
Deprecated()
Hidden() // run test, but do not add the operation to the spec
```

### Security

```go
BearerAuth()                    // uses "bearerAuth"
Security("apiKey")             // custom scheme
// Security("oauth2", "scope") // with scopes if needed
```

### Parameters

```go
Parameter("id", PathParam, String)
Parameter("limit", QueryParam, Integer)
Parameter("X-Request-ID", HeaderParam, String)
```

You can also define typed query params:

```go
type ListQuery struct {
    Search string `query:"search"`
    Page   int    `query:"page"`
}

QueryParamStruct(new(ListQuery))
```

### Request and response schema

```go
RequestBody(new(CreateUserRequest))

Response(201, "created", func() {
    ResponseSchema(new(User))
    ResponseHeader("X-Rate-Limit", "")

    SetBody(&CreateUserRequest{Name: "Alice"})
    SetHeader("X-Trace-ID", "abc")
    SetQueryParam("verbose", "true")

    RunTest()
})
```

Request value setters (for execution):

- `SetParam(name, value)`
- `SetQueryParam(name, value)`
- `SetHeader(name, value)`
- `SetBody(body)`
- `SetRawBody([]byte, contentType)`

### Content types

By default gswag documents request bodies as `application/json`. Use `Consumes` and `Produces` to override:

```go
Post("Upload file", func() {
    Consumes("multipart/form-data")
    Produces("application/json")
    RequestBody(new(UploadForm))

    Response(200, "uploaded", func() {
        SetRawBody(formData, "multipart/form-data")
        RunTest(...)
    })
})
```

`Produces` accepts multiple types when an endpoint can serve different formats:

```go
Get("Export data", func() {
    Produces("application/json", "text/csv")
    ...
})
```

### Shared request setup with BeforeRequest

Use `BeforeRequest` (a thin `BeforeEach` wrapper) to share `SetParam`, `SetHeader`, or `SetBody` calls across multiple `Response` blocks — similar to `let` in rswag:

```go
Get("Get order", func() {
    BeforeRequest(func() {
        SetHeader("Authorization", "Bearer test-token")
    })

    Response(200, "found", func() {
        SetParam("id", "42")
        RunTest(...)
    })

    Response(404, "not found", func() {
        SetParam("id", "999")
        RunTest(...)
    })
})
```

> **Note:** `BeforeRequest` runs during test execution (Ginkgo `BeforeEach`), so it is suited for values that can only be determined at runtime. Static values (known at test-tree build time) should be set directly inside the `Response` block.

## Config

```go
Init(&Config{
    Title:           "My API",           // required
    Version:         "1.0.0",            // required
    Description:     "Public API",
    TermsOfService:  "https://example.com/terms",
    Contact: &ContactConfig{
        Name:  "API Team",
        URL:   "https://example.com/support",
        Email: "api@example.com",
    },
    License: &LicenseConfig{
        Name: "Apache 2.0",
        URL:  "https://www.apache.org/licenses/LICENSE-2.0.html",
    },
    ExternalDocs: &ExternalDocsConfig{
        Description: "More docs",
        URL:         "https://example.com/docs",
    },
    Tags: []TagConfig{
        {Name: "users", Description: "User operations"},
    },
    OutputPath:      "./docs/openapi.yaml", // default: ./docs/openapi.yaml
    OutputFormat:    YAML,                   // YAML or JSON
    Servers: []ServerConfig{
        {URL: "https://api.example.com", Description: "prod"},
    },
    ExcludePaths: []string{
        "/internal/*",
        "/admin/health",
    },
    SecuritySchemes: map[string]SecuritySchemeConfig{
        "bearerAuth": BearerJWT(),
        "apiKey":     APIKeyHeader("X-API-Key"),
        "oauth2":     OAuth2Implicit("https://example.com/oauth/authorize", map[string]string{
            "read:users":  "read users",
            "write:users": "modify users",
        }),
    },

    EnforceResponseValidation: true,
    ValidationMode:            "warn", // "fail" (default) or "warn"

    CaptureExamples: true,
    MaxExampleBytes: 0, // 0 means default cap of 16384 bytes; set >0 to override
    Sanitizer: func(b []byte) []byte {
        return b // redact sensitive data here
    },

    MergeTimeout: 60 * time.Second, // how long MergeAndWriteSpec waits for slow nodes (default 30s)
})
```

`ExcludePaths` supports exact path matches and simple prefix patterns ending in `*`.
Excluded operations are still executed by tests when you hit them through `RunTest`; they are only omitted from spec generation.

Security helpers:

- `BearerJWT()`
- `APIKeyHeader(name)`
- `APIKeyQuery(name)`
- `APIKeyCookie(name)`
- `OAuth2Implicit(authURL, scopes)`

## Gomega Matchers

Matchers operate on `*http.Response` (the object passed to `RunTest` callback):

- `HaveStatus(code)`
- `HaveStatusInRange(lo, hi)`
- `HaveHeader(key, value)`
- `HaveJSONBody(expected)`
- `ContainJSONKey(key)`
- `MatchJSONSchema(model)`
- `HaveNonEmptyBody()`

Example:

```go
RunTest(func(resp *http.Response) {
    Expect(resp).To(HaveStatus(200))
    Expect(resp).To(HaveHeader("Content-Type", "application/json"))
    Expect(resp).To(ContainJSONKey("id"))
})
```

## Validation

### Validate in-memory spec

```go
issues := ValidateSpec()
for _, issue := range issues {
    fmt.Println(issue.String())
}
```

### Validate file

```go
issues, err := ValidateSpecFile("docs/openapi.yaml")
```

### Write then validate

```go
if err := WriteAndValidateSpec(); err != nil {
    // err wraps ErrSpecInvalid when error-level issues exist
    panic(err)
}
```

Validation highlights:

- `info.title` and `info.version` required (error)
- empty paths (warning)
- operation missing summary/tags (warning)
- undeclared security scheme references (error)

## Parallel Ginkgo Support

For `ginkgo -p`, each parallel process writes a partial spec and process 1 merges them all.

### Option A — helper (recommended)

```go
func TestAPI(t *testing.T) {
    gswag.RegisterParallelSuiteHandlers(&gswag.Config{
        Title:      "My API",
        Version:    "1.0.0",
        OutputPath: "./docs/openapi.yaml",
    }, "./tmp/gswag") // partialDir shared by all nodes
    RegisterFailHandler(Fail)
    RunSpecs(t, "API Suite")
}
```

`RegisterParallelSuiteHandlers` uses Ginkgo's `SynchronizedAfterSuite` internally, which guarantees that node 1 only merges *after* all other nodes have finished writing their partial files.

### Option B — manual `SynchronizedAfterSuite`

```go
var _ = SynchronizedAfterSuite(func() {
    // Runs on every node — write this node's partial spec.
    Expect(gswag.WritePartialSpec(GinkgoParallelProcess(), "./tmp/gswag")).To(Succeed())
}, func() {
    // Runs only on node 1, after all other nodes finish the block above.
    suiteCfg, _ := GinkgoConfiguration()
    Expect(gswag.MergeAndWriteSpec(suiteCfg.ParallelTotal, "./tmp/gswag")).To(Succeed())
})
```

Partial files are written as `./tmp/gswag/node-N.json`. `MergeAndWriteSpec` polls for each file with a configurable timeout (default 30 s, override via `Config.MergeTimeout`).

