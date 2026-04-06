package gswag_test

import (
	"net/http"
	"testing"

	"github.com/oaswrap/gswag"
)

func newRecorded(status int, body []byte, headers http.Header) *gswag.RecordedResponse {
	if headers == nil {
		headers = http.Header{}
	}
	return &gswag.RecordedResponse{
		StatusCode: status,
		BodyBytes:  body,
		Headers:    headers,
	}
}

// --- HaveStatus ---

func TestHaveStatus_Match(t *testing.T) {
	m := gswag.HaveStatus(200)
	ok, err := m.Match(newRecorded(200, nil, nil))
	if err != nil || !ok {
		t.Fatalf("expected match for status 200, err=%v ok=%v", err, ok)
	}
}

func TestHaveStatus_NoMatch(t *testing.T) {
	m := gswag.HaveStatus(200)
	ok, err := m.Match(newRecorded(404, nil, nil))
	if err != nil || ok {
		t.Fatalf("expected no match for status 404, err=%v ok=%v", err, ok)
	}
}

func TestHaveStatus_FailureMessage(t *testing.T) {
	m := gswag.HaveStatus(200)
	msg := m.FailureMessage(newRecorded(404, []byte("not found"), nil))
	if msg == "" {
		t.Fatal("expected non-empty failure message")
	}
}

func TestHaveStatus_NegatedFailureMessage(t *testing.T) {
	m := gswag.HaveStatus(200)
	msg := m.NegatedFailureMessage(newRecorded(200, nil, nil))
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
		ok, err := m.Match(newRecorded(code, nil, nil))
		if err != nil || !ok {
			t.Fatalf("expected match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_NoMatch(t *testing.T) {
	m := gswag.HaveStatusInRange(200, 299)
	for _, code := range []int{199, 300, 404, 500} {
		ok, err := m.Match(newRecorded(code, nil, nil))
		if err != nil || ok {
			t.Fatalf("expected no match for status %d", code)
		}
	}
}

func TestHaveStatusInRange_Messages(t *testing.T) {
	m := gswag.HaveStatusInRange(200, 299)
	if m.FailureMessage(newRecorded(500, nil, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveHeader ---

func TestHaveHeader_Match(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	m := gswag.HaveHeader("Content-Type", "application/json")
	ok, err := m.Match(newRecorded(200, nil, h))
	if err != nil || !ok {
		t.Fatalf("expected header match, err=%v ok=%v", err, ok)
	}
}

func TestHaveHeader_NoMatch(t *testing.T) {
	m := gswag.HaveHeader("Content-Type", "application/json")
	ok, err := m.Match(newRecorded(200, nil, nil))
	if err != nil || ok {
		t.Fatalf("expected no match for missing header")
	}
}

func TestHaveHeader_Messages(t *testing.T) {
	m := gswag.HaveHeader("X-Foo", "bar")
	if m.FailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveJSONBody ---

func TestHaveJSONBody_Match(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	body := []byte(`{"id":1,"name":"Widget"}`)
	m := gswag.HaveJSONBody(Item{ID: 1, Name: "Widget"})
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || !ok {
		t.Fatalf("expected JSON body match, err=%v ok=%v", err, ok)
	}
}

func TestHaveJSONBody_NoMatch(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	body := []byte(`{"id":2,"name":"Other"}`)
	m := gswag.HaveJSONBody(Item{ID: 1, Name: "Widget"})
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || ok {
		t.Fatalf("expected no match for different JSON body")
	}
}

func TestHaveJSONBody_InvalidJSON(t *testing.T) {
	m := gswag.HaveJSONBody(map[string]string{"k": "v"})
	_, err := m.Match(newRecorded(200, []byte("not json"), nil))
	if err == nil {
		t.Fatal("expected error for invalid JSON body")
	}
}

func TestHaveJSONBody_Messages(t *testing.T) {
	m := gswag.HaveJSONBody(map[string]string{"k": "v"})
	if m.FailureMessage(newRecorded(200, []byte(`{}`), nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- ContainJSONKey ---

func TestContainJSONKey_Match(t *testing.T) {
	body := []byte(`{"id":"123","name":"Alice"}`)
	m := gswag.ContainJSONKey("id")
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || !ok {
		t.Fatalf("expected ContainJSONKey to match 'id', err=%v ok=%v", err, ok)
	}
}

func TestContainJSONKey_NoMatch(t *testing.T) {
	body := []byte(`{"name":"Alice"}`)
	m := gswag.ContainJSONKey("id")
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || ok {
		t.Fatalf("expected no match for missing key 'id'")
	}
}

func TestContainJSONKey_NotObject(t *testing.T) {
	m := gswag.ContainJSONKey("id")
	_, err := m.Match(newRecorded(200, []byte(`[1,2,3]`), nil))
	if err == nil {
		t.Fatal("expected error for non-object JSON body")
	}
}

func TestContainJSONKey_Messages(t *testing.T) {
	m := gswag.ContainJSONKey("id")
	if m.FailureMessage(newRecorded(200, []byte(`{}`), nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- HaveNonEmptyBody ---

func TestHaveNonEmptyBody_Match(t *testing.T) {
	m := gswag.HaveNonEmptyBody()
	ok, err := m.Match(newRecorded(200, []byte(`{}`), nil))
	if err != nil || !ok {
		t.Fatalf("expected match for non-empty body")
	}
}

func TestHaveNonEmptyBody_NoMatch(t *testing.T) {
	m := gswag.HaveNonEmptyBody()
	ok, err := m.Match(newRecorded(204, []byte{}, nil))
	if err != nil || ok {
		t.Fatalf("expected no match for empty body")
	}
}

func TestHaveNonEmptyBody_Messages(t *testing.T) {
	m := gswag.HaveNonEmptyBody()
	if m.FailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}

// --- MatchJSONSchema ---

func TestMatchJSONSchema_Match(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	body := []byte(`{"id":1,"name":"Widget","extra":"ignored"}`)
	m := gswag.MatchJSONSchema(Item{})
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || !ok {
		t.Fatalf("expected MatchJSONSchema to match, err=%v ok=%v", err, ok)
	}
}

func TestMatchJSONSchema_NoMatch_MissingKey(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	body := []byte(`{"id":1}`) // missing "name"
	m := gswag.MatchJSONSchema(Item{})
	ok, err := m.Match(newRecorded(200, body, nil))
	if err != nil || ok {
		t.Fatalf("expected no match when required key missing")
	}
}

func TestMatchJSONSchema_NonObjectModel(t *testing.T) {
	// Model is a slice → structural check is skipped → always passes.
	m := gswag.MatchJSONSchema([]string{})
	ok, err := m.Match(newRecorded(200, []byte(`{}`), nil))
	if err != nil || !ok {
		t.Fatalf("expected pass for non-object model, err=%v ok=%v", err, ok)
	}
}

func TestMatchJSONSchema_Messages(t *testing.T) {
	type Item struct {
		ID int `json:"id"`
	}
	m := gswag.MatchJSONSchema(Item{})
	if m.FailureMessage(newRecorded(200, []byte(`{}`), nil)) == "" {
		t.Fatal("expected non-empty failure message")
	}
	if m.NegatedFailureMessage(newRecorded(200, nil, nil)) == "" {
		t.Fatal("expected non-empty negated failure message")
	}
}
