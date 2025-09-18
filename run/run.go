package run

import (
	"slices"
	"time"

	"github.com/gdt-dev/core/api"
	"github.com/gdt-dev/core/testunit"
	"github.com/samber/lo"
)

// Run stores state of a test run when tests are executed with the `gdt` CLI
// tool.
type Run struct {
	// scenarioResults is a map, keyed by the Scenario path, of slices of
	// TestUnitResult structs corresponding to the test specs in the scenario.
	// There is guaranteed to be exactly the same number of TestUnitResults in
	// the slice as scenarios in the scenario.
	scenarioResults map[string][]TestUnitResult
}

// OK returns true if all Scenarios in the Run had all successful test units.
func (r *Run) OK() bool {
	return !lo.SomeBy(lo.Values(r.scenarioResults), func(results []TestUnitResult) bool {
		return !lo.SomeBy(results, func(r TestUnitResult) bool {
			return len(r.failures) == 0
		})
	})
}

// ScenarioPaths returns a sorted list of Scenario Paths.
func (r *Run) ScenarioPaths() []string {
	paths := lo.Keys(r.scenarioResults)
	slices.Sort(paths)
	return paths
}

// ScenarioResults returns the set of TestUnitResults for a Scenario with the
// supplied path.
func (r *Run) ScenarioResults(path string) []TestUnitResult {
	return r.scenarioResults[path]
}

// StoreResult stores a test unit result to the Run for the supplied test unit.
func (r *Run) StoreResult(
	index int,
	path string, // the Scenario.Path
	tu *testunit.TestUnit,
	res *api.Result,
) {
	if _, ok := r.scenarioResults[path]; !ok {
		r.scenarioResults[path] = []TestUnitResult{}
	}
	r.scenarioResults[path] = append(
		r.scenarioResults[path],
		TestUnitResult{
			index:    index,
			name:     tu.Name(),
			elapsed:  tu.Elapsed(),
			skipped:  tu.Skipped(),
			failures: res.Failures(),
			detail:   tu.Detail(),
		},
	)
}

// TestUnitResult stores a summary of the test execution of a single test unit.
type TestUnitResult struct {
	// index is the 0-based index of the test unit within the test scenario.
	index int
	// name is the short name of the test unit
	name string
	// skipped is true if the test unit was skipped
	skipped bool
	// failures is the collection of assertion failures for the test spec that
	// occurred during the run. this will NOT include RuntimeErrors.
	failures []error
	// elapsed is the time take to execute the test unit
	elapsed time.Duration
	// detail is a buffer holding any log entries made during the run of the
	// test spec.
	detail string
}

func (u TestUnitResult) OK() bool {
	return len(u.failures) == 0
}

func (u TestUnitResult) Name() string {
	return u.name
}

func (u TestUnitResult) Index() int {
	return u.index
}

func (u TestUnitResult) Failures() []error {
	return u.failures
}

func (u TestUnitResult) Skipped() bool {
	return u.skipped
}

func (u TestUnitResult) Detail() string {
	return u.detail
}

func (u TestUnitResult) Elapsed() time.Duration {
	return u.elapsed
}
