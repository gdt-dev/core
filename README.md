# Core `gdt` library and specification

[![Go Reference](https://pkg.go.dev/badge/github.com/gdt-dev/core.svg)](https://pkg.go.dev/github.com/gdt-dev/core)
[![Go Report Card](https://goreportcard.com/badge/github.com/gdt-dev/core)](https://goreportcard.com/report/github.com/gdt-dev/core)
[![Build Status](https://github.com/gdt-dev/core/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/gdt-dev/core/actions)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)

`gdt` is a testing library that allows test authors to cleanly describe tests
in a YAML file. `gdt` reads YAML files that describe a test's assertions and
then builds a set of structures that can be run by either the `gdt`
command-line interface or the standard Go
[`testing`](https://golang.org/pkg/testing/) `go test` tool.


## `gdt` test scenario structure

A `gdt` test scenario (or just "scenario") is simply a YAML file.

All `gdt` scenarios have the following fields:

* `name`: (optional) string describing the contents of the test file. If
  missing or empty, the filename is used as the name
* `description`: (optional) string with longer description of the test file
  contents
* `defaults`: (optional) is a map of default options and configuration values
* `fixtures`: (optional) list of strings indicating named fixtures that will be
  started before any of the tests in the file are run
* `depends`: (optional) list of [`Dependency`][dependency] objects that
  describe a program or package that should be available in the host's `PATH`
  that the test scenario depends on.
* `depends.name`: string name of the program or package the test scenario
  depends on.
* `depends.when`: (optional) object describing any constraints/conditions that
  should apply to the evaluation of the dependency.
* `depends.when.os`: (optional) string operating system. if set, the dependency
  is only checked for that OS.
* `depends.when.version`: (optional) string version constraint. if set, the
  program or package must be available on the host `PATH` and must satisfy the
  version constraint.
* `skip-if`: (optional) list of [`Spec`][basespec] specializations that will be
  evaluated *before* running any test in the scenario. If any of these
  conditions evaluates successfully, the test scenario will be skipped.
* `tests`: list of [`Spec`][basespec] specializations that represent the
  runnable test units in the test scenario.

[basespec]: https://github.com/gdt-dev/core/blob/ecee17249e1fa10147cf9191be0358923da44094/types/spec.go#L30
[dependency]: https://github.com/gdt-dev/core/blob/e83cb2bcac4fbb73205c58851b22a8b0a6cc51b5/api/dependency.go

The scenario's `tests` field is the most important and the [`Spec`][basespec]
objects that it contains are the meat of a test scenario.

### `gdt` test spec structure

A spec represents a single *action* that is taken and zero or more
*assertions* that represent what you expect to see resulting from that action.

`gdt` plugins each define a specialized subclass of the base [`Spec`][basespec]
that contains fields that are specific to that type of test.

For example, there is an [`exec`][exec-plugin] plugin that allows you to
execute arbitrary commands and assert expected result codes and output. There
is an [`http`][http-plugin] that allows you to call an HTTP URL and assert that
the response looks like what you expect. There is a [`kube`][kube-plugin]
plugin that allows you to interact with a Kubernetes API, etc.

`gdt` examines the YAML file that defines your test scenario and uses these
plugins to parse individual test specs.

All test specs have the following fields:

* `name`: (optional) string describing the test unit.
* `description`: (optional) string with longer description of the test unit.
* `timeout`: (optional) a string duration of time the test unit is expected to
  complete within.
* `retry`: (optional) an object containing retry configurationu for the test
  unit. Some plugins will automatically attempt to retry the test action when
  an assertion fails. This field allows you to control this retry behaviour for
  each individual test.
* `retry.interval`: (optional) a string duration of time that the test plugin
  will retry the test action in the event assertions fail. The default interval
  for retries is plugin-dependent.
* `retry.attempts`: (optional) an integer indicating the number of times that a
  plugin will retry the test action in the event assertions fail. The default
  number of attempts for retries is plugin-dependent.
* `retry.exponential`: (optional) a boolean indicating an exponential backoff
  should be applied to the retry interval. The default is is plugin-dependent.
* `wait` (optional) an object containing [wait information][wait] for the test
  unit.
* `wait.before`: a string duration of time that gdt should wait before
  executing the test unit's action.
* `wait.after`: a string duration of time that gdt should wait after executing
  the test unit's action.
* `on`: (optional) an object describing actions to take upon certain
  conditions.
* `on.fail`: (optional) an object describing an action to take when any
  assertion fails for the test action.
* `on.fail.exec`: a string with the exact command to execute upon test
  assertion failure. You may execute more than one command but must include the
  `on.fail.shell` field to indicate that the command should be run in a shell.
* `on.fail.shell`: (optional) a string with the specific shell to use in executing the
  command to run upon test assertion failure. If empty (the default), no shell
  is used to execute the command and instead the operating system's `exec` family
  of calls is used.

[exec-plugin]: https://github.com/gdt-dev/core/tree/ecee17249e1fa10147cf9191be0358923da44094/plugin/exec
[http-plugin]: https://github.com/gdt-dev/http
[kube-plugin]: https://github.com/gdt-dev/kube
[wait]: https://github.com/gdt-dev/core/blob/2791e11105fd3c36d1f11a7d111e089be7cdc84c/types/wait.go#L11-L25

#### `exec` test spec structure

The `exec` plugin's test spec allows test authors to execute arbitrary commands and
assert that the command results in an expected result code or output.

In addition to all the base `Spec` fields listed above, the `exec` plugin's
test spec also contains these fields:

* `exec`: a string with the exact command to execute. You may execute more than
  one command but must include the `shell` field to indicate that the command
  should be run in a shell. It is best practice, however, to simply use
  multiple `exec` specs instead of executing multiple commands in a single
  shell call.
* `shell`: (optional) a string with the specific shell to use in executing the
  command. If empty (the default), no shell is used to execute the command and
  instead the operating system's `exec` family of calls is used.
* `var-stdout`: (optional) a string with the name of a variable to save the
  contents of the test spec's `stdout` stream. This named variable can then be
  referred from subsequent test specs. Note: this is a shortcut for the
  longer-form `var:{VAR_NAME}:from:stdout`
* `var-stderr`: (optional) a string with the name of a variable to save the
  contents of the test spec's `stderr` stream. This named variable can then be
  referred from subsequent test specs. Note: this is a shortcut for the
  longer-form `var:{VAR_NAME}:from:stderr`
* `var-rc`: (optional) a string with the name of a variable to save the
  contents of the test spec's return/exitcode value. This named variable can
  then be referred from subsequent test specs. Note: this is a shortcut for the
  longer-form `var:{VAR_NAME}:from:returncode`
* `var`: (optional) an object describing variables that can have values saved
  and referred to by subsequent test specs. Each key in the `var` object is the
  name of the variable to define. The `var.from` field contains a string
  describing where the value for the variable should be sourced.
* `var.$VARIABLE_NAME.from`: (required) a string describing where the variable
  with name `$VARIABLE_NAME` should source its value. The strings `stdout`,
  `stderr` and `returncode` refer to the corresponding stdout, stderr
  and return/exitcode values. All other string values for `var.from` indicate
  the name of the environment variable to read into the named variable.
* `assert`: (optional) an object describing the conditions that will be
  asserted about the test action.
* `assert.exit-code`: (optional) an integer with the expected exit code from the
  executed command. The default successful exit code is 0 and therefore you do
  not need to specify this if you expect a successful exit code.
* `assert.out`: (optional) a [`PipeExpect`][pipeexpect] object containing
  assertions about content in `stdout`.
* `assert.out.is`: (optional) a string with the exact contents of `stdout` you expect
  to get.
* `assert.out.all`: (optional) a string or list of strings that *all* must be
  present in `stdout`.
* `assert.out.any`: (optional) a string or list of strings of which *at
  least one* must be present in `stdout`.
* `assert.out.none`: (optional) a string or list of strings of which *none
  should be present* in `stdout`.
* `assert.err`: (optional) a [`PipeAssertions`][pipeexpect] object containing
  assertions about content in `stderr`.
* `assert.err.is`: (optional) a string with the exact contents of `stderr` you expect
  to get.
* `assert.err.all`: (optional) a string or list of strings that *all* must be
  present in `stderr`.
* `assert.err.any`: (optional) a string or list of strings of which *at
  least one* must be present in `stderr`.
* `assert.err.none`: (optional) a string or list of strings of which *none
  should be present* in `stderr`.

[execspec]: https://github.com/gdt-dev/core/blob/2791e11105fd3c36d1f11a7d111e089be7cdc84c/exec/spec.go#L11-L34
[pipeexpect]: https://github.com/gdt-dev/core/blob/2791e11105fd3c36d1f11a7d111e089be7cdc84c/exec/assertions.go#L15-L26

### Marking a program or package dependency for a test scenario

Often when creating `gdt` test scenarios, you will want to declare that the
scenario requires a particular application be available in the host's `PATH`.

Use the `depends` field to tell `gdt` about these requirements:

```yaml
name: scenario requiring `myapp` binary be available to execute

depends:
 - name: myapp

tests:
 - name: start-myapp
   exec: myapp start
```

Running the above test scenario when `myapp` is not available to execute
results in a runtime error:

```
$ gdt run myapp.yaml
Error: runtime error: dependency not satisfied: myapp
```

You can specify that a dependency only be applicable on a specific operating
system using the `depends.when.os` field:

```yaml
name: scenario requiring `myapp` binary be available to execute

depends:
 - name: myapp
   when:
     os: linux

tests:
 - name: start-myapp
   exec: myapp start
```

Running the above scenario on Linux yields the same runtime error:

```
$ gdt run myapp.yaml
Error: runtime error: dependency not satisfied: myapp (OS:linux)
```

However changing the `depends.when.os` to `windows`:

```yaml
name: scenario requiring `myapp` binary be available to execute

depends:
 - name: myapp
   when:
     os: windows

tests:
 - name: start-myapp
   exec: myapp start
```

Running the scenario will show a different runtime error that is dependent on
the operating system:

```
$ gdt run myapp.yaml
Error: runtime error: exec: "myapp": executable file not found in $PATH
```

### Passing variables to subsequent test specs

A `gdt` test scenario is comprised of a list of test specs. These test specs
are executed in sequential order. If you want to have one test spec be able to
use some output or value calculated or asserted in a previous step, you can use
the `gdt` variable system.

Here's an test scenario that shows how to define variables in a test spec and
how to use those variables in later test specs.

file: `plugin/exec/testdata/var-save-restore.yaml`:

```yaml
name: var-save-restore
description: a scenario that tests variable save/restore across multiple test specs
tests:
  - exec: echo 42
    var-stdout: VAR_STDOUT

  - exec: echo $$VAR_STDOUT
    var-rc: VAR_RC
    assert:
      out:
        is: 42

  - exec: echo $$VAR_RC
    assert:
      out:
        is: 0

  - exec: echo 42
    assert:
      out:
        is: $$VAR_STDOUT
```

In the first test spec, we specify that we want to store the value of the
`stdout` stream in a variable called `VAR_STDOUT`:

```yaml
  - exec: echo 42
    var-stdout: VAR_STDOUT
```

In the second test spec, we refer to the `VAR_STDOUT` variable using the
double-dollar-sign notation in the `exec` field and also specify a `VAR_RC`
variable to contain the value of the return/exitcode from the executed
statement (`echo 42`):

```yaml
  - exec: echo $$VAR_STDOUT
    var-rc: VAR_RC
    assert:
      out:
        is: 42
```

> **NOTE**: We use the double-dollar-sign notation because by default, `gdt`
> replaces all single-dollar-sign notations with environment variables *BEFORE*
> executing the test specs in a test scenario. Using the double-dollar-sign
> notation means that environment variable substitution does not impact the
> referencing of `gdt` variables referenced in a test spec.

In the third test spec, we simply echo out the value of that `VAR_RC` variable
and assert that the stdout stream contains the string "0" (since `echo 42`
returns 0.):

```yaml
  - exec: echo $$VAR_RC
    assert:
      out:
        is: 0
```

Finally, in the fourth step, we demonstrate that we can refer to the
`VAR_STDOUT` variable defined in the very first test spec from the
`assert.out.is` field. This shows the flexibility of the `gdt` variable system.
You can define variables using a simple declarative syntax and then refer to
the value of those variables using the double-dollar-sign notation in any
subsequent test spec.

### Timeouts and retrying assertions

When evaluating assertions for a test spec, `gdt` inspects the test's
`timeout` value to determine how long to retry the `get` call and recheck
the assertions.

If a test's `timeout` is empty, `gdt` inspects the scenario's
`defaults.timeout` value. If both of those values are empty, `gdt` will look
for any default `timeout` value that the plugin uses.

If you're interested in seeing the individual results of `gdt`'s
assertion-checks for a single `get` call, you can use the `gdt.WithDebug()`
function, like this test function demonstrates:

file: `testdata/matches.yaml`:

```yaml
name: matches
description: create a deployment and check the matches condition succeeds
fixtures:
  - kind
tests:
  - name: create-deployment
    kube:
      create: testdata/manifests/nginx-deployment.yaml
  - name: deployment-exists
    kube:
      get: deployments/nginx
    assert:
      matches:
        spec:
          replicas: 2
          template:
            metadata:
              labels:
                app: nginx
        status:
          readyReplicas: 2
  - name: delete-deployment
    kube:
      delete: deployments/nginx
```

file: `matches_test.go`

```go
import (
    "github.com/gdt-dev/core"
    _ "github.com/gdt-dev/kube"
    kindfix "github.com/gdt-dev/kube/fixture/kind"
)

func TestMatches(t *testing.T) {
	fp := filepath.Join("testdata", "matches.yaml")

	kfix := kindfix.New()

	s, err := gdt.From(fp)

	ctx := gdt.NewContext(gdt.WithDebug())
	ctx = gdt.RegisterFixture(ctx, "kind", kfix)
	s.Run(ctx, t)
}
```

Here's what running `go test -v matches_test.go` would look like:

```
$ go test -v matches_test.go
=== RUN   TestMatches
=== RUN   TestMatches/matches
=== RUN   TestMatches/matches/create-deployment
=== RUN   TestMatches/matches/deployment-exists
deployment-exists (try 1 after 1.303µs) ok: false, terminal: false
deployment-exists (try 1 after 1.303µs) failure: assertion failed: match field not equal: $.status.readyReplicas not present in subject
deployment-exists (try 2 after 595.62786ms) ok: false, terminal: false
deployment-exists (try 2 after 595.62786ms) failure: assertion failed: match field not equal: $.status.readyReplicas not present in subject
deployment-exists (try 3 after 1.020003807s) ok: false, terminal: false
deployment-exists (try 3 after 1.020003807s) failure: assertion failed: match field not equal: $.status.readyReplicas not present in subject
deployment-exists (try 4 after 1.760006109s) ok: false, terminal: false
deployment-exists (try 4 after 1.760006109s) failure: assertion failed: match field not equal: $.status.readyReplicas had different values. expected 2 but found 1
deployment-exists (try 5 after 2.772416449s) ok: true, terminal: false
=== RUN   TestMatches/matches/delete-deployment
--- PASS: TestMatches (3.32s)
    --- PASS: TestMatches/matches (3.30s)
        --- PASS: TestMatches/matches/create-deployment (0.01s)
        --- PASS: TestMatches/matches/deployment-exists (2.78s)
        --- PASS: TestMatches/matches/delete-deployment (0.02s)
PASS
ok  	command-line-arguments	3.683s
```

You can see from the debug output above that `gdt` created the Deployment and
then did a `kube.get` for the `deployments/nginx` Deployment. Initially
(attempt 1), the `assert.matches` assertion failed because the
`status.readyReplicas` field was not present in the returned resource. `gdt`
retried the `kube.get` call 4 more times (attempts 2-5), with attempts 2 and 3
failed the existence check for the `status.readyReplicas` field and attempt 4
failing the *value* check for the `status.readyReplicas` field being `1`
instead of the expected `2`. Finally, when the Deployment was completely rolled
out, attempt 5 succeeded in all the `assert.matches` assertions.

## Contributing and acknowledgements

`gdt` was inspired by [Gabbi](https://github.com/cdent/gabbi), the excellent
Python declarative testing framework. `gdt` tries to bring the same clear,
concise test definitions to the world of Go functional testing.

The Go gopher logo, from which gdt's logo was derived, was created by Renee
French.

Contributions to `gdt` are welcomed! Feel free to open a Github issue or submit
a pull request.
