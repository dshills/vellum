# macOS grid input and workload qualification

Stored-table warmup now refuses nonzero starting offsets, incomplete closed
trajectories, extra callbacks and changed completed geometry/scale before ordinary
measurements. A separate bounded scalar ledger makes admission, delivered wheel
arithmetic and owner transitions inspectable. This advances reliable Mac
performance qualification; it does not repair an established framework input
ordering defect or claim a rendering speedup.

## Implementation

The opt-in `NewFrameMetricsWithGridInputDetails` constructor preallocates
`3*tickets+256` records for at most4096 frame tickets. Ordinary metrics allocate
no ledger or additional input clocks. Bindings copy identifiers up to4096 bytes.
Records contain no model, widget, event, callback, cell values or native pointer.
Exhaustion/oversized identifiers explicitly refuse diagnostic evidence; normal
grid input continues with its existing behavior.

Admissions copy logical requests, encoded physical deltas, completed scale and
owner state. Actual native wheel handlers copy their current tag, handler scale,
integer steps and signed remainder. Posted owner transitions reuse existing state
reads and link the originating wheel sequence; scrollbar and layout paths remain
distinct. Native style/remainder reads never move into a peer-window owner callback.
An independent ledger mutex releases before posting and callouts. Admission
appends happen after the metrics lock releases; tags/ingress sequences establish
correlation, and append order is not FIFO or clock order. Tag zero means untraced,
not desktop/user origin. A ledger snapshot is not a native barrier.

`loadedgrid -perf-input-details=/new/path.json` saves the detached ledger after
native Run retires, including failed warmups. Files are exclusive and capped at
64MiB; overflow produces nonzero finalization. This diagnostic mode cannot combine
with CPU/source profiling. Stage/sample flags agree and ordinary validation
rejects diagnostic samples. Ordinary ready records now retain warm start/end
viewports and generations plus callback count under `warmup_guard: 1`.
See [the contract](../../../specs/grid-performance.md).

## Verification

Ten final source-bound checks pass: full offscreen/native driver race and vet,
three new input-ledger groups repeated three times, loaded-grid race/vet,
34 Python record/process/statistical tests, and grid/replay consumer race.
The input tests exercise actual untraced fractional wheel routing followed by a
reserved tagged FIFO event, detached concurrent snapshots, capacity refusal and
identifier bounds. Consumer tests refuse warm offset/layout/callback drift and
preserve exclusive failed-run evidence. The fractional fixture accounts for
float32 encoding precision and excludes the prior initial-layout notification.

Three faulty Go overlays fail explicit assertions: bypassing the closed warmup
guard, dropping the delivered wheel tag, and aliasing ledger snapshots. These
are behavioral failures, not build errors or panics. Final endpoints bind647
selected runtime/specification files and6139 prepared dependency files exactly.
Selected source/prepared identity is not a complete compiler/cache/SDK closure.

One initial release diagnostic and ten further fresh fixed-table diagnostic
processes (five10k/five million) each retain745 records within capacity1006:
248 admission/wheel/owner triples and one untraced startup layout transition.
All tags correlate; admission and handler scales/deltas agree; wheel remainders
are zero; warmup returns to zero offsets with128 callbacks and generation2→130.
At least400 cells are actually clipped after completion; recorder loss/pending
state is zero. These are diagnostic provenance assertions, not latency evidence.

The [preceding campaign](../macos-grid-publication/README.md) remains refused.
Its measured offset oscillation fits a prior signed wheel remainder, but its
warmup lacks delivered-scale and event-provenance records. The new diagnostic
runs do not reproduce that discrepancy, so its cause remains unproven. No
historical failure is relabeled and no viewport reset hides it.

## Ordinary performance

After regression/diagnostic children ended, the same optimized release artifact
ran with input/source diagnostics and CPU profiling disabled. Four fixed/variable
10k/million cases each use five fresh processes,128 unmeasured warmups and120
measured samples. All20 child exits and record validations pass. Every warmup
closes, and paired10k/million sample cardinalities/scales/normalized clips match.
All2400 observations, outliers and per-run statistics remain in the evidence.

Nearest-rank pooled elapsed UI-work milliseconds are:

| Case | p50 | p95 | p99 | Max |
| --- | ---: | ---: | ---: | ---: |
| fixed-10000 | 12.644 | 13.479 | 14.533 | 66.204 |
| fixed-1000000 | 12.334 | 13.274 | 13.880 | 24.570 |
| variable-10000 | 12.818 | 13.682 | 14.400 | 18.444 |
| variable-1000000 | 12.509 | 13.367 | 13.906 | 14.765 |

Every run still exceeds the unchanged4ms p95 /6ms p99 targets. The maximum is
retained and its cause is unclassified. No A/B speedup is inferred: this is a
current guarded workload measurement, separate from the preceding refused
artifact/campaign. Phase percentiles are not additive. CPU-side measured work
excludes queue delay/acquire/present-call waits; GPU execution and compositor
presentation remain unavailable. Process-wide Go allocations, retained pool
counts and end-only RSS/CPU observations do not attribute native/GPU memory.

Current release binary SHA256: `f8feeacb6d653a9d1042f76a3effc2314b531de8e5224581bdb6a1df7d918ecf`.
Mac completion, sustained table/tree P6, font/native/GPU memory, endurance,
human screen-reader/input workflows and broader professional acceptance remain
open. The next optimization must reduce actual rendering/owner work while
preserving this matched workload and its unchanged targets.

## Evidence and review

[records.tar.gz](records.tar.gz) contains final check/native/diagnostic/negative/
ordinary receipts, raw records, source snapshots, diagnostic historical evidence
and inert producer copies. [records-index.json](records-index.json) binds archive
membership; [audit.json](audit.json) binds current source/prepared inputs and
review/evidence linkage. Binaries are identified by hashes, not stored or claimed
byte-identical to a second canonical build.

Independent source review corrected sample-mode validation and diagnostic ID
retention before the final gate. Receipt reviews separately audit diagnostic and
ordinary workload evidence. One significant source/test/spec/report batch is
reviewed with Prism using Anthropic Claude Sonnet5.5; [raw review records](review/process.json)
preserve coverage, findings and process status, with local triage if needed.
No public release, Git push/tag or macOS preference change is part of this checkpoint.
