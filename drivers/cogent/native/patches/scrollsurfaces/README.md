# Ordered scrolling surfaces

Stage 39 adds optional metadata to the existing painter ContextPush. The complete
original command stream remains in place, so SVG, PDF and other renderers retain
their order and fallback. Native raster replay restores an opaque rectangular
surface at its original painter position and redraws disjoint exposed strips.
It captures the result before later siblings paint, preserving their overlap.

The per-renderer cache retains at most 64 groups, 32 MiB of owned RGBA pixels and
8 MiB of detached actual-paint descriptors. Descriptor accounting includes mask
headers/pixels, fixed path data and operation slots, excluding Go allocator/map
metadata. Replacement capture temporarily overlaps old plans and builder/trim
allocations. Neither retained pixels nor descriptors reference widgets, models,
font faces, shaped runs, commands or borrowed glyph masks. Resize, eviction and
clock exhaustion retire both. Offsets are absolute integer output pixels with
checked signed subtraction; large jumps use full replay.

Opted-in native Grid panes request VerifyPaint. Capture follows the existing text renderer
on its originating owner, preserving glyph-data precedence and exact subpixel,
monospace, superscript/subscript, background and selection placement. It copies
actual alpha masks or fixed outline control points, known value colors and clips.
Ordered old operations translated into the overlap must exactly equal current
operations affecting that overlap; unknown/changed paints fall back. Full-mask
and full-outline comparison can conservatively reject safe reuse. The accepted
current plan also drives full/strip replay, closing the font/color checking-to-
drawing gap. The guard stops promptly at its descriptor/operation budget.

Uniform sources must really cover their destination within their finite bounds.
Unknown images/colors, transforms, masks, painting paths, blur, uncertain opaque
coverage and unsupported text use ordinary replay. Standard-box paths with both
fill and stroke disabled are harmless. SVG/bitmap, stroked, decorated, vertical
and math text are refused. Zero text clips are refused because the legacy raster
scanner interprets that rectangle as disabling clipping. Other callers may use
an explicit epoch promise; native Grid always uses the exact paint proof.

Quadratic/cubic filler curves flatten relative to their integer control hull and
restore its origin and terminal point exactly. The preceding absolute weighted
float calculation can change edges under integer translation. Some prior
antialias edge pixels intentionally change. Shaping, layout, glyph identities,
bounds and semantic geometry remain unchanged. Existing BSD attribution remains;
new tests and render metadata are MIT.

Tests compare exact emitted edges and prior-dispatch full-replay pixels, including
selected/bidi/combining text at four DPIs and four fractional phases; dirty parent
repaint and later overlaps; source bounds; signed extremes; resize/eviction;
changed ink, background, selection, clipping, physical phase, glyph output and
actual consumed mask bytes; capture limits; and real native-adapter frames.
The ordinary GPU image upload remains complete. This contract does not establish
strip-only GPU writes, physical GPU/presentation timing, native memory reclamation,
4/6 ms latency acceptance or complete Mac qualification.

The owned pane buffer uses toroidal storage. Checked scroll deltas change a
bounded logical-to-storage offset; overlap restore splits at most once in each
axis. Only disjoint exposed strips are written back before later siblings paint.
Full repair resets the mapping and captures the complete pane. CapturedPixels
counts owned CPU cache writes, not GPU uploads. This avoids a second full-pane
copy after every reuse. Capture shares identical copied glyph masks within one
plan, checking current bytes again before sharing. A changed borrowed mask gets
a new owned copy; the temporary lookup is cleared before retaining the plan.
Detached replay preserves covered-clear elision: an owned uniform Src immediately
covered by an equal-rectangle opaque uniform Over is replaced by that Over.
Admission has already checked each original uniform's finite source coverage.

Managed main-window Grid reuse is opt-in through Driver.ExperimentalGridScrollReuse
or ManualConfig.ExperimentalGridScrollReuse. Both default to false and are copied
at adapter startup; additional modal scenes retain ordinary replay. Current
matched stored-table experiments regress latency even though pixels agree, so
normal managed rendering emits no cache hints and retains no pane cache. The
experimental mode remains available for further renderer work; this is not a
recommended performance setting or completed latency acceptance.
