// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	// ErrParse indicates a YAML definition is not valid
	ErrParse = errors.New("parse error")
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

// ParseError is a custom error type that stores the location of an error that
// occurred while parsing a gdt test specification.
type ParseError struct {
	// Path is the filepath to the parsed document.
	Path string
	// Line is the line number where the parse error occurred.
	Line int
	// Column is the column number where the parse error occurred.
	Column int
	// Message is the error message.
	Message string
	// Contents is the contents of the file read at Path.
	Contents string
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	contents := ""
	if e.Contents != "" {
		contents = fmt.Sprintf("\n%s\n", e.Contents)
	}
	return fmt.Sprintf(
		"error parsing %q: at line %d, column %d:\n%s\n%s",
		e.Path, e.Line, e.Column, e.Message, contents,
	)
}

// SetContents adds the detail to the error message for surrounding contents if
// the Path, Line and Column is set.
func (e *ParseError) SetContents() {
	if e.Path != "" {
		f, err := os.Open(e.Path)
		if err != nil {
			// just ignore...
			return
		}
		defer f.Close()

		b := &strings.Builder{}
		viewStartLine := max(0, e.Line-2)
		viewEndLine := e.Line + 2

		sc := bufio.NewScanner(f)
		x := 0
		for sc.Scan() {
			x++
			line := sc.Text()
			if x > viewEndLine {
				break
			}
			if x < viewStartLine {
				continue
			}
			_, _ = fmt.Fprintf(b, "%03d: %s\n", x, line)
			if x == e.Line {
				_, _ = fmt.Fprintf(b, "  %s^", strings.Repeat(" ", e.Column))
			}
		}
		if err := sc.Err(); err != nil {
			// just ignore...
			return
		}
		e.Contents = b.String()
	}
}

func (e *ParseError) Unwrap() error {
	return ErrParse
}

var (
	// ErrUnknownSourceType indicates that a From() function was called with an
	// unknown source parameter type.
	ErrUnknownSourceType = errors.New("unknown source argument type")
)

// UnknownSpecAt returns an ErrUnknownSpec with the line/column of the supplied
// YAML node.
func UnknownSpecAt(path string, node *yaml.Node) error {
	return &ParseError{
		Path:    path,
		Line:    node.Line,
		Column:  node.Column,
		Message: "no plugin could parse spec definition",
	}
}

// UnknownFieldAt returns an ErrUnknownField for a supplied field annotated
// with the line/column of the supplied YAML node.
func UnknownFieldAt(field string, node *yaml.Node) error {
	return fmt.Errorf(
		"%w: %q at line %d, column %d",
		ErrUnknownField, field, node.Line, node.Column,
	)
}

// ExpectedMapAt returns an ErrExpectedMap error annotated with the
// line/column of the supplied YAML node.
func ExpectedMapAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected map field",
	}
}

// ExpectedScalarAt returns an ErrExpectedScalar error annotated with
// the line/column of the supplied YAML node.
func ExpectedScalarAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar field",
	}
}

// ExpectedSequenceAt returns an ErrExpectedSequence error annotated
// with the line/column of the supplied YAML node.
func ExpectedSequenceAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected sequence field",
	}
}

// ExpectedIntAt returns an ErrExpectedInt error annotated
// with the line/column of the supplied YAML node.
func ExpectedIntAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected int value",
	}
}

// ExpectedScalarOrSequenceAt returns an ErrExpectedScalarOrSequence error
// annotated with the line/column of the supplied YAML node.
func ExpectedScalarOrSequenceAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar or sequence of scalars field",
	}
}

// ExpectedScalarOrMapAt returns an ErrExpectedScalarOrMap error annotated with
// the line/column of the supplied YAML node.
func ExpectedScalarOrMapAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar or map field",
	}
}

// ExpectedTimeoutAt returns an ErrExpectedTimeout error annotated
// with the line/column of the supplied YAML node.
func ExpectedTimeoutAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected timeout specification",
	}
}

// ExpectedWaitAt returns an ErrExpectedWait error annotated with the
// line/column of the supplied YAML node.
func ExpectedWaitAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected wait specification",
	}
}

// ExpectedRetryAt returns an ErrExpectedRetry error annotated with the
// line/column of the supplied YAML node.
func ExpectedRetryAt(node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected retry specification",
	}
}

// InvalidRetryAttempts returns an ErrInvalidRetryAttempts error annotated with
// the line/column of the supplied YAML node.
func InvalidRetryAttempts(node *yaml.Node, attempts int) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("invalid retry attempts: %d", attempts),
	}
}

// UnknownSourceType returns an ErrUnknownSourceType error describing the
// supplied parameter type.
func UnknownSourceType(source interface{}) error {
	return fmt.Errorf("%w: %T", ErrUnknownSourceType, source)
}

// FileNotFound returns ErrFileNotFound for a given file path
func FileNotFound(path string, node *yaml.Node) error {
	return &ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("file not found: %q", path),
	}
}

var (
	// RuntimeError is the base error class for all errors occurring during
	// runtime (and not during the parsing of a scenario or spec)
	RuntimeError = errors.New("runtime error")
	// ErrRequiredFixture is returned when a required fixture has not
	// been registered with the context.
	ErrRequiredFixture = fmt.Errorf(
		"%w: required fixture missing",
		RuntimeError,
	)
	// ErrTimeoutConflict is returned when the Go test tool's timeout conflicts
	// with either a total wait time or a timeout in a scenario or test spec
	ErrTimeoutConflict = fmt.Errorf(
		"%w: timeout conflict",
		RuntimeError,
	)
)

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
