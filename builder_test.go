package gswag_test

import (
	"testing"

	"github.com/oaswrap/gswag"
)

// --- DSL type constants ---

func TestParamLocationConstants(t *testing.T) {
	if gswag.PathParam != gswag.InPath {
		t.Error("PathParam should equal InPath")
	}
	if gswag.QueryParam != gswag.InQuery {
		t.Error("QueryParam should equal InQuery")
	}
	if gswag.HeaderParam != gswag.InHeader {
		t.Error("HeaderParam should equal InHeader")
	}
	if gswag.CookieParam != gswag.InCookie {
		t.Error("CookieParam should equal InCookie")
	}
}

func TestSchemaTypeConstants(t *testing.T) {
	cases := []struct {
		val  gswag.SchemaType
		want string
	}{
		{gswag.String, "string"},
		{gswag.Integer, "integer"},
		{gswag.Number, "number"},
		{gswag.Boolean, "boolean"},
		{gswag.Object, "object"},
		{gswag.Array, "array"},
	}
	for _, tc := range cases {
		if string(tc.val) != tc.want {
			t.Errorf("SchemaType %q: want %q, got %q", tc.val, tc.want, string(tc.val))
		}
	}
}

// --- SetTestServer ---

func TestSetTestServer_AcceptsServer(t *testing.T) {
	// SetTestServer should not panic for valid targets.
	gswag.SetTestServer("http://localhost:9999")
	// Restore to nil for other tests.
	gswag.SetTestServer(nil)
}

func TestSetTestServer_AllowsArbitraryTargetType(t *testing.T) {
	// SetTestServer stores the target as-is; validation happens when requests run.
	gswag.SetTestServer(12345)
	gswag.SetTestServer(nil)
}
