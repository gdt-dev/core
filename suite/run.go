// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package suite

import (
	"context"
)

// Run executes the tests in the test suite
func (s *Suite) Run(ctx context.Context, subject any) error {
	for _, sc := range s.Scenarios {
		if err := sc.Run(ctx, subject); err != nil {
			return err
		}
	}
	return nil
}
