// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

import (
	"context"
)

// Runnable are things that Run either a `*testing.T` or a `*run.Options` that
// tracks execution of test units being run.
type Runnable interface {
	// Run accepts a context and either a `*testing.T` or a `*run.Options` and
	// runs some tests within that context.
	//
	// Errors returned by Run() are **RuntimeErrors**, not failures in
	// assertions.
	Run(context.Context, any) error
}
