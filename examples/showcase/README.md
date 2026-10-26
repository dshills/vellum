# Vellum showcase apps

Three small macOS applications that demonstrate the existing public framework and
native driver APIs. These are working, in-memory examples—not screenshot mockups,
production design tools, photo managers or monitoring services.

From a prepared checkout or SDK, source `drivers/cogent/env.sh`, then:

```sh
cd examples/showcase
go run ./cmd/forma
go run ./cmd/luma
go run ./cmd/pulse
```

Use a window around 1400×880 for the full three-column compositions. The sidebar
and inspector have fixed widths; these desktop examples do not promise a compact
phone-size layout. No network access, user documents or personal photos are used.
All changes reset on launch. Closing a window exits its application.

## Forma — design studio

Select a layer in the sidebar or click a shape on the canvas. Change its color,
choose a size and click Apply size, nudge it, or add a new circle/square. Drag the
canvas to pan; scroll to zoom; Fit artboard restores the composition. Cmd-N adds
a circle; Cmd-Z / Shift-Cmd-Z use the same bounded command/history path as the
native full-color toolbar. Undo and Redo restore the same object identities.

The poster combines original generated typography with retained GPU vector
geometry and an image layer. Shapes are edited through the inspector; dragging
pans the canvas, rather than moving a shape. The visible corner marks indicate
selection and are not resize handles. This does not replace the separate
[placement example](../placement/README.md), which demonstrates palette dragging,
marquee selection, direct resize handles, files and grouping.

## Luma — image library

Choose images in the filmstrip, search, filter collections or favorites, and
change the active preview's Original/Warm/Mono look. Command/Shift selection
uses the collection's public selection proposals; the most recently selected
image drives the primary preview. Original image data stays immutable. Favorites
and color choices are in-memory view state, without an application Undo log.

Every landscape is original, deterministic mathematical artwork generated in Go.
The RGB histogram samples the actual selected image/variant. There is no photo
import, file management or destructive editing. Assets are generated before the
UI opens and shared as bounded immutable images; no decoder work runs in View.

## Pulse — telemetry explorer

Browse a virtualized source of 1,024 indexed fixture records. Filter services,
select a row to inspect it, change the chart time range, or click the chart to
inspect a sample. Sort by name or latency. Next sample / Cmd-R advances the
repeatable synthetic dataset; Reset filters / Cmd-0 restores its view settings.

The charts are retained GPU meshes, not web views. The data is entirely fictional:
there are no background timers, network calls, host metrics or implied production
availability/performance measurements. Read-only row selection uses the scoped
Mac palette; editable grids retain their distinct active-cell cursor.

## Build application bundles

The native toolbar uses AppKit, original full-color PNG icons and standard window
chrome. The managed UI uses Vellum's Mac theme. GPU MeshCanvas remains within the
framework's documented experimental graphics scope; these apps do not expand its
stable API or claim performance/release qualification.

After sourcing the driver environment, from this example directory:

```sh
python3 package.py --output /private/tmp/vellum-showcase
```

The output directory must not already exist. This builds Forma.app, Luma.app and
Pulse.app, original 1024px bundle icons, private ad-hoc signatures, ZIPs, hashes
and a manifest. The current private packaging flow does not grant public binary
publication approval or notarize apps. Prefer local build/extraction storage;
this host's synchronized Documents area adds metadata rejected by codesign.
The full source/README/screenshots may be reviewed independently of app binaries.

## Checks and screenshots

```sh
go test -race ./...
go vet ./...
# Public driver routing checks; no desktop windows:
go test -race -tags offscreen ./...
```

Model regressions exercise Undo/Redo identity, rejected edits, bounded objects,
collection proposals/filtering/favorites, immutable image looks and indexed grid
revisions/projections. Offscreen routing mounts the managed controls with custom
GPU extensions replaced by empty layout nodes: that harness explicitly refuses
unqualified custom surfaces. Native screenshots and checks are recorded separately in
[the screenshot guide](../../docs/screenshots/README.md); automated routing does
not establish physical-pointer or human accessibility usability.
