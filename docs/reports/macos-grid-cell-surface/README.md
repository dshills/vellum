# Managed grid body layouts — Mac checkpoint

The grid replaces per-cell body `core.Text` widgets with pane-owned, current-
projection layout values on the supported passive route. This removes body
widget update/layout/traversal while preserving public grid models, host pointer
and keyboard routing, completed geometry, native accessibility, panes, headers,
scrollbars and the native inline editor. Unsupported configurations recover the
ordinary widget implementation.

This is a measured rendering improvement. It does **not** close M3 or Mac
completion. Every ordinary run-p95 below still exceeds the unchanged 4 ms target.
No table/tree/general-UI matrix, presentation, human input, M4 resource/endurance,
M5 native workflow or M6 distribution gate closes.

## Measured private prototype

The clean calling repositories were `410a98c`; the retained private SDK was clean
`abe67c6`, stage0079, Go 1.25.6, with 7,879 payload files plus its manifest. The
prototype used compiler overlays; it changed neither checkout nor SDK/prepared
files. One ordinary baseline, candidate and separately instrumented candidate
binary are pinned in the plan. Ordinary binaries contain no surface counters.

The local display was LG HDR DQHD, 5120×1440, DPI 162/DPR 1; native render size was
1519×1266. The existing stored 10k fixed-row fixture retains twenty columns,
two frozen columns and at least 400 nonempty clipped cells, 50 row identities
and eight column identities. It routes initialized scroll events through the
native FIFO and framework owner. Each of five interleaved pairs per mode has
128 unmeasured warm events and 120 measured events. These are native automated
routing observations, not physical pointer or input-to-presentation evidence.

Per-run nearest-rank percentiles were recomputed from all 2,400 ordinary samples;
no pooling or sums of independent phase percentiles were used.

| Mode | Median run p95, original → prototype | Reduction | Improved pairs | Median run p99, original → prototype |
| --- | --- | --- | --- | --- |
| CPU | 5.929751 → 4.811373 ms | 1.118378 ms | 5/5 | 6.733457 → 5.093458 ms |
| GPU | 5.022791 → 4.249374 ms | 0.773417 ms | 5/5 | 5.712500 → 4.667084 ms |

Both modes pass the frozen prototype adoption bar: at least 0.4 ms lower median
run-p95, at least four of five improved pairs, exact paired geometry, no increase
in median run-p99 or count of p99-over-6-ms runs, and no more than 1 MiB additional
median terminal Go heap. Original → prototype terminal median HeapAlloc is
28,781,192 → 26,042,880 bytes on CPU and 30,173,592 → 27,378,816 on GPU. This is
post-GC **whole-process Go heap for this fixture**, including domain/demo data;
it does not establish baseline UI, native/font or GPU retention budgets.

There is an allocation tradeoff. Median-run process allocation p50 rises
3,611,576 → 4,160,904 bytes/frame on CPU (15.21%) and
5,218,168 → 5,748,888 on GPU (10.17%). This process-wide observation includes
measurement/inspection helpers and GC pressure; it is not a framework-only
allocation attribution or an allocation improvement. The original 64 KiB and
100-allocation baseline requirements remain open.

Two separate 120-measured-frame diagnostics match each surface observation to
the exact completed generation/frame. All 240 show 468–540 current layout values
and nonnil shapes, equal to the materialized projection, with zero body native
children and zero native body Text entries. CPU records prove fresh raster/copy
and an attempted image upload with checked RGBA bytes; GPU records prove fresh
accepted SceneHost drawing without fallback or reuse. Diagnostic timing is
excluded from the table. `RetainedCells` counts current cell representations and
alone cannot prove widget removal.

All 22 runs have complete bounded stdout/stderr, stable source/binary pins,
terminal/reaped child and absent owned process group. The shared native lease
was held through cleanup; per-case stops and the final scope stop are sealed.
No physical input, placement operator retry, input source, VoiceOver or OS
preference changes occurred.

## Ownership and fallback

Only the admitted projection's values and shaped lines are retained. Stable
row/column keys preserve unchanged overlapping shapes; departed values, text,
styles and references are cleared on replacement/disposal. No hidden Text
surrogates, offscreen-row cache, font-key bank or cross-frame paint-fragment
cache is added. Escaped text uses the existing rich-text/shaper route. Global
edge rounding, inside-border content positions, clipping and painter order
match the original implementation.

Admission requires completed fixed-allocation geometry, no editor, managed
absolute inherited units, supported opaque styles, no foreign body children and
an explicit text-batch renderer. Custom widget initialization, debug tracing,
both Inspector modes and experimental scroll reuse retain the widget route.
Selection/cursor styles are resolved separately; computed paint backgrounds use
a render-local copy so painting cannot invalidate retained shape epochs.

A late refusal revokes completed measurement, marks layout pending and queues
owner reconstruction before the same-generation fast return. Failed/incomplete
values are never painted as valid cells. One invalidated paint can omit body
values before the owner restores the complete ordinary frame; no stale cell
measurement or action admission survives that transition. Native editing uses
the ordinary widget/editor route.

## Corrected source integration

The measured overlay remains frozen. The production integration additionally
handles a paint-only `Restyle`: dirty inherited styles fail closed instead of
painting old templates without a layout pass. The legacy oracle selector moves
from production `NewManual` to widget-specific **test helpers**, so normal Manual
execution uses the same value route as native windows. The direct default-Manual
promotion and paint-only restyle recovery have dedicated regressions. Native
children or Parts added after admission also revoke the value route; fallback
retains the unowned child. The Parts regression uses a parented separate tree
with zero Children membership, so the Parts guard is independently exercised.

Stage0080 owns the small Core renderer/Inspector eligibility hook and its direct
refusal tests through the existing original/output hash and ordered patch chain.
The staged driver/prepared-source correspondence is recorded separately.
Prototype timings are not relabeled as exact integrated-binary acceptance.
The corrected integration passes the complete driver-root race suite (275.166 s
outer), example race tests, vet, Core admission tests, eleven preparation test
methods and the exact stage0080 prepared correspondence. The final Parts fixture
was then corrected to Core's separate-tree shape; the final focused race suite
passes that test-only delta without changing production bytes. Both final native
functional diagnostics pass 120 measured frames each on the corrected source,
with actual value layouts, zero body widgets and the required fresh CPU/GPU
route. Their timing is not used for adoption or full matrix qualification. Every
owned native group is terminal and absent, with final stops sealed.

## Verification and reviews

The frozen prototype's focused race tests compare original pixels, completed
geometry and ScenePacket replay across 96/120/144/162/192 DPI. They cover current
shape retention, mixed scrolling, selected rows and cursor borders, Unicode,
escaped markup, tabs/newlines, edit fallback, disabled state, variable heights,
empty/frozen panes, absolute typography, routed pointer/key/fractional wheel,
stale releases and AX handlers, fresh exactly-once GridChange, late unsupported
style/renderer/WidgetInit recovery and disposal. A broad driver race suite,
example race suite, vet and both native builds pass. Widget-specific legacy
tests explicitly exercise the reference implementation; their pass is not
claimed as value-route coverage.

Failed attempts remain evidence: the original selected/cursor pixel failure,
the initial incomplete selection/edit test sources, test import errors and the
legacy widget-index panics are retained with their source versions. They are
not relabeled as passes. The actual selected-background and fallback defects
were corrected before native admission. Fourteen malformed pure validator
controls reject missing/duplicate/stale observations, body widgets, incomplete
shapes, invalid CPU source records, workload/callback drift and diagnostic data
in ordinary timing. Historical/synthetic parser controls create no native proof.

Internal source review, exact native admission and independent raw-result review
are retained. The latter confirms the gains, all cleanup chains and the
allocation regression. Integration review is retained. One complete Anthropic
Claude Sonnet 5.5 Prism call reviewed all 19 staged files (68,290 diff bytes,
one chunk); it retained one medium font/DPI invalidation finding (CLI exit 1).
The existing complete-style comparison includes the unit context and font, so
the reported stale-shape path was refuted. A post-review test bypasses normal
Style/dirty notifications and verifies epoch advancement and replacement of
current shapes for DPI, font size and font family changes, then restores the
context and compares original-route pixels and geometry. All focused race tests
pass. An isolated negative control disabling just the complete-style comparison
fails all three cases as expected. The initial enum-fixture compile failure is
retained. Production bytes are unchanged from the reviewed, race-tested and
native-verified integration. This is resolved-context evidence; physical display
transitions remain open. Raw review output, the finding disposition and internal
closure are indexed in `checkpoint.json`; no second provider call was made.

The archive index, exact identities and final verification/review result are
listed in `checkpoint.json`. No public push, release, tag, signing or notarization
is part of this checkpoint.
