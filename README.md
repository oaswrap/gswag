# gswag

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

## Config

```go
Init(&Config{
    Title:           "My API",           // required
    Version:         "1.0.0",            // required
    Description:     "Public API",
    OutputPath:      "./docs/openapi.yaml", // default: ./docs/openapi.yaml
    OutputFormat:    YAML,          // or JSON
    OpenAPIVersion:  V30,           // or V31
    Servers: []ServerConfig{
        {URL: "https://api.example.com", Description: "prod"},
    },
    SecuritySchemes: map[string]SecuritySchemeConfig{
        "bearerAuth": BearerJWT(),
        "apiKey":     APIKeyHeader("X-API-Key"),
    },

    EnforceResponseValidation: true,
    ValidationMode:            "warn", // "fail" (default) or "warn"

    CaptureExamples: true,
    MaxExampleBytes: 16384, // 0 means no cap
    Sanitizer: func(b []byte) []byte {
        return b // redact sensitive data here
    },
})
```

Security helpers:

- `BearerJWT()`
- `APIKeyHeader(name)`
- `APIKeyQuery(name)`
- `APIKeyCookie(name)`

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

For `ginkgo -p`, write partial specs from every process and merge on process 1.

```go
var _ = AfterSuite(func() {
    node := GinkgoParallelProcess()
    suiteCfg, _ := GinkgoConfiguration()

    Expect(WritePartialSpec(node, "./tmp/gswag")).To(Succeed())

    if node == 1 {
        Expect(MergeAndWriteSpec(suiteCfg.ParallelTotal, "./tmp/gswag")).To(Succeed())
    }
})
```

Partial files are stored as `tmp/gswag/node-N.json`.

## CLI

Install:

```sh
go install github.com/oaswrap/gswag/cmd/gswag@latest
```

### Validate

```sh
gswag validate docs/openapi.yaml
gswag validate --strict docs/openapi.yaml
```

### Diff

```sh
gswag diff base.yaml head.yaml
gswag diff --json base.yaml head.yaml
gswag diff --json --no-fail base.yaml head.yaml
```

### Version

```sh
gswag version
```
