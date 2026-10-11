# macOS source cost checkpoint

Vellum now has an explicit native source diagnostic mode. Ordinary frame metrics
retain their call path. The source measurements locate the current fixed-grid
cost primarily in CPU scene rasterization, guiding the next bounded repaint
experiment. This checkpoint implements measurement, not a latency optimization
or completion of the Mac framework milestone.

## Implementation

Pinned stage 36 follows the committed empty-sprite baseline `381f031` and binds
six changed and eight new prepared files. Known scene/sprite sources time their
actual lazy raster and Copy calls. Scrim/insets time actual Copy calls; unknown
sources and drawers expose explicit unavailable breakdowns. Source kinds are
captured numeric metadata. The detailed composer draws every source once in its
original order and submits once even if diagnostic capacity is exceeded.

An opt-in GPUDrawer pointer collects actual existing-and-unchanged cache hits,
SetFromGoImage API attempts/time, and checked logical full-input 4WH RGBA bytes.
Begin/Read/End use separate mutex acquisitions; the originating renderContext
and native-window ownership serialize the batch. Deferred cleanup covers source
and drawer panic/Goexit. The backend Value API logs errors without returning them;
the instrumentation does not invent successful uploads or failure counts.

`NewFrameMetricsWithSourceDetails` accepts 1..4096 tickets. At most 32 scalar
records span each complete ticket, including intermediate composition before
its final armed frame. Records have composition/source ordinals and retain no
source/widget/image/user-text references. Overflow refuses diagnostic completion
without changing drawing. Snapshots deeply detach the nested records. Ordinary
`NewFrameMetrics` retains its 1..65536 bound and no per-source collection.

The loaded-grid flag `-perf-source-details` requires `-perf`; all record stages
agree on the mode. The runner checks scalar schema, bounds, order, nested
intervals, full-image attempt bytes and ticket sums. Normal warm-budget validation
refuses detailed records, independently of profiling. Diagnostic CLI results
always leave warm_target_met false. Raster/Copy/API costs never add to UI work.

## Validation

Fresh prepare and manifest/order tests pass with 6,137 prepared files. Focused
core/composer/GPUDrawer and public recorder race checks pass three times, including
unchanged/changed real paint pixels and copy order, actual gathered kinds,
default-path selection, unknown availability, shared prefix/final bounds,
cancellation/retirement, aggregate overflow, and detached snapshots. Source and
drawer panic/Goexit tests require the exact abnormal terminal, rather than merely
accepting an arbitrary exit.

Four mutation overlays produce real assertion failures when source drawing is
duplicated, timer cleanup is removed, the whole-ticket bound is disabled, or an
actual cache hit is omitted. They demonstrate regression sensitivity; the old
stage35 did not expose these diagnostic APIs. Full native adapter/commands and
loaded-grid offscreen race, native default/offscreen vet, loaded-grid vet, runner
and SDK packaging/check unit suites pass. These software tests do not measure
GPU transfer or physical display behavior.

Consumer author checks independently preserve 32 Python tests and 15 focused
Go race occurrences. A qualification wrapper initially refused setup because it
incorrectly required a root go.sum; no Go child ran. The first native reporting
helper then used `native` instead of `native_sample` after a passing ordinary
child. That attempt is retained and excluded from the final combined observation;
the corrected helper completed all four final native processes.

## Native observation

An optimized Go 1.27.2 arm64 executable ran on the local M4 Pro with macOS 26.6.2.
One ordinary default-mode process and three fresh detailed processes each ran
128 warmup interactions and 120 measured fixed-10k scrolls. All retained the
same 400 clipped cells, stored values and geometry. The ordinary samples omit
SourceDetails. Detailed samples are independent of ordinary warm qualification.

The 360 detailed tickets retain 361 compositions/source records, all scene kind.
One intermediate unchanged composition contributes an actual image-cache hit;
the 360 changed scene images each enter SetFromGoImage. Each input is 3072×1152,
or 14,155,776 logical RGBA bytes. The total 5,096,079,360 bytes describes attempted
inputs across all API calls, not successful or physical GPU transfers.

| Same-ticket/source observation | Pooled value |
| --- | ---: |
| CPU raster share of source Draw | 78.91% |
| Copy share of source Draw | 21.08% |
| Image API share of source Draw | 21.02% |
| Mean CPU raster per detailed ticket | 5.383 ms |
| Mean Copy per detailed ticket | 1.438 ms |
| Mean image API work per detailed ticket | 1.434 ms |
| Ordinary single-process UI-work p95 / p99 | 13.213 / 13.333 ms |
| Detailed three-process UI-work p95 / p99 | 13.136 / 13.602 ms |

Shares divide pooled nested interval sums by pooled source Draw, with all prefix
work retained. Percentiles of distinct phases are not additive. The one ordinary
probe is not a paired overhead/optimization comparison; elapsed detailed samples
include extra clocks and observation effects. Both unchanged 4ms p95/6ms p99
targets remain unmet. These fixed-10k observations do not meet the five-process,
four-case full P6 protocol or qualify native/GPU memory, cold caches, long strings,
tree latency, endurance or physical presentation.

This native fixture exercises scene rasterization, image API attempts and one
cache hit. It does not qualify detailed counters under actual sprite, uniform,
resident-texture or custom-source GPU workloads. Those paths have separately
scoped software/source proofs and remain broader native acceptance work.

## Evidence and review

Raw logs, original attempts, selected source/prepared maps, exact patch inputs,
inert producer/source copies and internal review certificates are preserved in
[records.tar.gz](records.tar.gz), with member hashes in
[records-index.json](records-index.json). Source snapshots use text suffixes;
the archive is evidence rather than another compiled package tree. The selected
bindings are not a complete compiler, module cache or SDK closure.

[Independent audit](audit.json) records the final checks and evidence bindings.
Internal source review approved the implementation after the unwind-test fix.
The significant implementation/test/spec/report batch receives one Prism review
using Anthropic Claude Sonnet 5.5; [review records](review/process.json) preserve
the exact payload and coverage. Public module namespace github.com/dshills/vellum
and root MIT license remain in place. No source push, release tag, signing,
notarization, OS input source, VoiceOver or preference change occurs here.
