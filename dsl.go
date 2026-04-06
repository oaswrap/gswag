package gswag

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/onsi/ginkgo/v2"
)

// dslServer is the HTTP test target set by SetTestServer.
var dslServer interface{}

// dslPathStack tracks nested path prefixes during Ginkgo tree construction.
var dslPathStack []string

// dslOpStack tracks the current operation context during tree construction.
var dslOpStack []*dslOp

// dslRespExecStack tracks response execution contexts during tree construction.
var dslRespExecStack []*dslRespExec

var dslPendingMu sync.Mutex
var dslPendingOps []*dslOp

// dslOp holds spec-level metadata for one HTTP operation.
type dslOp struct {
	method       string
	path         string
	summary      string
	description  string
	tags         []string
	operationID  string
	deprecated   bool
	hidden       bool
	security     []map[string][]string
	params       []dslParam
	reqBodyModel interface{}
	queryStruct  interface{}
	responses    map[int]*dslRespSpec
}

// dslParam is a declared parameter (spec-side only; values come from dslRespExec).
type dslParam struct {
	name     string
	location ParamLocation
	typ      SchemaType
	required *bool
	explode  *bool
	enumVals []interface{}
	defVal   interface{}
	hasDef   bool
}

// ParameterOption customizes an operation parameter declared by Parameter.
type ParameterOption func(*dslParam)

// ParamRequired marks a parameter as required/optional in the generated spec.
func ParamRequired(required bool) ParameterOption {
	return func(p *dslParam) {
		p.required = &required
	}
}

// ParamExplode controls the OpenAPI explode flag for the parameter.
func ParamExplode(explode bool) ParameterOption {
	return func(p *dslParam) {
		p.explode = &explode
	}
}

// ParamEnum constrains parameter values to the provided enum values.
func ParamEnum(values ...interface{}) ParameterOption {
	return func(p *dslParam) {
		p.enumVals = append([]interface{}(nil), values...)
	}
}

// ParamDefault sets the parameter default value in the generated schema.
func ParamDefault(value interface{}) ParameterOption {
	return func(p *dslParam) {
		p.defVal = value
		p.hasDef = true
	}
}

// dslRespSpec holds spec-side response metadata for one status code.
type dslRespSpec struct {
	description string
	bodyModel   interface{}
	headers     map[string]interface{}
}

// dslRespExec holds test-execution values for one Response block.
type dslRespExec struct {
	status          int
	pathParams      map[string]string
	queryParams     map[string]string
	headers         map[string]string
	body            interface{}
	bodyRaw         []byte
	bodyContentType string
}

// SetTestServer registers the HTTP target used by RunTest.
func SetTestServer(target interface{}) {
	dslServer = target
}

// Path wraps fn in a Ginkgo Describe node and pushes template onto the path stack.
// Use at package level with var _ = Path(...).
func Path(template string, fn func()) bool {
	ginkgo.Describe(template, func() {
		dslPathStack = append(dslPathStack, template)
		defer func() {
			dslPathStack = dslPathStack[:len(dslPathStack)-1]
		}()
		fn()
	})
	return true
}

// Get declares a GET operation on the current path.
func Get(summary string, fn func()) { dslVerb(http.MethodGet, summary, fn) }

// Post declares a POST operation on the current path.
func Post(summary string, fn func()) { dslVerb(http.MethodPost, summary, fn) }

// Put declares a PUT operation on the current path.
func Put(summary string, fn func()) { dslVerb(http.MethodPut, summary, fn) }

// Patch declares a PATCH operation on the current path.
func Patch(summary string, fn func()) { dslVerb(http.MethodPatch, summary, fn) }

// Delete declares a DELETE operation on the current path.
func Delete(summary string, fn func()) { dslVerb(http.MethodDelete, summary, fn) }

func dslVerb(method, summary string, fn func()) {
	ginkgo.Describe(method+" "+summary, func() {
		path := strings.Join(dslPathStack, "")
		op := &dslOp{
			method:    method,
			path:      path,
			summary:   summary,
			responses: make(map[int]*dslRespSpec),
		}
		dslOpStack = append(dslOpStack, op)
		defer func() {
			dslOpStack = dslOpStack[:len(dslOpStack)-1]
		}()

		fn()
		enqueuePendingDSLOp(copyDslOp(op))
	})
}

// Tag appends one or more tags to the current operation.
func Tag(tags ...string) {
	topOp().tags = append(topOp().tags, tags...)
}

// Description sets the description of the current operation.
func Description(desc string) {
	topOp().description = desc
}

// OperationID sets the operationId of the current operation.
func OperationID(id string) {
	topOp().operationID = id
}

// Deprecated marks the current operation as deprecated in the spec.
func Deprecated() {
	topOp().deprecated = true
}

// Hidden excludes the current operation from the generated spec while still
// allowing RunTest to execute the underlying HTTP request.
func Hidden() {
	topOp().hidden = true
}

// Security adds a named security requirement to the current operation.
func Security(schemeName string, scopes ...string) {
	op := topOp()
	op.security = append(op.security, map[string][]string{schemeName: scopes})
}

// BearerAuth adds a Bearer JWT security requirement to the current operation.
func BearerAuth() {
	op := topOp()
	op.security = append(op.security, map[string][]string{bearerAuthSchemeName: {}})
}

// Parameter declares a named parameter for the current operation.
func Parameter(name string, in ParamLocation, typ SchemaType, opts ...ParameterOption) {
	op := topOp()
	p := dslParam{name: name, location: in, typ: typ}
	for _, opt := range opts {
		if opt != nil {
			opt(&p)
		}
	}
	op.params = append(op.params, p)
}

// RequestBody sets a typed struct as the request body schema for the current operation.
func RequestBody(model interface{}) {
	topOp().reqBodyModel = model
}

// QueryParamStruct registers a struct with query tags as query parameter schemas.
func QueryParamStruct(v interface{}) {
	topOp().queryStruct = v
}

// Response declares a response for the current operation and wraps fn in a Ginkgo Context.
func Response(status int, description string, fn func()) {
	ginkgo.Context(fmt.Sprintf("%d %s", status, description), func() {
		op := topOp()
		op.responses[status] = &dslRespSpec{description: description}

		respExec := &dslRespExec{
			status:      status,
			pathParams:  make(map[string]string),
			queryParams: make(map[string]string),
			headers:     make(map[string]string),
		}
		dslRespExecStack = append(dslRespExecStack, respExec)
		defer func() {
			dslRespExecStack = dslRespExecStack[:len(dslRespExecStack)-1]
		}()

		fn()
	})
}

// ResponseSchema sets the expected response body schema model for the current response.
func ResponseSchema(model interface{}) {
	re := topRespExec()
	op := topOp()
	if op.responses[re.status] == nil {
		op.responses[re.status] = &dslRespSpec{}
	}
	op.responses[re.status].bodyModel = model
}

// ResponseHeader declares a response header schema for the current response.
func ResponseHeader(name string, model interface{}) {
	re := topRespExec()
	op := topOp()
	if op.responses[re.status] == nil {
		op.responses[re.status] = &dslRespSpec{}
	}
	if op.responses[re.status].headers == nil {
		op.responses[re.status].headers = make(map[string]interface{})
	}
	op.responses[re.status].headers[name] = model
}

// SetParam sets a path parameter value for the current test case.
func SetParam(name, value string) {
	topRespExec().pathParams[name] = value
}

// SetQueryParam sets a query parameter value for the current test case.
func SetQueryParam(name, value string) {
	topRespExec().queryParams[name] = value
}

// SetHeader sets a request header for the current test case.
func SetHeader(name, value string) {
	topRespExec().headers[name] = value
}

// SetBody sets a typed request body for the current test case.
func SetBody(body interface{}) {
	topRespExec().body = body
}

// SetRawBody sets a raw request body for the current test case.
func SetRawBody(body []byte, contentType string) {
	re := topRespExec()
	re.bodyRaw = body
	re.bodyContentType = contentType
}

// RunTest registers a Ginkgo It block that fires the HTTP request and calls fn if provided.
func RunTest(fn ...func(*http.Response)) {
	method := topOp().method
	path := strings.Join(dslPathStack, "")
	respExecSnap := copyDslRespExec(topRespExec())

	ginkgo.It("executes request and validates response", func() {
		flushPendingDSLOps()

		target := dslServer
		if target == nil {
			ginkgo.Fail("gswag: test server not set - call gswag.SetTestServer() in BeforeSuite")
			return
		}

		b := dslBuildRequest(method, path, respExecSnap)
		recorded := b.do(target)

		if globalCollector != nil {
			globalCollector.injectInferredRequestSchema(b, recorded)
			globalCollector.injectRecordedResponseSchema(method, path, recorded)
			globalCollector.appendExamples(b, recorded)
		}

		if len(fn) > 0 && fn[0] != nil {
			fn[0](recordedToHTTPResponse(recorded))
		}
	})
}

// dslBuildRequest constructs a requestBuilder from execution-side snapshot only.
func dslBuildRequest(method, path string, re *dslRespExec) *requestBuilder {
	b := newRequestBuilder(method, path)
	for k, v := range re.pathParams {
		b.pathParams[k] = v
	}
	for k, v := range re.queryParams {
		b.queryParams[k] = v
	}
	for k, v := range re.headers {
		b.headers[k] = v
	}
	if re.body != nil {
		b.body = re.body
	}
	if len(re.bodyRaw) > 0 {
		b.bodyRaw = re.bodyRaw
		b.bodyContentType = re.bodyContentType
	}
	return b
}

// recordedToHTTPResponse converts a recordedResponse to an *http.Response.
func recordedToHTTPResponse(r *recordedResponse) *http.Response {
	return &http.Response{
		StatusCode: r.StatusCode,
		Header:     r.Headers,
		Body:       io.NopCloser(bytes.NewReader(r.BodyBytes)),
	}
}

func topOp() *dslOp {
	if len(dslOpStack) == 0 {
		panic("gswag: Tag/Parameter/RequestBody/QueryParamStruct/Response must be called inside Get/Post/Put/Patch/Delete")
	}
	return dslOpStack[len(dslOpStack)-1]
}

func topRespExec() *dslRespExec {
	if len(dslRespExecStack) == 0 {
		panic("gswag: SetParam/SetQueryParam/SetHeader/SetBody/SetRawBody/ResponseSchema/RunTest must be called inside Response")
	}
	return dslRespExecStack[len(dslRespExecStack)-1]
}

func copyDslOp(op *dslOp) *dslOp {
	cp := *op
	cp.tags = append([]string(nil), op.tags...)
	cp.params = append([]dslParam(nil), op.params...)
	for i := range cp.params {
		cp.params[i].required = cloneBoolPtr(op.params[i].required)
		cp.params[i].explode = cloneBoolPtr(op.params[i].explode)
		cp.params[i].enumVals = append([]interface{}(nil), op.params[i].enumVals...)
		cp.params[i].defVal = op.params[i].defVal
		cp.params[i].hasDef = op.params[i].hasDef
	}
	cp.security = nil
	for _, s := range op.security {
		m := make(map[string][]string, len(s))
		for k, v := range s {
			m[k] = append([]string(nil), v...)
		}
		cp.security = append(cp.security, m)
	}
	cp.responses = make(map[int]*dslRespSpec, len(op.responses))
	for k, v := range op.responses {
		c := *v
		if v.headers != nil {
			c.headers = make(map[string]interface{}, len(v.headers))
			for hk, hv := range v.headers {
				c.headers[hk] = hv
			}
		}
		cp.responses[k] = &c
	}
	return &cp
}

func cloneBoolPtr(v *bool) *bool {
	if v == nil {
		return nil
	}
	b := *v
	return &b
}

func copyDslRespExec(r *dslRespExec) *dslRespExec {
	cp := *r
	cp.pathParams = make(map[string]string, len(r.pathParams))
	for k, v := range r.pathParams {
		cp.pathParams[k] = v
	}
	cp.queryParams = make(map[string]string, len(r.queryParams))
	for k, v := range r.queryParams {
		cp.queryParams[k] = v
	}
	cp.headers = make(map[string]string, len(r.headers))
	for k, v := range r.headers {
		cp.headers[k] = v
	}
	cp.bodyRaw = append([]byte(nil), r.bodyRaw...)
	return &cp
}

func enqueuePendingDSLOp(op *dslOp) {
	dslPendingMu.Lock()
	defer dslPendingMu.Unlock()
	dslPendingOps = append(dslPendingOps, op)
}

func flushPendingDSLOps() {
	if globalCollector == nil {
		return
	}

	dslPendingMu.Lock()
	pending := dslPendingOps
	dslPendingOps = nil
	dslPendingMu.Unlock()

	for _, op := range pending {
		globalCollector.RegisterDSLOperation(op)
	}
}
