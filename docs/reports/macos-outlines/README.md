# Renderer-owned static outline reuse

Repeated static outline construction now reuses privately owned font-unit
segments in both ordinary CPU rendering and the opt-in GPU scene compiler. The
matched full stored-table campaign observes about 7.1–7.4 MB less process-wide Go
allocation per event and lower pooled p95 in every case. All builds still miss
the unchanged p95 ≤4 ms / p99 ≤6 ms targets. This is a verified optimization,
not full Mac/performance/memory acceptance or permission to enable GPU scenes by
default.

## Implementation and ownership

Owned Cogent stage 43 adds a renderer-private LRU shared with that renderer's
short-lived scene/scroll captures. Immutable Font identity plus full GID keys raw
font-unit outlines; different Font instances do not share entries. Size, DPI,
paint, subpixel placement and transforms do not change this geometry. Variable,
CFF and encoded bitmap/SVG resources retain their existing routes.

Only the bounded getter's successful newly owned segments enter the cache, and
Rasterx consumes them read-only. Ordinary public font methods still return their
own mutable output; Sideways/public mutations cannot reach cache storage. The
CPU path uses original retrieval on bounded refusal or unsupported resources.
Cache retention refusal alone never rejects an otherwise successful GPU outline.

Typesetting stage 4 supplies an opaque cost token measured from the existing
bounded getter's successful aggregate point/operation charges. A warm hit replays
that full logical cost before consuming segments, including recursive composite
traversal, phantom points, copies and segment bounds. Per-glyph/segment/finite
coordinate/extent-product checks remain unchanged. Cache warmth cannot bypass a
shared budget or resume a refused batch; partial consumption after terminal
refusal need not match a fresh traversal.

Each renderer retains at most 2 MiB of segment backing capacity, 1,024 entries,
eight Font references and 64 KiB of segment capacity per entry. These are payload
and reference bounds; map/list/allocator metadata, transient retrieval,
Font/backing storage and the existing glyph-mask cache are separate. Resize keeps
size-independent outlines. CPU and GPU Release drop the new cache references
while retaining CPU image snapshot behavior. General concurrent Face mutation
and concurrent Renderer use remain unsupported; the existing owner serialization
and immutable backing-reader contract still apply.

## Verification

Both patch stages have exact original/patch/output pins. Fresh preparation checks
all ordered stages and produces 6,171 files. Internal source review reconstructs
the two patches and verifies all ten changed outputs plus metadata. Preparation
registration now includes Core stages 1–43 and four typesetting stages.

Race checks pass for the font extension, complete Rasterx/sceneplan/GPU renderer
packages, the full offscreen driver and root framework. Eleven preparation tests
and relevant vet checks pass. Regressions cover real simple/composite/empty
geometry, original public-output independence, exact accepted budget deltas and
cold/warm boundaries, capacity/entry/font eviction, unsupported-font refusal,
variable CPU fallback, concurrent read-only cache calls, private captures, resize
and retirement. Original-versus-cached raster pixels match at four font sizes,
fractional placement, identity/transformed coordinates and fill/stroke routes.
Those sizes are not physical cross-display DPI qualification.

Three explicitly admitted native GPU groups pass integer composition,
shared-drawer order/fallback and real mixed Latin/Hebrew text, selections, paths
and icons. Recompiling identical scenes produces exactly equal cold/warm encoded
bytes. Warm packets then run on the actual native compute path and match CPU
pixel readback. This does not measure compositor presentation or device-loss
behavior. The initial native observer counted nested subtests as main groups;
its wrapper failure remains historical. The corrected run preserves native
terminal status and source maps before assertions and qualifies separately.

One Prism review uses Anthropic `claude-sonnet-5-5`: 14 implementation/pin/check/
contract files, two chunks, complete coverage, zero skipped/truncated inputs and
no findings. Runtime histories and host metadata are excluded from that payload.
Internal evidence review separately covers the final measurements and report.

## Ordinary native measurements

Baseline and candidate are separate optimized builds. The baseline predates
production edits; the candidate is rebuilt after fresh preparation. Exact build
artifacts and source/prepared identities are bound by receipts; external compiler,
module-cache and dynamic-library closure are not claimed. No profiling or source
instrumentation runs during the ordinary campaign.

The matrix includes both artifacts, CPU/GPU-selected modes, fixed/variable row
heights and 10,000/1,000,000 stored rows: five repeats per case, 120 measured native
events after 128 closed warmups. Baseline/candidate order alternates by repetition.
All 80 processes exit normally, all 9,600 events validate, and paired artifact and
small/large row-count geometry match. Every raw endpoint has warmup_guard=1.
Statistics use same-ticket UI work and nearest-rank percentiles, retaining every
outlier and per-run distribution. The table shows pooled p95/p99 from 600 events
per artifact/case and mean decimal MB allocated per event.

| Mode / height / rows | UI p95 ms, before → after | UI p99 ms, before → after | Process Go MB/event, before → after |
|---|---:|---:|---:|
| cpu-fixed-10000 | 12.37 → 10.88 | 12.74 → 11.71 | 16.675 → 9.309 |
| cpu-fixed-1000000 | 11.80 → 10.60 | 12.59 → 11.17 | 16.050 → 8.934 |
| cpu-variable-10000 | 12.36 → 11.30 | 12.99 → 12.00 | 16.620 → 9.178 |
| cpu-variable-1000000 | 11.96 → 10.75 | 12.71 → 11.31 | 16.080 → 8.786 |
| gpu-fixed-10000 | 11.32 → 10.67 | 12.05 → 11.30 | 21.887 → 14.518 |
| gpu-fixed-1000000 | 11.19 → 10.23 | 11.94 → 11.04 | 21.228 → 14.055 |
| gpu-variable-10000 | 11.95 → 10.55 | 12.41 → 11.49 | 21.867 → 14.456 |
| gpu-variable-1000000 | 11.25 → 10.40 | 12.00 → 11.36 | 21.254 → 13.904 |

All per-run warm target checks fail for both artifacts/modes. Lower observed
latency is not representative sustained/tree/workbench or physical-presentation
qualification. Process-wide Go allocation includes measurement helpers and
concurrent adapter work; it is not exact renderer, retained, native or GPU memory.

The campaign producer's provisional MEASURED label alone is insufficient. An
independent final validator requires observed producer exit 0, exact build hashes,
all closed-warmup guards, complete 80-process/case coverage, current source maps,
paired/row-count geometry and independently recomputed raw statistics. Its
separate PASS_CAMPAIGN_VALIDATION receipt closes this measurement scope.

## Separate allocation attribution

Four further fixed-10k native processes record differential sampled Go allocations
for baseline/candidate ×CPU/GPU: 120 observations and 128 warmups each, with
MemProfileRate=524288 and before/after profiles. They are diagnostic and excluded
from ordinary latency targets. Go pprof uses after minus before; regex-focused
stacks avoid interpreting absence from a top-40 table as zero work.

The baseline's inclusive outline retrieval is 789.75 MiB for CPU and 819.27 MiB for
GPU. Candidate profiles contain zero sampled bytes in the focused outline
retrieval/point-construction stacks. That supports the expected attribution; it
does not prove literally zero allocation or retained/native/GPU memory. Cold
admission and oversized/unsupported fallback still allocate when required.

A separate real composite-glyph microbenchmark has five repeats. Ordinary
retrieval costs 4,624 B and 12 allocations/op at 916–934 ns; warm lookup costs
0 B and 0 allocations/op at about 14 ns. This isolates lookup behavior and is not
a native-frame or representative UI speedup measurement.

The archive retains raw native results, profiles, command/status/hash records,
build/preparation/test/review receipts and inert producer recipes. It excludes
executables, compiler caches and app bundles. Initial setup/test/observer failures
remain separate records: incorrect baseline package path, isolated dependency
module resolution, corrected original snapshot verified against the pre-edit
build, stage registration, and native subgroup counting. Full Rasterx tests write
PNG fixtures; those test outputs are not a prepared-source closure. The current
candidate and native campaigns use fresh, pin-verified preparation afterward.

Menus, clipboard, physical Finder drag, human input/accessibility, cross-display,
sustained/native/GPU/font memory, broader tree/workbench performance and public
SDK distribution gates remain open. Windows/Linux qualification is deferred.
