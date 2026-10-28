# Vellum Forge — Design Workbench

An original Fusion-inspired macOS example consuming public Vellum APIs. It is a
small editable assembly and UI demonstration, not an Autodesk product or CAD kernel.

After the repository native preparation, run from this directory:

```sh
. ../../drivers/cogent/env.sh
go run .
```

- Drag the viewport to orbit, Shift-drag to pan, scroll to zoom. Option-drag also
  orbits. Click a body to select. Select mode disables orbit while selecting.
- Click a labeled orientation-cube face, use the orientation buttons, or focus
  the viewport and use arrows/Home. Inspect provides additional view controls.
- The browser has separate selection, visibility, disclosure and activation.
- Add/delete bodies, move by 4 mm, or Apply a width such as `2 in`, `50 mm` or
  `width / 2`. Use native toolbar Undo/Redo or Cmd-Z / Shift-Cmd-Z.
- The bottom feature strip has selection, earlier/later reordering and a marker
  that previews body creation visibility. These changes are undoable. This is a
  mock feature history, with no geometric dependency replay or feature solver.
- Right-click opens the visible action wheel; Escape dismisses it. The Inspect
  ribbon also opens it for keyboard access. Invisible flick gestures are absent.
- Resize the window to see whole ribbon groups overflow into More. Split-pane
  dividers resize the browser and inspector. Artwork is original, full-color.

Studies use versioned `.vforge` files through native Open/Save (Cmd-O/Cmd-S;
Cmd-Shift-S for Save As). Closing or opening a dirty study asks before discard.
An atomic private recovery copy is offered after interruption; only one Forge
instance may own recovery. Saved studies are never replaced by recovery.
No network or cloud service is used. Display geometry uses opaque boxes and a round shaft; picking follows box and capped-cylinder primitives,
with a 48-body demo limit. There are no sketch constraints, B-rep modeling,
assemblies at Fusion scale, cross-window capture, or secondary-button capture.
The reusable SceneView3D and viewport helpers are experimental additions; the
released v1 API baseline/backend and performance policy are unchanged.

Checks: `go test -race -tags offscreen ./...` and `go vet -tags offscreen ./...`.
`go run . --smoke` performs a bounded real GPU readiness/readback check and
synthetic owner-routed orbit/pan/zoom/history checks. Those synthetic events are
separate from physical pointer/native-delivery verification in the delivery report.
Private packages are ad-hoc signed, unnotarized, and not public release artifacts.

Build a private package from a clean, committed checkout (native environment
prepared as above): `python3 -B package.py --output /private/tmp/forge-new-build`.
The output is exclusive and contains the app, ZIP, checksum, source manifest and
this README. It verifies the signature again after extracting the exact ZIP.
No publication, Developer ID identity or notarization is requested.

A rejected dirty checkout does not reserve the output directory. Later failures
retain partial output for diagnosis; only a finished producer writes manifest.json
and SHA256SUMS. Choose a fresh output directory for another attempt.

The Fusion fidelity pass keeps commands frozen while using a shorter grouped
ribbon with separators, a single compact browser heading/outline, and one feature
icon strip. Hover/focus feature icons for names and paused state. The two straight
arrows move the marker; the two curved arrows move the selected feature earlier
or later. These remain separate history actions with Undo/Redo; no new timeline
drag behavior or CAD feature calculation is introduced. Visibility and activation
retain independent keyboard/pointer targets and accessible checked states.

The inspector now uses compact aligned property controls and scrolls independently
when the window is short. The orientation row scrolls horizontally if a split
pane is too narrow. Cube faces carry original stroke names; the same six face
identities and camera/picking math remain. The existing action panel has compact
flat full-color cells, tooltips and accessible names; its actions, Escape dismissal
and modal lifecycle remain unchanged. The viewport is still a small box-proxy
demonstration, and the presentation pass adds no CAD operations.

## Try the completed preview

1. Start on Solid. The blue tool treatment and viewport caption identify Orbit.
2. Drag to orbit, choose Pan and drag to pan, or scroll to zoom. Select admits
   clicks while leaving plain drags inert; Shift-drag pans and Option-drag orbits.
3. Click a cube face to orient. Fit assembly restores the initial camera/target;
   Home/orientation buttons change orientation while preserving pan/zoom.
4. Select a body in the viewport or browser. Its material stays visible inside
   a blue outline of its picking bounds. Hover gives a lighter preview.
5. Select Drive shaft to see cylindrical display geometry. Its inspector explicitly
   identifies precise cylinder picking. The blue outline still shows its bounding box.
6. Apply a width, Move or Delete, then Undo/Redo. Right-click opens existing actions;
   Escape dismisses them. Navigation/selection does not add history entries.

The app intentionally retains a light design on light/dark macOS hosts. Ribbon,
action panel title/content/footer and Close use matching readable colors. Full-color
icons and native focus feedback remain. It does not offer automatic dark theming.

The completion checks physically verified navigation tools, pan, inert Select
drag, cube Front, right-click, Escape and Fit dismissal on the private candidate.
Earlier attempts affected by desktop activity are excluded. Automated routing,
geometry, winding, host-palette and history tests are separate evidence; these
checks do not qualify other hardware/displays, human accessibility or CAD scale.

Shift-click bodies in the viewport or browser to toggle membership in a selection.
Plain click replaces it. Selected bodies retain their materials and blue outlines.
Move +4 mm and Delete operate on the full selection as one Undo/Redo action.
The inspector width field edits the primary body only.
