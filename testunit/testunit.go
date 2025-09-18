// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package testunit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
)

const (
	indent = "   "
)

// TestUnit contains state about a unit under test. This class is used by the
// `gdt` CLI tool instead of the `*testing.T` struct which is used when a `gdt`
// test case is executed by the `go test` tool. An `*api.Spec` is converted
// into a TestUnit by the `Scenario.Run()` method. TestUnit implements `api.T`.
type TestUnit struct {
	sync.RWMutex
	ctx       context.Context
	cancelCtx context.CancelFunc
	// detail is our log stream the test unit can write results/messages to.
	detail *strings.Builder
	// name is the name/title of the test unit
	name string
	// parent points at another test unit if it's a subtest.
	parent *TestUnit
	// failed is true if the test unit has been marked as failed.
	failed bool
	// failures is a collection of assertion failures encountered for the test
	// unit.
	failures []error
	// skipped is true if the test unit has been marked as skipped.
	skipped bool
	// done is true if the test unit is finished and any subtests have
	// completed.
	done bool
	// start is the timestamp of when the test unit was started.
	started time.Time
	// elapsed is the amount of time spent executing the test unit.
	elapsed time.Duration
}

func (u *TestUnit) finish() {
	u.Lock()
	u.elapsed += time.Since(u.started)
	u.done = true
	u.Unlock()
}

// Name returns the full name of the test unit. The test unit name is a
// concatenation of the parent(s) name and this test unit's name.
func (u *TestUnit) Name() string {
	return u.name
}

// Elapsed returns the duration the test took to execute.
func (u *TestUnit) Elapsed() time.Duration {
	return u.elapsed
}

// Detail returns the saved log entries.
func (u *TestUnit) Detail() string {
	if u.detail != nil {
		return u.detail.String()
	}
	return ""
}

// Fail marks the function as having failed but continues execution.
func (u *TestUnit) Fail() {
	if u.parent != nil {
		u.parent.Fail()
	}
	u.Lock()
	defer u.Unlock()
	// u.done needs to be locked to synchronize checks to u.done in parent tests.
	if u.done {
		panic("Fail called after " + u.name + " has completed")
	}
	u.failed = true
}

// Failed reports whether the test unit has failed.
func (u *TestUnit) Failed() bool {
	u.RLock()
	defer u.RUnlock()

	return u.failed
}

// FailNow marks the function as having failed and stops its execution
// Execution will continue at the next test unit.
func (u *TestUnit) FailNow() {
	u.Fail()
	u.finish()
}

func (u *TestUnit) log(s string) {
	if u.detail == nil {
		return
	}
	s = strings.TrimSuffix(s, "\n")
	// Second and subsequent lines are indented 4 spaces. This is in addition to
	// the indentation provided by outputWriter.
	s = strings.ReplaceAll(s, "\n", "\n"+indent)
	s += "\n"
	u.detail.WriteString(s)
}

// Log writes an entry to the detail log, ensuring a newline at the end of the
// log line.
func (u *TestUnit) Log(args ...any) {
	u.log(fmt.Sprintln(args...))
}

// Logf writes an entry to the detail log.
func (u *TestUnit) Logf(format string, args ...any) {
	u.log(fmt.Sprintf(format, args...))
}

// Error adds the supplied errors or error message strings to the test unit's
// collected assertion failures. Execution will continue after marking the test
// unit as failed.
func (u *TestUnit) Error(args ...any) {
	errs := lo.Map(args, func(arg any, _ int) error {
		switch arg := arg.(type) {
		case error:
			return arg
		default:
			return fmt.Errorf("%s", arg)
		}
	})
	u.failures = append(u.failures, errs...)
	u.Fail()
}

// Errorf adds an error to the test unit's collected assertion failures.
// Execution will continue after marking the test unit as failed.
func (u *TestUnit) Errorf(format string, args ...any) {
	u.failures = append(u.failures, fmt.Errorf(format, args...))
	u.Fail()
}

// Fatal adds the supplied errors or error message strings to the test unit's
// collected assertion failures and immediately stops execution of the test
// unit.
func (u *TestUnit) Fatal(args ...any) {
	u.Error(args...)
	u.FailNow()
}

// Fatalf adds an error to the to the test unit's collected assertion failures
// and immediately stops execution of the test unit.
func (u *TestUnit) Fatalf(format string, args ...any) {
	u.Errorf(format, args...)
	u.FailNow()
}

// Skip is equivalent to Log followed by SkipNow.
func (u *TestUnit) Skip(args ...any) {
	u.Log(args...)
	u.SkipNow()
}

// Skipf is equivalent to Logf followed by SkipNow.
func (u *TestUnit) Skipf(format string, args ...any) {
	u.Logf(format, args...)
	u.SkipNow()
}

// SkipNow marks the test unit as having been skipped and stops its execution.
func (u *TestUnit) SkipNow() {
	u.Lock()
	defer u.RUnlock()
	u.skipped = true
	u.finish()
}

// Skipped reports whether the test was skipped.
func (u *TestUnit) Skipped() bool {
	u.RLock()
	defer u.RUnlock()
	return u.skipped
}
