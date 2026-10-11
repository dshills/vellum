# Renderer-owned row tracking

Stage 37 follows source-cost diagnostics. Three changed files and two new
software test files are bound by exact input, patch and output hashes.

The native raster renderer and its mutex-owned glyph-mask renderer select
NewScannerWithRowTracking. Its private half-open range covers rows allocated
by scanner methods, independently of clipping. Draw saves a pending final cell
before restricting row traversal. Clear resets the whole bounding range,
including clipped-out rows; repeated Draw and clip changes retain accumulation.
SetBounds always initializes the full current row-head buffer, including reuse
of capacity after shrinking and regrowth.

The general NewScanner retains full-row Draw/Clear and direct exported
Cell/CellIndex behavior. The new constructor requires exclusive method-based
accumulation; the two owned callers meet that contract. Winding, coverage,
span order, colors, transforms and clipping semantics are unchanged.

Differential tests copy the baseline full-row Draw/Clear/SetBounds/alpha oracle,
feed actual scanner methods, and compare exact ordered spans and saved state.
Pending cells, both winding modes, negative/offscreen/empty clipping, distant
contours, repeat drawing, capacity reuse and default direct mutation are covered.
The copied oracle retains original upstream attribution; the new harness is MIT.
Owned-renderer tests compare actual Render fill/stroke, curves, transforms,
clipping, replay and resizing with an ordinary full-row reference. Independently
initialized local glyph caches compare real-font cold/warm masks, placement and
software composition without mutating the global cache. These establish software
pixel parity, not GPU execution or complete native text composition.
Small-path benchmarks compare the same operation and zero-allocation span
consumer at multiple surface heights. They establish scanner work separately
from real native frame latency; this does not make a partial GPU upload or
change widget invalidation.
