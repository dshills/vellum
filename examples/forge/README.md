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

Documents exist only in memory; no Save/Open, user files, network or cloud service
is used. Closing discards the study. Geometry/picking uses opaque box proxies,
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
