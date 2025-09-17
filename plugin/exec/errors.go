// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package exec

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/gdt-dev/core/api"
)

// ExecEmpty returns an ErrExecEmpty with the line/column of the supplied YAML
// node.
func ExecEmpty(node *yaml.Node) error {
	return &api.ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: "expected non-empty exec field",
	}
}

// ExecInvalidShellParse returns an ErrExecInvalid with the error from
// shlex.Split
func ExecInvalidShellParse(err error, node *yaml.Node) error {
	return &api.ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("cannot parse shell args: %s", err),
	}
}

// ExecUnknownShell returns a wrapped version of ParseError that indicates the
// user specified an unknown shell.
func ExecUnknownShell(shell string, node *yaml.Node) error {
	return &api.ParseError{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("unknown shell %q", shell),
	}
}

// ExecRuntimeError returns a RuntimeError with an error from the Exec() call.
func ExecRuntimeError(err error) error {
	return fmt.Errorf("%w: %s", api.RuntimeError, err)
}
