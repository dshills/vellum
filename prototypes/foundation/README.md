# Disposable Cogent foundation experiment

This macOS arm64 fixture checks P1–P3 extension boundaries. It is not Vellum's production core. See [the results](../../docs/reports/FOUNDATION_RESULTS.md) and [ADR 0002](../../docs/decisions/0002-foundation-gaps.md).

## Build and run

Requires Go 1.27.1, Rust/Cargo (edition 2024, Rust ≥1.87), Xcode/macOS SDK, Python 3 and `install_name_tool`. Dependency downloads require network access. Recorded versions and source hashes are in [evidence/manifest.json](evidence/manifest.json). Ensure Cargo is on PATH; this machine used `$HOME/.cargo/bin`.

```sh
./run.sh test
./run.sh build
./run.sh run -report "$PWD/evidence/my-run.json" -exit-after 90s
./run.sh run -accessibility=false -idle-sample 60s -exit-after 67s \
  -report "$PWD/evidence/my-idle.json"
./run.sh run -lifecycle -exit-after 18s -report "$PWD/evidence/my-resize.json"
```

The runner isolates Go and Cargo caches under `/tmp`; override `P0_CACHE_DIR`, `FOUNDATION_CARGO_HOME` or `FOUNDATION_RUST_TARGET` if needed. `env.sh` narrowly permits the owned executable-relative linker rpath. The vendored AccessKit C source is unmodified; Cargo.lock pins its transitive dependencies. Rust builds both static and dynamic artifacts, but this fixture packages only the dylib: linking both AccessKit and wgpu statically produced a duplicate Rust runtime symbol.

Run the built bundle executable when capturing stdout and flags. Its dylib resolves from `Contents/MacOS/native`, without a development-directory runtime path. The bundle is for local experimentation; signing, notarization and clean-machine distribution remain untested. `build/` is ignored. Tests validate snapshot ownership/graph structure, removed-row rejection and 1,000 lossless actions through Cogent's real event deque; they do not load the native bridge.

## Native checks

Inspect the window with macOS accessibility tooling. The adapter activates when queried. Native button, checkbox, field and row actions should update the visible status and JSON report. Set value to `22k`, rotate, toggle off, rotate again and page forward: expect one accepted rotation, one rejection, value `22k`, and eight replacement rows starting at R9. Collapse then expand the tree; use the canvas object's native Pick action. Native multiline set-value is supported as a whole buffer replacement, with Cogent's observed newline normalization.

The native adapter publishes owned snapshots with stable IDs and enqueues unique custom events. Callbacks never access widgets. Attach/update/free run on Cocoa's main thread while the NSView is alive. This fixture attaches through Cogent's public OnShow hook; upstream AccessKit recommends attaching before the first show/focus. Production use requires a proper creation/focus/lifetime hook and bounded action transport. Text ranges/selection/caret and native virtual-range discovery are absent.

For deletion probes, focus the single-line field, select all and paste the probe. Inspect its exact value before Backspace; then invoke **Record text using its native accessibility action**, which preserves keyboard focus. The keyboard Record shortcut clicks the visible button and can move focus, so do not use it before deleting. `AB` → `A` is the control; `é` and `👩‍💻` should each delete as a whole grapheme, but the pinned toolkit leaves `e` and `👩‍` respectively. This is a recorded P2 failure.

The fixed editor corpus includes Latin, Arabic, Hebrew, Indic, CJK, combining sequences and emoji, using uncontrolled system font fallback. It is not a validated font/shaping corpus. Real Japanese/Chinese composition, mixed-direction caret/selection, screen-reader completion, scale/monitor changes and Windows/Linux adapters still need execution. Do not infer those passes from native AX inspection. VoiceOver and system input-source settings were not changed automatically.

`-idle-sample` logs one whole-process Getrusage CPU interval after three seconds settling. Canvas callbacks are a repaint proxy, not compositor presents or wakeups. Memory is HeapAlloc after forced GC plus Mach current resident size; private/native/GPU memory is not isolated. The report also writes a sampled heap profile. Shown-callback timing is not presentation timing. These diagnostic runs do not satisfy the plan's five-run performance workload.
