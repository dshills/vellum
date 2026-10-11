# macOS raster row-range checkpoint

Vellum's native raster and private glyph-mask renderers now bound scanner row
index work to the accumulated path's row interval. This removes surface-height
scaling for small paths while preserving ordinary scanner compatibility. The
full four-case release comparison observes small, mixed native frame differences;
it does not establish a general frame-latency improvement. The unchanged 4ms
p95/6ms p99 targets still fail, and Mac framework completion remains open.

## Implementation

Pinned stage 37 follows `1be608c`. The two exclusively owned renderer callers
select NewScannerWithRowTracking. Its private half-open interval covers allocated
rows independently of clipping. Draw flushes the pending final cell before
narrowing traversal. Clear resets the complete bounding interval, including
clipped-out rows; SetBounds fully initializes row heads across shrinking,
capacity reuse and regrowth. Repeated Draw and clip changes retain accumulation.

General NewScanner keeps complete-row Draw/Clear, including directly supplied
exported cell lists. The opt-in constructor requires method-owned accumulation.
Winding, coverage, span order, colors, transforms and composition are unchanged.
Sparse distant contours can still span most of the surface. No partial image
upload, widget invalidation or new scheduler policy is introduced.

## Software qualification

Seven focused groups compare exact tracked span streams and saved state against
an independently copied baseline full-row Draw/Clear/SetBounds/alpha oracle.
They cover pending cells, clipping and clip widening, reverse insertion, disjoint
and seeded contours, both winding rules, zero/negative/offscreen bounds, repeated
drawing, Clear and shrink/regrow/reallocation. A separate default-constructor
case supplies exported cell lists directly. All groups pass under race three times.

The full scanner/raster package, native adapter, loaded grid/tree and replay race
suites pass, as do native adapter offscreen vet, prepare/order tests and check
script syntax. The complete-directory wrapper initially rejected 17 PNG files
created by the raster suite. All 587 selected sources and 6,138 prepared inputs
remained unchanged. Recovery retains actual logs and parent-attested zero exit
codes; it does not reconstruct missing elapsed times. These 17 PNGs also match
exact baseline production overlays byte for byte. Their existing tests use general
scanners, so they prove ordinary raster compatibility, separately from tracked
span equivalence.

Additional tests exercise actual owned Renderer.Render through 35 fill/stroke,
curve, transform, clipping and disjoint-path scenes at heights 64/1152/32/8192/64,
each drawn twice. Two independently initialized local glyph caches compare 64
cold and 128 warm/integer-translated real-font lookups per cache, including exact
alpha masks, placement and placed software RGBA output. These tests pass under
race three times. They leave the global cache untouched and do not qualify the
complete RenderText compositor or physical GPU/display output.

Fresh preparation reproduces all five pinned outputs and 6,139 prepared files;
prepare tests, the additive owned-pixel race tests and an optimized rebuild pass.
The earlier native campaign binds its recorded candidate binary. The final
prepared runtime files match that candidate's recorded inputs exactly; changes
are the added test and five patch metadata files, with the 17 generated PNGs
removed. The rebuild has a different binary hash despite equal Go build settings
and input/action build-ID components. Exact binary equality was refused and
retained as evidence. The rebuilt artifact is not independently native-measured;
this checkpoint links the results by unchanged runtime sources, not binary identity.

A missing pending-cell save produces explicit span/state assertion failures.
An incomplete Clear leaves invalid row heads and produces a detected runtime
index panic. The original negative heuristic called both assertion failures;
a separate correction preserves their actual terminal classifications.

The first registration wrapper assumed one manifest stage per patch number.
That attempt restored stage36 and its targeted command found no new tests; it
was interrupted and excluded. A later preparation endpoint refusal is retained
without inventing its uncaptured initial difference. Final checks require the
named new tests to actually execute and bind freshly reproduced patch outputs.

## Small-path and native observations

Three optimized software microbenchmark repetitions draw and clear the same
nine-row rectangle with the same span consumer, without image allocation.
Median ns/op at surface heights 64 / 1152 / 8192 are respectively
140.0 / 140.3 / 139.6 for tracking and 178.7 / 1004 / 6199 for the full-row oracle.
Both report zero allocations. This isolates scanner row work, not complete
raster rendering, upload or native latency.

One baseline CPU profile retains 240 fixed-10k samples but is excluded from
latency qualification. Its process-wide samples include substantial native and
runtime work. Source diagnostic elapsed intervals cannot be interpreted as
kernel-thread CPU accounting or identify a hot function by themselves.

The ordinary comparison uses optimized Go 1.27.2 arm64 builds on the local M4
Pro with macOS 26.6.2. Each build runs five fresh processes for every fixed/variable
10k/million-row case, with 128 warmups and 120 samples per process. The balanced
sequential order alternates which build runs first; all 40 processes and 4,800
samples pass actual native-input validation. No race/profile/source-details
instrumentation runs during measurement. Same-case baseline/candidate and
10k/million pairs have identical cell cardinalities, scale and layout hashes.

| Case | UI-work p95, baseline → candidate (ms) | UI-work p99 (ms) |
| --- | ---: | ---: |
| Fixed 10k | 13.023 → 12.857 | 13.224 → 13.084 |
| Variable 10k | 13.176 → 13.030 | 13.562 → 13.468 |
| Fixed million | 12.906 → 12.888 | 13.460 → 13.368 |
| Variable million | 13.048 → 13.087 | 13.359 → 13.389 |

Every case retains all outliers and individual-run variation. The variable-million
case increases slightly; its maximum increases from 13.561 to 14.386ms, and the
variable-10k maximum increases from 14.068 to 14.294ms. Other independent stages
can increase, and their percentiles are not additive. These small mixed changes
are not evidence of a general UI latency reduction. All individual runs fail the
unchanged targets. This completes the declared ordinary table measurement protocol
for this comparison, not P6 tree/native-memory/cold-cache/endurance acceptance,
physical GPU execution/presentation or the complete Mac milestone.

## Evidence and review

[records.tar.gz](records.tar.gz) preserves exact logs, historical refusals,
source/prepared maps, profiler and image observations, inert source/producer
copies and internal review certificates. [records-index.json](records-index.json)
binds its regular-file members. [audit.json](audit.json) records final linkage
and review scope. Binary hashes identify artifacts without claiming full
compiler, module-cache or SDK closure.

Internal reviews qualify source contracts, test integration and native statistics
separately. The significant source/test/spec/report batch receives one Prism
review using Anthropic Claude Sonnet 5.5; [review records](review/process.json)
preserve its payload and findings. Public namespace github.com/dshills/vellum and
MIT licensing remain in place. Broader Mac performance and acceptance work continues.
