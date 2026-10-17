# Scene composition and GPU tile-culling decision

17 October 2026. Base `31afa90` uses the `bd672a6` renderer/loaded-grid runtime
and unchanged stage0079 prepared dependencies (6,260 files). This checkpoint
changes documentation only. All experiment code and tests use private Go overlays;
no checkout or prepared production source is changed. Tile culling is rejected.

## Decision

The fixed adoption bar requires at least 0.4 ms lower median run-p95 and an
improvement in at least four of five matched pairs without correctness, retention
or tail regression. Only **one of five pairs improves**. Median run-p95 increases
**5.002915 → 5.028750 ms** (+0.025835 ms, +0.52%). Median of per-run nearest-rank
p50 process allocation increases **5,300,696 → 5,334,584 bytes** (+0.64%). Reject
the entire candidate and stop local tile-culling tuning. Smaller encoded buffers
do not establish lower total frame latency.

| Pair | Baseline p95 ms | Candidate p95 ms | Baseline p99 ms | Candidate p99 ms |
| --- | ---: | ---: | ---: | ---: |
| 1 | 5.062914 | 5.025666 | 5.239833 | 5.634670 |
| 2 | 4.996790 | 5.114209 | 5.804749 | 5.654250 |
| 3 | 5.001708 | 5.037040 | 5.579209 | 5.716417 |
| 4 | 5.007585 | 5.028750 | 5.328625 | 5.509959 |
| 5 | 5.002915 | 5.020293 | 6.072335 | 5.886709 |

All ten ordinary runs fail the unchanged **4 ms p95** target. Four baseline and
all five candidate runs pass **6 ms p99**. Baseline pair 5 p99 is 6.072335 ms
and fails that target. No table or full Mac gate closes. The narrow adoption comparison is
five ordinary release pairs, alternating order, 128 warmup operations and 120
samples each, fixed 10k rows, matched visible geometry with at least 400 nonempty
clipped cells, GPU scene rendering requested and normal runtime settings.
No race, profiler, source-details instrumentation or GOMAXPROCS/GOMEMLIMIT/GOFLAGS
override is used. The already verified serializer release baseline is reused;
loaded-grid runtime inputs are unchanged through `31afa90`.

A separate 120-sample source-details candidate run positively observes GPU
acceptance without fallback/reuse and actual tile-reference reduction. Median
references fall **55,388 → 19,405**; checked input is **416,056 bytes** (four
buffers plus 16-byte config). Each packet's resource and input byte ledger is
checked exactly. These are diagnostics, never pooled into ordinary timing.

## Measured composition before the experiment

Four separate diagnostic cases cover fixed/variable 10k/1M rows with 120 measured
packets each. Every packet correlates one-to-one with an accepted GPU source,
without fallback/reuse; buffer and data-resource ledgers reconcile exactly.
The surface is 1519×1266 pixels, 7,600 tiles, scale 1.6875 (162 DPI).

| Case | Uniform / Mask / Pixels / Spans, p50 counts | Distinct masks, every packet | Tile refs p50 | Checked input p50 bytes |
| --- | --- | ---: | ---: | ---: |
| Fixed 10k | 64 / 2452 / 0 / 207 | 83 | 55,388 | 559,988 |
| Fixed 1M | 64 / 2452 / 0 / 207 | 83 | 55,386 | 559,900 |
| Variable 10k | 72 / 2495 / 0 / 207 | 83 | 55,434 | 563,580 |
| Variable 1M | 72 / 2495 / 0 / 207 | 83 | 55,434 | 563,448 |

Per-packet mask references reuse **83 distinct packet-local mask resources**,
accounting for **12,488 mask-data bytes**, roughly 2.2% of complete input. Ops and
Refs jointly account for 70.70–70.88% of mean input. No Pixels operation occurs
in the 480 measured packets. Span-data p50 is 90.7–90.8 KB; its producers have not
been classified, so it is not attributed entirely to glyphs. Deduplication here
is within one packet, not proof of persistent resource/heap identity.

10k/1M cases match per-step kind counts and Ops/Rows sizes. Refs/Data can differ
slightly; these observations do not explain the differences or replace complete
scaling qualification. Table entries are marginal nearest-rank p50 values and
do not describe one common frame. Exact totals are checked per frame. All four
runs use instrumented diagnostic builds; their timings cannot qualify performance.

## Candidate and verification

The private candidate records the last original operation ordinal that replaces
each complete actual tile: unmasked Uniform Src, including transparent Src, or
opaque Uniform Over. GPU encoding keeps the ordered suffix beginning there.
Partial coverage, translucent Over and masked operations cannot reset a tile.
Final partial surface tiles use their real bounds. Original operation, data,
reference and byte admission charges remain in force before culling. Legacy
Encode, immutable Packet, CPUReplay, resources, shader and transport stay unchanged.

The frozen original encoder remains an independent admission/stream oracle.
Tests decode its output and independently prune each tile backwards, compare
painter order and exact Ops/Data, then compare encoded GPU pixels with CPUReplay.
Coverage includes partial edge tiles, transparent Src, masks, overlapping text,
translucency, rejection boundaries and detached results. The initial focused
failure was a stale exact-reference expectation; it and its corrected full run
are preserved separately. Full sceneplan/rasterx/gpuscene race checks pass
(20.051 seconds), as do vet, diagnostic-decoder checks and both release builds.

Actual native GPU tests pass (2.652 seconds), including culling and retained
partial-batch readback against CPUReplay, unchanged five writes/exact input bytes
and one submit. The strengthened retention fixture uses alpha 32768, requires
!Independent and proves retained-target replay differs from fresh replay. An
earlier passing but weak alpha-65534 fixture is preserved, not used as the final
retention proof. No native GPU tests are skipped. These assertions do not prove
physical input, presentation or sustained/native memory budgets.

Full rasterx tests generated 17 new PNG outputs in the ignored prepared tree.
Native plan admission refused that drift before launch. The outputs were
preserved outside the prepared tree, every original file was proved identical
to the canonical checkout, and only the generated leaves were removed. Both
prepared trees again hash to stage0079 before and after the comparison. No
production code changed. Internal source, admission and result reviews are
recorded; the rejected implementation is evidence only. One completed Prism
Anthropic Claude Sonnet 5.5 call reviews this report and the two completion specs.
Its two discarded medium comments were inspected; explicit p99 wording and a
per-case distinct-mask column clarify the already reconciled results.

## Evidence and remaining work

[records.json](records.json) hashes every member of [records.tar.gz](records.tar.gz).
`composition/assessment.json` and `composition/native/*/{native.log,composition.json}`
retain all 480 original diagnostic observations. `culling/assessment.json`
independently recomputes all 1,200 ordinary samples' per-run p95/p99 and allocation
p50 using integer nearest rank; `culling/native/{pair}-{baseline,candidate}/native.log`
contains the raw runs. `culling/diagnostic/` is separate. Source overlays, original
source, commands/stdout/stderr, pins, reviews and failed checks are also retained.
No binary or SDK is duplicated in this checkpoint.

The fresh live lease, source/binary pins and stop scope were held. All comparison
app invocations exited zero and the runner attempted owned-group cleanup; this
does not establish strict per-PID absence. No preferences, input sources or
VoiceOver settings changed. Cross-display, physical pointer, human input and
screen-reader workflows remain open. The failed placement operator stays sealed;
its next verification is direct/manual, with no operator retry.

M2 native, M3, M4, M5, M6 rights/final SDK and M7 remain open. Continue the
[Mac delivery plan](../../../specs/macos-delivery.md) with one bounded retained
managed-row paint-fragment prototype shared by CPU/GPU paths. It must avoid
repeated glyph/path capture through actual-input proof, preserve clipping,
fractional placement, original admission and invalidation, and pass the same
material full-frame adoption bar on both paths. Reassess feasibility after
90 minutes; no further AX/packed-upload/tile-culling microtuning is queued.
The prototype is not implemented or qualified by this checkpoint.
