package gswag

import "testing"

// TestDSLFunctions_UpdateOpAndRespExec verifies that the DSL helper functions
// update the current operation and response execution stacks correctly.
func TestDSLFunctions_UpdateOpAndRespExec(t *testing.T) {
	// prepare op and resp exec stacks
	dslOpStack = nil
	dslRespExecStack = nil

	op := &dslOp{method: "PUT", path: "/p", responses: make(map[int]*dslRespSpec)}
	dslOpStack = append(dslOpStack, op)

	re := &dslRespExec{status: 200, pathParams: make(map[string]string), queryParams: make(map[string]string), headers: make(map[string]string)}
	dslRespExecStack = append(dslRespExecStack, re)

	// call various DSL helpers
	Description("descr")
	OperationID("opid")
	Deprecated()
	BearerAuth()
	Parameter("id", InPath, String)
	RequestBody(struct{ A string }{})
	QueryParamStruct(struct {
		Q string `query:"q"`
	}{})

	// response-side setters
	ResponseSchema(struct{ X int }{})
	ResponseHeader("X-Count", 1)
	SetParam("id", "1")
	SetQueryParam("q", "v")
	SetHeader("H", "v")
	SetBody(map[string]interface{}{"a": 1})
	SetRawBody([]byte("raw"), "text/plain")

	// verify op updated
	if op.description != "descr" || op.operationID != "opid" || !op.deprecated {
		t.Fatalf("op metadata not set correctly")
	}
	if len(op.security) == 0 {
		t.Fatalf("security not set on op")
	}
	if op.reqBodyModel == nil || op.queryStruct == nil {
		t.Fatalf("request body or query struct not set")
	}

	// verify response spec updated
	rs := op.responses[200]
	if rs == nil || rs.bodyModel == nil {
		t.Fatalf("response schema not set")
	}
	if _, ok := rs.headers["X-Count"]; !ok {
		t.Fatalf("response header not set")
	}

	// verify resp exec updated
	if re.pathParams["id"] != "1" || re.queryParams["q"] != "v" || re.headers["H"] != "v" {
		t.Fatalf("resp exec values not set")
	}

	// cleanup stacks
	dslRespExecStack = nil
	dslOpStack = nil
}

// TestFlushPendingDSLOpsRegisters ensures that pending DSL ops are flushed into
// the spec collector.
func TestFlushPendingDSLOpsRegisters(t *testing.T) {
	// create collector
	cfg := &Config{Title: "T", Version: "v"}
	sc := newSpecCollector(cfg)
	globalCollector = sc

	// ensure pending empty
	dslPendingOps = nil

	op := &dslOp{method: "GET", path: "/x", summary: "s", responses: make(map[int]*dslRespSpec)}
	enqueuePendingDSLOp(op)

	flushPendingDSLOps()

	// check spec has path
	if _, ok := sc.reflector.Spec.Paths.MapOfPathItemValues["/x"]; !ok {
		t.Fatalf("expected pending op to be registered into spec")
	}

	// reset collector
	globalCollector = nil
}

// TestDSLMethodWrappers_EnqueueOps verifies that the DSL wrapper helpers
// (Put/Patch/Delete) enqueue operations without panicking when called outside
// of a Ginkgo runtime.
func TestDSLMethodWrappers_EnqueueOps(t *testing.T) {
	// reset pending ops
	dslPendingOps = nil

	// ensure path stack present so ops get a path
	dslPathStack = []string{"/wraps"}

	// call wrappers — ensure they don't panic when invoked outside Ginkgo runtime.
	Put("put op", func() {})
	Patch("patch op", func() {})
	Delete("delete op", func() {})

	// cleanup
	dslPendingOps = nil
	dslPathStack = nil
}
