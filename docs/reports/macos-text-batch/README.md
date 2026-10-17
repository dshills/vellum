# Transient grid text batching and immutable glyph masks

Managed grid panes can now batch contiguous plain passive Text commands after
the existing pane/background proof. Normal widget identity, shaping, current
mutable PaintText layout, clipping, input/accessibility and painter order remain.
Pending/selected/highlighted/decorated/custom content, debug rendering and other
renderers retain ordinary traversal. Both rasterx and its GPU wrapper advertise
batch support; CPU fallback understands the same commands. Capture retains full
logical work admission and original glyph placement arithmetic.

Only private bitmap glyph-cache masks enter a per-renderer immutable snapshot
bank, bounded to 1,024 entries and 8 MiB of charged original/copy backing and
entry structs. Queue/map/allocator metadata is separate. Every packet charges
its immutable resources and remains valid after cache eviction or renderer
resize/release. Public mutable GlyphMask/image sources retain their original
copy-and-compare behavior. Small-outline text also batches; its existing exact
span cache and the GPU shader are unchanged. No cross-frame text identity cache
or asynchronous work is introduced.

## Native result

Five interleaved release pairs use fixed 10k rows, 400 nonempty clipped cells,
128 warm and 120 measured events per process. Paired geometry checks pass:
50 rows/eight columns, 477 retained cells/nine headers, 162 logical DPI (1.6875),
1519x1266 client on Apple M4 Pro/LG HDR DQHD at 60 Hz, macOS 26.6.2 (25G83).
Go 1.27.2 ARM64 builds have no race detector, profiler or compiler overlay;
ordinary runs leave GOMAXPROCS/GOMEMLIMIT unset.

| Run | Baseline p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| 1 | 5.036290 | 4.882873 |
| 2 | 5.344917 | 4.985878 |
| 3 | 5.167541 | 4.951418 |
| 4 | 5.135375 | 4.904084 |
| 5 | 5.481376 | 5.085417 |

All five paired p95s improve. Median run-p95 is **5.168→4.951 ms** (4.2% lower).
Median per-run median process-wide Go allocation observations are
**6,646,664→5,311,496 B** (20.1% lower); helpers/runtime remain included.
All candidate p99s are below 6 ms, but every p95 still exceeds 4 ms. **No Mac gate
closes.** The broader workload matrix remains open after this failed first case.
Allocation observations do not qualify retained memory or the baseline UI budget.

Separate source-details cases record 120 captures per build, all GPU accepted,
zero fallback; the baseline has one additional reuse record. Their median
layout/paint is 0.736→0.594 ms and capture 1.114→1.008 ms. Median logical encoded input
stays 559,988 B, as expected for unchanged painter operations. Diagnostics are not
pooled into ordinary timing; nested phases and separate percentiles are not added.
These records do not establish presentation or physical-pointer behavior.

The first comparison was stopped after discovering that the GPU wrapper did not
advertise batch capability. Its incomplete records remain archived; they do not
enter the five pairs above. Delegation and a positive GPU-wrapper regression
were added before the corrected release binary and fresh comparison.

## Checks and reviews

Original CPU pixels and captured replay match across 96/120/144/162/192 DPI,
light/dark, scroll/frozen panes, selection/disabled state, decorations, background
changes and isolated text repaint. A positive 162 DPI test installs the actual GPU
wrapper and observes 16 batches/70 entries before mutating public PaintText layouts;
current CPU/capture pixels still match the original painter after mutation.
Further tests cover ordered overlapping text, actual warm bitmap reuse, nil or
oversized batches, refusal leaving CPU state untouched and exactly-once ordinary
fallback, detached masks, clipping/nonzero source origins, exact legacy encoding,
malformed/budget input, cache eviction with live packets, resize and release.

Complete affected sceneplan/rasterx/gpuscene race suites pass (19.547s command),
as does the full driver offscreen race suite (275.778s). Vet and release build
pass. Admitted native GPU compositor text/path, pixel/order/fallback and resource
regressions pass under the shared live lease. Eleven preparation regressions pass;
fresh preparation matches all 6,260 files exactly. The inherited full rasterx tests
write 17 output PNGs; their removal is recorded before source comparison.

Internal review finds no source blocker. One fresh complete Prism Anthropic
Claude Sonnet 5.5 call reviews all 11 staged source/spec/patch files: 41,030 B,
one chunk, no truncation. CLI exit 0 has **zero retained findings, two discarded
medium concerns and zero suppressed**. The patch-evidence validator discarded
the two concerns; both were also independently refuted from source. Actual
*Text/This guards preserve subclass dispatch, IsVisible matches StartRender,
pending/nil layouts fall back, and unchanged work.text checks nil before any
dereference. The explicit nil-layout regression passes. No production edit or
repeated provider call follows this review. Raw concerns/dispositions are retained.

## Identity and remaining work

Baseline source is 0c59fe0; its reused verified release binary SHA is
543fb562553c38d74d5d02d94474a26e22e3c19857f0654820f7e57bca5312d1.
The candidate binary SHA is
38a98dcaddb7855a2476943aa468ae6f8294f11f665e0b88946ea65e677fb9da.
The baseline compiled driver leaves match committed source and its stage0078
receipt; this is not a fresh historical rebuild. The measured candidate uses
stage0079 with 6,259 prepared files. One additional test-only file follows native
measurement; fresh final preparation contains 6,260 files with aggregate SHA
1d9861ac10a0d83f607a36fd70dc79f00ce51d427c792b30b6a1331c0c3451ba.
Production runtime leaves are unchanged after the corrected native build.

Records contain raw samples, commands, failures, reviews, source diffs and hashes,
indexed in records.json with bounded records.tar.gz. No SDK/binary is copied.
The existing runner observes direct-child exit/EOF and attempts group cleanup;
an endpoint query finds no matching owned benchmark executable. Strict per-PID/
PGID absence proof is not claimed. No input source, clipboard, VoiceOver or system
preferences changed. Cross-display, physical/human input and complete application
workflow, performance, memory/endurance, rights/final SDK and audit gates remain.

Adopt the scoped reduction; approximately 0.95 ms remains above the median run-p95
budget in this case. Next investigate the remaining full refresh/publication and
capture costs using the current complete-frame evidence. Moving work outside the
measured frame would not establish a real reduction. M2 native/M3/M4/M5/M6/M7
remain open; the eight-hour final soak awaits a stable qualifying candidate.
