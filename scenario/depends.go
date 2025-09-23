package scenario

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gdt-dev/core/api"
	gdtcontext "github.com/gdt-dev/core/context"
	"github.com/gdt-dev/core/debug"
)

// checkDependencies examines the scenario's set of dependencies and returns a
// runtime error if any dependency isn't satisfied.
func (s *Scenario) checkDependencies(
	ctx context.Context,
) error {
	if len(s.Depends) == 0 {
		return nil
	}
	ctx = gdtcontext.PushTrace(ctx, "scenario.check-deps")
	defer func() {
		ctx = gdtcontext.PopTrace(ctx)
	}()

	for _, dep := range s.Depends {
		if err := s.checkDependency(ctx, dep); err != nil {
			return err
		}
	}
	return nil
}

// checkDependency returns an error if the supplied Dependency isn't satisfied.
func (s *Scenario) checkDependency(
	ctx context.Context,
	dep *api.Dependency,
) error {
	if dep == nil {
		return nil
	}

	when := dep.When
	if when != nil {
		if when.OS != "" {
			if !strings.EqualFold(runtime.GOOS, when.OS) {
				return nil
			}
		}
	}

	_, err := exec.LookPath(dep.Name)
	if err != nil {
		execErr, ok := err.(*exec.Error)
		if ok && execErr.Err == exec.ErrNotFound {
			return api.DependencyNotSatisfied(dep)
		} else {
			return fmt.Errorf(
				"error checking for program %q: %w",
				dep.Name, err,
			)
		}
	}

	if when != nil && when.Version != "" {
		// TODO(jaypipes): do some robust version checking for
		// applications/packages here.
		_ = when.Version
	}

	debug.Printf(ctx, "dependency %q satisfied", dep.Name)
	return nil
}
