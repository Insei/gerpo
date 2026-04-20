# Static analysis — gerpolint

gerpo's WHERE operators (`EQ`, `In`, `Contains`, and the 20 other methods on [`types.WhereOperation`](where.md)) accept `any`. That is a deliberate trade-off — methods on Go interfaces cannot be generic — but it means the compiler will not catch a mismatch like:

```go
h.Where().Field(&m.Age).EQ("18")       // field is int, arg is string
h.Where().Field(&m.Age).Contains("a")  // Contains requires string/*string
```

Both fail at runtime (either when gerpo rejects the option or when PostgreSQL refuses the coercion). **gerpolint** is a `go/analysis` checker that catches these at `go vet` time, either as a standalone binary or as a `golangci-lint` plugin.

## The rule

- Field `T` → argument must be assignable to `T`.
- Field `*T` → argument may be `T`, `*T`, or untyped `nil`.
- Untyped constants use spec-level representability: `EQ(18)` is fine on `type Age int`, but `EQ(3.14)` on `int` is rejected.

gerpolint identifies gerpo calls by package path (`github.com/insei/gerpo/types`) plus receiver-method shape, so unrelated `EQ` / `In` methods in other packages are left alone.

## Rules

| ID | Trigger | Example |
|---|---|---|
| `GPL001` | Scalar operator, argument type mismatch | `Field(&m.Age).EQ("18")` |
| `GPL002` | Variadic operator, element type mismatch | `Field(&m.Age).In(1, "2", 3)` |
| `GPL003` | String-only operator on non-string field | `Field(&m.Age).Contains("x")` |
| `GPL004` | Field pointer cannot be resolved statically (e.g., via a variable) | `p := &m.Age; Field(p).EQ(...)` |
| `GPL005` | Argument's static type is `any` — static check skipped | `var v any = 18; EQ(v)` |

## Standalone binary

```bash
go install github.com/insei/gerpo/cmd/gerpolint@latest
gerpolint ./...
```

From a clone, `make lint-gerpolint` does the same via `go run ./cmd/gerpolint ./...`.

Flags:

| Flag | Values | Default | Purpose |
|---|---|---|---|
| `-unresolved-field` | `skip` / `warn` / `error` | `skip` | How to treat `Field(ptr)` whose argument cannot be resolved to a concrete field |
| `-any-arg` | `skip` / `warn` / `error` | `warn` | How to treat arguments whose static type is `any` |
| `-disabled-rules` | `GPL001,GPL002,…` | (empty) | Comma-separated rule IDs to skip entirely |

Example:

```bash
gerpolint -unresolved-field=error -disabled-rules=GPL005 ./...
```

## golangci-lint plugin

gerpolint registers as a [golangci-lint v2 module plugin](https://golangci-lint.run/plugins/module-plugins/). The golangci-lint `custom` subcommand builds a bespoke binary with gerpolint embedded alongside your other linters.

### 1. Point golangci-lint at the plugin

Drop `.custom-gcl.yml` at your repo root:

```yaml title=".custom-gcl.yml"
version: v2.5.0         # your golangci-lint version
name: custom-gcl
destination: ./bin
plugins:
  - module: github.com/insei/gerpo
    import: github.com/insei/gerpo/gerpolintplugin
    version: latest
```

### 2. Enable the linter

```yaml title=".golangci.yml"
version: "2"
linters:
  enable:
    - gerpolint
  settings:
    custom:
      gerpolint:
        type: module
        description: Type-safe check of gerpo WHERE filter arguments.
        original-url: https://github.com/insei/gerpo
        settings:
          unresolved-field: skip      # skip | warn | error
          any-arg: warn               # skip | warn | error
          disabled-rules: []          # e.g. [GPL004, GPL005]
```

### 3. Build and run

```bash
golangci-lint custom           # produces ./bin/custom-gcl
./bin/custom-gcl run ./...
```

From a clone of `insei/gerpo`, `make lint-gerpolint-plugin` wraps both steps.

Diagnostics surface with category prefixes so you can filter on them in CI, e.g. `| grep GPL001` to fail a build only on scalar mismatches.

## Directives

Suppress diagnostics inline without changing the linter configuration:

| Directive | Scope |
|---|---|
| `//gerpolint:disable-line[=GPL001,…]` | the current line |
| `//gerpolint:disable-next-line[=GPL001,…]` | the line below |
| `//gerpolint:disable[=GPL001,…]` | from here until `//gerpolint:enable` or EOF |
| `//gerpolint:enable` | close the most recent `disable` block |

Without `=…`, the directive disables *all* gerpolint rules on its scope. Unknown rule IDs trigger a one-shot `GPL-DIRECTIVE-UNKNOWN` warning — the directive itself is ignored so the underlying rule keeps firing.

```go
// Legitimate []any spread — static types are erased by design, so skip GPL005.
h.Where().Field(&m.ID).In(wanted...) //gerpolint:disable-line=GPL005

//gerpolint:disable
// Generated code below — bypass gerpolint wholesale.
...
//gerpolint:enable
```

## When to reach for which knob

- **CI-fail on real bugs**: leave defaults; `GPL001`–`GPL003` fire on concrete type errors.
- **Field-pointer helpers**: if your code routes field pointers through helper functions, gerpolint cannot resolve them — either keep `-unresolved-field=skip` (default) or flip to `error` to force inlined usage.
- **`[]any` slices**: inline a `//gerpolint:disable-line=GPL005` at the call site. Disabling `GPL005` globally via `disabled-rules: [GPL005]` silences every `any`-typed argument, which is usually too broad.
- **Generated code**: bracket the file with `//gerpolint:disable` / `//gerpolint:enable` at the top and bottom.

## What gerpolint does *not* do

- It does not check `OrderBy().Field(...)` — there is no value to type-check.
- It does not validate that a field pointer resolves to a column configured in the repository builder; a runtime "option is not available" error remains the safety net for misconfigured columns.
- It does not run without type information. `analysis.LoadMode` is `TypesInfo`; for golangci-lint that is already the case.
