# macOS keyed Grid and fixed-allocation layout checkpoint

Managed grids retain native text widgets by row/column identity and lay out their
changed descendants inside a verified fixed allocation. This removes unrelated
scene sizing while preserving full scene painting and ordinary raster upload.
The implementation uses the default path; the separate CPU pane-cache experiment
remains disabled. This checkpoint adds a retained layout boundary. Its measured full-frame results
are mixed; it does not qualify the GPU renderer or complete Mac/P6 acceptance.

## Implementation

The pool contains exactly the admitted bounded projection. Overlapping row/column
keys keep their widgets and shaped text; departing widgets serve entering keys,
and surplus widgets are deleted. Headers use column identity. Snapshot order is
restored per pane without native OnAdd, with the private editor last. No rows,
widgets or shaped runs outside the current projection are retained. Native cell
incarnations still change with portable generation; widget reuse grants no stale
input rights. The owner still materializes every bounded Snapshot cell and checks
source health; no reduction in source Cell callbacks is claimed.

The host records its completed scene, parent, geometry/allocation, resolved style,
scale, viewport, frozen/header geometry, configuration, effective disabled state
and help. An unchanged fixed allocation can request local layout. Only changed
text/style/physical size is updated and sized; all moved descendants receive
current position and clipping. Physical sizes use rounded global edges so
fractional-DPI widths cannot reuse an invalid shape. New widgets are initialized;
cross-pane moves, editor sessions, styles/themes, scale/host allocation or parent
changes fall back to ordinary full layout.

Stage 40 introduces Core LocalLayoutRoot and Scene.NeedsLocalLayout. The queue
holds at most 64 initialized roots and checks actual identity and membership on
at most 256 ancestor edges, including parts/scrollbars. It runs under the existing
originating render-context ownership. Requests are detached before callouts;
new local/global/render requests survive for a later frame. Full rebuild/restyle/
layout takes precedence over preexisting local work. An uncertain or changed
outer envelope falls back to full layout. Successful local layout records the
whole scene in painter order, clearing old positions and preserving later
siblings. Destruction clears retained queue references before and after cleanup.

An atomic GlobalLayoutPending fence refuses old Grid geometry during requested
or active full layout, including another scene's inspection. It clears only after
normal completed paint with no newer full request. Panic/Goexit retains a fenced
full retry. Publication identity is rechecked; this does not authorize foreign
widget mutation. The earlier unconditional flag clearing lost new requests from
callbacks and was corrected during internal review. Preserving render requests
also exposed Slider.Render requesting its own next paint; thumb positioning in
Render is now pure while external setters retain invalidation.

Completed capture schedules one owner accessibility publication even for a
same-generation full layout. Two integration corrections keep that completion
transaction coherent and quiescent. A private field can repaint preedit alone;
actual editor text/focus/session/visibility/bounds differing from completed data
requests a bounded host repaint and capture after child painting, without full
field update. Identical completed owner refresh requests nothing. Trees compare
complete validated bounded snapshots, including same-generation Pending/Cancel,
effective disabled state and help before full updates, while restoring only a
different native scrollbar value. Draft/selection/restore and controlled semantics
remain synchronized.

## Checks and source boundaries

The preceding clean canonical commit is 391a50f. Final v2 contains 666 selected
source files and 6,148 prepared files. The stage changes four existing Core files
and adds two; literal patch and input/output pins bind all six. Twenty final
checks pass: preparation integrity/fresh preparation, portable race tests, scoped
Core local-layout/lifetime race tests and two scoped Core vets, full driver
race/vet under default and offscreen builds, eight consumer race suites and two
native build artifacts. Core vet explicitly excludes inherited composites and
unreachable diagnostics. This is not a full Core test/complete compiler closure.

Actual local/full software frame comparisons cover fixed and indexed variable
rows at 96/120/144/192 DPI and owner batches 1/7/64, mixed-axis scrolling,
fractional-DPI physical sizes, retained shapes and an unrelated layout probe.
Pool tests check current painter/editor-last order and shrink/zero/repopulation
retirement. Pending geometry, same-generation full completion, private preview,
restore and shared Tree+Grid quiescence have real regression coverage. Core tests
cover requeue, full precedence, capacity, envelopes, lifecycle, abnormal unwind,
render quiescence and existing originating timer delivery. Pre-correction overlays
compile and fail actual requeue/slider/identity/editor/tree assertions; setup and
failed draft receipts remain historical evidence, not final qualifications.

Candidate v1's broader driver gate failed the completed private-preview and shared
Tree+Grid cases. Frozen sources and refused review/check receipts are retained.
The independently reviewed v2 fixes both; final tests bind the new bytes rather
than relabelling the v1 result. Three live Cocoa processes pass on the exact final
artifacts: fixed-10k and variable-million 400-cell accessibility/identity/action
checks, plus generic accessibility. No input sources, VoiceOver or OS preferences
are changed. Human IME/screen-reader, physical GPU pixels and native-memory
acceptance remain separate.

## Matched ordinary measurements

Forty interleaved native processes retain 4,800 ordinary frame observations:
five processes per build for fixed/variable stored 10k/million tables. Every
process completes 128 guarded warmups and 120 measured frames. Diagnostics are
disabled, all outliers are retained, and paired geometry matches exactly across
builds and row counts. Candidate source/prepared/binary bytes remain qualified
throughout. The existing final default-v6 artifact matches clean canonical
391a50f and its prior qualified source/prepared/build receipt; this is an explicit
baseline artifact link, not a newly built baseline or the old cache experiment.
Pooled nearest-rank UI-work p95 / p99 in milliseconds (negative change is faster):

| Case | Baseline p95 / p99 | Candidate p95 / p99 | p95 change |
| --- | ---: | ---: | ---: |
| fixed-10000 | 12.553 / 13.150 | 12.220 / 13.032 | -2.66% |
| fixed-1000000 | 12.389 / 12.650 | 12.208 / 12.514 | -1.46% |
| variable-10000 | 12.676 / 13.010 | 12.856 / 13.661 | +1.43% |
| variable-1000000 | 12.275 / 12.752 | 12.486 / 12.885 | +1.72% |

Pooled median UI work falls from 11.464–11.867 to
9.086–9.301 ms across the four cases. Recorded reconciliation
medians fall from 3.834–3.958 to 1.419–1.565
ms; layout/paint medians change from 2.101–2.178 to
1.794–1.860 ms. These are separately measured phase
distributions, not summed independent percentiles. They show lower typical work
in this comparison despite the mixed tail result; they do not identify a single
cause for the remaining slow frames.

0 of 40 processes meet both unchanged 4 ms p95 / 6 ms p99 targets;
40 fail both. Full p50/p95/p99/max, per-process variation, all phases
and queue distributions remain in the raw summary. Fixed-row pooled p95 improves slightly while variable-row pooled p95 regresses;
no general full-frame speedup is established. These observations qualify this
local comparison, not physical presentation, GPU transfer/retirement,
native allocations, full P6 data-view coverage or the broader professional-app
workloads. Raster replay and image upload remain ordinary and must be addressed
by an ordered GPU scene-source renderer and its own correctness/acceptance work.

## Evidence and review

The archive and exact per-file index retain original Core inputs, frozen source
versions, refused drafts/gates, meaningful compiled negatives, final checks,
source/prepared/binary links, native output and every ordinary observation.
Independent source and receipt reviews precede one significant-batch Prism
review using Anthropic Claude Sonnet 5.5. Raw provider output, disposition and
separate final closure are kept in review/. Current endpoint maps establish
selected input identity, not all compiler/dependency/SDK artifacts. Distribution
notices, native/GPU memory, human native services and full Mac acceptance remain
open. No public push, tag, signing or release is performed.
