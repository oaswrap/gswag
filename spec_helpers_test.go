package gswag

import (
    "encoding/json"
    "reflect"
    "testing"

    "github.com/swaggest/openapi-go/openapi3"
)

func TestDslSchemaTypeToReflect(t *testing.T) {
    if dslSchemaTypeToReflect(Integer) != reflect.TypeOf(int64(0)) {
        t.Fatalf("Integer did not map to int64")
    }
    if dslSchemaTypeToReflect(Number) != reflect.TypeOf(float64(0)) {
        t.Fatalf("Number did not map to float64")
    }
    if dslSchemaTypeToReflect(Boolean) != reflect.TypeOf(false) {
        t.Fatalf("Boolean did not map to bool")
    }
    if dslSchemaTypeToReflect(String) != reflect.TypeOf("") {
        t.Fatalf("String did not map to string")
    }
}

func TestDslSchemaParamAndStringParamJSON(t *testing.T) {
    p := dslSchemaParam("limit", Integer, openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
    if reflect.ValueOf(p).IsZero() {
        t.Fatalf("expected non-zero parameter")
    }

    sp := stringParam("q", openapi3.ParameterLocation{QueryParameter: &openapi3.QueryParameter{}})
    if reflect.ValueOf(sp).IsZero() {
        t.Fatalf("expected non-zero string param")
    }
}

func TestBuildPathParamsStructFromDSLAndInfer(t *testing.T) {
    params := []dslParam{{name: "id", location: InPath, typ: Integer}}
    v := buildPathParamsStructFromDSL("/users/{id}", params)
    if v == nil {
        t.Fatalf("expected non-nil struct")
    }
    rt := reflect.TypeOf(v)
    if rt.Kind() != reflect.Ptr {
        t.Fatalf("expected pointer to struct, got %v", rt.Kind())
    }
    f := rt.Elem().Field(0)
    if f.Type != reflect.TypeOf(int64(0)) {
        t.Fatalf("expected field type int64, got %v", f.Type)
    }

    // buildPathParamsStruct should infer int when concrete value looks numeric
    v2 := buildPathParamsStruct("/items/{itemId}", map[string]string{"itemId": "42"})
    if v2 == nil {
        t.Fatalf("expected non-nil struct from buildPathParamsStruct")
    }
    rt2 := reflect.TypeOf(v2)
    f2 := rt2.Elem().Field(0)
    if f2.Type != reflect.TypeOf(int64(0)) {
        t.Fatalf("expected inferred int64 field, got %v", f2.Type)
    }
}

func TestCopyDslOpAndRespExecDeepCopy(t *testing.T) {
    op := &dslOp{
        method: "GET",
        path:   "/z",
        tags:   []string{"a"},
        params: []dslParam{{name: "p", location: InQuery, typ: String}},
        responses: map[int]*dslRespSpec{
            200: {description: "ok", headers: map[string]interface{}{"X": "v"}},
        },
    }

    copy := copyDslOp(op)
    // mutate original
    op.tags[0] = "b"
    op.params[0].name = "q"
    op.responses[200].headers["X"] = "changed"

    if copy.tags[0] != "a" {
        t.Fatalf("tags were not copied deeply")
    }
    if copy.params[0].name != "p" {
        t.Fatalf("params were not copied deeply")
    }
    if copy.responses[200].headers["X"] != "v" {
        t.Fatalf("response headers were not copied deeply")
    }

    // resp exec
    re := &dslRespExec{status: 200}
    re.pathParams = map[string]string{"id": "1"}
    re.queryParams = map[string]string{"q": "v"}
    re.headers = map[string]string{"H": "v"}
    re.bodyRaw = []byte("x")
    rcopy := copyDslRespExec(re)
    re.pathParams["id"] = "2"
    re.bodyRaw[0] = 'y'
    if rcopy.pathParams["id"] != "1" {
        t.Fatalf("resp exec pathParams not deeply copied")
    }
    if string(rcopy.bodyRaw) != "x" {
        t.Fatalf("resp exec bodyRaw not deeply copied")
    }
}

// jsonContains is a lightweight helper to check if a marshaled JSON blob contains a key or value string.
func jsonContains(b []byte, s string) bool {
    if !json.Valid(b) {
        return false
    }
    return containsString(string(b), s)
}

func containsString(hay, needle string) bool { return len(hay) > 0 && (func() bool { return (len(needle) > 0 && (len(hay) >= len(needle))) })() && (reflect.DeepEqual(true, true)) && (func() bool { return len(needle) == 0 || (len(hay) >= len(needle) && (func() bool { return stringIndex(hay, needle) >= 0 })()) })() }

func stringIndex(s, sep string) int {
    for i := 0; i+len(sep) <= len(s); i++ {
        if s[i:i+len(sep)] == sep {
            return i
        }
    }
    return -1
}
