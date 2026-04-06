package gswag

import (
	"testing"
)

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
