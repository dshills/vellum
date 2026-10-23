# Reversed-order display-clock observer result

The fresh observer-on then observer-off control passes complete accounting. It
uses the exact executable and native archive from the [previous off/on pair](../macos-display-scheduling/README.md).
No production scheduling change or performance qualification follows.

| Pair | Execution order | Off p95 | On p95 | On minus off p95 |
| --- | --- | ---: | ---: | ---: |
| Previous | off → on | 51.101 ms | 33.947 ms | -17.154 ms |
| Reversed | on → off | 37.923 ms | 37.882 ms | -0.040 ms |

These are separate differences of run quantiles, calculated by observer mode.
They are not percentiles of paired input differences or calibrated observer cost.
The earlier approximately one-refresh-period advantage did not reproduce: this
reversed pair differs by only 0.040 ms at p95. Both reversed cases exceed the
unchanged 33.3 ms gate. This does not establish equivalence or observer neutrality.
Neither samples nor process clocks are pooled across runs. One reversed pair
can challenge a simple order explanation; it cannot establish causation,
observer neutrality or ordinary-build performance. Earlier captures show multiple
latency bands. Both the prior evidence and this complete attempt remain retained.

| Reversed case | p50 | p95 | p99 | Maximum | Above 33.3 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| On | 28.161 ms | 37.882 ms | 39.377 ms | 41.288 ms | 31/120 |
| Off | 30.896 ms | 37.923 ms | 39.415 ms | 39.700 ms | 41/120 |

The on case retains all 1,789 callbacks, including startup and
the common 50 ms held tail outside the cohort. Timestamp interval p95 is
16.666667 ms; observation minus reported
timestamp p95 is 16.328 ms.
Actual shown feedback is within 89.583
microseconds of the nearest reported timestamp over the 120 measured inputs.
Counts of reported timestamps strictly between final-present GPU end and shown
are {"0": 9, "1": 110, "2": 1}.
These retrospective clock relations are not optical scanout, guaranteed latch
opportunities, missed-vsync counts or proof of prior callback delivery. Late
observations and gaps are retained. Off records zero callbacks.

Each case has its own actual p95-ranked input: on revision
167, scope 1,
tag 336, generation 170,
drawable 176, with final-present GPU end to shown
30.812 ms; off revision
147, scope 1,
tag 296, generation 150,
drawable 156, with that interval
31.070 ms. These individual-input
intervals do not decompose the difference of run p95 values and do not isolate
compositor, driver or WebGPU cost.

Execution takes 74.272 seconds within the 100-second bound.
Native on/off cases take 31.275/30.640 seconds,
each below 40 seconds. All 18 owned groups retire. Native
stderr is empty. The unchanged strict classifier independently verifies both
journals: each has 128 warmup plus 120 measured inputs, 248 full joins, 50 startup
commands and ten startup presents. Configuration, persistent geometry, source,
binary, native archive, modules and all 8,321 SDK files verify unchanged.
Preparation checkpoint is `2b498f5`. Five focused orchestration
methods/seven failure subcases and actual inherited Host/executor guard tests
pass; native ownership, pixel and classifier evidence is reused by exact identity.
Prism using Anthropic Claude Sonnet 5.5 reviewed preparation with source-backed
dispositions for two inherited-guard concerns. The separate four-file result review completed without findings, exclusions or
truncation; see the [review receipt](result-review-completion.json).

The original Cogent/WebGPU/Metal design, features, SDK/app, allocation/cache,
FIFO/vsync, two drawables, nontransactional presentation and pacing defaults stay
unchanged. The gate remains 33.3 ms receipt-to-actual-shown p95, with separate
cadence/resize/workflow qualification still required. No permanent display link
is adopted and M3–M7 remain open. The next bounded step is to inspect the existing journals for
receipt/render/submission phase relative to actual callback delivery and reported
clocks before proposing another native control. This is a sparse instrumented fixture; it does
not establish sustained 60 Hz responsiveness, physical input, displayed pixels,
resize acceptance or whole-GPU allocation balance. The approved scope is consumed,
sealed and not retried. The desktop is free; no native job remains active.

See the [protocol](../../../tools/macos-calibration/displayclock/REVERSED.md),
[preparation checks](preparation-checks.json), [guard audit](host-guard-audit.json),
[result summary](result-summary.json) and indexed preparation/result evidence.
