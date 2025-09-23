// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package scenario

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff"

	"github.com/gdt-dev/core/api"
	gdtcontext "github.com/gdt-dev/core/context"
	"github.com/gdt-dev/core/debug"
	"github.com/gdt-dev/core/run"
	"github.com/gdt-dev/core/testunit"
)

// Run executes the scenario. The error that is returned will always be derived
// from `api.RuntimeError` and represents an *unrecoverable* error.
//
// Test assertion failures are *not* considered errors. The Scenario.Run()
// method controls whether the test runner calls `Fail()` or `Skip()` which
// will mark the test units failed or skipped if a test unit evaluates to
// false.
func (s *Scenario) Run(ctx context.Context, subject any) error {
	if s.Path != "" {
		// NOTE(jaypipes): This is necessary to allow relative path lookups for
		// file loads *within* the test scenario itself.
		cwd, _ := os.Getwd()
		if err := os.Chdir(filepath.Dir(s.Path)); err != nil {
			return err
		}
		defer func() {
			_ = os.Chdir(cwd)
		}()
	}
	if err := s.checkDependencies(ctx); err != nil {
		return err
	}
	switch subject := subject.(type) {
	case *testing.T:
		return s.runGo(ctx, subject)
	case *run.Run:
		return s.runExternal(ctx, subject)
	default:
		return fmt.Errorf("unknown run type %T", subject)
	}
}

// runExternal executes the scenario using the `gdt` CLI tool as the underlying
// test runner and a `*RunState` to track test run state. The error that is
// returned will always be derived from `api.RuntimeError` and represents an
// *unrecoverable* error.
func (s *Scenario) runExternal(ctx context.Context, run *run.Run) error {
	ctx = gdtcontext.PushTrace(ctx, s.Title())
	defer func() {
		ctx = gdtcontext.PopTrace(ctx)
	}()

	rootUnit := testunit.New(
		ctx,
		testunit.WithName(s.Title()),
	)
	ctx = gdtcontext.SetTestUnit(ctx, rootUnit)

	if len(s.Fixtures) > 0 {
		fixtures := gdtcontext.Fixtures(ctx)
		for _, fname := range s.Fixtures {
			lookup := strings.ToLower(fname)
			fix, found := fixtures[lookup]
			if !found {
				return api.RequiredFixtureMissing(fname)
			}
			if err := fix.Start(ctx); err != nil {
				return err
			}
			defer fix.Stop(ctx)
		}
	}

	// If the test author has specified any pre-flight checks in the `skip-if`
	// collection, evaluate those first and if any failed, skip the scenario's
	// tests.
	for _, skipIf := range s.SkipIf {
		res, err := skipIf.Eval(ctx)
		if err != nil {
			return err
		}
		if len(res.Failures()) == 0 {
			rootUnit.Skipf(
				"skip-if: %s passed. skipping test.",
				skipIf.Base().Title(),
			)
			return nil
		}
	}

	var runErr error

	scenCleanups := []func(){}
	scenOK := true
	for idx, t := range s.Tests {
		tu := testunit.New(
			ctx,
			testunit.WithName(
				fmt.Sprintf(
					"%s/%s",
					s.Title(),
					t.Base().Title(),
				),
			),
		)
		ctx = gdtcontext.SetTestUnit(ctx, tu)
		res, err := s.runSpec(ctx, tu, idx)
		if err != nil {
			runErr = err
			break
		}

		scenCleanups = append(scenCleanups, res.Cleanups()...)

		// Results can have arbitrary run data stored in them and we
		// save this prior run data in the top-level context (and pass
		// that context to the next Run invocation).
		if res.HasData() {
			ctx = gdtcontext.SetRun(ctx, res.Data())
		}
		if len(res.Failures()) > 0 {
			tu.FailNow()
		}
		scenOK = scenOK && !tu.Failed()

		run.StoreResult(idx, s.Path, tu, res)
	}
	slices.Reverse(scenCleanups)
	if scenOK {
		for _, cleanup := range scenCleanups {
			cleanup()
		}
	}
	return runErr
}

// runGo executes the scenario using the `go test` tool as the underlying test
// runner and the Go `*testing.T` to track test run state. The error that is
// returned will always be derived from `api.RuntimeError` and represents an
// *unrecoverable* error.
func (s *Scenario) runGo(ctx context.Context, t *testing.T) error {
	ctx = gdtcontext.PushTrace(ctx, s.Title())
	defer func() {
		ctx = gdtcontext.PopTrace(ctx)
	}()

	if s.hasTimeoutConflict(ctx, t) {
		return api.TimeoutConflict(s.Timings)
	}

	if len(s.Fixtures) > 0 {
		fixtures := gdtcontext.Fixtures(ctx)
		for _, fname := range s.Fixtures {
			lookup := strings.ToLower(fname)
			fix, found := fixtures[lookup]
			if !found {
				return api.RequiredFixtureMissing(fname)
			}
			if err := fix.Start(ctx); err != nil {
				return err
			}
			defer fix.Stop(ctx)
		}
	}

	// If the test author has specified any pre-flight checks in the `skip-if`
	// collection, evaluate those first and if any failed, skip the scenario's
	// tests.
	for _, skipIf := range s.SkipIf {
		res, err := skipIf.Eval(ctx)
		if err != nil {
			return err
		}
		if len(res.Failures()) == 0 {
			t.Skipf(
				"skip-if: %s passed. skipping test.",
				skipIf.Base().Title(),
			)
			return nil
		}
	}

	var res *api.Result
	var err error

	t.Run(s.Title(), func(tt *testing.T) {
		for idx := range s.Tests {
			res, err = s.runSpec(ctx, tt, idx)
			if err != nil {
				break
			}

			for _, cleanup := range res.Cleanups() {
				t.Cleanup(cleanup)
			}

			// Results can have arbitrary run data stored in them and we
			// save this prior run data in the top-level context (and pass
			// that context to the next Run invocation).
			if res.HasData() {
				ctx = gdtcontext.SetRun(ctx, res.Data())
			}

			for _, fail := range res.Failures() {
				tt.Fatal(fail)
			}
		}
	})
	return err
}

type runSpecRes struct {
	r   *api.Result
	err error
}

// runSpec wraps the execution of a single test spec
func (s *Scenario) runSpec(
	ctx context.Context, // this is the overall scenario's context
	t api.T, // T specific to the goroutine running this test spec
	idx int, // index of the test spec within Scenario.Tests
) (res *api.Result, err error) {
	// Create a brand new context that inherits the top-level context's
	// cancel func. We want to set deadlines for each test spec and if
	// we mutate the single supplied top-level context, then only the
	// first deadline/timeout will be used.
	specCtx, specCancel := context.WithCancel(ctx)
	defer specCancel()

	defaults := s.getDefaults()
	spec := s.Tests[idx]
	sb := spec.Base()

	specTraceMsg := strconv.Itoa(idx)
	if sb.Name != "" {
		specTraceMsg += ":" + sb.Name
	}
	specCtx = gdtcontext.PushTrace(specCtx, specTraceMsg)
	defer func() {
		specCtx = gdtcontext.PopTrace(specCtx)
	}()

	plugin := sb.Plugin
	rt := getRetry(specCtx, defaults, plugin, spec)
	to := getTimeout(specCtx, defaults, plugin, spec)
	ch := make(chan runSpecRes, 1)

	wait := sb.Wait
	if wait != nil && wait.Before != "" {
		debug.Printf(specCtx, "wait: %s before", wait.Before)
		time.Sleep(wait.BeforeDuration())
	}

	if to != nil {
		specCtx, specCancel = context.WithTimeout(specCtx, to.Duration())
		defer specCancel()
	}

	go s.execSpec(specCtx, ch, rt, idx, spec)

	select {
	case <-specCtx.Done():
		t.Fatalf("assertion failed: timeout exceeded (%s)", to.After)
	case runres := <-ch:
		res = runres.r
		err = runres.err
	}
	if err != nil {
		return nil, err
	}

	if wait != nil && wait.After != "" {
		debug.Printf(specCtx, "wait: %s after", wait.After)
		time.Sleep(wait.AfterDuration())
	}
	return res, nil
}

// execSpec executes an individual test spec, performing any retries as
// necessary until a timeout is exceeded or the test spec succeeds
func (s *Scenario) execSpec(
	ctx context.Context,
	ch chan runSpecRes,
	retry *api.Retry,
	idx int,
	spec api.Evaluable,
) {
	if retry == nil || retry == api.NoRetry {
		// Just evaluate the test spec once
		res, err := spec.Eval(ctx)
		if err != nil {
			ch <- runSpecRes{nil, err}
			return
		}
		debug.Printf(
			ctx, "spec/run: single-shot (no retries) ok: %v",
			!res.Failed(),
		)
		ch <- runSpecRes{res, nil}
		return
	}

	// retry the action and test the assertions until they succeed,
	// there is a terminal failure, or the timeout expires.
	var bo backoff.BackOff
	var res *api.Result
	var err error

	if retry.Exponential {
		bo = backoff.WithContext(
			backoff.NewExponentialBackOff(),
			ctx,
		)
	} else {
		interval := api.DefaultRetryConstantInterval
		if retry.Interval != "" {
			interval = retry.IntervalDuration()
		}
		bo = backoff.WithContext(
			backoff.NewConstantBackOff(interval),
			ctx,
		)
	}
	ticker := backoff.NewTicker(bo)
	maxAttempts := 0
	if retry.Attempts != nil {
		maxAttempts = *retry.Attempts
	}
	attempts := 1
	start := time.Now().UTC()
	success := false
	for tick := range ticker.C {
		if (maxAttempts > 0) && (attempts > maxAttempts) {
			debug.Printf(
				ctx, "spec/run: exceeded max attempts %d. stopping.",
				maxAttempts,
			)
			ticker.Stop()
			break
		}
		after := tick.Sub(start)

		res, err = spec.Eval(ctx)
		if err != nil {
			ch <- runSpecRes{nil, err}
			return
		}
		success = !res.Failed()
		debug.Printf(
			ctx, "spec/run: attempt %d after %s ok: %v",
			attempts, after, success,
		)
		if success {
			ticker.Stop()
			break
		}
		for _, f := range res.Failures() {
			debug.Printf(
				ctx, "spec/run: attempt %d failure: %s",
				attempts, f,
			)
		}
		attempts++
	}
	ch <- runSpecRes{res, nil}
}

// hasTimeoutConflict returns true if the scenario or any of its test specs has
// a wait or timeout that exceeds the go test tool's specified timeout value
func (s *Scenario) hasTimeoutConflict(
	ctx context.Context,
	t *testing.T,
) bool {
	d, ok := t.Deadline()
	if ok && !d.IsZero() {
		now := time.Now()
		s.Timings.GoTestTimeout = d.Sub(now)
		debug.Printf(
			ctx, "scenario/run: go test tool timeout: %s",
			(s.Timings.GoTestTimeout + time.Second).Round(time.Second),
		)
		if s.Timings.TotalWait > 0 {
			if s.Timings.TotalWait.Abs() > s.Timings.GoTestTimeout.Abs() {
				return true
			}
		}
		if s.Timings.MaxTimeout > 0 {
			if s.Timings.MaxTimeout.Abs() > s.Timings.GoTestTimeout.Abs() {
				return true
			}
		}
	}
	return false
}

// getTimeout returns the timeout configuration for the test spec. We check for
// overrides in timeout configuration using the following precedence:
//
// * Spec (Evaluable) override
// * Spec's Base override
// * Scenario's default
// * Plugin's default
func getTimeout(
	ctx context.Context,
	defaults *Defaults,
	plugin api.Plugin,
	eval api.Evaluable,
) *api.Timeout {
	evalTimeout := eval.Timeout()
	if evalTimeout != nil {
		debug.Printf(
			ctx, "using timeout of %s",
			evalTimeout.After,
		)
		return evalTimeout
	}

	sb := eval.Base()
	baseTimeout := sb.Timeout
	if baseTimeout != nil {
		debug.Printf(
			ctx, "using timeout of %s",
			baseTimeout.After,
		)
		return baseTimeout
	}

	if defaults != nil && defaults.Timeout != nil {
		debug.Printf(
			ctx, "using timeout of %s [scenario default]",
			defaults.Timeout.After,
		)
		return defaults.Timeout
	}

	pluginInfo := plugin.Info()
	pluginTimeout := pluginInfo.Timeout

	if pluginTimeout != nil {
		debug.Printf(
			ctx, "using timeout of %s [plugin default]",
			pluginTimeout.After,
		)
		return pluginTimeout
	}
	return nil
}

// getRetry returns the retry configuration for the test spec. We check for
// overrides in retry configuration using the following precedence:
//
// * Spec (Evaluable) override
// * Spec's Base override
// * Scenario's default
// * Plugin's default
func getRetry(
	ctx context.Context,
	defaults *Defaults,
	plugin api.Plugin,
	eval api.Evaluable,
) *api.Retry {
	evalRetry := eval.Retry()
	if evalRetry != nil {
		if evalRetry == api.NoRetry {
			return evalRetry
		}
		msg := "using retry"
		if evalRetry.Attempts != nil {
			msg += fmt.Sprintf(" (attempts: %d)", *evalRetry.Attempts)
		}
		if evalRetry.Interval != "" {
			msg += fmt.Sprintf(" (interval: %s)", evalRetry.Interval)
		}
		msg += fmt.Sprintf(" (exponential: %t)", evalRetry.Exponential)
		debug.Println(ctx, msg)
		return evalRetry
	}

	sb := eval.Base()
	baseRetry := sb.Retry
	if baseRetry != nil {
		if baseRetry == api.NoRetry {
			return baseRetry
		}
		msg := "using retry"
		if baseRetry.Attempts != nil {
			msg += fmt.Sprintf(" (attempts: %d)", *baseRetry.Attempts)
		}
		if baseRetry.Interval != "" {
			msg += fmt.Sprintf(" (interval: %s)", baseRetry.Interval)
		}
		msg += fmt.Sprintf(" (exponential: %t)", baseRetry.Exponential)
		debug.Println(ctx, msg)
		return baseRetry
	}

	if defaults != nil && defaults.Retry != nil {
		defaultRetry := defaults.Retry
		if defaultRetry == api.NoRetry {
			return defaultRetry
		}
		msg := "using retry"
		if defaultRetry.Attempts != nil {
			msg += fmt.Sprintf(" (attempts: %d)", *defaultRetry.Attempts)
		}
		if defaultRetry.Interval != "" {
			msg += fmt.Sprintf(" (interval: %s)", defaultRetry.Interval)
		}
		msg += fmt.Sprintf(" (exponential: %t) [scenario default]", defaultRetry.Exponential)
		debug.Println(ctx, msg)
		return defaultRetry
	}

	pluginInfo := plugin.Info()
	pluginRetry := pluginInfo.Retry

	if pluginRetry != nil {
		if pluginRetry == api.NoRetry {
			return pluginRetry
		}
		msg := "using retry"
		if pluginRetry.Attempts != nil {
			msg += fmt.Sprintf(" (attempts: %d)", *pluginRetry.Attempts)
		}
		if pluginRetry.Interval != "" {
			msg += fmt.Sprintf(" (interval: %s)", pluginRetry.Interval)
		}
		msg += fmt.Sprintf(" (exponential: %t) [plugin default]", pluginRetry.Exponential)
		debug.Println(ctx, msg)
		return pluginRetry
	}
	return nil
}

// getDefaults returns the Defaults parsed from the scenario's YAML
// file's `defaults` field, or nil if none were specified.
func (s *Scenario) getDefaults() *Defaults {
	scDefaultsAny, found := s.Defaults[DefaultsKey]
	if found {
		return scDefaultsAny.(*Defaults)
	}
	return nil
}
