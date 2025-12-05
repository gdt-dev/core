package scenario

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/gdt-dev/core/api"
	gdtcontext "github.com/gdt-dev/core/context"
	"github.com/gdt-dev/core/debug"
)

var defaultVersionSelectorArgs = []string{"-v"}

// looseSemVerRegex is a regular expression that lets invalid semver
// expressions through. Taken from semver library.
const defaultVersionSelectorFilter string = `v?([0-9]+)(\.[0-9]+)?(\.[0-9]+)?` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?`

var (
	defaultVersionSelector = &api.DependencyVersionSelector{
		Args:   defaultVersionSelectorArgs,
		Filter: defaultVersionSelectorFilter,
	}
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

	binPath, err := exec.LookPath(dep.Name)
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

	dv := dep.Version
	if dv != nil {
		vc := dv.SemVerConstraints
		if vc != nil {
			verStr, err := versionStringFromDependency(binPath, dv.Selector)
			if err != nil {
				return err
			}
			ver, err := semver.NewVersion(verStr)
			if err != nil {
				return api.DependencyNotSatisfied(dep)
			}
			if !vc.Check(ver) {
				return api.DependencyNotSatisfiedVersionConstraint(
					dep, dv.Constraint,
				)
			}
		}
	}

	debug.Printf(ctx, "dependency %q satisfied", dep.Name)
	return nil
}

// versionStringFromDependency returns a version string from the supplied
// dependency binary path and an optional version selector struct that
// instructs us how to get the version from the binary.
func versionStringFromDependency(
	binPath string,
	selector *api.DependencyVersionSelector,
) (string, error) {
	if selector == nil {
		selector = defaultVersionSelector
	}
	if selector.Filter == "" {
		selector.Filter = defaultVersionSelectorFilter
		selector.FilterRegex = regexp.MustCompile(defaultVersionSelectorFilter)
	}
	args := selector.Args
	out, err := exec.CommandContext(context.TODO(), binPath, args...).Output()
	if err != nil {
		return "", err
	}
	if selector.FilterRegex != nil {
		if !selector.FilterRegex.MatchString(string(out)) {
			return "", fmt.Errorf(
				"unable to determine version string from %q using regex %q",
				string(out), selector.FilterRegex.String(),
			)
		}
		return selector.FilterRegex.FindString(string(out)), nil
	}
	return string(out), nil
}
