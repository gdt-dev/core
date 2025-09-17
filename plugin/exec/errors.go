// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package exec

import (
	"fmt"

	"github.com/gdt-dev/core/api"
)

// ExecRuntimeError returns a RuntimeError with an error from the Exec() call.
func ExecRuntimeError(err error) error {
	return fmt.Errorf("%w: %s", api.RuntimeError, err)
}
