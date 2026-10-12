# Development

## Modules and dependencies

The root `github.com/dshills/vellum` module contains portable public packages and uses
only the standard library. The `drivers/cogent` module depends on that local root
module and the pinned native stack. Keep driver-specific dependencies in that
module. Each prototype has its own module and records its experimental inputs.

The driver module's repository-relative replacements point to prepared Cogent and
go-text source copies under `drivers/cogent/build/`. `drivers/cogent/prepare.sh` downloads
exact pseudo-versions, verifies patch-input SHA-256 hashes and applies committed
patches from driver-owned `native/` inputs to owned copies. Preparation has no
prototype pipeline dependency; this is a repository-local development distribution.

## Checks

Run portable checks from the repository root:

```sh
go test -race ./...
go vet ./...
```

For the full portable and native check, bootstrap first:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./drivers/cogent/prepare.sh
./scripts/check.sh
```

The script runs root race tests and vet, native-module tests and vet, then builds
`drivers/cogent/build/vellum-workbench`. It also race-tests, vets and builds the
independent hello, drawing, 3D, imaging, editor and grid modules, validates dependency
preparation inputs and race-tests upstream lifecycle regressions. Exact text and
shaped range fixtures run with the prepared typesetting replacement, installed
font cache and offscreen display initialization through
`scripts/check-text-backend.py`; native consumers separately exercise AppKit. The macOS CI workflow follows the same
bootstrap and check sequence from a fresh checkout. Native checks need the C
compiler, graphics libraries and GPU dependencies described in the root README.

To run the native example after preparation:

```sh
cd drivers/cogent
. ./env.sh
go run ./cmd/workbench
```

Generated `build/` directories and scratch reviews under `docs/reviews/tmp/`
are ignored. Committed
prototype evidence and review records remain part of the repository. Capture-time
paths and hashes in evidence are historical records; moving documentation does
not regenerate measurements or certify additional platform behavior.

## Editor native acceptance

After building `examples/editor`, run its `-smoke` mode for deterministic Cocoa
input, composition, native range geometry and document persistence checks. Run
native desktop fixtures sequentially. The portable and offscreen tests do not
qualify these AppKit behaviors.

A separate read-only accessibility inspector can qualify exact text, UTF-16
ranges, screen bounds, point lookup and RTL geometry using an already trusted
caller:

```sh
python3 scripts/check-editor-accessibility.py \
  --binary /tmp/vellum-editor \
  --output /tmp/vellum-editor-external-ax
```

The output directory must be new. The wrapper compiles its optional Swift
inspector, checks existing accessibility trust without prompting, and runs the
editor's dedicated 60-second `-ax-fixture`. It records binary/source hashes and
both process results. It refuses unsupported platforms or missing trust before
launching the application. It never changes input sources, VoiceOver or privacy
preferences. Swift is needed only for this acceptance tool. Human VoiceOver and
IME candidate-panel workflows remain separate checks. See the
[editor acceptance record](reports/macos-editor/README.md).

## Grid acceptance

`examples/grid` is a separate consumer module with independent 10,000-row and
million-row procedural windows. Its offscreen tests route primitive pointer,
wheel and keyboard events through the managed widgets and compare completed
software frames at different owner drain batches. `scripts/check.sh` includes
these tests; `scripts/check-example.sh` race-tests, vets and builds the module.

After preparing dependencies, source `drivers/cogent/env.sh` and run
`go run . -smoke` from `examples/grid` for sequential Cocoa measurement and
two-window lifecycle checks. This desktop smoke uses explicit owner commands;
offscreen routing, desktop rendering and physical input are separate evidence.
See the [grid example acceptance](reports/macos-grid-example/README.md).

`examples/loadedgrid` stores actual million-row ×20 numeric records and runs
generation, sort/filter indexing and immutable edit copies on bounded workers.
From the repository root, source `drivers/cogent/env.sh`, then run `go run .`
inside that module. `-rows 10000` selects a smaller input and `-variable-rows`
adds indexed heights. Its offscreen race tests and vet are included in
`scripts/check.sh`. Apply waits for the filter text field's native input barrier
before reading its controlled value; Save uses the inline-cell barrier.
Run `go run -race . -smoke` and then `go run -race . -smoke -variable-rows`
sequentially for the bounded Cocoa workflow. See the
[loaded-data report](reports/macos-grid-loaded/README.md) for qualification scope
and the remaining sustained performance and large-tree gates.

`examples/loadedtree` uses `ui.DataTree` and a worker-built `virtual.Outline` with
100,000 actual stored nodes. Load, expansion and ancestor filtering run on bounded
workers; owner navigation and admission use accepted indexes. Run its offscreen
tests for deterministic routing or `go run -race . -nodes 100000 -smoke` for the
bounded Cocoa owner workflow. This does not qualify tree frame latency or human
screen-reader workflows.

`examples/loadedgrid -perf` and `scripts/benchmark-grid.py` opt into one-window
native work tracing for a 400-cell stored table. See the
[measurement contract](../specs/grid-performance.md) for exact work boundaries,
raw records, paired 10k/million geometry and the unchanged 4 ms/6 ms warm targets.
Performance runs use optimized builds, separately from race-instrumented checks.

## Standalone applications and benchmarks

`examples/hello` is a separately versioned consumer module. To create a consumer
outside the repository, source `drivers/cogent/env.sh` and run
`python3 scripts/new-app.py /path/to/my-app --module example.com/my-app`.
Its four explicit replacements point to the local framework and prepared native
modules. Run `go test -race ./...`, `go vet ./...` and `go build .` in the result.

Run `scripts/benchmark.sh` for portable allocation/time baselines, or add `--native`
to launch three desktop processes and measure verified GPU startup, synchronous
render/readback, RSS and sampled idle CPU. The [SDK contract](../specs/sdk.md)
defines measurement boundaries and residual release/platform gates.

Collect Mac SDK distribution inputs without launching a window:

```sh
python3 scripts/distribution-inventory.py --output /tmp/vellum-distribution.json
```

Use a new output path outside the checkout. The report records SDK file hashes,
resolved embedded assets, listed cgo sources and native linker-search candidates
against the verified module notice graph. Review-required asset/native rights
and prebuilt archive component provenance remain separate from this inventory.
The collector does not approve or publish a distribution.

Set `VELLUM_CACHE_DIR` to reuse a chosen Go cache across preparation/check commands.
The legacy `P0_CACHE_DIR` variable is accepted for existing local environments.
Without either variable, sourcing the environment creates an isolated temporary cache.
