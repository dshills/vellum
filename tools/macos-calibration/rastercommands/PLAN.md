# Bounded original raster command metadata

This private observer leaves render execution unchanged. Before Renderer.Render,
record scalar types, rectangles/source positions/operations, builtin uniform RGBA
values and path length/transform/fill/stroke flags. No source images, widgets,
commands, handles or borrowed slices are retained. No custom Color.RGBA/At/Bounds
or Stringer calls: type strings use %T, color values only built-in value types.
Mutex-protected snapshots detach row slices; reset releases metadata. Maximum
32renderer calls/1024total items; overflow must refuse rather than accept a prefix.
All Go code is under testdata, outside ordinary Go package discovery.

Use a fresh exact Core SDK copy and one Render-entry hook, private module/test
compiler overlays, original headless routed-pixel fixture. Original0/1/2state,
exact full image pixels and unrelated-key no-op still pass. Four render calls/
80items record1500x1250 geometry:20items/frame,8imageDraw parameters,4unpainted
paths and4push/4pop contexts. Recorded operations may be skipped/cropped by the
existing renderer; do not claim8fullfills or attribute0.85ms native raster time.
Manual96DPI geometry differs from native1519x1266. No native execution is claimed.

Two focused private recorder methods cover custom-callback avoidance, detached
snapshots, caps/overflow and reset. Initial recorder-test setup stopped because
the copied Core module's ../typesetting sibling was absent; corrected modfile
resolves the immutable SDK typesetting source. Retain failure; no network/SDK
changes. Offscreen fixture uses the correct existing module replacements.

Next prepare a separately reviewed single native three-transition command capture
with original library/defaults, exact locked native geometry and strict owned/pin/
foreground/lifecycle bounds. Add actual skip/crop decision metadata if needed
before choosing any repair. This is a structural question, not another unchanged
performance campaign. Do not infer ordinary/native equivalence or change gates.
Features frozen, experimental rendering/pacing disabled,33.3ms gate unchanged,
M3–M7 open. Prism Anthropic Claude Sonnet5.5 review and commit each checkpoint.
