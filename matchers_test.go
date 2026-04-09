package gswag

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// newTestResponse builds a minimal *http.Response for matcher tests.
func newTestResponse(status int, body string, headers http.Header) *http.Response {
	if headers == nil {
		headers = http.Header{}
	}
	return &http.Response{
		StatusCode: status,
		Header:     headers,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

//nolint:unparam // headers may be passed in future tests; keep parameter for flexibility
func respWith(body string, status int, headers map[string]string) *http.Response {
	h := http.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{StatusCode: status, Header: h, Body: io.NopCloser(strings.NewReader(body))}
}

type errReadCloser struct{ err error }

func (e *errReadCloser) Read(p []byte) (int, error) { return 0, e.err }
func (e *errReadCloser) Close() error               { return nil }

// --- HaveStatus ---

func TestHaveStatus_Match(t *testing.T) {
	m := HaveStatus(200)
	ok, err := m.Match(newTestResponse(200, "", nil))
	if err != nil || !ok {
		t.Fatalf("expected match for status 200, err=%v ok=%v", err, ok)
	}
}

func TestHaveStatus_NoMatch(t *testing.T) {
	m := HaveStatus(200)
	ok, _ := m.Match(newTestResponse(404, "", nil))
	if ok {
		t.Fatal("expected no match for status 404")
	}
}

func TestHaveStatus_FailureMessage(t *testing.T) {
	m := HaveStatus(200)
	msg := m.FailureMessage(newTestResponse(404, "not found", nil))
	if msg == "" {
		t.Fatal("expected non-empty failure message")
	}
}

func TestHaveStatus_NegatedFailureMessage(t *testing.T) {
	m := HaveStatus(200)
	msg := m.NegatedFailureMessage(newTestResponse(200, "", nil))
	if msg == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestHaveStatus_InvalidActual(t *testing.T) {
	m := HaveStatus(200)
	_, err := m.Match("not a response")
	if err == nil {
		t.Fatal("expected error for invalid actual type")
	}
}

// extra failure-message-focused test.
func TestHaveStatus_FailureMessages(t *testing.T) {
	resp := respWith("oops", 500, nil)
	m := HaveStatus(200)
	ok, err := m.Match(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected not ok")
	}
	fm := m.(*haveStatusMatcher).FailureMessage(resp)
	if !strings.Contains(fm, "Expected status 200 but got 500") {
		t.Fatalf("unexpected failure message: %s", fm)
	}
}

// --- HaveStatusInRange ---

func TestHaveStatusInRange_Match(t *testing.T) {
	m := HaveStatusInRange(200, 299)
	for _, code := range []int{200, 201, 204, 299} {
		ok, err := m.Match(newTestResponse(code, "", nil))
		if err != nil || !ok {
			t.Fatalf("expected match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_NoMatch(t *testing.T) {
	m := HaveStatusInRange(200, 299)
	for _, code := range []int{199, 300, 404, 500} {
		ok, _ := m.Match(newTestResponse(code, "", nil))
		if ok {
			t.Fatalf("expected no match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_Messages(t *testing.T) {
	m := HaveStatusInRange(200, 299)
	if m.FailureMessage(newTestResponse(500, "", nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestHaveStatusInRange_Failure(t *testing.T) {
	resp := respWith("", 302, nil)
	m := HaveStatusInRange(200, 299)
	ok, _ := m.Match(resp)
	if ok {
		t.Fatalf("expected not ok")
	}
	fm := m.(*haveStatusRangeMatcher).FailureMessage(resp)
	if !strings.Contains(fm, "Expected status in [200, 299] but got 302") {
		t.Fatalf("unexpected msg: %s", fm)
	}
}

// --- HaveHeader ---

func TestHaveHeader_Match(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	m := HaveHeader("Content-Type", "application/json")
	ok, err := m.Match(newTestResponse(200, "", h))
	if err != nil || !ok {
		t.Fatalf("expected match for Content-Type header, err=%v ok=%v", err, ok)
	}
}

func TestHaveHeader_NoMatch(t *testing.T) {
	m := HaveHeader("Content-Type", "application/json")
	ok, _ := m.Match(newTestResponse(200, "", nil))
	if ok {
		t.Fatal("expected no match for missing header")
	}
}

func TestHaveHeader_Messages(t *testing.T) {
	m := HaveHeader("X-Key", "val")
	if m.FailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestHaveHeaderAndJSONMatchers_FailureMessages(t *testing.T) {
	resp := respWith(`notjson`, 200, nil)

	// header missing
	hm := HaveHeader("X-Req", "v")
	ok, _ := hm.Match(resp)
	if ok {
		t.Fatalf("expected header match to fail")
	}
	hfm := hm.(*haveHeaderMatcher).FailureMessage(resp)
	if !strings.Contains(hfm, "Expected header") {
		t.Fatalf("unexpected header failure: %s", hfm)
	}

	// JSON body equality failure message
	jb := HaveJSONBody(map[string]any{"a": 1})
	ok2, err := jb.Match(resp)
	if err == nil && ok2 {
		t.Fatalf("expected JSON match to fail")
	}
	jfm := jb.(*haveJSONBodyMatcher).FailureMessage(resp)
	if !strings.Contains(jfm, "Expected response body to equal") {
		t.Fatalf("unexpected json failure message: %s", jfm)
	}

	// ContainJSONKey failure message
	ck := ContainJSONKey("x")
	ok3, _ := ck.Match(resp)
	if ok3 {
		t.Fatalf("expected contain json key to fail")
	}
	if !strings.Contains(ck.(*containJSONKeyMatcher).FailureMessage(nil), "contain JSON key") {
		t.Fatalf("unexpected contain json key failure")
	}
}

// --- HaveJSONBody ---

func TestHaveJSONBody_Match(t *testing.T) {
	m := HaveJSONBody(map[string]any{"id": float64(1)})
	ok, err := m.Match(newTestResponse(200, `{"id":1}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match for JSON body, err=%v ok=%v", err, ok)
	}
}

func TestHaveJSONBody_NoMatch(t *testing.T) {
	m := HaveJSONBody(map[string]any{"id": float64(1)})
	ok, _ := m.Match(newTestResponse(200, `{"id":2}`, nil))
	if ok {
		t.Fatal("expected no match for different JSON body")
	}
}

func TestHaveJSONBody_MultipleReads(t *testing.T) {
	// Body should be re-readable after first matcher read.
	resp := newTestResponse(200, `{"id":1}`, nil)
	m := HaveJSONBody(map[string]any{"id": float64(1)})
	ok1, _ := m.Match(resp)
	ok2, _ := m.Match(resp) // second read must also work
	if !ok1 || !ok2 {
		t.Fatalf("expected both reads to match, ok1=%v ok2=%v", ok1, ok2)
	}
}

func TestHaveJSONBody_InvalidJSON(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("notjson"))}
	ok, err := HaveJSONBody(map[string]any{"a": 1}).Match(resp)
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got ok=%v", ok)
	}
}

// --- ContainJSONKey ---

func TestContainJSONKey_Match(t *testing.T) {
	m := ContainJSONKey("id")
	ok, err := m.Match(newTestResponse(200, `{"id":1,"name":"x"}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match, err=%v ok=%v", err, ok)
	}
}

func TestContainJSONKey_NoMatch(t *testing.T) {
	m := ContainJSONKey("missing")
	ok, _ := m.Match(newTestResponse(200, `{"id":1}`, nil))
	if ok {
		t.Fatal("expected no match for absent key")
	}
}

func TestContainJSONKey_Messages(t *testing.T) {
	m := ContainJSONKey("id")
	if m.FailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestContainJSONKey_NonObject(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("[]"))}
	_, err := ContainJSONKey("k").Match(resp)
	if err == nil {
		t.Fatalf("expected error for non-object JSON in ContainJSONKey")
	}
}

// --- HaveNonEmptyBody ---

func TestHaveNonEmptyBody_Match(t *testing.T) {
	m := HaveNonEmptyBody()
	ok, err := m.Match(newTestResponse(200, `{"ok":true}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match for non-empty body, err=%v ok=%v", err, ok)
	}
}

func TestHaveNonEmptyBody_NoMatch(t *testing.T) {
	m := HaveNonEmptyBody()
	ok, _ := m.Match(newTestResponse(200, "", nil))
	if ok {
		t.Fatal("expected no match for empty body")
	}
}

func TestHaveNonEmptyBody_Empty(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(""))}
	ok, err := HaveNonEmptyBody().Match(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected HaveNonEmptyBody to be false for empty body")
	}
}

func TestHaveNonEmptyBodyAndMatchJSONSchemaFailures(t *testing.T) {
	empty := respWith("", 200, nil)
	nb := HaveNonEmptyBody()
	ok, _ := nb.Match(empty)
	if ok {
		t.Fatalf("expected empty body to be false")
	}
	if !strings.Contains(nb.(*haveNonEmptyBodyMatcher).FailureMessage(nil), "non-empty") {
		t.Fatalf("unexpected non-empty failure")
	}

	// MatchJSONSchema: model has field 'id' but body lacks it
	body := respWith(`{"name":"x"}`, 200, nil)
	ms := MatchJSONSchema(struct {
		ID string `json:"id"`
	}{})
	ok2, err := ms.Match(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok2 {
		t.Fatalf("expected schema match to fail")
	}
	if !strings.Contains(ms.(*matchJSONSchemaMatcher).FailureMessage(body), "structurally match schema") {
		t.Fatalf("unexpected schema failure message")
	}
}

// --- MatchJSONSchema ---

func TestMatchJSONSchema_Match(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	m := MatchJSONSchema(&Item{})
	ok, err := m.Match(newTestResponse(200, `{"id":1,"name":"x","extra":"y"}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected structural match, err=%v ok=%v", err, ok)
	}
}

func TestMatchJSONSchema_NoMatch(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	m := MatchJSONSchema(&Item{})
	ok, _ := m.Match(newTestResponse(200, `{"id":1}`, nil)) // missing "name"
	if ok {
		t.Fatal("expected no match when required key is absent")
	}
}

func TestMatchJSONSchema_Messages(t *testing.T) {
	type Item struct {
		ID int `json:"id"`
	}
	m := MatchJSONSchema(&Item{})
	if m.FailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestMatchJSONSchema_ModelNotObject(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(`{"x":1}`))}
	// model is a primitive -> should return true, nil (skips structural check)
	ok, err := MatchJSONSchema("primitive").Match(resp)
	if err != nil || !ok {
		t.Fatalf("expected true,nil when model is not an object, got %v,%v", ok, err)
	}
}

// --- Negative / error-path tests ---

func TestReadBody_NilAndError(t *testing.T) {
	// nil Body
	r := &http.Response{Body: nil}
	b, err := readBody(r)
	if err != nil || b != nil {
		t.Fatalf("expected nil,nil for nil body, got %v,%v", b, err)
	}

	// error on Read
	r2 := &http.Response{Body: &errReadCloser{err: errors.New("boom")}}
	_, err = readBody(r2)
	if err == nil {
		t.Fatalf("expected error from readBody when Read fails")
	}
}

func TestMatchers_ToHTTPResponseTypeError(t *testing.T) {
	// passing wrong type should produce an error from toHTTPResponse via Match
	_, err := HaveStatus(200).Match("not a response")
	if err == nil {
		t.Fatalf("expected error when passing wrong type to matcher")
	}
}
