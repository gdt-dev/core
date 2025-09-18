// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

// T is the shared interface surface area between an "internal" runnable test
// case when the `go test` tool is used as the test runner and an "external"
// runnable test case when the `gdt` CLI tool is used as the test runner. It's
// essentially a subset of Go's `testing.TB` interface methods.
type T interface {
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	Name() string
	Skip(args ...any)
	SkipNow()
	Skipf(format string, args ...any)
	Skipped() bool
}
