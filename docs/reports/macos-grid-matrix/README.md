# Current Mac table matrix and M3 decision

Clean runtime source `bd672a6` completes the existing ordinary fixed/variable
10k/1M stored-table matrix in both CPU and GPU-requested modes. Each of the eight
cases has five 120-event release runs after 128 warm events: **40 runs and 4,800
ordinary samples**. Every run finishes its fixture and exits zero. **Every p95
fails the unchanged 4 ms limit.** This is a complete measurement of this table
matrix, not a pass for it or the broader Mac framework.

| Mode | Case | Median run-p95 ms | Run-p95 range ms | Worst run-p99 ms | p95 passes | p99 passes |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| GPU | fixed-10000 | 4.707 | 4.686–4.950 | 5.843 | 0/5 | 5/5 |
| GPU | fixed-1000000 | 4.730 | 4.575–5.070 | 8.617 | 0/5 | 4/5 |
| GPU | variable-10000 | 4.819 | 4.794–4.882 | 5.532 | 0/5 | 5/5 |
| GPU | variable-1000000 | 4.772 | 4.662–4.847 | 5.349 | 0/5 | 5/5 |
| CPU | fixed-10000 | 5.703 | 5.504–5.977 | 6.518 | 0/5 | 1/5 |
| CPU | fixed-1000000 | 5.608 | 5.395–5.680 | 6.354 | 0/5 | 3/5 |
| CPU | variable-10000 | 5.766 | 5.565–5.831 | 6.703 | 0/5 | 1/5 |
| CPU | variable-1000000 | 5.764 | 5.134–5.829 | 6.077 | 0/5 | 2/5 |

All five 10k/1M geometry pairs match per trajectory phase for each height/mode:
clipped cell count, row/column cardinalities, scale and normalized layout hash.
Each sample has at least 400 nonempty clipped cells, 50 rows and eight columns.
The pooled 1M/10k p95 ratios are GPU fixed 0.995/variable 0.990 and CPU fixed
0.974/variable 1.003. Those descriptive ratios indicate matched visible work;
they do not replace per-run limits. The GPU fixed-1M run with p99 8.617 ms is
preserved. Nineteen of twenty GPU p99s and seven of twenty CPU p99s meet 6 ms.

Per-run p95/p99 values are independently recomputed from all raw native logs in
archive member `raw-log-percentiles.json`. The 8.617124 ms observation is
`gpu/fixed-1000000/run-1/native.log` (and its result.json); each mode's summary.json
also retains every per-run distribution. [records.json](records.json) indexes all
members and hashes in [records.tar.gz](records.tar.gz).

The same release binary is used throughout: the prior verified serializer build,
whose compiled production leaves match committed bd672a6. No fresh rebuild is
claimed. Both checkouts were clean at the recorded case-admission and terminal
snapshots of that commit; source/binary identity is checked before every case. Stage0079 dependencies remain unchanged. Go 1.27.2
ARM64, Apple M4 Pro, macOS 26.6.2 (25G83), LG HDR DQHD 60 Hz, 162 logical DPI
(scale 1.6875), 1519x1266 client. Runtime GOMAXPROCS/GOMEMLIMIT/GOFLAGS are unset;
ordinary cases have no profiler, overlay, race detector or source-detail timing.
GPU mode runs first, then CPU; modes are not interleaved, so these observations
are not a controlled causal comparison of renderer speed or thermal state.

## Separate route and profile evidence

Four fresh source-detail cases each have 120 actual GPU captures, all accepted,
zero fallback/reuse. They establish route coverage for the workload family, not
every uninstrumented frame. Their timings are excluded from the ordinary 4,800
samples even when faster. Requested GPU mode by itself would not establish route.
The existing runner delivers initialized native scroll events through the actual
handler/owner/frame ticket. It does not exercise physical pointer delivery or
measure compositor presentation. Other input/pacing/display/human gates remain.

A separate 600-event Go CPU profile and eight-second native wall-stack sample
are diagnostic only. Go CPU has 6.57 s sampled over 20.12 s, including native wait:
cgocall 4.02 s, glfwWaitEvents 2.46 s cumulative; syncAccessibility 0.52 s and
ScenePacket 0.22 s cumulative. These overlapping entries cannot be added or
converted to additive critical-path frame costs. TryWriteBuffer's 0.80 s is not
a reason to repeat the already-rejected packed-upload change.

The native sample's exact-path assertion fails because Apple redacts tmp paths.
That raw failure remains in profile-result.json. A separate assessment matches
the Popen-owned PID 77008, sample header and exact Mach-O UUID
CC47E241-12BD-36D9-A610-4D45BC85B1F3 against the SHA-pinned executable. No sample is
rerun to erase the failure. Native stack siblings show 515 and 234 observations
in the AX-update main-queue block under 6,559 main-thread wall-stack observations.
These are wall-stack observations, not CPU accounting or additive UI-work time.
Source confirms Cocoa JSON parse/commit is dispatched asynchronously; it is not
synchronous Go ReconcileNS. The owner snapshot/cache/serialization and full scene
capture remain the synchronous investigation boundaries.

## Bounded engineering decision

The initial 90-minute M3 investigation starts at 19:56:03 UTC on 17 October and
requires reassessment by 21:26:03. It yields the adopted scoped serializer change,
this complete matrix and the following concrete next implementation experiment:

1. Build an owned segmented ready accessibility snapshot. Construct each eligible
   managed grid once into its final detached group; serialize through an ordered
   reference view and share that immutable group with its ready cache/baseline.
   Keep handlers and activation guards separate. This removes full-snapshot
   concatenation and the additional cache clone. It is a deeper ownership change
   than the rejected direct-append/stamping experiment a899d9b.
2. Initially admit ready leaf grids with provable completed-measurement coherence.
   Pending, inline-editor, custom, namespaced and uncertain cases take the complete
   old route before bridge/cache/generation mutation. Preserve custom callbacks
   exactly once. Keep per-grid storage independent so a pending sibling cannot
   retain a retired grid. Preserve nil/empty wire distinctions and capture guard
   maps before stripping callbacks from immutable cache groups.
3. Positively prove the eligible fixture uses the new route. Compare all wire
   fields/order, interleaved scopes, generations/actions, pending transitions,
   namespaces, detached retention and retirement against an independent old-path
   oracle. Run full driver race, vet and relevant native checks.
4. Compare five ordinary interleaved pairs on the unchanged native workload.
   Adoption requires at least 0.4 ms lower median run-p95 and four of five paired
   improvements, without correctness, retention or tail regression. This is an
   experiment-adoption bar, separate from the unchanged 4/6 ms acceptance limits.
   If it fails, reject it and stop small follow-up tuning of that candidate.
   A local benchmark/allocation gain is insufficient. Re-run the complete CPU/GPU
   matrix after an adopted material change; every required row must still qualify.

Internal review rejects choosing a new coverage-span shader from aggregate input
bytes: at this measured DPI the private bitmap glyph route returns before the
small-outline span cache. Any later resource/shader redesign first requires actual
current-route operation/resource counts and encoded-buffer composition. Naive
packet pooling is also excluded: public packet retention and the replay journal
require immutable ownership beyond a frame. Do not move work out of measurement
or change thresholds to obtain a pass. No completion estimate is issued while the
material performance experiment remains unresolved.

This is the next bounded implementation checkpoint, followed by an explicit
adopt/reject result before more optimization. M2 native/M3/M4/M5/M6/M7 remain open;
no final soak, rights/SDK or audit requirement is waived. No system preference,
input source, VoiceOver or clipboard changes occurred. The runner sees exit/EOF
and attempts group cleanup; the endpoint query is not strict per-PID/PGID proof.

## Checkpoint review

Internal review approves the complete raw result accounting and bounded M3-A
plan. One fresh Prism Anthropic Claude Sonnet 5.5 call covers all three staged
report/spec files, 17,099 B, one chunk, no truncation/cache. CLI exit 0 has zero
retained findings, one discarded medium and zero suppressed. The raw concern
requests discoverable per-run p99 evidence; its own arithmetic agrees with the
counts. Independent integer nearest-rank recomputation from all 40 raw logs
confirms every count and identifies the exact failed run. Explicit archive/index
and member pointers are added above. No source/target change or repeated provider
call follows. This documentation checkpoint does not alter the tested runtime.
