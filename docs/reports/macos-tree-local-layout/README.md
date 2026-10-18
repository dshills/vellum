# Rejected retained tree layout prototype

The private implementation passes its offscreen/race/build checks and separate native
route diagnostics, but is **rejected**: median run-p95 and run-p99 worsen in both modes,
with zero improved pairs. No runtime, public API, prepared dependency or SDK change is
adopted. All 20 ordinary p95s still fail 4 ms.

Base: clean `e154763b7083d23d0c7b09e2433baa65da5fa9b6` in both working checkouts. The
private overlay uses the relocated SDK built from clean
`abe67c625a498d27b78f9719e6038cc1d6c23e88`, Go 1.25.6, stage0079 and the unchanged
stored-tree performance consumer. Later base changes are documentation only. Canonical
source, SDK and prepared libraries remain unchanged during the experiment.

## Implementation

The candidate keeps overlapping native rows by domain key within the current admitted
projection, uses each departing row as a donor once and retires surplus rows. It retains
no offscreen row or paint cache. Each installed row receives its current generation's
OutlineRef; widget identity never serves as input authority.

`OutlineSnapshot.IndexIdentity` captures a comparable opaque identity with State/Rows.
Its private byte token holds no index, data or model reference. Equal source
stamps/counts do not establish index identity. The token is process-local, excluded from
JSON and invalid as input authority.

The footer is made fully passive because Core's local text updater refuses every
interactive text ability; row text already uses that same passive list. This
intentionally changes the private footer's interaction abilities. Pointer actions remain
owned by the outline host/pager. The frozen-constructor oracle shares the candidate
mount styling. The oracles establish software pixels, completed row geometry/semantics
and routed row actions; they do not prove unchanged footer interaction behavior. This
change is rejected with the rest of the experiment.

Only proved viewport changes may use the existing fixed-allocation local-layout
mechanism. Actual scene, parent, geometry, resolved style, scale, dimensions, row
configuration and immutable index must match. Pending work, selection, topology, changed
pager states, theme/font/width/ownership changes or uncertain proof take complete
layout. Changed passive row/footer text is styled and sized; fractional physical row
sizes are checked. Complete scene paint preserves painter order and clears old pixels.
Closing clears retained layout-stamp/text references. `TreeInspection.LocalLayouts`
counts successful local-root passes, not presentations.

## Correctness and build evidence

All commands use Go 1.25.6 and the private overlay. Virtual/root and full driver
regressions use `-race -count=1`; driver runs use `-tags=offscreen`. IP networking is
denied and local Unix IPC remains allowed for headless compilation/tests. Every accepted
receipt records stable pins, terminal child and owned process-group absence.

- Whole virtual and root-module race suites pass.
- Focused tree and complete driver-root package (`.`) offscreen race suites pass; the
  driver-root run takes 274.807 seconds.
- Driver vet and ordinary/diagnostic native builds pass.
- At DPI 96/120/144/162/192, local layout matches exact software pixels and
  geometry/semantics after forced full layout, with unrelated sibling sizing skipped on
  successful local passes.
- A separate sequential owner/model uses the frozen original visible-index constructor.
  Captures and completed geometry/semantics match the new constructor across scrolls and
  scales; its different slot-reassignment behavior is asserted.
- Two scrolls before one frame refuse pending geometry and stale accessibility actions.
  A press armed before scrolling cannot select on release. After a positively observed
  local pass, current accessibility or routed pointer input selects once; queued
  text/pending work is cleared.
- Zero/foreign equal-shape index identity, pending state, selection,
  width/row-height/help changes and forced full-layout cases refuse local admission.
  Existing tree tests cover native geometry, keyboard, disclosure, pending publication,
  modal/disabled/retired scopes and close/remount behavior.

The first independent fixture attempted simultaneous Manual instances and was rejected
by the global render lease. The sequential correction first had an unused import; both
failures are preserved. A later passing body is excluded because the runner changed
before terminal pin validation. The final accepted focused receipt reruns stable bytes.
Cocoa deprecation and duplicate `-lobjc` build warnings are retained.

These are software routing/offscreen correctness checks. They do not establish physical
pointer, VoiceOver, IME, cross-display or actual GPU pixel equivalence.

## Native observation and adoption

The frozen plan completes two separate candidate route diagnostics, followed by five
original/candidate pairs per CPU/GPU mode, alternating build order. Each process has 128
warmups; diagnostics record 120 measured frames and ordinary processes record 240. A
shared desktop lease, fresh stop markers, watchdog/log bounds and owned process-group
retirement apply. The ordinary consumer is byte-identical between builds. A separate
diagnostic adds only the warm local-layout count to JSON.

The predetermined adoption bar is at least 0.4 ms lower median run-p95 and improvement
in at least four of five pairs in each mode, no increased median run-p99 or number of
p99 failures over 6 ms, matched geometry and bounded retention. The terminal Go-heap
comparison is only a prototype guard, not native/GPU memory or M4 qualification. The
original 4 ms tree target and all other framework gates remain unchanged. The 90-minute
investigation starts at 01:02:52 UTC and the explicit rejection is recorded at
2026-10-18T01:27:53.231725+00:00 on 18 October 2026, before its 02:32:52 UTC deadline.

 Both candidate diagnostics observe 40 successful local-root passes by the completed
warmup snapshot (initial and first-scroll counts are zero). The CPU diagnostic records
120 software scene sources; the GPU diagnostic records 120 fresh accepted scene-host
draws without fallback or reuse. This proves actual local-root/renderer route use for
these diagnostic corpora. It does not establish GPU pixel equivalence, GPU completion,
presentation or a local pass for every ordinary frame.

All 22 native processes exit zero with complete raw logs, empty stderr, stable input
pins, successful owned cleanup and observed group absence. Each case and the overall
scope are sealed. The plan binds 30 source/helper/prerequisite pins, exact
plan/admission bytes, all three executables, unchanged SDK vector and clean calling
checkouts. The reviewed ordinary runtime environment removes GOMAXPROCS, GOMEMLIMIT,
GOFLAGS and GORACE. Direct binary execution makes no enforced network-denial claim. All
4,800 ordinary samples are recomputed independently with integer nearest-rank
percentiles per run; diagnostics supply 240 separate measured samples.

| Mode | Original median run-p95 ms | Candidate median run-p95 ms | Original median run-p99 ms | Candidate median run-p99 ms | Improved pairs | Adoption |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| CPU | 7.293207 | 7.848000 | 7.949916 | 8.596251 | 0/5 | Reject |
| GPU | 7.071166 | 7.531169 | 8.300042 | 8.741667 | 0/5 | Reject |

| Mode / pair | Original p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| CPU / 1 | 7.293207 | 7.979166 |
| CPU / 2 | 7.272084 | 7.760165 |
| CPU / 3 | 7.458001 | 8.023209 |
| CPU / 4 | 7.340749 | 7.848000 |
| CPU / 5 | 7.259542 | 7.663251 |
| GPU / 1 | 7.385042 | 7.875290 |
| GPU / 2 | 6.914669 | 7.449419 |
| GPU / 3 | 7.516209 | 7.879960 |
| GPU / 4 | 7.006250 | 7.272834 |
| GPU / 5 | 7.071166 | 7.531169 |

Geometry matches in every pair: 68–69 nonempty clipped rows, bounded
materialized/retained rows, identical effective bounds, scale, normalized row rectangles
and local font. No samples are pooled across modes/builds; no failed limit is discarded.
Median terminal post-GC Go heap is 44,565,688/44,565,168 bytes for original/candidate
CPU and 46,607,712/46,617,800 bytes for GPU. These satisfy the experiment's +1 MiB
comparison guard but include 100k model nodes/caches/recorders, and say nothing about
the empty-workbench memory target or native/GPU deallocation. The largest sample is
10.235/9.764 ms on CPU and 8.823/9.798 ms on GPU for original/candidate; median p99
regresses in both modes.

The native local path exists but does not improve the tail. This experiment does not
isolate which remaining full-layout, map/style, raster or scene-source costs cause the
regression. The earlier ordinary boundary's phase costs remain historical evidence, not
a subtraction from this candidate. The entire private overlay, including new
identity/counter APIs, is retired. No local-guard relaxation, extra retained rows or
cache tuning is queued. A different material scene-publication/rendering decision is
needed for M3; the next independent implementation backlog item is M6's exact
shipped-component and rights inventory.

## Review, evidence and remaining gates

Internal review covers source, frozen native admission, complete outcomes and the
rejection. Prism reviews all twelve staged paths (59,966 bytes) in one fresh Anthropic
Claude Sonnet 5.5 call, with complete declared coverage and no excluded/truncated files.
It retains one high and two medium findings; CLI exit one is preserved. The high
compile-error finding quotes a `[REDACTED]` token initializer. Actual source line 207 is
syntactically complete (`IndexIdentity: TreeIndexIdentity{token: o.index.token}`) and is
bound to the passing race/build receipts; the marker appears only in the quoted finding
evidence, not the actual source. Its introduction is not established by this record.
This finding does not identify an actual compile failure. The footer-passivity finding
is addressed by the explicit behavior/oracle qualification above. The pending-error
concern is reviewed against the existing fail-closed projection/measurement fence and
recovery path; errors invalidate projection and measurement, and recovery requires
complete layout before recertification. The internal review refutes this as a
correctness defect; its disposition is retained in the evidence. No rejected source is
adopted and no provider rerun is used to obtain a clean result.

[Raw records and private sources](records.tar.gz) and their [per-file
manifest](records.json) retain the original/candidate overlay, frozen original oracle,
source versions, headless failures and accepted receipts, exact native plan/admission,
22 native logs/owned records, independent assessment and reviews. SDK/executable/font
payloads remain in the preceding SDK artifact and local experiment; they are not
duplicated in this archive. The code was temporarily staged only in the results worktree
for checkpoint review. All nine candidate source paths were then restored to the exact
base; the final staged changes contain documentation/evidence only.

This is a failed bounded update/layout experiment, not Mac completion. Unchanged table
failures, full 10k/100k wide/deep and original 500-visible/P6 expansion/filter
qualification, general UI/text/canvas/startup/input/pacing/resize, native/GPU memory,
idle/endurance, native M2 and physical/human/IME/VoiceOver/display workflows,
rights/final distribution and M7 remain open. No system preferences, clipboard, input
source, VoiceOver, placement operator or physical pointer input are used.
