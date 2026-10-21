# Repository layout

Vellum follows the Go project's [module organization guidance](https://go.dev/doc/modules/layout)
for a library with multiple public packages and separate executable consumers.

```text
Vellum/
  go.mod, vellum.go       portable framework module and version package
  app/, ui/, canvas/     public packages; tests live beside implementation
  command/, history/     commands and Undo/Redo
  document/, editor/     persistence and text models
  execution/, tasks/     scheduling and asynchronous work
  placement/, replay/    placement and deterministic input records
  automation/            opt-in semantic automation
  virtual/, workbench/   data views and dock layouts
  drivers/cogent/        native driver module with its own go.mod/go.sum
    cmd/                 workbench, gallery and native diagnostic commands
    internal/            driver-private supporting packages
    native/              pinned patches, assets and original notices
    build/               ignored prepared sources and build outputs
  examples/              independent application and contract modules
    placement/           internal drawing test app (go run .)
    macapps/cmd/         calculator, notes, weather and reminders commands
  docs/                  guides, decisions, research, reports and reviews
    reports/go.mod       excludes historical evidence from framework downloads
  specs/                 framework contracts and validation scope
  scripts/               build, verification and packaging tools
  prototypes/            historical experiments in separate modules
  release/               release metadata and dependency notices
  .github/workflows/     CI entry points
  LICENSE, README.md, CONTRIBUTING.md
```

## Module boundaries

The root module is `github.com/dshills/vellum` and uses only the standard library.
Its public package directories intentionally remain at the module root: imports
such as `github.com/dshills/vellum/app` are the API. There is no need for a `src/`
or `pkg/` wrapper around them. The root has no `go.sum` because it has no external
module requirements.

`drivers/cogent` is a separate module so native dependencies do not enter the
portable framework graph. Applications in `examples/` each declare their own
module. A single-command example may keep `package main` in its module root;
modules containing several commands use `cmd/<name>/`. Supporting packages that
are part of the intended example API can remain next to those commands.

Root `go test ./...` checks only the root module. Native-driver and example checks
must run in their own modules; see [CONTRIBUTING](../CONTRIBUTING.md) and the
[example index](../examples/README.md). Module-local `replace` directives point to
the local framework and prepared native sources. A `go.work` file is optional and
local; repository checks use the documented module-local setup.

## Evidence and generated files

Historical evidence stays at its recorded paths. `docs/reports/go.mod` is an
archive boundary, with no framework dependency relationship or published API.
This uses Go's [nested-module exclusion rule](https://go.dev/ref/mod#zip-files)
to keep large reports out of downloads of the parent framework module. Without
that boundary, tracked root-module contents exceeded the 500 MiB uncompressed
module-archive limit. It does not shrink Git clones or rewrite history.

SDK packaging already excludes historical reports, reviews and prototypes through
its source whitelist. The report boundary does not change runtime behavior or
native SDK packaging. Generated dependency copies and binaries belong under
ignored `build/`; local workspaces and scratch reviews are also ignored. Committed
patches, licenses, test fixtures and source-bound historical evidence remain in
Git.
