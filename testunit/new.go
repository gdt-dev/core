// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package testunit

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	nameSeparator = "/"
)

type Option func(*TestUnit)

// WithParent creates TestUnit pointing at a parent test unit.
func WithParent(parent *TestUnit) Option {
	return func(u *TestUnit) {
		u.parent = parent
		if u.name != "" {
			u.name = fmt.Sprintf("%s%s%s", parent.name, nameSeparator, u.name)
		} else {
			u.name = u.parent.name
		}
	}
}

// WithName creates TestUnit with a specified test unit name.
func WithName(name string) Option {
	return func(u *TestUnit) {
		if u.parent != nil {
			u.name = fmt.Sprintf("%s%s%s", u.parent.name, nameSeparator, name)
		} else {
			u.name = name
		}
	}
}

// New returns a new initialized *TestUnit
func New(ctx context.Context, opts ...Option) *TestUnit {
	u := &TestUnit{
		detail:   &strings.Builder{},
		failures: []error{},
	}
	for _, opt := range opts {
		opt(u)
	}
	u.ctx, u.cancelCtx = context.WithCancel(ctx)
	u.started = time.Now()
	return u
}
