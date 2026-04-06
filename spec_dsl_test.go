package gswag

import (
	"testing"
)

func TestRegisterDSLOperation_AppendsDSLParamsAndResponseHeaders(t *testing.T) {
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)

	op := &dslOp{
		method:  "GET",
		path:    "/things/{id}",
		summary: "get thing",
		params: []dslParam{
			{name: "q", location: InQuery, typ: String},
			{name: "X-Req", location: InHeader, typ: String},
		},
		responses: map[int]*dslRespSpec{
			200: {description: "ok", headers: map[string]interface{}{"X-Count": 1}},
		},
	}

	sc.RegisterDSLOperation(op)

	pi, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/things/{id}"]
	if !ok {
		t.Fatalf("path not registered")
	}
	oper, ok := pi.MapOfOperationValues["get"]
	if !ok {
		t.Fatalf("get operation not found")
	}

	// check params appended (query + header)
	foundQ := false
	foundH := false
	for _, p := range oper.Parameters {
		if p.Parameter.Name == "q" {
			foundQ = true
		}
		if p.Parameter.Name == "X-Req" {
			foundH = true
		}
	}
	if !foundQ || !foundH {
		t.Fatalf("expected DSL params appended q=%v h=%v", foundQ, foundH)
	}

	// check response header present
	r := oper.Responses.MapOfResponseOrRefValues["200"].Response
	if r == nil || r.Headers == nil {
		t.Fatalf("response or headers missing")
	}
	if _, ok := r.Headers["X-Count"]; !ok {
		t.Fatalf("expected X-Count header in response")
	}
}

func TestTopOpTopRespExecPanics(t *testing.T) {
	// ensure stacks empty
	dslOpStack = nil
	dslRespExecStack = nil

	// topOp should panic
	did := false
	func() {
		defer func() {
			if recover() != nil {
				did = true
			}
		}()
		_ = topOp()
	}()
	if !did {
		t.Fatalf("expected topOp to panic when called outside an operation")
	}

	did = false
	func() {
		defer func() {
			if recover() != nil {
				did = true
			}
		}()
		_ = topRespExec()
	}()
	if !did {
		t.Fatalf("expected topRespExec to panic when called outside a Response")
	}
}
