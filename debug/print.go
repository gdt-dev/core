// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package debug

import (
	"context"
	"fmt"
	"strings"

	gdtcontext "github.com/gdt-dev/core/context"
)

// Printf writes a message with optional message arguments to the context's
// Debug output. The behaviour is analogous to `fmt.Printf`.
func Printf(
	ctx context.Context,
	format string,
	args ...any,
) {
	tu := gdtcontext.TestUnit(ctx)
	writers := gdtcontext.Debug(ctx)
	if len(writers) == 0 && tu == nil {
		return
	}

	trace := gdtcontext.Trace(ctx)

	prefix := gdtcontext.DebugPrefix(ctx)
	msg := prefix
	if trace != "" {
		msg += " [" + trace + "] "
	}
	msg += fmt.Sprintf(format, args...)
	msg = strings.TrimSuffix(msg, "\n") + "\n"
	for _, w := range writers {
		//nolint:errcheck
		w.Write([]byte(msg))
	}
	if tu != nil {
		tu.Log(strings.TrimSuffix(msg, "\n"))
	}
}

// Println writes a message with optional message arguments to the context's
// Debug output, ensuring there is a newline in the message line. This is
// analogous to `fmt.Println` behaviour.
func Println(
	ctx context.Context,
	args ...any,
) {
	tu := gdtcontext.TestUnit(ctx)
	writers := gdtcontext.Debug(ctx)
	if len(writers) == 0 && tu == nil {
		return
	}

	trace := gdtcontext.Trace(ctx)

	prefix := gdtcontext.DebugPrefix(ctx)
	msg := prefix
	if trace != "" {
		msg += " [" + trace + "] "
	}
	msg += fmt.Sprintln(args...)
	for _, w := range writers {
		//nolint:errcheck
		w.Write([]byte(msg))
	}
	if tu != nil {
		tu.Log(strings.TrimSuffix(msg, "\n"))
	}
}
