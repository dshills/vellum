# Sparse scanner-span staging bands

The GPU scene compiler now builds spans in a bounded geometric band instead of
allocating full-target row headers for every scanner draw. This reduces observed
Go allocation in the native numeric-table workload by about 5.7–5.9 MB per
observation and reduces its observed GC cycles. Measured tail differences favor
the candidate in this campaign, but both builds still fail the unchanged p95 ≤4 ms
and p99 ≤6 ms warm targets. The scene renderer remains disabled by default; full
Mac/P4/P6 and native-memory acceptance remain open.

## Implementation and invariants

BeginSpans creates no row headers. The first accepted span allocates one header;
later out-of-band rows grow capacity geometrically within the existing target
height cap. Headroom can move for ascending, descending or arbitrary row order.
Only touched rows are copied when moving the staging band, preserving their span
order. Combined row-capacity/span reservation precedes allocation. Flush refunds
unused headers and copies exact relative rows/spans into the sealed packet.

The existing packet/shader/encoding format, byte blend arithmetic, painter/tile
order, independent-reset proof, detached ownership and fallback policy remain.
The production delta is only Builder. Stage 42 also includes the CPU reference/
allocation fixtures and extra multi-row native cases, with original/output pins.
There are no new GPU resource calls or caller callbacks. Previously accepted
construction needs no more header credit. Empty/clipped streams now retain no
discarded full-target reservation; the old empty-stream accounting leak is not
claimed as byte-accounting parity. Nonempty relative-row encoding and retained
bytes remain identical for the tested reference corpus.

Independent internal source review approves origin enclosures, copy indices,
growth charging, flush refunds and arbitrary order. Fresh preparation reproduces
all 6,167 files exactly. Seventeen final race/vet/preparation/consumer checks pass.
CPU tests compare all encoded Ops/Rows/Refs/Data bytes with independently built
prior full-target rows and pixels with the reference scanner, for both operators,
overlap, zero coverage, gaps, clipped rows, random arrival and target extremes.
They also check empty streams, charged one-row admission, geometric headroom,
exact sealed capacity and refusal before allocation.

Four explicitly admitted native GPU groups pass with individual RUN/PASS logs,
including the 96-packet integer compositor corpus with added end/middle/first-row
growth and relocation, real text/path parity, shared-device ordering/fallback/
mutable image behavior and borrowed-to-owned slot recycling. Native readback is
confined to these qualification tests. Backend adapter type is not recorded.

The initial stage producer incorrectly equated grouped-stage count with patch
number; its assertion stopped before manifest mutation. The failed helper is
preserved, its expectation was corrected to the preceding patch identity, and
fresh preparation verifies the final pins. No failed helper output qualifies code.

## Scoped allocation benchmark

The same 400-path ×12-row corpus on 128×4096 targets is run five times per Builder,
using explicit inert overlays. Only Builder differs; both versions use the same
benchmark-only test fixture and successfully seal the corpus. Allocation drops
from about 39.9 MB to 0.875 MB per operation. The baseline reports 10,814
allocations per operation and the candidate 12,413: geometric growth creates more
small containers while eliminating large full-height arrays. Raw time/allocation
samples are preserved. This construction benchmark is not a native frame result.

## Matched ordinary native results

Both builds select the GPU scene renderer. This compares the preceding qualified
ced30ee artifact with sparse bands, not CPU mode with GPU mode. The existing
baseline artifact is explicitly linked to its earlier source/prepared/build
qualification and current clean canonical source. Only the candidate's Builder
production bytes differ; the two test changes are excluded from the executable.

Forty interleaved native processes cover five repetitions per build for fixed/
variable stored 10k/million rows. Each completes 128 closed guarded warmups and
120 measured events. All 4,800 ordinary observations and outliers are retained,
with no diagnostic profiling/source/input flags or GPU readback. Full geometry
matches across builds and row counts. Source, prepared and binary hashes remain
stable through the campaign. This is not a complete external compiler/cache/
dynamic-library closure.

Pooled nearest-rank total UI-work p95 / p99 in milliseconds; negative is faster.
Phase percentiles are not added together.

| Case | Previous GPU p95 / p99 | Sparse bands p95 / p99 | p95 change |
| --- | ---: | ---: | ---: |
| fixed-10000 | 11.809 / 12.126 | 11.699 / 12.041 | -0.93% |
| fixed-1000000 | 11.104 / 12.165 | 10.857 / 11.739 | -2.22% |
| variable-10000 | 12.120 / 12.601 | 11.854 / 12.258 | -2.19% |
| variable-1000000 | 12.039 / 12.363 | 11.188 / 11.566 | -7.06% |

The targets remain failed for both builds. Per-run statistics, variation and all
maxima remain in the archive. The campaign does not establish a general speedup,
physical GPU execution/presentation, broad scene eligibility or full P6 acceptance.

Median process-wide Go allocation per observation, in decimal MB, with GC-cycle
sums across the five runs. These intervals include measurement helpers and other
Go activity, so they do not precisely isolate framework allocation.

| Case | Previous MB | Sparse bands MB | Previous / candidate GC cycles |
| --- | ---: | ---: | ---: |
| fixed-10000 | 27.48 | 21.67 | 480 / 372 |
| fixed-1000000 | 26.54 | 20.85 | 75 / 60 |
| variable-10000 | 27.68 | 21.79 | 476 / 369 |
| variable-1000000 | 26.68 | 20.86 | 75 / 55 |

## Review and remaining work

Four separate diagnostic processes verify the candidate's numeric-table main-scene
GPU route across 480 samples and 481 source records; the extra prefix composition
is retained. All main-scene records have zero Go-image upload attempts. These are
excluded from ordinary results and count API attempts rather than physical
transfers. Retained Go heap and end-only process RSS/CPU are kept,
but do not attribute native/GPU memory. Human services/input/accessibility,
cross-display, broader window/modal/sprite/workbench/tree support, physical GPU
loss, native memory and public distribution remain separate requirements.

One Prism review uses Anthropic Claude Sonnet 5.5 for the completed implementation,
contracts and report; raw results/disposition are preserved with independent
source and evidence reviews. No input sources, VoiceOver, display/power settings
or other OS preferences were changed. Further allocation work should be driven by
the remaining measured costs while keeping the renderer opt-in.
