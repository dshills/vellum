# Cogent foundation trial: native accessibility works; acceptance remains on hold

9 October 2026. A bounded **local macOS P1–P3 feasibility trial** followed P0. It establishes a real native accessibility extension seam, reproduces a text correctness failure, and identifies scheduling and memory costs. **G1 has not passed.** Continue with focused remediation feasibility before building the workbench API; see [ADR 0002](../decisions/0002-foundation-gaps.md).

A subsequent [bounded remediation trial](REMEDIATION_RESULTS.md) locally fixes the representative field failure and measures scheduling/spelling improvements. This document preserves the original baseline outcomes.

## What changed

The disposable [foundation fixture](../../prototypes/foundation/README.md) extends the same pinned Cogent revision used in P0, through public APIs, with an AccessKit C bridge. It exposes a button, checkbox, error-bearing property field, expandable tree, eight rows from a synthesized 100,000-row range, multiline editor and two selectable canvas objects. It also includes Cogent's public xyzcore scene. Unlike P0, macOS now exposes actual native controls and accepts their actions.

This smaller fixture does not allocate P0's full domain dataset, implement a general adapter or replace the toolkit's text/backend code. No toolkit fork or private import was needed for the demonstrated seam.

| Requirement | Local outcome | Evidence / limitation |
|---|---|---|
| Native accessibility tree and actions | Demonstrated | Native roles/values/focus, checkbox changes, field edits, tree collapse/expand, table paging/selection and canvas Pick |
| Action integrity and validation | Demonstrated for fixture | 18 callbacks / 18 handled native actions; disabled rotation and empty edit rejected |
| Virtual identity | Demonstrated for visible range | Eight semantic rows replaced by eight new rows; removed IDs reject actions in a unit test; no native off-screen range discovery |
| Grapheme deletion | **Failed** | Backspace splits combining and emoji sequences in ordinary property field |
| Real IME / native text ranges | Open | Marked-text/caret integration gap in inspected source; no live Japanese/Chinese composition or text-range bridge |
| Screen-reader usability | Open | Real AX inspected and invoked; VoiceOver/NVDA/Orca completion not exercised |
| Idle CPU | **Diagnostic target exceeded** | Two 60-second samples about 2.2% of one CPU; zero canvas repaint callbacks |
| Memory | **Diagnostic concern** | About 403–404 MiB retained Go heap in a small fixture; font parsing and spelling dominate sampled profile |
| Native resize / cleanup | Demonstrated locally | Six matching content sizes; normal timed shutdown with attached adapter |
| Relative native library packaging | Demonstrated locally | Relocated bundle runs with development/dylib environment variables removed |
| Full platform/lifecycle/distribution matrix | Open | Windows/Linux, clean machines, signing, DPI/monitor changes, loss/recreation, occlusion and restore unverified |

## P3: real native bridge

The adapter obtains the exported GLFW Cocoa view/window from Cogent's public desktop window, attaches AccessKit C and uses `RunOnMain` for Cocoa operations. Native callbacks read mutex-protected owned snapshots; they never inspect widgets. A numeric `runtime/cgo.Handle` identifies bridge state without retaining a Go pointer in native memory. Action payloads are copied before the native request is freed, queued to the UI, validated against current nodes, and dispatched through the same command validation used by visible controls. The adapter is freed before view destruction and the handle deleted afterward.

The [native AX capture](../../prototypes/foundation/evidence/native-verified-ax.txt) shows controls, selected canvas R18, the replacement R9–R16 rows, multilingual editor values and focused property field. In the [verified report](../../prototypes/foundation/evidence/native-verified.json), value `22k` is retained after an empty edit is rejected; one rotation succeeds and another is rejected while disabled. Native multiline set-value works, with two newline characters appended by the observed editor buffer normalization. It is not an exact-preservation pass.

Two integration issues were resolved in this disposable adapter:

- Linking AccessKit and wgpu as static Rust libraries produced a duplicate `_rust_eh_personality` symbol. AccessKit now uses a bundled dylib with an executable-relative rpath; wgpu keeps its existing linkage. Runtime dependencies are recorded in [native-libraries.txt](../../prototypes/foundation/evidence/native-libraries.txt).
- Ordinary Cogent custom events may coalesce consecutive events. An earlier burst lost commands because the adapter used that path. Native commands now enqueue explicit **unique** events. A race-enabled test sends 1,000 actions through Cogent's actual deque and verifies FIFO delivery; the final native run handles all 18 callbacks. This is an adapter correction, not evidence that Cogent's documented compression is broken.

Focus forwarding and explicit view-focus updates make native focus observable. However, AccessKit's Rust subclassing-adapter documentation recommends attachment before the view's first show/focus; the public OnShow hook attaches afterward. Production integration still needs suitable creation/focus/destruction hooks. The prototype's action queue is lossless but unbounded. Full snapshots and explicit field mappings are sufficient here, not an optimized general semantic adapter.

Text ranges, selection/caret, menus inside the application, native virtual-range discovery, assistive-technology announcement behavior and sustained callback/lifetime stress remain unimplemented or unverified. Encoding an invalid field state and polite status does not establish that a screen reader announces them correctly. AX inspection and invocation are useful feasibility evidence, not an accessibility compliance claim.

## P2: reproducible editing failure and native input gap

The fixed corpus includes Latin, Arabic, Hebrew, Indic, CJK, combining characters and emoji using system fallback fonts. Native values contain the characters; complex-script visual shaping, mixed-direction caret behavior and controlled font coverage have not been validated.

For each property-field probe, the exact input was checked in native AX, then one actual Backspace was sent. Native Record text preserved keyboard focus and saved the resulting field value:

| Input | Required whole-grapheme result | Observed result |
|---|---|---|
| `é` (`e` + U+0301) | Empty | `e` — **fail** |
| `👩‍💻` | Empty | `👩‍` — **fail** |
| `AB` (control) | `A` | `A` — pass |

The three outputs are in `text_samples` in the verified report. Pinned `core/textfield.go` deletes by rune position; grapheme navigation/deletion requires a change in the text implementation. A shaping library alone cannot fix this editing contract.

The pinned GLFW Cocoa input client stores marked text locally, returns no attributed substring and supplies a zero-size rectangle at the view-frame origin for `firstRectForCharacterRange`. The inspected Cogent path handles committed character events, without a demonstrated preedit/selection/caret bridge into these Go widgets. This is source evidence of an integration gap, not a substitute for a live IME test. Relevant excerpts are preserved in [source-boundaries.txt](../../prototypes/foundation/evidence/source-boundaries.txt).

Actual Japanese/Chinese composition, cancellation, candidate positioning, Unicode selection, undo grouping, scale changes and clipboard round-trip tests remain required. System input sources and VoiceOver settings were not changed during this trial.

## P1: diagnostic CPU, memory and lifecycle results

Recorded host: Apple M4 Pro, 64 GiB RAM, macOS 26.6.2 / 25G83, external display 5120×1440 at 60 Hz, Go 1.27.1, Rust/Cargo 1.99.0 and Xcode 27.0. See [environment.txt](../../prototypes/foundation/evidence/environment.txt). Power mode and compositor presentation timestamps were not recorded. The initial content size is 1120×960 physical pixels; this is not the plan's 1,000-control benchmark.

| 60-second static sample | CPU seconds | % of one logical CPU | Canvas repaint callbacks | HeapAlloc after forced GC | Current process resident |
|---|---:|---:|---:|---:|---:|
| Adapter disabled | 1.339 | 2.232% | 0 | 403.0 MiB | 760.5 MiB |
| Adapter enabled and queried | 1.322 | 2.204% | 0 | 403.1 MiB | 747.0 MiB |

Each sample starts after three seconds settling, with no focused text caret, commands, animation or background domain tasks. CPU is whole-process `Getrusage` time divided by elapsed wall time. These are one diagnostic interval per condition, not a statistically controlled comparison. The small difference does not establish an adapter overhead estimate. Raw [baseline](../../prototypes/foundation/evidence/idle-baseline-sample.json) and [adapter-active](../../prototypes/foundation/evidence/idle-accessibility-sample.json) samples are retained.

Both exceed the provisional <0.5% idle CPU target numerically. Zero canvas repaint callbacks do not mean zero presents or wakeups; neither was instrumented. Cogent's pinned base window loop starts an unconditional refresh-rate ticker and sends paint events until closure. That is a concrete obstacle to the intended ≤2 app-originated wakeups/s contract and a plausible source of idle work, but CPU attribution requires further profiling.

The interactive report retains 403.8 MiB HeapAlloc after forced GC. Its sampled [heap profile summary](../../prototypes/foundation/evidence/heap-top.txt) attributes about 213 MiB flat to OpenType table-buffer loading and about 115 MiB cumulative to spell-suggestion key construction. Font parsing accounts for about 252 MiB cumulatively. Profile estimates are sampled; do not add overlapping cumulative totals. These costs substantially exceed the plan's 40 MiB empty-workbench heap target, but this fixture includes a multilingual editor and spelling defaults, so it is not a formal baseline benchmark. Test lazy spelling and bounded font loading with preserved text coverage before deciding whether the cost can be reduced through maintainable hooks.

Resident size includes Go, native libraries and other mapped/native resources; private and GPU allocations were not isolated. Startup OnShow callbacks ranged roughly 337–400 ms in these runs, with the native lifecycle run at 354 ms. These are neither presented-frame measurements nor five-run percentiles and cannot pass/fail the startup budget.

The [lifecycle report](../../prototypes/foundation/evidence/lifecycle.json) records six public resize requests matching content sizes: 1000×850, 1180×980, 950×820, 1120×960, 980×840 and 1120×960. The trace records each native resize and normal closure. A minimize action was attempted near the timed exit; restore was not verified, so minimize/restore remains open. Device loss, surface recreation, DPI changes and long lifecycle stress also remain open.

The rebuilt `.app` ran from `/tmp/vellum-foundation-relocated` with Go/Cargo cache and dylib search variables removed, then closed normally; see [relocated.json](../../prototypes/foundation/evidence/relocated.json). Its accessibility adapter was disabled for that packaging smoke test. This validates local relative-library loading, not clean-machine installation, signing/notarization, a fresh font cache or another OS. Windows and Linux builds were not run.

## Reproduce and decide next

[run.sh](../../prototypes/foundation/run.sh) builds the locked Rust snapshot/dylib, vets and builds Go, packages the local app, and runs the fixture with explicit flags. Four race-enabled tests pass: malformed semantic graphs, removed-row action rejection, snapshot ownership including mutable bool pointers, and 1,000 ordered unique commands. Native bridge behavior was tested separately through the real macOS adapter; unit tests do not establish native lifetime safety.

The [manifest](../../prototypes/foundation/evidence/manifest.json) records pins, hashes, evidence scope and earlier diagnostic artifacts. Current sources, notices, logs, profiles and screenshots are retained alongside the runner. The bounded next task is to determine whether text/IME/accessibility hooks, demand-driven scheduling and memory controls can be implemented as supported extensions or small upstream changes. If they require broad replacement or a persistent private backend fork, test B against the same gates. Keep D provisional until that evidence and the platform matrix justify acceptance.

Primary source references: [Cogent native window](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/system/driver/desktop/window.go), [paint scheduler](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/system/driver/base/window.go), [property-field editing](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/textfield.go), [event compression](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/events/deque.go), [AccessKit C API](https://github.com/AccessKit/accesskit-c/blob/8b6ed37c20ed4c59390e253407983333053662ba/include/accesskit.h), and [AccessKit macOS adapter lifecycle](https://docs.rs/accesskit_macos/0.27.1/accesskit_macos/struct.SubclassingAdapter.html).
