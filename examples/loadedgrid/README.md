# Loaded stored-data grid

The separate `-accessibility-smoke -rows 10000` (optionally `-variable-rows`)
acceptance mode uses the same stored performance fixture with at least 400 actual
clipped cells. It queries Cocoa values, indices, screen frames and opaque object
identity across routed scrolls, replaces the binding under the same node ID and
exercises a current native cell action. It runs separately from `-perf` and
`-smoke`; its queries are not latency measurements or screen-reader certification.

This independent module starts with two usable empty DataGrid windows. **Load records** generates an immutable array of signed numeric records and builds its initial projections on a bounded task worker. The large window uses one million stored records by default; the small window uses the first 10,000 records from the same initial backing array. Each record contains twenty actual `int64` values, generated with seed 7. Subsequent edits produce independent immutable domain revisions for the affected window.

From a prepared checkout, run:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. drivers/cogent/env.sh
cd examples/loadedgrid
go run .
# A smaller data set or indexed variable heights:
go run . -rows 10000 -variable-rows
```

See [native setup](../../drivers/cogent/native/README.md). The module uses repository-relative replacements; its prepared native dependencies belong to the SDK or checkout, rather than a machine-specific development overlay.

**Sort** changes numeric column 1 order. The controlled filter field applies a case-sensitive Contains filter to column 2. Apply first settles that field through its originating window FIFO, so admitted native text and composition reach the controlled value before the worker query captures it. **Clear query** restores the complete domain. **Cancel work** cancels an admitted query, load, or edit without pretending its worker has already terminated. Each window has two task credits; a credit returns only after both worker termination and owner completion. A refused new task leaves an already accepted query intact.

**Edit active cell** (or F2/Return in the grid) opens the real managed cell field. Column 0 is read-only; other columns accept canonical signed `int64` text. Return or **Save cell** settles that field, then copies the stored array and builds the new projection off-owner. The owner checks the captured domain pointer, query token, edit session, cell version and request before accepting the candidate; it publishes its domain pointer only after `CompleteEdit` succeeds. An edit that stops matching a filter can close successfully and remove its row from the view. **External +1** constructs a new domain revision on a worker; an older draft remains recoverable as a conflict. Cancel discards a draft explicitly. Changing sort/filter preserves stable identities, cursor/viewport policy and the captured selection universe; the selected count describes the captured identities, rather than just currently visible filtered rows.

The immutable full-domain source uses dense stable keys and arithmetic reverse lookup. Worker projections store indexed order/inverse arrays. Domain revision and accepted view revision are separate. Variable heights depend on stable domain row identity and stay attached to that identity after sorting/filtering. Source snapshots and retained results are concurrent-read safe; diagnostic counters are atomic and selection captures have no mutable counter state.

The example admits at most 1,000,000 rows, four sort/filter clauses, 4,096 query bytes, 64 bytes per numeric cell, 4,096 edited-cell version entries and 4,096 materialized visible cells. Builder index and working budgets are 64 MiB and 256 MiB respectively, **excluding borrowed stored-domain payloads**. One million records hold 160,000,000 logical value bytes; an edited large-domain revision makes another full array copy on its worker. Current domains, the shared initial base and bounded in-flight immutable revisions can coexist. These bounds are not a total process/RSS or peak-memory estimate.

Portable tests check physical storage, dense indices, immutable edits, cancellation, refusal, domain/query/draft CAS and dual-terminal task credits. A separately reported Go `HeapAlloc` delta observes one retained stored million-row array; it is not an RSS, native/GPU-memory or latency measurement. Offscreen tests declare the actual controlled fields, buttons and DataGrid, then route primitive keys/pointers through Cogent's real managed scene/hit/focus path with supplied fonts/theme, explicit workers and owner batches 1/7/64. They include queued filter input followed immediately by Apply, out-of-order canceled jobs, window closure, settled async editing, conflict recovery and an edit removing the sole filtered row.

The bounded desktop fixture is:

```sh
go run -race . -smoke
go run -race . -smoke -variable-rows
```

It starts empty, loads the stored million records, checks completed current-stamp/frame geometry, runs worker sort/filter/clear, preserves captured selection, saves through the actual originating cell-field settlement and tests an external conflict. Its commands are owner operations, rather than physical desktop typing/pointer automation. Its terminal check closes domain windows; it does not independently measure private decoder cleanup or compositor teardown. Human IME, VoiceOver, broad accessibility, peak memory and end-to-end professional workload performance remain separate acceptance work. Source-bound qualification and preserved exploratory failures are handed to the integration owner for [the durable report](../../docs/reports/macos-grid-loaded/README.md).

The separate release-only performance mode compares actual stored 10k/million
fixed/variable tables at ≥400 measured visible cells. Its bounded originating
scroll trace records same-ticket UI work phases, source callbacks, Go allocation
observations and end-only process RSS. External `ps` runs after the completed
end record, outside measured native frames; baseline/warm process RSS and process
CPU deltas are unavailable. The native post-warm Go heap snapshot remains.
It excludes dataset generation and projection
building from warm UI work. See [the measurement contract](../../specs/grid-performance.md)
for the explicit five-run command, boundaries and unchanged P6 targets.
Performance mode disables native window-geometry persistence for the fixture
lifetime and reports the actual switch in its setup record, so measurement
windows do not restore or save application geometry. The previous process-local
switch is restored after teardown. Ordinary interactive mode is unchanged.
`-perf-novel` reports newly visited row strings separately from warm scrolling;
no screenshot/readback duration is presented as UI CPU work.

For an explicitly separate diagnostic, `-perf -rows 10000 -perf-cpu-profile
/private/tmp/new-grid-profile.pprof` records a bounded process-wide Go CPU profile
after warm-up. The output file must be new. Profiled samples are labeled and
rejected by the ordinary warm-budget runner; they are not qualification timings.

`-perf -rows 10000 -perf-source-details` enables bounded scalar native source
costs. The runner's matching `--source-details` mode reports DIAGNOSTIC and always
leaves warm_target_met false. Nested raster/Copy/image-API times do not add to
the existing UI-work total. Image counters mean API attempts and logical full
RGBA input bytes; upload success/failure and physical GPU traffic are unavailable.
See [the source-cost evidence](../../docs/reports/macos-source-costs/README.md).

Warmup must return to its zero-offset starting viewport with one callback per
admitted event. Offset, scale, geometry or callback drift refuses the run without
resetting the model. `-perf -rows=10000 -perf-input-details=/private/tmp/new-input.json`
selects a separate bounded scalar input ledger saved after native teardown,
including failed warmups. It observes admissions, actual wheel remainder/scale
and owner transitions; untraced means tag zero, not a desktop-origin assertion.
The file must be new. These diagnostic samples cannot qualify warm latency.

## Native presentation preflight

`-perf -rows=10000 -perf-presentation=/private/tmp/new-loadedgrid-presentation.json`
records a separate bounded raw journal using the public native presentation API.
It excludes other diagnostic modes and requires a new file. Every performance
event is labeled as diagnostic. The strict validator joins receipt, completed CPU
generation and actual native presentation, retaining losses and teardown state.
These serial runs cannot qualify sustained 60 Hz cadence or physical input latency.
See the [capture contract](../../specs/presentation-capture.md) and current
[Mac release policy](../../specs/macos-responsiveness.md); the earlier 4 ms budgets
remain optimization targets. No real presentation journal has been qualified by
this checkpoint.

## Fixed-deadline 60 Hz workload

`-perf -rows=10000 -perf-samples=600 -perf-60hz
-perf-presentation=/private/tmp/new-loadedgrid-60hz.json` selects a separate
10-second planned workload. Use the flags on one command line and a new file.
It retains every planned slot, late backlog, generation-bound viewport change
and actual native feedback. Per-event stdout/memory profiling stays outside
the cohort; validation/owner polling stays included. The version 2 analyzer uses
N-1 final-state intervals and reports collapsed states and overruns explicitly.
See the [paced contract](../../specs/paced-presentation.md) for exact bounds,
commands and qualification limits. These tools have not yet qualified live 60 Hz
performance, physical input or the remaining Mac release workloads.

Performance fixtures initialize Core settings before disabling window geometry
persistence, and restore the loaded process-local policy after teardown. Saved
settings are read, never saved by this guard. Settings errors or invalid zoom
refuse the fixture before native setup. The ordinary interactive demo is unchanged.

### Opt-in display backpressure experiment

The Mac-only `-experimental-display-backpressure=1s` option requires `-perf`
and a new `-perf-presentation` journal. Its timeout must be 100 ms through
5 s. It enables one display credit per native window before its first native
acquisition; a missing or skipped display callback refuses the run. The journal
records `Config.DisplayBackpressureNS` only when enabled. Zero keeps the existing
rendering and journal behavior.

This is a prospective experiment, with no native performance qualification yet.
The same input workload, acquisition prefix, final presentation requirements and
latency/cadence thresholds still apply. It does not drain startup paint work or
permit late unarmed skipped acquisitions. Run it in a separately reviewed native
scope; existing captures retain their original results.
