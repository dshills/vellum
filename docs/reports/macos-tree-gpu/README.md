# Stored-tree renderer investigation

The stored-tree consumer now exposes `-perf-gpu-scene` and
`-perf-source-details` without source overlays. Both require `-perf`; every JSON
event records the requested GPU mode and diagnostic mode. GPU mode permits the
existing fallback. Source details add clocks/counters and cannot qualify ordinary
latency. Detailed capacity is checked before stored-data/UI construction, including
warmup and both first/reverse scrolls. Defaults remain CPU/ordinary.

Consumer race/vet/build checks pass. Five native preflights cover CPU, GPU, both
source-detail combinations, and deep-chain GPU navigation. All 600 measured events
pass sequence/generation/callback/geometry guards: 360 ordinary and 240 diagnostic
samples are kept separate. Four invalid CLI/capacity cases refuse as expected.
The first sandboxed native flag check timed out during initialization; its owned
process was terminated. The admitted native repeat passes. This failed attempt
is preserved and does not qualify the flags.

Separate source diagnostics observe one full 3072×1152 main-scene source per
frame. In the final CPU diagnostic, mean raster work is 3.407 ms and mean image
copy work 1.796 ms, with one SetFromGoImage attempt and 14,155,776 logical input
RGBA bytes per frame. In the GPU diagnostic all 120 frames report CPU raster/copy
unavailable and zero image API uploads; mean source draw work is 3.393 ms. These
are elapsed host/API observations, not physical upload bytes, GPU execution or
presentation. A separate temporary upload counter also observes exactly one
attempt in each of 120 scroll intervals, rejecting the duplicate-upload hypothesis.

A separately profiled CPU run completes 2048 guarded scrolls. Substantial sampled
Go stacks fall under cgo/WriteTexture, while raster/text paths also appear. Cgo
sampling does not attribute native CPU execution reliably; its cumulative profile
values must not be added to recorded UI work or inferred as extra uploads.
Additional-clock, counter and profiled runs are excluded from ordinary timing.
See the [native mode summary](records/native-summary.json) for final preflights.

A scanline compaction prototype was rejected. It seals a stream's spans into one
exact allocation with per-row capacity fences instead of one allocation per
nonempty row. Source review finds ownership/order/admission preserved; the
allocation regression fails on the original with 40 allocations and passes on
the prototype with a limit of eight. Existing independent builder/scanner tests
pass, preserving encoded bytes and software pixels. These checks do not establish
native GPU pixel qualification for the prototype.

The paired ordinary campaign has 20 sequential release processes: five per
build/view, alternating build order, each with 128 closed guarded warmups and
240 native scrolls on sealed 100k-node wide/deep trees. Both artifacts use the
optional GPU renderer and identical consumer bytes; only compaction differs.
All 4800 measured events pass independent validation with identical phase
geometry and 71 clipped rows. Every source/prepared/overlay endpoint remains
stable. These measurements precede the shipping CLI-only consumer change.

Mean of five per-run measurements; Go allocation includes observers and other
process goroutines. Negative changes mean smaller values.

| View | Original / packed p95 | Allocation objects change | Allocated bytes change |
| --- | ---: | ---: | ---: |
| Wide | 6.833 / 6.838 ms | −13.08% | +0.21% |
| Deep | 7.172 / 7.398 ms | −12.25% | +0.37% |

Deep p95 is 3.14% worse in this campaign. Fewer allocation objects did not produce
a useful byte or latency gain, so no production renderer patch was adopted.
The [comparison summary](records/comparison-summary.json) retains per-run p95,
p99, mean, maxima and allocation values. No outlier or failed target is removed.
This is a scoped GPU-mode comparison, not a repeated CPU/GPU comparison proving
a default renderer decision.

Internal reviews approve the bounded consumer modes and prototype source
invariants. Prism reviews the shipping source/spec/report checkpoint with
Anthropic Claude Sonnet 5.5; its raw result and disposition accompany the evidence.
The archive retains prototypes, exact build/overlay/consumer fingerprints,
raw logs/receipts, original failing regression, passing tests, profile and
producer scripts. Executables are excluded. All 6204 prepared files are unchanged.
An unused patch-adoption script remains historical only; it was never executed.
An initial summary script syntax error is preserved beside its corrected version;
only the successful final validator supplies the summary.

The unchanged 4 ms tree target fails in all ten packed-prototype runs and is not
qualified by the five mode preflights. Full tree interaction, cold/cache and
maximum-row behavior, complete P6/P11, native/GPU memory, physical display,
overnight soak, human accessibility and Mac distribution remain open.
