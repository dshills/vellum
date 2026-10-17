# Current native grid viewport measurements — macOS

The current ordinary stored-table workload still misses its unchanged 4 ms p95 /
6 ms p99 UI-work target. All 40 native processes completed valid measurements;
all 40 failed p95 and 32 failed p99. Successful collection does not qualify the
performance gate. P6, P11 and full macOS completion remain open.

The fixture now records actual native window geometry after warm-up and after
measurement. Native queries wait for main-queue publication and run on main;
geometry is copied under the native lock outside measured work. The consumer
refuses partial or changed endpoints, invalid extents, grid clips outside render
or DPI/scale mismatch. Historical records without this guard retain their old
scope. This checkpoint changes fixture observation, validation, tests and specs;
shipping framework implementation and timing regions are unchanged.

## Observed workload and results

An optimized CGO build from base `23ea03fc70febdd1a8dd9f32e02cc482a32bfb72`
plus six exact fixture/parser/test/spec changes ran without overlays, race or
profiling instrumentation. Its original dirty VCS metadata, complete source and
prepared-file hashes, binary hash and launch arguments are preserved. The tested
binary is not relabeled as a clean build of this later checkpoint.

The requested initial client was 1500 × 1250 Dp. Observed native client and render
were 1519 × 1266 pixels on LG HDR DQHD, screen 0, 5120 × 1440, logical DPI 162,
DPR 1. Window-manager client units were also 1519 × 1266. The intended 768 × 624
logical grid viewport rounded to an observed 767 × 624; initial grid bounds were
(9, 32)–(1325, 1106) at scale 1.6875. The required minimum 400 clipped cells,
50 rows and eight columns was maintained. Native endpoint geometry and complete
per-sample grid geometry agreed for every paired 10k/million run.

Five serial repeat sets covered CPU and GPU-requested modes, fixed and stable-key
indexed variable heights, and 10,000 / 1,000,000 stored rows. Mode, height and row
order reversed on even repeats. Each process warmed 128 routed events and
measured 120, totaling 4,800 tickets and 2,400 row-count geometry comparisons.
The table reports milliseconds; each pooled group contains 600 samples. Per-run
p95/p99, tails, variation, pairs and raw samples remain in the evidence.

| Mode | Heights | Rows | Pooled p95 ms | Pooled p99 ms | Median process p95 ms |
|---|---|---:|---:|---:|---:|
| CPU | fixed | 10,000 | 6.627 | 7.071 | 6.295 |
| CPU | fixed | 1,000,000 | 6.158 | 6.685 | 6.066 |
| CPU | variable | 10,000 | 6.899 | 7.240 | 6.654 |
| CPU | variable | 1,000,000 | 5.992 | 6.499 | 5.963 |
| GPU requested | fixed | 10,000 | 6.184 | 6.565 | 6.031 |
| GPU requested | fixed | 1,000,000 | 5.927 | 6.309 | 5.871 |
| GPU requested | variable | 10,000 | 5.755 | 6.191 | 5.624 |
| GPU requested | variable | 1,000,000 | 5.769 | 6.242 | 5.562 |

## Checks and review

The Python consumer suite passed 53 tests, including guard completeness, endpoint
changes, invalid native geometry, scale and clipping rejection, plus historical
compatibility. The loaded-grid offscreen race tests and vet passed. Native
observation tests cover queued execution, lock ordering, detached values and
closed/wrong/multiple-window refusal. Worker completion joins native teardown.

Internal source, launch-readiness and result reviews cover this fixture checkpoint.
Prelaunch review corrected evidence handling: parse and hash the same review
bytes; validate saved launch/owned-process/binary/source/review/log bindings;
consume both stdout and stderr; preserve analysis with exclusive creation.
These corrections preceded native execution. The previous sizing implementation
received one Prism review using Anthropic Claude Sonnet 5.5; no additional
provider call is claimed for this measurement checkpoint.

All 40 owned groups exited 0, were reaped and observed absent. Streams reached
EOF with matching counts/hashes; source, prepared bytes and reviewed inputs stayed
stable. An exclusive native-test lease, stop markers, process deadline and output
bounds governed collection. This campaign is sealed against further execution.
No input-source, VoiceOver, clipboard, power, permission or other OS preferences
were changed. Scrolling entered the native event deque programmatically; physical
pointer/key interaction was not verified.

See [structured results](result.json), [manifest](evidence.manifest.json),
[raw evidence](evidence.tar.gz) and [measurement contract](../../../specs/grid-performance.md).

## Limits and next work

These measurements use the corrected current native viewport. There is no old
binary/viewport comparison, algorithmic-speedup claim or retroactive qualification
of historical failures. Native observations are endpoints and do not prove every
intervening extent or continuous host quietness. GPU-requested mode does not
establish GPU execution, native/GPU allocation or presentation latency. End-only
process observations are not retained-control memory attribution.

The next performance work should isolate remaining frame costs at this same
workload, then compare reviewed implementation changes with matched ordinary
runs. The complete table/tree target, 1,000-control/canvas stress, native/GPU
memory, endurance, physical interaction, cross-display and broader accessibility
qualification remain open. Public release remains gated.
