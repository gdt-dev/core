// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package parse

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// ErrParseUnknownField indicates that there was an unknown field in the
	// parsing of a spec or scenario. This is a sentinel error we use in
	// parsing gdt test scenarios in the plugin system.
	ErrParseUnknownField = errors.New("unknown field")
)

// Error is a custom error type that stores the location of an error that
// occurred while parsing a gdt test specification.
type Error struct {
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

// Error implements the error interface for Error.
func (e *Error) Error() string {
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
func (e *Error) SetContents() {
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

// UnknownSpecAt returns an ErrUnknownSpec with the line/column of the supplied
// YAML node.
func UnknownSpecAt(path string, node *yaml.Node) error {
	return &Error{
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
		ErrParseUnknownField, field, node.Line, node.Column,
	)
}

// ExpectedMapAt returns a parse error for when a field that can contain a
// map[string]interface{} did not contain that.
func ExpectedMapAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected map field",
	}
}

// ErrExpectedMapOrYAMLString returns a parse error for when a field that can
// contain a map[string]interface{} or an embedded YAML string did not contain
// either of those things.
func ErrExpectedMapOrYAMLString(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected either map[string]interface{} or a string with embedded YAML",
	}
}

// ExpectedScalarAt returns an ErrExpectedScalar error annotated with
// the line/column of the supplied YAML node.
func ExpectedScalarAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar field",
	}
}

// ExpectedSequenceAt returns a parse error for when a field that can contain a
// []interface{} did not contain that.
func ExpectedSequenceAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected sequence field",
	}
}

// ExpectedSequenceAt returns a parse error for when a field that can contain an
// integer did not contain that.
func ExpectedIntAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected int value",
	}
}

// ErrExpectedScalarOrSequenceAt returns a parse error for when a field that
// can contain either a scalar or a []interface{} did not contain either of
// those things.
func ExpectedScalarOrSequenceAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar or sequence of scalars field",
	}
}

// ExpectedScalarOrMapAt returns an ErrExpectedScalarOrMap error annotated with
// the line/column of the supplied YAML node.
func ExpectedScalarOrMapAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected scalar or map field",
	}
}

// ExpectedTimeoutAt returns an ErrExpectedTimeout error annotated
// with the line/column of the supplied YAML node.
func ExpectedTimeoutAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected timeout specification",
	}
}

// ExpectedWaitAt returns an ErrExpectedWait error annotated with the
// line/column of the supplied YAML node.
func ExpectedWaitAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected wait specification",
	}
}

// ExpectedRetryAt returns an ErrExpectedRetry error annotated with the
// line/column of the supplied YAML node.
func ExpectedRetryAt(node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected retry specification",
	}
}

// InvalidRetryAttempts returns an ErrInvalidRetryAttempts error annotated with
// the line/column of the supplied YAML node.
func InvalidRetryAttempts(node *yaml.Node, attempts int) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("invalid retry attempts: %d", attempts),
	}
}

// FileNotFound returns ErrFileNotFound for a given file path
func FileNotFound(path string, node *yaml.Node) error {
	return &Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("file not found: %q", path),
	}
}
