# Building a Mac design workbench

[Forge](../examples/forge/README.md) is a runnable consumer in its own Go module.
The owner authorized this post-release track; existing v1 signatures, published
artifacts, rendering/pacing defaults and release/performance policy are preserved.
These additions do not retroactively qualify new workloads or published builds.

## Reusable public pieces

| API | Responsibility and limit |
| --- | --- |
| `ui.Ribbon` | Stable tabs/groups/action identities, contextual tabs, full-color icon-above-label buttons and whole-group overflow; retained per owner/window. |
| `ui.ModelBrowser` | Separate disclosure, visibility, selection and activation proposals; up to 256 items, at most 16 ancestors. |
| `viewport.Camera`, `Ray`, `Box`, `Gesture` | Portable camera matrices, scale-consistent rays, box-proxy picking and sequence-aware click/drag admission. |
| `viewport.Scene`, `cogent.SceneView3D` | Experimental depth-rendered preview of up to 256 opaque box/cylinder display primitives with box picking; the same retained GPU device/resource contracts as `GPUView`. |
| `ui.ViewCube` | Keyboard/accessible view choices using the same face identifiers as the rendered in-canvas cube. |
| `quantity.Length`, `ui.QuantityField` | Bounded arithmetic and named length parameters, canonical millimeters, explicit Apply and application validation/command admission. |
| `ui.FeatureTimeline` | Controlled feature selection, earlier/later reorder and rollback-marker proposals. No solver or dependency replay. |
| `ui.MarkingMenu` | Visible eight-direction action panel, managed focus and keyboard buttons; no invisible flick gestures, nested radial tracking or pointer-anchored native popup. |

Use keyed literals. Keep each adapter/control and its model on one application
owner. Mount a fresh GPU adapter for another native window. Never modify the
model concurrently, reenter `Application.Do` from an owner callback, or release
borrowed WebGPU resources. Revisions must advance when scene data changes.

```go
scene := &viewport.Scene{
    Camera: viewport.DefaultCamera(), Grid: true, ShowCube: true,
    Objects: []viewport.Object{{
        ID: "body-1", Name: "Body", Visible: true,
        Bounds: viewport.Box{
            Min: viewport.Vec3{X: -1, Y: -1, Z: -1},
            Max: viewport.Vec3{X: 1, Y: 1, Z: 1},
        },
        Color: color.NRGBA{R: 60, G: 120, B: 220, A: 255},
    }},
}
surface := &cogent.SceneView3D{Owner: application, Window: "design", Model: scene}
// Return ui.Node{ID:"scene", Kind:ui.Extension, Grow:true, Custom:surface}
// from that registered window's View. Selection is a proposal:
surface.Select = func(id string) {
    scene.SelectedID = id
    _ = scene.Touch()
    _ = application.Invalidate("design")
}
```

Selection and camera state are transient. Document edits belong in the app's
command registry and `history.Journal`. Forge captures immutable before/after
snapshots so a body has the same ID after Undo/Redo. A marker changes preview
visibility; it does not reconstruct earlier body dimensions. Width values are
millimeters; Forge uses one scene unit for 20 mm.

Length examples: `2 in + 5 mm`, `(1 + 2) cm`, `width / 2`. Variables are lengths;
addition requires matching dimensions. A scalar final expression uses the chosen
default unit. Parsing refuses unknown names/units, code, functions, division by
zero, nonfinite results, excessive nesting/input and non-length results. The
field checks application bounds before Commit. Owners display error messages.

## Input and verification boundaries

Primary drag capture is same-window. Escape, focus/capture loss and outside
release cancel selection; camera cancellation restores the pre-drag view. Window
closure suppresses late callbacks. Shift-drag pans, Option/plain drag orbits,
and scrolling zooms. Select mode leaves drags inert and admits inside clicks.
Secondary/middle events are targeted only and do not capture outside the view.

GPUEvent positions and Size use matching full-content physical pixels; Inside
reflects the visible clip. Cube picking/rendering shares camera/rectangle math.
Clipping, scaling, exactly-once routing and cancellation have focused regressions.

The app's `--smoke` joins actual native GPU readiness/readback and buffer growth
with explicitly synthetic owner-routed interaction/history checks. It has a
60-second watchdog, fixture-only geometry settings and safe shutdown. It does
not establish physical OS targeting, human VoiceOver/IME usability, CAD-scale
performance or another display/hardware configuration. Read the final delivery
report for actual packaged-app pointer checks and limitations.

Private packaging uses `examples/forge/package.py`, an exclusive fresh output,
clean source commit, original procedural icon, ad-hoc signature, ZIP/checksum and
signature verification after exact ZIP extraction. No public download, release
tag, audit change, signing identity or notarization is performed.

## Forge fidelity integration

The remaining presentation pass keeps the original commands and public shapes:
compact inspector fields/actions, centered original cube face names, quiet
orientation alternatives, a balanced eight-action panel, and app-local scroll
containers for short/narrow panes. Source checkpoints are recorded in the
phase14–phase17 review folders under docs/reviews/fusion-workbench.

The complete app-description layout regression runs in the Forge module with
`go test -race -tags offscreen ./...`; CI's scripts/check.sh includes it. It uses
the app's theme factory and a measured placeholder only for the GPU viewport,
so geometry evidence does not claim GPU appearance, physical input, or human
accessibility qualification. Native private-package checks remain separate.

Integration fast-forwards main from the owner-approved branch when possible,
pushes source to the existing repository, and checks the exact commit's push CI.
A private package is then built from clean committed source using the unchanged
producer. Public release tags/assets, audit changes, signing identities and
notarization remain outside this source-integration/private-delivery request.
Final CI outcomes, package hashes and native observations belong to the final
source-bound delivery handoff; they are not preclaimed by this integration guide.

## Completed Forge feedback and delivery

`viewport.Object.Shape` defaults to `viewport.BoxShape`; the experimental
`viewport.CylinderXShape` renders a capped 48-segment cylinder along X, inscribed
in Bounds. Y/Z extents define its cross-section; Forge uses equal radii for the
shaft. All shapes deliberately retain Bounds for picking. This is display geometry,
not CAD topology or exact cylinder intersection. Use keyed literals as above.

Selected geometry keeps its original material with a thin blue outline of picking
bounds; hovered geometry is lighter. Forge's existing tool buttons/panel expose
one active Orbit/Pan/Select mode, with matching hints and caption. The demo keeps
an intentional light application theme under either host palette. Dialog-specific
colors theme native dialog chrome; generic content defaults preserve historical
native chrome. Existing frozen replay images/inputs remain unchanged.

See [the current native screenshot](screenshots/forge-completion.jpg) and its
[capture identity](screenshots/forge-completion-capture.json). This capture comes
from the tested implementation candidate; the following documentation checkpoint
adds no runtime code. Final exact-commit CI, ZIP checksums/signature results and
launch observations belong to the private delivery handoff. Source/review
checkpoints are phase22–phase27 under docs/reviews/fusion-workbench.
