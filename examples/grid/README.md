# Managed two-dimensional grid example

This independent module declares two managed `ui.DataGrid` windows: 10,000 and
1,000,000 **procedural** rows, each with 20 columns. It does not store a million
rows or construct a row-key index. A row key identifies the domain row even when
the indexed projection is reversed or filtered to even rows. Cell text is a
fixed integer function of that key, column and seed 7.

The Grid uses 24 logical-pixel rows, a frozen 28-pixel header and two frozen
leading columns. Reverse/filter replaces the immutable indexed source revision
in constant time. The buttons also demonstrate first/last key jumps, paging and
panning; the managed Grid handles pointer selection, two-axis scrolling and
focused arrow/page keys. Selection is a single active cell. The two models and
window lifetimes are independent.

From the repository root, prepare the pinned native dependency as described in
[development](../../docs/development.md), then run:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. drivers/cogent/env.sh
cd examples/grid
go run .
```

The default keeps the fixed-height source and its original callback-count
fixtures. Opt into variable rows with:

```sh
go run . -variable-rows
go run . -smoke -variable-rows
```

This mode wraps the indexed source with optional `virtual.GridRowMetrics`.
The stable domain key selects a repeating logical-pixel height pattern of
23/35/47/29; reverse and even-row filtering preserve each retained key's height.
Offsets and pixel-to-row lookup use a period of at most four, rather than a
row-height array or a million-row scan/index. Metric callback counters are
separate from the fixed row/cell counters. The per-window **Toggle row heights**
button is offered only in this startup mode. Each toggle installs a new source
revision, including toggling back to the fixed 24-pixel height; it retains an
independent active key and best-effort viewport anchor through `ReplaceSource`.
The old immutable metric revision remains unchanged.

`go run . -smoke` is a bounded desktop check. It waits for current measured Grid
frames, verifies both row counts and bounded cell pools, jumps to the millionth
row and column 19, checks frozen geometry, replaces the indexed source while
retaining a stable selected key, verifies inspection refusal after logical close,
and closes each domain window. It does not separately observe desktop compositor
teardown; the native runtime return and process exit are recorded separately. Each completed
stage emits JSON; the final stage is `pass`. This smoke uses explicit owner model
commands. It does not synthesize desktop pointer or keyboard input, measure
performance, or change macOS preferences. The returned native runtime and
adapter error paths also determine the exit status.

With `-variable-rows`, the desktop smoke additionally computes a Page destination
from the active row's top plus the current body height in logical pixels. It
checks the resulting stable selection and completed measured frame. This is
still an explicit owner command, without synthesized desktop input.

Portable tests verify indexed key inverses, immutable revisions, invalid keys,
projection selection/ref retirement, and the same fixed viewport for 10k/1M rows.
At 960×1228 logical pixels they require exactly 400 visible cells, 520 materialized
cells with overscan, 52 indexed row-key calls, 52 indexed reverse lookups and 520
cell calls for either source size. These are access counts, not a latency or
memory benchmark.

The `offscreen` tests exercise the actual public `cogent.NewManual` router with
supplied pinned font bytes and theme, explicit frames and fixed virtual time:

```sh
go test -race -tags offscreen -count=1 ./...
```

They route real primitive mouse/key/wheel events to measured controls, rather
than invoking `Activate`/`GridChange` callbacks directly. The routed acceptance
covers far scrolling, frozen geometry at 96/120/144/192 DPI (including fixed 23-pixel rows and 121-pixel columns
at fractional scales), key paging, source replacement
with stable identities, stale presses, disabled/modal routing, independent window
retirement, software image capture and equivalent results at owner batches
1/7/64. Source-bound qualification is recorded in
[Grid example report](../../docs/reports/macos-grid-example/README.md);
the integration owner retains the handoff, tested source graph and raw check logs there.
Desktop execution is recorded separately by the integration owner. A desktop
build alone does not qualify the native smoke.

Variable-row portable tests compare prefix and pixel inverse queries to a
small brute-force projection, check endpoints and malformed sources, verify
stable-key heights across reverse/filter and immutable height toggles, and
count bounded indexed work for 10k/1M rows. The additional `offscreen` tests route
actual measured row targets, pixel Page keys, wheel deltas and height-toggle
buttons; they check odd-height clips/shared edges and stale press retirement.
These tests do not use UI callbacks as an input surrogate. The previously
published report above describes the fixed-height checkpoint; variable-mode
acceptance is recorded separately by the integration owner.

All-row/range selection, cell editing/conflicts, tree expansion, loaded-data
storage, native services, human accessibility/IME checks and performance
qualification remain outside this example's evidence. See the public [Grid specification](../../specs/grid.md).
