# M3-A segmented ready snapshot experiment — rejected

17 October 2026. Clean base `9d3c97f` uses runtime `bd672a6` and unchanged stage0079
prepared dependencies (6,260 files). This checkpoint changes documentation only.
The entire experimental implementation, tests and source-details diagnostic field
were restored to HEAD after the fixed comparison. No experiment code is shipped.

## Decision and native measurements

Reject M3-A and stop small follow-up tuning of this candidate. The committed
adoption bar required at least 0.4 ms lower median run-p95 and improvements in at
least four of five pairs without correctness, retention or tail regression.
All five pairs were worse. Median run-p95 increased **4.749206 → 4.866957 ms**
(+0.117751 ms, +2.48%). The median of per-run nearest-rank p50 process allocation decreased
**5,300,464 → 4,727,104 bytes** (10.82%); allocation alone does not qualify latency.
Every ordinary run fails the unchanged 4 ms p95 target; all ten runs pass the
6 ms p99 target. The table gate requires both, so it remains open. M2 native,
M3, M4, M5, M6 rights/final SDK and M7 also remain open.

| Pair | Baseline p95 ms | Candidate p95 ms | Baseline p99 ms | Candidate p99 ms |
| --- | ---: | ---: | ---: | ---: |
| 1 | 4.891168 | 4.968460 | 5.534083 | 5.308375 |
| 2 | 4.686252 | 4.746001 | 5.060834 | 5.073999 |
| 3 | 4.749206 | 4.835670 | 5.288124 | 5.463540 |
| 4 | 4.676960 | 5.226874 | 5.590916 | 5.807958 |
| 5 | 4.763170 | 4.866957 | 5.327836 | 5.371209 |

Five ordinary release runs per build alternate order by pair, with 128 warmup
operations and 120 samples per run, GPU scene rendering requested and normal
runtime settings (no GOMAXPROCS/GOMEMLIMIT/GOFLAGS override, race or profiler).
The existing runner checks at least 400 nonempty clipped cells, matched geometry
and identical trajectories at 10k fixed rows. The previous verified serializer
release binary is reused: `9d3c97f` is documentation-only after `bd672a6`.
This is a narrow adoption comparison, not a repeat of the full CPU/GPU table
matrix, display/input/presentation qualification or a physical pointer test.

Two separate source-details runs are never pooled into ordinary timing. Each
records 120 captures, all accepted, no fallback or reuse. Candidate cumulative
segmented update counts increase 130 → 249 across 120 distinct samples; baseline
counts are zero. This positively observes the proposed route, not per-frame cost
attribution. The diagnostic field was reverted with the experiment.

## What was tested and rejected

The private experiment constructed eligible ready leaf grids once into final
independent detached groups, shared those immutable arrays with completed caches
and serialized an ordered reference vector. Callbacks/guards stayed separate;
pending/editor/custom/namespaced/unsupported shapes retained the whole old route
before mutation. Shape admission preceded row-map allocation and final coherence
checks remained. Native transport and prepared dependencies did not change.

A frozen original visitor/grid producer supplied an independent exact wire and
geometry oracle. Actual two-grid layouts with ordinary peers passed at five DPIs,
enabled and disabled; duplicate cache/epoch, pending stale actions, source/target
removal and close cleanup passed. Detached row-selection/empty-first geometry
parity and a fresh successful press requiring selection, exactly one GridChange
and focus also passed. The complete offscreen driver race suite passed
(279.326 seconds); focused race checks, both JSON implementations, vet and the
release build passed. Initial failures in expected stale-error handling,
simultaneous Manual lifetime and setup-notification counting were test-fixture
errors; original failed logs and corrected passing logs are retained separately.

Internal implementation and native-admission reviews found no remaining
correctness/ownership blocker. Passing tests do not override the failed timing
criterion. One Prism Anthropic Claude Sonnet 5.5 review covers the final rejection
report and delivery/completion updates; the rejected source received internal
review and is retained as evidence, not resubmitted as a proposed production change.

## Evidence and limits

[records.json](records.json) hashes every member of [records.tar.gz](records.tar.gz).
The archive contains raw ordinary native logs under
`native/native/{pair}-{baseline,candidate}/native.log`, the comparison and binary
receipts, source-details logs, test commands/stdout/stderr, internal reviews,
`rejected-source.diff` plus all eleven changed source/test leaves, the exact
revert receipt, final Prism response and assessment. `assessment.json` lists all
pair p95/p99 values and the fixed decision. Percentiles use each 120-sample run's
integer nearest-rank p95/p99; medians aggregate the five run values per build.

The live lease, fresh stop scope and source/binary pins were held throughout;
all app invocations exited zero and the runner attempted owned-group cleanup.
This does not establish strict per-PID absence, physical input, actual presentation,
resource-budget qualification or human screen-reader behavior. Existing compiler
warnings remain in the logs. No preferences or system input sources were changed.

## Next work

Continue the finite [Mac delivery plan](../../../specs/macos-delivery.md).
Use the available desktop for the pending composed-workflow and placement checks,
with a fresh current-source probe/stop scope and one bounded input-delivery attempt.
If delivery fails again, retain the failure and require direct/manual verification;
do not spend another open-ended cycle constructing operators. Human/display cases
stay open. Before another M3 renderer/scene redesign, collect actual operation and
resource-kind counts, encoded-buffer composition and source phase evidence for
the failing fixture. Neither aggregate bytes nor Go cgocall samples prove a new
shader/transport bottleneck. No further M3-A tuning is queued. M2 native, M3,
M4, M5, M6 rights/final SDK and M7 remain open.
