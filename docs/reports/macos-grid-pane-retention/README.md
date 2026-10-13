# Mac grid pane retention

Combined row/column scrolling could recycle a departing frozen-cell widget into a
scrolling pane. That parent change forced a full scene layout, including unrelated
siblings. Retention now reserves same-pane donors for every hole before using the
stable old-order fallback. Exact domain-key matches retain priority. The admitted
widget pool, native capacity limits and full-layout safety guards are unchanged.

## Functional acceptance

The new regression fails against the original algorithm: each combined-scroll
direction moves four donors across panes and sizes an unrelated sibling. The final
two-pass implementation passes three race repetitions at 96/120/144/192 DPI. The
tests check actual widget parents, overlapping domain-key identity, local-layout
admission, pane child order and exact pixels against the ordinary full-layout
oracle. Growth, donor shortage, true pane population changes, stable fallback,
ordered retirement and duplicate-key rejection are also checked.

Default and offscreen driver race suites and the loaded-grid offscreen race suite
pass. The full adapter commands originally exit 1 because the sandbox denied the
workbench package's temporary Unix sockets. Both affected workbench packages pass
authorized unsandboxed race retries; original failure receipts remain unchanged.
Default/offscreen vet passes. Four separate real Cocoa grid AX runs cover fixed and
variable rows at 10,000 and 1,000,000 records: bounded 400-cell semantics, geometry,
overlapping object identity, replacement retirement and native press admission.
These queries do not qualify human VoiceOver or physical pointer input.

## Native performance comparison

Each of four serial campaigns uses five fresh optimized processes per fixed/
variable 10k/million case, 128 warmup callbacks and 120 measured frames per process.
Together they record 80 processes and 9,600 measured frames. CPU scene and opt-in
GPU scene modes remain separate. Every paired measured frame has the same actually
clipped workload, layout identity, materialized/retained pool and source callback
counts before and after the fix. Fixed rows show 400 cells and 50 rows; variable
rows show 400/408 cells and 50/51 rows, always eight columns. Stored values are
verified by the production fixture. Selected source and binary hashes are bound
to each campaign; the selected source list is not a complete compiler closure.

The following pooled elapsed UI-work measurements are milliseconds. Per-run
distributions and raw native tickets remain in the archive.

| Mode / case | Mean before | Mean after | p95 before | p95 after | p99 before | p99 after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| cpu / fixed-10000 | 7.279 | 7.560 | 10.360 | 8.614 | 10.939 | 9.427 |
| cpu / fixed-1000000 | 7.117 | 7.413 | 9.981 | 8.472 | 10.184 | 8.744 |
| cpu / variable-10000 | 7.037 | 7.342 | 10.275 | 8.661 | 11.127 | 9.184 |
| cpu / variable-1000000 | 6.853 | 7.203 | 10.093 | 8.515 | 10.330 | 8.904 |
| gpu / fixed-10000 | 7.284 | 7.629 | 10.010 | 8.702 | 10.786 | 9.424 |
| gpu / fixed-1000000 | 7.071 | 7.357 | 9.788 | 8.470 | 10.337 | 8.925 |
| gpu / variable-10000 | 7.058 | 7.443 | 10.180 | 8.821 | 10.951 | 9.584 |
| gpu / variable-1000000 | 6.796 | 7.204 | 9.902 | 8.528 | 10.166 | 8.821 |

Tail latency decreases in this comparison, while average frame work increases.
This is a measured tradeoff on one host across serial runs, not a controlled
estimate of a causal speedup. The unchanged p95 <=4 ms and p99 <=6 ms targets still
fail in every final process. This checkpoint fixes pane ownership and preserves
local layout; it does not pass P6 performance acceptance. UI work excludes acquire
wait and queued delay, and does not measure native/GPU CPU time or compositor
presentation. No percentile sums are used.

A separate 2,048-frame process-wide CPU profile is diagnostic only. Its CGO/wait
attribution cannot identify precise native CPU costs. A geometry-publication cache
experiment produced no persuasive full-frame gain and was removed before the
final implementation. Historical source and receipts are archived with that scope;
they are excluded from final source qualification and Prism review.

## Review and evidence

Internal source review approves the exact three implementation/test/spec files.
Prism ran once with Anthropic Claude Sonnet 5.5: complete three-file coverage,
8,110 input and 1,278 output tokens, raw exit 0 and no findings. Report/evidence
validation is a separate internal review, without another provider call.

[Checkpoint](records/checkpoint.json) binds the [archive](records/evidence.tar.gz)
and [member index](records/evidence-index.json). The archive preserves raw exits,
stdout/stderr, source manifests, profile, review and comparison evidence. Executable
binaries and raw serial/UUID-bearing hardware metadata are excluded; redacted
metadata retains the raw file hash. All 6,174 prepared dependency files remain
byte-identical. Checks ran within the authorized fifteen-hour quiet window and
changed no OS preferences or general clipboard contents. Wider Mac qualification,
hardware coverage, performance/endurance and public distribution remain open.
