// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package exec

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/gdt-dev/core/api"
	"github.com/gdt-dev/core/debug"
)

const (
	varFromStdout = "stdout"
	varFromStderr = "stderr"
	varFromRC     = "returncode"
)

type VarEntry struct {
	// From is a string that indicates where the value of the variable will be
	// sourced from. `stdout`, `stderr` and `returncode` indicate to source the
	// value of the variable from the output buffer for stdout, stderr or the
	// returncode value. All other strings indicate the value of the variable
	// should be sourced from an envvar of the same name.
	From string `yaml:"from"`
}

// Variables allows the test author to save arbitrary data to the test scenario,
// facilitating the passing of variables between test specs potentially
// provided by different gdt Plugins.
type Variables map[string]VarEntry

// saveVars examines the supplied Variables and what we got back from the
// Action.Do() call and sets any variables in the run data context key.
func saveVars(
	ctx context.Context,
	vars Variables,
	outbuf *bytes.Buffer,
	errbuf *bytes.Buffer,
	ec int,
	res *api.Result,
) {
	for varName, entry := range vars {
		switch entry.From {
		case varFromStdout:
			debug.Printf(ctx, "save.vars: %s -> <stdout>", varName)
			res.SetData(varName, strings.TrimSpace(outbuf.String()))
		case varFromStderr:
			debug.Printf(ctx, "save.vars: %s -> <stderr>", varName)
			res.SetData(varName, strings.TrimSpace(errbuf.String()))
		case varFromRC:
			debug.Printf(ctx, "save.vars: %s -> <returncode>", varName)
			res.SetData(varName, ec)
		default:
			extracted := os.Getenv(entry.From)
			debug.Printf(ctx, "save.vars: %s -> %s", varName, extracted)
			res.SetData(varName, extracted)
		}
	}
}
