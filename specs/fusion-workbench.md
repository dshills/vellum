# Fusion-style Mac workbench development track

The owner authorized this post-release feature work on27 October2026 after the
build13 publication. This narrow track permits the reusable controls and input
additions below; it supersedes the earlier feature freeze only for this work.
Published tags/archives, existing stable APIs, backend, rendering/pacing defaults,
release/performance policy and historical audit records remain unchanged.

1. Typed GPU input and portable camera/ray/box picking helpers; preserve legacy
   GPUView.Input. Primary capture, cancellation and visible-clip hit testing reuse
   the existing same-window router. Alternate buttons have targeted events only.
2. Contextual grouped ribbon and rich model-browser controls using public nodes,
   immutable full-color artwork and controlled application-owned callbacks.
3. Reusable viewport tools: orientation control/ViewCube, unit-expression entry
   and marking-menu presentation with keyboard alternatives.
4. Reusable feature timeline plus a runnable independently consumable workbench
   with a small editable assembly, selection/visibility, pan/orbit/zoom, commands
   and Undo/Redo. Timeline entries are a mock feature history, not a CAD solver.
5. Native Mac layout/interaction acceptance, private ad-hoc package/ZIP, README,
   signature/exact-byte checks and final launch. No new public release assets.

Every implementation phase updates this contract/docs, runs focused meaningful
regressions and appropriate compatibility checks, receives Prism review using
Anthropic Claude Sonnet5.5 and is committed. Reuse unchanged checks. Automated
routing/picking evidence is separate from physical pointer/native presentation.
Human IME/VoiceOver, Fusion-scale assemblies, cross-window pointer dragging and
parametric geometry remain outside this demonstration's qualification.

GPUEvent coordinates/Size are matching full-content physical pixels; Inside uses
visible clipping. Primary release is consumed once. Escape, focus loss, hidden
or removed owners cancel using existing capture; closed windows suppress late
callbacks. Secondary/middle input does not claim out-of-widget capture. Camera
helpers are renderer-independent, Y-up, float64 model math with float32 WebGPU
matrices at the upload boundary. Picking boxes are application-supplied proxies.
