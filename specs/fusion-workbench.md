# Fusion-style Mac workbench development track

The owner authorized this post-release feature work on 27 October 2026 after the
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
Anthropic Claude Sonnet 5.5 and is committed. Reuse unchanged checks. Automated
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
limits remain 64 entries/four levels. Source menus and instance caches preserve
submenu identity across ordinary redraws; removed cached branches are pruned.

ui.ModelBrowser provides separate disclosure, visibility, select and active-item
targets on a bounded 256-item hierarchy (16 ancestors). The application accepts events;
removed, disabled or newly collapsed targets refuse stale callbacks. This control
does not replace the large indexed DataTree or claim CAD-scale rich-row virtualization.

## Small-scene 3D view and orientation

viewport.Scene admits 256 opaque named box proxies. SceneView3D is an experimental
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
up to 128 stable features and a rollback marker, with accessible earlier/later
reorder proposals. It does not solve or replay CAD features. MarkingMenu supplies
an eight-direction, focus-scoped action panel and external direction admission;
this version has visible buttons/keyboard access, not invisible flick gestures,
nested radial tracking or pointer-positioned native popup windows.

## Private Forge delivery

The independent examples/forge module uses the public additions in a small box
assembly. Geometry edits, visibility, creation/deletion and feature-marker/reorder
changes use snapshot history. Selection, active component, camera and split widths
are transient. Body creation is capped at48 and feature entries at96. No files
or cloud services are used; closing discards the in-memory study.

A clean-commit private producer builds a new original-icon app, ad-hoc signs it,
creates a ZIP/checksum/source manifest and verifies the exact extracted signature.
Native readiness/readback and GPU-buffer growth are tested separately from
physical pointer input. Published build13 artifacts remain immutable and contain
none of these additions. No public distribution or new audit decision is implied.

Forge tool icons use distinct original antialiased silhouettes on transparent
backgrounds (body-plus, trash, transform arrows, grid, projection, orbit, hand,
cursor, orientation faces and curved Undo/Redo), not generic colored tiles. Ribbon
buttons use a flat transparent tool appearance while retaining native hover/focus
feedback. The tool style avoids the text-button hover shadow creating overflow.

## Frozen-feature Fusion fidelity pass

The owner approved a presentation pass limited to the ribbon, model browser and
feature timeline. Existing command IDs, item identities, event validation,
primary capture, marker/reorder semantics, Undo/Redo and lifecycle stay intact.
No new widgets, drag interactions, CAD operations, context-popup behavior,
backend options, release/performance gates or public release assets are added.

The fixed visual reference is Autodesk's desktop interface overview at
https://help.autodesk.com/cloudhelp/ENU/Fusion-GetStarted/images/diagram/ui-overview-sept-26.png,
with Mac window/toolbar references from the prior visual comparison. These are
qualitative references across different versions/scales, not a pixel specification.
Ribbon buttons are shorter with complete-group dividers; overflow still owns
whole groups. Browser rows use compact spacing, quiet independent targets and
one heading; the app removes duplicate project headings and its outer scroller.
The timeline becomes one icon strip with an explicit blue marker. Existing
previous/next and earlier/later actions remain as named, tooltip-bearing buttons;
feature names/paused state remain in tooltips and accessibility. Missing artwork
retains a text fallback. The mock body-creation history does not become a solver.
The inspector, orientation controls, model, renderer and modal action panel are
outside this pass. Native appearance and targeting are checked separately from
automated layout/routing, using the existing approved Forge-only mouse helper.

The packaged visual check exposed the native tab host's 96dp content-page floor
leaving an empty band under compact ribbon tools. Only tab descriptors whose
pages are all grouped toolbars clear that internal minimum; ordinary tabs keep
their native floor, restored on descriptor reuse. A three-DPI geometry/reuse
regression pins this correction. No public API or input behavior changes.

The first packaged minimum-only correction collapsed the scrollable page despite
the offscreen upper-height check. It is superseded: grouped-tool pages also
contribute intrinsic vertical content (visible Y overflow and zero vertical grow);
ordinary pages retain their native scroll/grow defaults. The regression now
requires a lower height bound and visible, contained ribbon action geometry.

## Remaining frozen-feature fidelity checkpoints

The owner approved inspector, ViewCube/orientation, existing action-panel and
whole-window polish, followed by integration/CI and private Forge packaging.
No commands, CAD operations, public API fields, backend/performance policy,
public release assets or audit decisions change. Each checkpoint has focused
tests, Prism Anthropic Claude Sonnet 5.5 review and a commit.

The inspector uses a compact property row with the existing QuantityField IDs,
Apply/validation path and command/history. Original full-color action artwork,
smaller headings, aligned fields and restrained dividers replace oversized gaps.

The orientation polish replaces disconnected letter pixels with centered,
original continuous stroke face names and clearer neutral faces/dark lettering.
The existing mesh pipeline, resource ownership, cube rectangle, six face IDs and
picking math stay unchanged. Accessible orientation alternatives use quieter
Mac tool styling and explicit names/tooltips, with the same callbacks/IDs.

The existing eight-direction action panel keeps its modal lifecycle, direction
order, stable cell/action IDs, validation, focus scope and callbacks. Smaller
flat tool cells align full-color icons/labels and use tighter balanced spacing;
the noninteractive center labels the panel. No flick gestures, pointer-anchored
popup, nested menu or additional command is introduced.

Whole-window polish keeps the inspector independently scrollable in short
windows and orientation alternatives in a bounded horizontal scroller when
panes narrow. The existing projection state is accurately labeled. The shared
app-local theme factory is unchanged in value; offscreen layout tests use those
same descriptors with only the GPU viewport replaced by a measured placeholder.
Actual GPU appearance/picking remain separate native checks.

The packaged native check found a dark native-toolbar surface under light
caption colors when the host uses dark appearance. Forge now explicitly pairs
its existing light ribbon page surfaces and ink, including native tab-label ink,
without changing global appearance or framework theme policy.

Framework CI exposed an SDK source inventory omission: ui now imports quantity
and viewport, so those owned portable packages must also be exported and copied
into the relocated-consumer fixture. The consumer exercises their public APIs
after deleting source. Existing distribution/audit/notices gates remain intact;
no public SDK asset is built or published by this repair. A process-local dark
native-palette pixel regression pins the light ribbon surface and dark captions.

The native action-panel check exposed the same host-dark mismatch on the
ordinary dialog's separate Body. Its resolved existing background/foreground
pair now reaches that native surface; absent theme colors retain native defaults.
Popover/sheet policy, commands, layout, public shape and global system appearance
are unchanged. The regression switches the process-local native palette after
manual initialization and before opening the dialog, matching the actual failure
sequence that an initialization-only palette diagnostic did not reproduce.

## Frozen-feature completion pass

The owner approved five bounded closeout items: existing CI completion; physical
cube/context targeting; consistent appearance across native host palettes;
existing hover/selection/navigation feedback; and a green-source private package
with updated screenshots and controls. Features, APIs, backend/pacing, release
policy and distribution decisions remain frozen.

Physical build-10 checks now establish a visible Front cube-face click, viewport
right-click opening the existing action panel and Escape dismissal. The repeated
sequence preserves body selection, camera and history except for the requested
orientation. Detached 1/1.5/2-scale routing regressions separately pin cube versus
object admission and secondary-button context without primary capture. No
cross-window or secondary capture claim is introduced.

Forge retains its explicit light application design under either host appearance.
The ordinary dialog's enclosing scene/title/footer and Close button use the same
resolved pair/theme as its managed content; sheets/popovers remain unchanged.
Ribbon action tooltips expose existing labels. Menus retain the existing Mac
palette; native tooltips retain their paired inverse surface/ink. Process-local
light/dark host regressions include dialog chrome and disabled action styling;
no global settings or runtime appearance option is added.

The owner identified the rectangular drive shaft as a visual defect. This narrow
correction permits one optional X-axis cylinder display primitive in the
experimental preview model, with default box behavior and bounded 48-segment
geometry. Forge's shaft has equal Y/Z radii. Box picking proxies remain explicit;
no exact cylinder intersections, CAD topology, modeling command or stable-v1 API
is introduced. Snapshot history preserves the primitive during move/resize/delete.
