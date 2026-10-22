# Phase-matched presentation / pending-feedback control

One new native-reference then framework pair, after preparation tests, internal
review, Prism (Anthropic Claude Sonnet 5.5) and commit. Existing Cogent/WebGPU/Metal
backend, library archive10e39953, Core, SDK/app, FIFO/vsync/two drawables,
nontransactional presentation, allocation/cache/render/pacing defaults and
33.3ms receipt-to-actual-shown gate remain unchanged. No feature or SDK API ships.
All changes are private consumer/compiler-overlay diagnostics.

Both use the locked match, opaque red/green rectangle on white and a view-linked
CADisplayLink observer. Native retains its cached-texture blit; framework retains
normal routing/reconcile/paint/upload and the existing submission observers.
Input APIs/algorithms and submission topology differ; subtraction is not pure
framework cost. Native gains command-buffer GPU/commit observations, without
changing command count/order or presentation policy. Shared observer source and
scalar recorder are identical in both. Observer overhead remains uncalibrated.

## Phase schedule and admission

Retain 248 routed requests:128warmup,120measured. Four fractions of reported
period repeat in sequence:1/8,3/8,5/8,7/8. That gives32warmups and30measured per
phase. Request anchors advance by8reported callbacks. Target is anchor.timestamp
plus3periods plus the selected fraction; anchor observation must precede target.
No request is generated from a fabricated timestamp lattice. Period/duration and
adjacent reported intervals must remain within0.1ms of60Hz; no gaps are discarded.
This deliberately changes diagnostic request spacing from120ms to approximately
8refresh periods while keeping response deadlines120ms and population fixed.
It does not alter renderer pacing or qualify changing-state60Hz cadence.

Both workers use the same bounded scalar anchor/wait code. Framework keeps its
Go deadline fields via a separately recorded Go/native clock-read bridge <=0.1ms;
actual phase admission uses the native receipt clock, not that approximate bridge.
Every actual receipt must be0..2ms after its target, preserving the chosen phase
bucket. Across implementations each phase's receipt-error p50 and p95 must differ
by<=1ms. These predeclared matching criteria determine capture comparability;
they neither waive nor replace the release gate. Reject the pair on mismatch;
retain every input, with no filtering, replacement, unchanged retry or new cutoff.
Retain80ms common held tail outside the cohort for timestamp coverage.

## Pending-state observations

Record all own-layer acquisition attempts from startup to shutdown, with bounded
scalar IDs and clocks: begin, acquired/unavailable, presented callback; record
receipts at the exact measured recorder path before FIFO publication. Keep
512attempts,4096events,8192ticks and at most8simultaneously pending/acquiring
states, with explicit overflow/refusal and no eviction. One mutex serializes
scalar state/census updates; no drawable, window, Go pointer or scene is retained.
All callbacks must drain; after closure only serialization allocates JSON.

Replay every transition and verify before/after pending IDs and acquiring counts.
Join each cohort receipt, actual acquisition/drawable identity and presented
feedback exactly. Account unmatched acquisitions only as retained pre-cohort
startup. No attempt may silently disappear. Preserve zero-time startup feedback
as non-displayed evidence under the unchanged parent rules. Retain nil/duplicate/
overflow/pending failures; none passes attribution. Both native GPU journals and
framework source/raster/upload/GPU joins remain complete through old validators.

A recorded pending drawable means acquisition awaiting its presented callback.
It does not measure Metal pool availability, retained object lifetime, available
credits, compositor queue depth or physical scanout. The receipt hook census is
sampled inside its recorded observation bracket; derive state at earlier native
receipt, present-request and GPU clocks separately from last observed transition.
Keep reported display timestamps and actual callback observations distinct.

## Bounds and evidence

One newly approved3minute quiet scope, max100seconds overall,40seconds per native
process, existing exclusive lease/owned retirement, full input/SDK/config/geometry
checks. Native then framework, stop first invalid, seal finally, no retry.
Report each process and30sample phase separately; no pooled samples, independent
quantile sums, cross-process clock joins, ordinary equivalence or causal fix claim.
Source/build/recipe/log/SDK and focused controls precede actual capture. Review
and commit actual results separately, including rejection. M3–M7 remain open;
no permanent display link, policy candidate, public release or gate change follows.
