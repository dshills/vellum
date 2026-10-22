# Existing-backend display scheduling: off/on result

Both cases of the single approved control pass complete accounting. The existing
Cogent/WebGPU renderer, allocation/cache policy, FIFO/vsync, two drawables and
nontransactional presentation stay unchanged. No production fix is adopted.

| Display-clock observer | Receipt-to-shown p50 | p95 | p99 | Maximum | Above 33.3 ms |
| --- | ---: | ---: | ---: | ---: | ---: |
| Off | 43.409 ms | 51.101 ms | 53.018 ms | 53.094 ms | 120/120 |
| On | 26.817 ms | 33.947 ms | 36.504 ms | 36.576 ms | 14/120 |

The on-minus-off difference of run p95 values is
**-17.154 ms**, approximately one refresh period.
This is one ordered off/on pair, not a distribution of paired input differences,
a repeated estimate or proof of causation. Process/order and temporal variation
remain possible explanations. Earlier captures already showed more than one
latency band. **The display-clock observer cannot be assumed neutral.** Its
registration/wake behavior is now a concrete scheduling hypothesis to control,
not an accepted repair. Both cases still fail the unchanged 33.3 ms p95 number.

## Joined display-clock observations

The on case retains all 1,789 callbacks, including startup and the common
held tail. All 1,788 adjacent reported timestamp intervals are approximately
16.667 ms; none is discarded for lateness or a gap. Across all callbacks, host
observation minus reported timestamp has p95
10.902 ms, while target minus timestamp is approximately
16.667 ms. Callback observation is not GPU completion, latch time or scanout.

The 120 measured actual presentations lie within
10.542 microseconds of the nearest reported display timestamp.
There are 107 measured responses with one reported timestamp strictly between
final-present GPU end and shown, and 13 with none. These are retrospective
reported-event counts, not missed-vsync counts or guaranteed latch opportunities.
No timestamp lattice is fitted to presentation feedback.

Each process has its own p95-ranked input. Off selects accepted revision
243, scope 1, opaque tag 488, generation
246, drawable 252; its final-present GPU end to shown is
46.831 ms. On selects revision 199, scope 1, tag
400, generation 202, drawable 208; that interval is
29.497 ms. These are actual individual inputs, not a phase-percentile sum
or a decomposition of the difference between run p95 values. No compositor-only,
driver-only, fixed-frame or WebGPU-specific cause is established.

## Checks and scope

Execution takes **78.604 seconds** overall, within the 100-second bound.
Native cases take 31.384 and 30.732 seconds, below their 40-second bounds.
All 18 owned native/probe groups retire; native stderr is empty. Each case retains
128 warmup and 120 measured accepted inputs with complete source/raster/upload/
GPU/drawable joins. Each also retains 50 startup commands and ten startup presents.
Off records zero display callbacks; on drains/invalidate/releases cleanly. Both
cases use the same binary/archive and identical 50 ms held tails outside the cohort.
Configuration, persistent geometry, all source/binary/module/overlay/native pins,
raw logs and all 8,321 SDK files verify unchanged within and across cases.

Preparation passed eight focused classifier methods with 20 negative cases,
exact timestamp-boundary and late-callback/gap controls, the original routed
complete-pixel fixture, native recorder/ownership tests and fresh consumer builds.
The native JSON Boolean issue was fixed before desktop use; a counterfactual
with the original implementation is refused by the final native fixture with an
ordinary exit and no window or crash. The exact reviewed native archive/Core and
source-detail overlays are reused; no Rust or SDK rebuild occurs.

Prism using Anthropic Claude Sonnet 5.5 reviewed all 15 preparation files and ten
focused follow-up files. Boundary semantics are tested; source-backed dispositions
retain strict focus identity, unconditional pre-start cleanup and the fixed-on
CPU/GPU observer CLI. Complete review records and the unused initial revision
remain in indexed preparation evidence. Source checkpoint is
`633d5b9`. The [protocol](../../../tools/macos-calibration/displayclock/PLAN.md),
[source audit](source-audit.json), [preparation checks](preparation-checks.json) and
[result summary](result-summary.json) retain the basis.

The next focused control should reverse observer order and test whether the
latency band follows observer presence before proposing any scheduling change.
Do not adopt a permanent display link or treat these instrumented results as
shipping performance. This sparse routed fixture does not establish sustained
60 Hz cadence, resize/workflow acceptance, physical input, displayed pixels or
whole-GPU allocation balance. Features, SDK/app, defaults and the release gate
are unchanged; M3–M7 remain open. The approved desktop scope is consumed, the
desktop is free and no native job remains active.
