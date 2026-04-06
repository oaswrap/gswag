package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/swaggest/openapi-go/openapi3"
)

func TestHasFlag(t *testing.T) {
	args := []string{"--foo", "--Bar", "baz"}
	if !hasFlag(args, "--foo") {
		t.Fatalf("expected --foo present")
	}
	if !hasFlag(args, "--bar") {
		t.Fatalf("expected case-insensitive --bar present")
	}
	if hasFlag(args, "--missing") {
		t.Fatalf("did not expect --missing")
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outC <- buf.String()
	}()
	f()
	w.Close()
	return <-outC
}

func TestPrintUsage(t *testing.T) {
	out := captureStdout(func() { printUsage() })
	if out == "" {
		t.Fatalf("expected usage output")
	}
	if !bytes.Contains([]byte(out), []byte("Commands:")) {
		t.Fatalf("usage missing Commands section")
	}
}

func TestPathMethodSetAndParamStatusAndOpChanges(t *testing.T) {
	// build base spec with GET /p and POST /p
	base := &openapi3.Spec{}
	base.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}

	pi := openapi3.PathItem{MapOfOperationValues: map[string]openapi3.Operation{}}
	// base GET with one required query param and responses 200,201
	getOp := openapi3.Operation{}
	req := true
	por := openapi3.ParameterOrRef{Parameter: &openapi3.Parameter{Name: "q", In: openapi3.ParameterIn("query"), Required: &req}}
	getOp.Parameters = []openapi3.ParameterOrRef{por}
	getOp.Responses = openapi3.Responses{MapOfResponseOrRefValues: map[string]openapi3.ResponseOrRef{"200": {}, "201": {}}}
	pi.MapOfOperationValues["get"] = getOp

	// next spec: GET /p without param and only 200; and added POST /p with required param
	next := &openapi3.Spec{}
	next.Paths = openapi3.Paths{MapOfPathItemValues: map[string]openapi3.PathItem{}}
	npi := openapi3.PathItem{MapOfOperationValues: map[string]openapi3.Operation{}}
	nget := openapi3.Operation{}
	nget.Responses = openapi3.Responses{MapOfResponseOrRefValues: map[string]openapi3.ResponseOrRef{"200": {}}}
	npi.MapOfOperationValues["get"] = nget
	// POST with new required param 'x'
	post := openapi3.Operation{}
	r2 := true
	por2 := openapi3.ParameterOrRef{Parameter: &openapi3.Parameter{Name: "x", In: openapi3.ParameterIn("query"), Required: &r2}}
	post.Parameters = []openapi3.ParameterOrRef{por2}
	post.Responses = openapi3.Responses{MapOfResponseOrRefValues: map[string]openapi3.ResponseOrRef{"200": {}}}
	npi.MapOfOperationValues["post"] = post

	base.Paths.MapOfPathItemValues["/p"] = pi
	next.Paths.MapOfPathItemValues["/p"] = npi

	// pathMethodSet
	s := pathMethodSet(base)
	if !s["get /p"] {
		t.Fatalf("expected get /p in set")
	}

	// diffSpecs
	dr := diffSpecs(base, next)
	// removed: none (get still present), added: post
	foundAdded := false
	for _, a := range dr.added {
		if a == "post /p" {
			foundAdded = true
		}
	}
	if !foundAdded {
		t.Fatalf("expected post /p to be listed as added")
	}

	// opChanges for GET should include breaking removed parameter and removed status 201
	changes := opChanges("/p", "get", getOp, nget)
	foundBreakingParam := false
	foundRemovedStatus := false
	for _, c := range changes {
		if containsBreaking(c) && (len(c) > 0) {
			if bytes.Contains([]byte(c), []byte("removed parameter")) {
				foundBreakingParam = true
			}
			if bytes.Contains([]byte(c), []byte("removed response status")) {
				foundRemovedStatus = true
			}
		}
	}
	if !foundBreakingParam {
		t.Fatalf("expected breaking removed parameter in opChanges: %v", changes)
	}
	if !foundRemovedStatus {
		t.Fatalf("expected removed response status in opChanges: %v", changes)
	}

	// containsBreakingResult
	if !containsBreakingResult(diffResult{removed: []string{"x"}}) {
		t.Fatalf("expected containsBreakingResult true when removed present")
	}
}

func TestRunInitCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	code := runInitNoExit([]string{dir})
	if code != 0 {
		t.Fatalf("runInitNoExit failed with code %d", code)
	}
	// expect suite_test.go and workflow file
	if _, err := os.Stat(filepath.Join(dir, "suite_test.go")); err != nil {
		t.Fatalf("suite_test.go not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", "gswag.yml")); err != nil {
		t.Fatalf("workflow not created: %v", err)
	}
}

func TestRunValidateNoExitBehaviours(t *testing.T) {
	// valid spec
	dir := t.TempDir()
	valid := `openapi: 3.0.3
info:
  title: X
  version: v1
paths: {}
`
	vp := filepath.Join(dir, "valid.yaml")
	os.WriteFile(vp, []byte(valid), 0o644) //nolint:errcheck
	if code := runValidateNoExit([]string{vp}); code != 0 {
		t.Fatalf("expected valid spec to return 0, got %d", code)
	}

	// invalid spec (missing info.title)
	bad := `openapi: 3.0.3
info:
  version: v1
paths: {}/n`
	bp := filepath.Join(dir, "bad.yaml")
	os.WriteFile(bp, []byte(bad), 0o644) //nolint:errcheck
	if code := runValidateNoExit([]string{bp}); code == 0 {
		t.Fatalf("expected invalid spec to return non-zero")
	}
}

func TestRunInitForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	// create existing suite_test.go with sentinel content
	sp := filepath.Join(dir, "suite_test.go")
	if err := os.WriteFile(sp, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	// run without force should succeed and skip overwrite
	if code := runInitNoExit([]string{dir}); code != 0 {
		t.Fatalf("runInitNoExit failed: %d", code)
	}
	b, err := os.ReadFile(sp)
	if err != nil {
		t.Fatalf("read suite_test: %v", err)
	}
	if string(b) == "" {
		t.Fatalf("expected suite_test.go to exist")
	}

	// run with --force should overwrite
	if code := runInitNoExit([]string{"--force", dir}); code != 0 {
		t.Fatalf("runInitNoExit --force failed: %d", code)
	}
	b2, err := os.ReadFile(sp)
	if err != nil {
		t.Fatalf("read suite_test after force: %v", err)
	}
	if bytes.Equal(b, b2) {
		t.Fatalf("expected file to be overwritten with --force")
	}
}

func TestRunInitMkdirAllError(t *testing.T) {
	dir := t.TempDir()
	// create a regular file at the target path so MkdirAll should fail
	target := filepath.Join(dir, "notadir")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	code := runInitNoExit([]string{target})
	if code == 0 {
		t.Fatalf("expected non-zero exit code when MkdirAll fails")
	}
}

func TestRunValidateNonexistentFile(t *testing.T) {
	// pick a clearly non-existent path
	code := runValidateNoExit([]string{"/no/such/path/definitely-not-exist.yaml"})
	if code == 0 {
		t.Fatalf("expected non-zero exit code for nonexistent spec file")
	}
}
