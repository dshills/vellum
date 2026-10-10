# Vellum P0: existing Go foundations

Disposable comparison, recorded 9 October 2026. Two native engineering-workspace prototypes share one Go domain model. These are experiments, not the initial production framework.

Read [P0 results](../../docs/reports/P0_RESULTS.md) and [the provisional decision](../../docs/decisions/0001-p0-foundation.md).

## Run

From this directory, with Go 1.27.1 and a working native CGO toolchain:

```sh
./run.sh gio
./run.sh cogent
./run.sh gpu
./run.sh test
./run.sh build
```

On macOS, the tested toolchain is Xcode with the macOS SDK. Module downloads require network access to the Go proxy/checksum service. `env.sh` isolates module/build caches under `/tmp/vellum-p0-cache`; override `P0_CACHE_DIR` if desired. `GOWORK=off` keeps the two dependency graphs independent. The minimum module Go versions differ; the recorded run used 1.27.1 for both.

Flags pass through to the selected application:

```sh
./run.sh gio -rows 100000 -exit-after 30s -report /tmp/vellum-gio.json
./run.sh cogent -rows 100000 -exit-after 30s -report /tmp/vellum-cogent.json
./run.sh gpu -exit-after 30s
```

`-rows` must be at least 96. `-report` writes structural counters on close; it is not a performance benchmark. Cogent also accepts `-trace` for event/focus diagnostics. Native binaries and local macOS app bundles from the recorded builds are under ignored `build/`; they are developer artifacts, unsigned for distribution and not notarized. The source runner rebuilds from the pinned modules. No private packages or modified toolkit source are required.

## Shared fixture

Each board has 100,000 deterministic components by default. Select a row or tree entry, edit its inspector value, rotate it, sort/filter the projection, and run/cancel the checksum worker. Selection uses component IDs, not row positions. The tree displays at most the first 64 matching components; the canvas displays 96 component glyphs. The worker deliberately pauses 5 ms per 256 components to make cancellation and stale-result handling observable.

The two document tabs preserve each board's edits and selected ID. Filter/sort and task status belong to the shared session; changing tabs does not create a second task. Filtering scans the model and sorting runs synchronously on explicit updates. This is not P6's million-row indexing experiment. No docking, undo, file persistence, task pool or semantic automation transport is implemented.

## Controls

Gio: use the visible buttons and fields; Cmd+P toggles its inline command palette. Cogent provides the same visible actions plus these shortcuts on macOS (Meta on other platforms, untested):

| Shortcut | Action |
|---|---|
| Cmd+E / Cmd+R | Commit inspector value / rotate |
| Cmd+1 / Cmd+2 | Sort by reference / value |
| Cmd+B / Cmd+K | Start / cancel calculation |
| Cmd+0 / Cmd+I | Focus filter / inspector value |
| Cmd+P | Toggle searchable command palette |
| Cmd+3 / Cmd+4 | Switch to Board A / Board B |
| Cmd+Shift+X | Finish smoke run |

After opening/focusing a Cogent editor, allow its next frame to apply the deferred focus before injecting paste. In the palette, Tab then Return invokes the remaining matching command. Commands revalidate availability in the shared model even if a shortcut button remains enabled. Real multilingual IME behavior is untested.

## Evidence

`evidence/manifest.json` records pins, platform scope, source/artifact hashes and verification. Selected screenshots, native accessibility snapshots, package lists, native-library lists and logs are retained. Earlier diagnostic logs are explicitly labeled in the manifest; they include prototype bugs that were fixed. Do not use them as evidence of the final behavior.

The native automation cursor query differs between Gio and Cogent/GLFW: Cogent's injected pointer clicks landed at the physical cursor location rather than the injected event coordinate. Keyboard interaction was used to verify Cogent. Real mouse scrolling/selection/resizing in Cogent still requires manual validation. Native accessibility snapshots exposed window chrome only for both candidates. Neither is accepted for production.
