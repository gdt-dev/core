// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrUnknownField indicates that there was an unknown field in the parsing
	// of a spec or scenario.
	ErrUnknownField = errors.New("unknown field")
	// ErrFailure is the base error class for all errors that represent failed
	// assertions when evaluating a test.
	ErrFailure = errors.New("assertion failed")
	// ErrTimeoutExceeded is an ErrFailure when a test's execution exceeds a
	// timeout length.
	ErrTimeoutExceeded = fmt.Errorf("%s: timeout exceeded", ErrFailure)
	// ErrNotEqual is an ErrFailure when an expected thing doesn't equal an
	// observed thing.
	ErrNotEqual = fmt.Errorf("%w: not equal", ErrFailure)
	// ErrIn is an ErrFailure when a thing unexpectedly appears in an
	// container.
	ErrIn = fmt.Errorf("%w: in", ErrFailure)
	// ErrNotIn is an ErrFailure when an expected thing doesn't appear in an
	// expected container.
	ErrNotIn = fmt.Errorf("%w: not in", ErrFailure)
	// ErrNoneIn is an ErrFailure when none of a list of elements appears in an
	// expected container.
	ErrNoneIn = fmt.Errorf("%w: none in", ErrFailure)
	// ErrUnexpectedError is an ErrFailure when an unexpected error has
	// occurred.
	ErrUnexpectedError = fmt.Errorf("%w: unexpected error", ErrFailure)
)

// TimeoutExceeded returns an ErrTimeoutExceeded when a test's execution
// exceeds a timeout length. The optional failure parameter indicates a failed
// assertion that occurred before a timeout was reached.
func TimeoutExceeded(duration string, failure error) error {
	if failure != nil {
		return fmt.Errorf(
			"%w: timed out waiting for assertion to succeed (%s)",
			failure, duration,
		)
	}
	return fmt.Errorf("%s (%s)", ErrTimeoutExceeded, duration)
}

// NotEqualLength returns an ErrNotEqual when an expected length doesn't
// equal an observed length.
func NotEqualLength(exp, got int) error {
	return fmt.Errorf(
		"%w: expected length of %d but got %d",
		ErrNotEqual, exp, got,
	)
}

// NotEqual returns an ErrNotEqual when an expected thing doesn't equal an
// observed thing.
func NotEqual(exp, got interface{}) error {
	return fmt.Errorf("%w: expected %v but got %v", ErrNotEqual, exp, got)
}

// In returns an ErrIn when a thing unexpectedly appears in a container.
func In(element, container interface{}) error {
	return fmt.Errorf(
		"%w: expected %v not to contain %v",
		ErrIn, container, element,
	)
}

// NotIn returns an ErrNotIn when an expected thing doesn't appear in an
// expected container.
func NotIn(element, container interface{}) error {
	return fmt.Errorf(
		"%w: expected %v to contain %v",
		ErrNotIn, container, element,
	)
}

// NoneIn returns an ErrNoneIn when none of a list of elements appears in an
// expected container.
func NoneIn(elements, container interface{}) error {
	return fmt.Errorf(
		"%w: expected %v to contain one of %v",
		ErrNoneIn, container, elements,
	)
}

// UnexpectedError returns an ErrUnexpectedError when a supplied error is not
// expected.
func UnexpectedError(err error) error {
	return fmt.Errorf("%w: %s", ErrUnexpectedError, err)
}

var (
	// ErrUnknownSourceType indicates that a From() function was called with an
	// unknown source parameter type.
	ErrUnknownSourceType = errors.New("unknown source argument type")
)

// UnknownSourceType returns an ErrUnknownSourceType error describing the
// supplied parameter type.
func UnknownSourceType(source interface{}) error {
	return fmt.Errorf("%w: %T", ErrUnknownSourceType, source)
}

var (
	// RuntimeError is the base error class for all errors occurring during
	// runtime (and not during the parsing of a scenario or spec)
	// nolint:staticcheck
	RuntimeError = errors.New("runtime error")
	// ErrRequiredFixture is returned when a required fixture has not
	// been registered with the context.
	ErrRequiredFixture = fmt.Errorf(
		"%w: required fixture missing",
		RuntimeError,
	)
	// ErrDependencyNotSatisfied is returned when a required fixture has not
	// been registered with the context.
	ErrDependencyNotSatisfied = fmt.Errorf(
		"%w: dependency not satisfied",
		RuntimeError,
	)
	// ErrTimeoutConflict is returned when the Go test tool's timeout conflicts
	// with either a total wait time or a timeout in a scenario or test spec
	ErrTimeoutConflict = fmt.Errorf(
		"%w: timeout conflict",
		RuntimeError,
	)
)

// DependencyNotSatified returns an ErrDependencyNotSatisfied with the supplied
// dependency name and optional constraints.
func DependencyNotSatisfied(dep *Dependency) error {
	constraintsStr := ""
	constraints := []string{}
	progName := dep.Name
	if dep.When != nil {
		if dep.When.OS != "" {
			constraints = append(constraints, "OS:"+dep.When.OS)
		}
		if dep.When.Version != "" {
			constraints = append(constraints, "VERSION:"+dep.When.Version)
		}
		constraintsStr = fmt.Sprintf(" (%s)", strings.Join(constraints, ","))
	}
	return fmt.Errorf("%w: %s%s", ErrDependencyNotSatisfied, progName, constraintsStr)
}

// RequiredFixtureMissing returns an ErrRequiredFixture with the supplied
// fixture name
func RequiredFixtureMissing(name string) error {
	return fmt.Errorf("%w: %s", ErrRequiredFixture, name)
}

// TimeoutConflict returns an ErrTimeoutConflict describing how the Go test
// tool's timeout conflicts with either a total wait time or a timeout value
// from a scenario or spec.
func TimeoutConflict(
	ti *Timings,
) error {
	goTestTimeout := ti.GoTestTimeout
	totalWait := ti.TotalWait
	maxTimeout := ti.MaxTimeout
	msg := fmt.Sprintf(
		"go test -timeout value of %s ",
		(goTestTimeout + time.Second).Round(time.Second),
	)
	if totalWait > 0 {
		if totalWait.Abs() > goTestTimeout.Abs() {
			msg += fmt.Sprintf(
				"is shorter than the total wait time in the scenario: %s. "+
					"either decrease the wait times or increase the "+
					"go test -timeout value.",
				totalWait.Round(time.Second),
			)
		}
	} else {
		if maxTimeout.Abs() > goTestTimeout.Abs() {
			msg += fmt.Sprintf(
				"is shorter than the maximum timeout specified in the "+
					"scenario: %s. either decrease the scenario or spec "+
					"timeout or increase the go test -timeout value.",
				maxTimeout.Round(time.Second),
			)
		}
	}
	return fmt.Errorf("%w: %s", ErrTimeoutConflict, msg)
}
