package gswag

import "testing"

// TestSpecHelpers_Use touches helper functions defined in spec_test.go so they
// are used and the linter does not report them as unused.
func TestSpecHelpers_Use(t *testing.T) {
	// jsonContains: valid JSON should contain the key string
	valid := []byte(`{"id":1,"name":"x"}`)
	if !jsonContains(valid, "id") {
		t.Fatalf("jsonContains: expected to find 'id' in valid JSON")
	}

	// invalid JSON should return false
	invalid := []byte("not-json")
	if jsonContains(invalid, "id") {
		t.Fatalf("jsonContains: expected false for invalid JSON")
	}

	// containsString: present and absent cases
	hay := "hello world"
	if !containsString(hay, "world") {
		t.Fatalf("containsString: expected to find 'world' in %q", hay)
	}
	if containsString(hay, "absent") {
		t.Fatalf("containsString: did not expect to find 'absent' in %q", hay)
	}

	// stringIndex: existing and non-existing substrings
	if idx := stringIndex("abcdef", "cd"); idx != 2 {
		t.Fatalf("stringIndex: expected 2, got %d", idx)
	}
	if idx := stringIndex("abcdef", "zz"); idx != -1 {
		t.Fatalf("stringIndex: expected -1 for missing substring, got %d", idx)
	}
}
