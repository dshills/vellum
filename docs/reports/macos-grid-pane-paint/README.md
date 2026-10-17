# Pane-aware managed grid painting

The managed pane now avoids cell/header StandardBox construction when it would
only redraw the pane's freshly painted opaque uniform background. It retains
normal Text widgets, shaping, clipping, selection/cursor/editor behavior, input,
measurement and accessibility. Strict style, actual-geometry, normal-leaf,
ancestor-context and disjoint-cell guards retain the original painter for
unsupported cases. Temporary flags restore after the full pane traversal;
isolated Text repaints still clear their old contents. No new shader, cache,
resource lifetime or public API is introduced. The existing experimental scroll
protocol keeps its original painter.

## Native result

Five interleaved release pairs use fixed 10k rows, 128 warm events and 120
measured events per process. Each paired geometry comparison passes with 400
nonempty clipped cells, 50 rows/eight columns, 477 retained cells/nine headers,
scale 1.6875 and a 1519x1266 client on this M4 Pro/LG HDR DQHD host. Go 1.27.2
ARM64 binaries have no race detector, profiler or compiler overlay;
GOMAXPROCS/GOMEMLIMIT are unset.

| Run | Baseline p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| 1 | 5.459041 | 5.277626 |
| 2 | 5.414833 | 5.476002 |
| 3 | 5.465916 | 4.859876 |
| 4 | 5.425832 | 5.106624 |
| 5 | 5.336294 | 5.140168 |

Four of five paired p95s are lower. Median run-p95 is **5.426→5.140 ms**, about
5.3% lower. Median per-run median process-wide Go allocation observations are
**7,504,192→6,905,904 B**, about 8% lower; helpers/runtime are included. These
are scoped observations, not a general no-regression or retained-memory claim.
Every p95 still exceeds 4 ms; one candidate p99 also exceeds 6 ms. **No Mac gate
closes.** The broader workload matrix stops at this failed first workload.

Separate source-details cases have 120 captures per build, all accepted GPU draws,
zero fallback and zero reuse. Their median layout/paint is 0.838→0.744 ms,
capture 1.165→1.121 ms, and logical input 618,092→559,988 B. These instrumented
cases are diagnostic and never pooled into ordinary timing. Nested durations
and separate phase percentiles are not added. This proves the GPU route only
in those diagnostic processes, not presentation or physical pointer behavior.

## Correctness and reviews

Original CPU pixels and captured packet replay match across 96/120/144/162/192 DPI,
light/dark backgrounds, scrolling/frozen panes, selection and disabled state.
The 162 DPI cases remove 1,389 render commands across eight frames per theme;
this covers the bitmap glyph route rather than only small-outline spans.
Optimized-first background changes after a prior frame, decorated/foreign
content, isolated long-to-short Text repaint, and positive-admission-before-
refusal for overlaps/masks/clip paths/bounds paths/transforms pass with race
checks. The existing verified-scroll pixel regression passes. Final native
build/vet and the complete driver offscreen race suite pass (278.659s command).

Internal review approved the guards and requested the positive/refusal tests.
One fresh Prism Anthropic Claude Sonnet 5.5 call reviews all four implementation,
test and contract files: 18,640 B, one complete chunk, no truncation. It returns
**two retained medium findings and CLI exit 1**. Independent internal review
refutes both for this guarded path: normal Text clips all foreground, borders
and shadows to its disjoint TotalBBox; opacity 1/state layer 0 returns exactly the
explicit/inherited background used by the proof. An invisible pane paints no
children and deferred flags restore. Targeted added race tests prove oversized
border/shadow pixel parity while neighboring boxes actually elide, and explicit
state-layer refusal. These additional tests run after the full suite/provider;
production leaves and measured binary are unchanged. No unchanged provider
re-review is requested. Raw findings and scoped dispositions remain in records.

## Identity, limits and next implementation

Source base is clean ae99e65; the measured candidate is that base plus the
recorded four-file working diff. Final test-only additions are separately
recorded. Binary SHA:
543fb562553c38d74d5d02d94474a26e22e3c19857f0654820f7e57bca5312d1.
Baseline reuses binary 478cedd825bbcc48709fe45210c448588b72fa82ce3df41b4485ed1293d745ba,
compiled from f641776 plus the recorded three shared-baseline implementation
leaves; those exactly match ae99e65. This is not a fresh historical clean build.
All 6,253 prepared files match the preceding receipt and remain unchanged.

Records include exact commands, raw samples, source diffs, hashes, failures and
reviews, indexed in records.json with a bounded records.tar.gz. No SDK/binary
is copied. All cases use the existing runner under a fresh stop/shared lease.
The runner observes direct-child exit/EOF and attempts group cleanup; a terminal
endpoint query finds no matching native executable. Strict per-PID/PGID absence
receipts, human accessibility, displays, physical input and presentation are not
claimed. No input sources, clipboard, VoiceOver or preferences are changed.

Adopt the scoped paint/allocation reduction. About 1.14 ms remains above the p95
budget in this workload. Next implement a pane text batch of immutable glyph
instances covering both 162 DPI bitmap masks and small-outline spans, preserving
widgets/shaping/geometry/input/AX and original fallback. It must remove repeated
per-cell context/text-command construction and general glyph recapture, then
prove original pixels and matched native cost. Async AX fallback is deferred:
it needs sequencing/acknowledgement and bounded snapshot lifetime; merely moving
work outside the recorded frame would not establish a real reduction. M2 native,
M3, M4 memory/endurance, M5 human/display, M6 rights/final SDK and M7 remain open.
