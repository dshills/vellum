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

## Ribbon and model browser

ui.Ribbon retains contextual tab/group/action identities and full-color artwork.
Its nested measured group frames use the existing overflow adapter; complete
hidden groups become submenus, preserving split-button choices and disabled
ancestry. The ordinary toolbar/customizer contract remains unchanged. Menu
limits remain64 entries/four levels. Source menus and instance caches preserve
submenu identity across ordinary redraws; removed cached branches are pruned.

ui.ModelBrowser provides separate disclosure, visibility, select and active-item
targets on a bounded256-item hierarchy (depth16). The application accepts events;
removed, disabled or newly collapsed targets refuse stale callbacks. This control
does not replace the large indexed DataTree or claim CAD-scale rich-row virtualization.

## Small-scene 3D view and orientation

viewport.Scene admits256 opaque named box proxies. SceneView3D is an experimental
shared-device extension using GPUView, depth, retained geometry, camera-only
uniform updates, explicit revision changes and the same typed capture lifecycle.
Click proposes one nearest visible object; drag or Option-drag orbits; Shift-drag
pans; scrolling zooms. Hover is transient. Selection/geometry commands remain
application-owned. The in-canvas ViewCube renders and picks the same six labeled
faces; dragging it orbits, and ui.ViewCube supplies keyboard/accessible face
alternatives. Small views omit the cube. This is a preview toolkit, not CAD
geometry, arbitrary-mesh picking or an assembly performance qualification.

## Controlled editing and feature controls

quantity.Length parses bounded arithmetic, explicit mm/cm/m/in/ft and named
length parameters; it never evaluates code. Canonical values are millimeters.
QuantityField admits only finite, in-range candidates through Apply; applications
own drafts/errors and commit through their commands. FeatureTimeline presents
up to128 stable features and a rollback marker, with accessible earlier/later
reorder proposals. It does not solve or replay CAD features. MarkingMenu supplies
an eight-direction, focus-scoped action panel and external direction admission;
this version has visible buttons/keyboard access, not invisible flick gestures,
nested radial tracking or pointer-positioned native popup windows.
