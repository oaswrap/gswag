package gswag_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/oaswrap/gswag"
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

// --- HaveStatus ---

func TestHaveStatus_Match(t *testing.T) {
	m := gswag.HaveStatus(200)
	ok, err := m.Match(newTestResponse(200, "", nil))
	if err != nil || !ok {
		t.Fatalf("expected match for status 200, err=%v ok=%v", err, ok)
	}
}

func TestHaveStatus_NoMatch(t *testing.T) {
	m := gswag.HaveStatus(200)
	ok, _ := m.Match(newTestResponse(404, "", nil))
	if ok {
		t.Fatal("expected no match for status 404")
	}
}

func TestHaveStatus_FailureMessage(t *testing.T) {
	m := gswag.HaveStatus(200)
	msg := m.FailureMessage(newTestResponse(404, "not found", nil))
	if msg == "" {
		t.Fatal("expected non-empty failure message")
	}
}

func TestHaveStatus_NegatedFailureMessage(t *testing.T) {
	m := gswag.HaveStatus(200)
	msg := m.NegatedFailureMessage(newTestResponse(200, "", nil))
	if msg == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

func TestHaveStatus_InvalidActual(t *testing.T) {
	m := gswag.HaveStatus(200)
	_, err := m.Match("not a response")
	if err == nil {
		t.Fatal("expected error for invalid actual type")
	}
}

// --- HaveStatusInRange ---

func TestHaveStatusInRange_Match(t *testing.T) {
	m := gswag.HaveStatusInRange(200, 299)
	for _, code := range []int{200, 201, 204, 299} {
		ok, err := m.Match(newTestResponse(code, "", nil))
		if err != nil || !ok {
			t.Fatalf("expected match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_NoMatch(t *testing.T) {
	m := gswag.HaveStatusInRange(200, 299)
	for _, code := range []int{199, 300, 404, 500} {
		ok, _ := m.Match(newTestResponse(code, "", nil))
		if ok {
			t.Fatalf("expected no match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_Messages(t *testing.T) {
	m := gswag.HaveStatusInRange(200, 299)
	if m.FailureMessage(newTestResponse(500, "", nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveHeader ---

func TestHaveHeader_Match(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	m := gswag.HaveHeader("Content-Type", "application/json")
	ok, err := m.Match(newTestResponse(200, "", h))
	if err != nil || !ok {
		t.Fatalf("expected match for Content-Type header, err=%v ok=%v", err, ok)
	}
}

func TestHaveHeader_NoMatch(t *testing.T) {
	m := gswag.HaveHeader("Content-Type", "application/json")
	ok, _ := m.Match(newTestResponse(200, "", nil))
	if ok {
		t.Fatal("expected no match for missing header")
	}
}

func TestHaveHeader_Messages(t *testing.T) {
	m := gswag.HaveHeader("X-Key", "val")
	if m.FailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, "", nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveJSONBody ---

func TestHaveJSONBody_Match(t *testing.T) {
	m := gswag.HaveJSONBody(map[string]interface{}{"id": float64(1)})
	ok, err := m.Match(newTestResponse(200, `{"id":1}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match for JSON body, err=%v ok=%v", err, ok)
	}
}

func TestHaveJSONBody_NoMatch(t *testing.T) {
	m := gswag.HaveJSONBody(map[string]interface{}{"id": float64(1)})
	ok, _ := m.Match(newTestResponse(200, `{"id":2}`, nil))
	if ok {
		t.Fatal("expected no match for different JSON body")
	}
}

func TestHaveJSONBody_MultipleReads(t *testing.T) {
	// Body should be re-readable after first matcher read.
	resp := newTestResponse(200, `{"id":1}`, nil)
	m := gswag.HaveJSONBody(map[string]interface{}{"id": float64(1)})
	ok1, _ := m.Match(resp)
	ok2, _ := m.Match(resp) // second read must also work
	if !ok1 || !ok2 {
		t.Fatalf("expected both reads to match, ok1=%v ok2=%v", ok1, ok2)
	}
}

// --- ContainJSONKey ---

func TestContainJSONKey_Match(t *testing.T) {
	m := gswag.ContainJSONKey("id")
	ok, err := m.Match(newTestResponse(200, `{"id":1,"name":"x"}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match, err=%v ok=%v", err, ok)
	}
}

func TestContainJSONKey_NoMatch(t *testing.T) {
	m := gswag.ContainJSONKey("missing")
	ok, _ := m.Match(newTestResponse(200, `{"id":1}`, nil))
	if ok {
		t.Fatal("expected no match for absent key")
	}
}

func TestContainJSONKey_Messages(t *testing.T) {
	m := gswag.ContainJSONKey("id")
	if m.FailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveNonEmptyBody ---

func TestHaveNonEmptyBody_Match(t *testing.T) {
	m := gswag.HaveNonEmptyBody()
	ok, err := m.Match(newTestResponse(200, `{"ok":true}`, nil))
	if err != nil || !ok {
		t.Fatalf("expected match for non-empty body, err=%v ok=%v", err, ok)
	}
}

func TestHaveNonEmptyBody_NoMatch(t *testing.T) {
	m := gswag.HaveNonEmptyBody()
	ok, _ := m.Match(newTestResponse(200, "", nil))
	if ok {
		t.Fatal("expected no match for empty body")
	}
}

// --- MatchJSONSchema ---

func TestMatchJSONSchema_Match(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	m := gswag.MatchJSONSchema(&Item{})
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
	m := gswag.MatchJSONSchema(&Item{})
	ok, _ := m.Match(newTestResponse(200, `{"id":1}`, nil)) // missing "name"
	if ok {
		t.Fatal("expected no match when required key is absent")
	}
}

func TestMatchJSONSchema_Messages(t *testing.T) {
	type Item struct {
		ID int `json:"id"`
	}
	m := gswag.MatchJSONSchema(&Item{})
	if m.FailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newTestResponse(200, `{}`, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}
