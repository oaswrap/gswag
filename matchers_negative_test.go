package gswag

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type errReadCloser struct{ err error }

func (e *errReadCloser) Read(p []byte) (int, error) { return 0, e.err }
func (e *errReadCloser) Close() error               { return nil }

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

func TestHaveJSONBody_InvalidJSON(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("notjson"))}
	ok, err := HaveJSONBody(map[string]interface{}{"a": 1}).Match(resp)
	if err == nil {
		t.Fatalf("expected error for invalid JSON, got ok=%v", ok)
	}
}

func TestContainJSONKey_NonObject(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("[]"))}
	_, err := ContainJSONKey("k").Match(resp)
	if err == nil {
		t.Fatalf("expected error for non-object JSON in ContainJSONKey")
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

func TestMatchers_ToHTTPResponseTypeError(t *testing.T) {
	// passing wrong type should produce an error from toHTTPResponse via Match
	_, err := HaveStatus(200).Match("not a response")
	if err == nil {
		t.Fatalf("expected error when passing wrong type to matcher")
	}
}
