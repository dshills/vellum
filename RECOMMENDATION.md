# Recommendation: a Go-native professional workbench

Research date: 9 October 2026. Decision status: **proceed to bounded architectural experiments; do not begin production framework development yet.**

Current experiment status: P0 selected **D (extend Cogent Core)** for the bounded foundation trial, with A (Gio) as control and B as fallback. The B hypothesis below records the original research recommendation; it is superseded for the current experiment by [ADR 0001](decisions/0001-p0-foundation.md). [Remaining foundation checks](FOUNDATION_CHECKS_RESULTS.md) and [exact-text/bounded-mailbox contracts](FOUNDATION_CONTRACTS_RESULTS.md) keep G1 and workbench expansion on hold; no production foundation/API is accepted.

## Recommended direction

Design Vellum as an ordinary-Go, immediate-style workbench API backed by retained UI records, an explicit command model, and an event-driven scheduler. Keep application domain state independent. Make custom GPU rendering, text editing, native accessibility, and deterministic tests core requirements. Put optional AI/automation transports above commands and semantics.

The leading implementation hypothesis is **Architecture B**: an independent Go UI core using go-text for text, an existing GPU abstraction, a reused window/input layer, and AccessKit adapters. Prototype SDL3 with wgpu-native through the maintained Go binding; compare SDL3 GPU before selecting a renderer. This is a conditional recommendation based on integration boundaries, not proof that this stack is faster or cheaper than Gio.

**Architecture A (full Gio) and D (extend Cogent Core) must be tested first as lower-cost controls.** If either passes the hard requirements with modest extensions, build the workbench there. Do not build a new core merely to obtain preferred syntax. Gio already solves much of the Go immediate UI problem; Cogent Core already couples Go UI with WebGPU. [Gio API][GIO-APP], [Gio GPU boundary][GIO-GPU], [Cogent Core widget model][CC-WIDGET], [Cogent Core dependencies][CC-MOD]

## Why the research justifies this direction

PhotoCraft's professional behavior comes from a separated editing engine, commands, snapshot jobs/history, custom panels, a design system, and specialized GPU composition. Egui supplies useful UI mechanisms, but it does not account for all of those features. That combination is the relevant target. [Application][PC-APP], [commands][PC-COMMANDS], [jobs][PC-JOBS], [canvas][PC-GPUCANVAS]

The inspected immediate-mode systems retain identity, interaction, layout/text caches and window state. Slint demonstrates how retained dependencies and geometry can avoid unnecessary work. A hybrid internal architecture is therefore a practical synthesis, not a novel claim that immediate UI has finally discovered state. [Egui memory][EG-MEMORY], [Gio state][GIO-BUTTON], [Slint partial rendering][SL-PARTIAL]

Gio's public renderer processes Gio operations; its generic driver is internal, and desktop accessibility/native DnD remain material gaps in this source audit. That makes an independent renderer and platform adapter worth testing. It does not prove that a Gio extension is infeasible. [GPU interface][GIO-GPU], [internal driver][GIO-DRIVER], [semantic roles][GIO-SEMANTIC], [Wayland handlers][GIO-WAYLAND]

## What to build, reuse, and defer

| Area | Build in Vellum | Reuse | Explicit boundary |
|---|---|---|---|
| API and state | Small Go packages, typed widget/edit state, stable IDs, explicit errors and resource ownership | Go language/runtime and standard tooling | No required DSL, reflection-based binding, or application code generation |
| Workbench | Dock tree, panes, toolbar/menu/palette integration, virtual table/tree, inspectors, themes | Existing layout/control implementations if A/D wins | Not an image editor, CAD kernel, PCB model, or IDE engine |
| Rendering | Ordered display list, modest UI batching/renderer, damage/caching rules, custom surface contract | wgpu-native or SDL3 GPU | No new Vulkan/Metal/D3D backend family or shader compiler |
| Text | Editing, caret/selection/IME integration, cache ownership and font policy | go-text, Unicode helpers, established raster components | No new shaping engine; advanced code/rich editor separately scoped |
| Platform | Thin ownership/lifecycle/service adapters, native menu/notification gaps | SDL3 or validated Gio shell; OS facilities | Do not reimplement all window systems; desktop first |
| Accessibility | Semantic schema and updates, Go bridge, custom-widget conventions | AccessKit native adapters | A node tree alone does not establish screen-reader support |
| Tasks | Scoped task service, cancellation/progress/results, UI wakeups | Goroutines, contexts, channels, ordinary functions | No hidden goroutine per widget; no arbitrary cross-thread mutation |
| Commands | Typed dispatcher, enablement, target scope, history hooks | Application's domain operations | Framework does not automatically make every edit reversible |
| Testing | Logic harness, deterministic clock, stable lookup, event replay and benchmarks | Native graphics/test tooling where appropriate | UI, domain, GPU and OS tests remain distinct |
| Automation | Optional CLI/socket/pipe/MCP adapters | Same command and semantic layer | No AI dependency in rendering; no automatic full application-state export |

## UI lifecycle and ownership

Start with **one serialized UI owner across windows**, coordinated with the native main thread. Background workers never mutate live widgets or documents directly. Keeping a single owner is a deliberate simplification; parallel layout and independent window owners are deferred until evidence warrants their synchronization cost.

The proposed lifecycle is:

1. Drain bounded native events and worker completions; route actions against the last presented hit/semantic geometry and current focus/modal scope.
2. Validate commands against current state and commit domain/UI changes once. Reject stale worker results by explicit document revision and lifetime token.
3. Invoke application UI declaration once for each invalidated window, producing frame records keyed by stable IDs. A returned click is a consumed action, not a condition re-fired during measurement.
4. Solve measurement/arrangement internally, using retained cache keys; never rerun arbitrary application side effects to resolve layout.
5. Produce paint and semantic updates, encode shared-device GPU work, and present according to the frame scheduler.
6. Sleep until input, a worker wakeup, a timer, an animation deadline or platform repaint requires work.

This is a proposed contract to test, not an API promise. A retained record can be a compact struct rather than a heap object with callbacks. Use stable entity keys, not visible text, row indices, memory addresses or localized labels. Explicitly retain focused editors and captured drag sources when offscreen; expire other transient state under a bounded policy. Persist only versioned user-facing layout/preferences.

Cache invalidation should distinguish **layout**, **paint**, **semantics**, and **platform effects**. A tooltip fade should not rebuild the document model. A value change can require semantic output even if the window is occluded. Accessibility queries must not depend on a new paint occurring.

## Custom drawing contract

Offer ordinary high-level 2D canvas drawing and an advanced renderer-specific package. The latter should provide:

- Shared device/queue access with an explicit backend/version and capability set.
- A preparation phase for uploads, compute and offscreen passes; an ordered composition phase for UI/canvas rendering.
- Defined physical size, logical scale, clip/transform, sample count, target format, premultiplied-alpha and color-space conventions.
- Persistent resources with explicit `Close`/release, deferred GPU-safe destruction and device-generation checks.
- Bounded frames in flight and asynchronous readback; no hidden CPU round trip for normal composition.

Applications own geometry, texture content, spatial indexes, culling/LOD and domain-specific color transforms. Framework callbacks cannot independently present its swapchain or retain transient pass handles after the callback. Device loss must produce an explicit recoverable application state; a software fallback is a separate capability, not an untested promise. Egui-wgpu and PhotoCraft show why ordered preparation and retained resources are useful. [Callback contract][EG-WGPU], [canvas resource management][PC-GPUCANVAS]

## Tasks and commands as Go advantages

A task is started by an accepted command or explicit application event. The caller supplies `context.Context`, immutable inputs and an execution function; receives a task handle; and renders its state. Progress delivery is coalesced and bounded. Completion is lossless while the owner exists. Cancellation remains cooperative, and closing a document invalidates pending commits even if the worker cannot stop immediately.

Keep the task API separate from `*UI` so tasks can be exercised headlessly. An API analogous to `tasks.Start(ctx, spec, work)` is safer than a `ui.Go` call whose placement in a draw closure can accidentally spawn on every frame. This is a naming/lifecycle illustration, not implemented code. Panic recovery is confined to the worker boundary and never substitutes for transactional mutation safety.

Commands have stable IDs; typed argument/result definitions; target document/window scope; availability with disabled reasons; check state; execution and optional cancellation; and history/automation policy. Menus and palettes bind to command references. A registry can expose explicit schemas for IPC without making ordinary Go calls serialize through JSON. Both human and automation invocation recheck preconditions at commit time. [PhotoCraft command precedent][PC-COMMANDS], [snapshot job precedent][PC-JOBS], [Qt actions][QT-ACTION]

Undo belongs to the domain transaction layer. The framework can group drag gestures, label/coalesce edits, expose history state, and ensure consistent command entry. It should not clone arbitrary application graphs or claim any callback can be undone.

## Semantics, accessibility, and AI

Provide stable role/label/value/action records independent of paint buffers. Add relations, focus, selection, validation errors, table coordinates/ranges, and text-position mappings as needed. Widgets supply standard semantics; custom canvases can supply a semantic projection of selected/visible domain objects rather than one node per drawn primitive.

Accessibility adapters consume that projection and return actions through the UI owner. Optional automation also uses it, plus an explicit domain query/command interface. An inspector field may expose `component.value`, but the framework cannot infer electrical meaning, unit constraints, authorization or whether a PCB edit is safe. Applications must declare those rules.

An opt-in local IPC adapter should have a versioned schema, request IDs, target/revision checks, bounded payloads and selected command exposure. Unix sockets and Windows named pipes can use OS access controls; HTTP/MCP requires authentication and local-origin policy. Keep sensitive values redacted and preserve an auditable command/result path. PhotoCraft and Slint demonstrate this architectural opportunity, so differentiation depends on robust Go integration and guarantees. [PhotoCraft UI control][PC-CONTROL], [Slint MCP implementation][SL-MCP]

## Largest risks and reversal conditions

| Risk | Evidence / uncertainty | Required experiment and decision |
|---|---|---|
| Internationalized editing | Shaping is available; editor/IME/native-index integration is not automatically solved | P2: real Japanese/Chinese IME, Arabic/Hebrew mixed text, Indic/emoji selection across target OSes; stop UI API stabilization on failure |
| Desktop accessibility | Gio bridge gap; AccessKit Go integration unproven; virtual tables/editors require richer semantics | P3: real VoiceOver, NVDA and Orca behavior; no “accessible” claim based only on tree inspection |
| GPU/native lifetime | Binding ABI, surface destruction, color formats, resource disposal and device loss | P1/P4/P7: shared-device custom work and lifecycle stress on Metal, D3D12 and Vulkan |
| Ownership and GC | Goroutines do not enforce immutability; foreign memory invisible to Go heap limits | P5/P8/P11: stale-result tests, race testing, allocation/retention measurements, cancellation under load |
| Platform and packaging | Linux protocols vary; multiple native libraries and CGO complicate builds | P1/P10: clean-machine packages, Wayland/X11, mixed DPI, dialogs/DnD/clipboard; document minimum supported matrix |
| Duplicate ecosystem investment | Gio/Cogent Core/Fyne/GoGPU cover substantial territory | P0: choose an extension if it passes with less maintenance; do not optimize solely for API taste |
| Sustainable maintenance | Desktop support is permanent work, not a one-time bootstrap | Require named owners for rendering, text/platform and testing before a production commitment |

## Answers to the brief's success questions

1. **Technically justified?** Feasible; a new core is conditionally justified by unmet professional integration requirements, not yet economically validated.
2. **Reuse what?** Text shaping/fallback, GPU abstraction/backends, native window/input services, accessibility adapters and Go infrastructure.
3. **Build what?** Missing workbench controls, explicit ownership/lifecycle, commands/tasks, semantic/test contracts, and the minimum core/adapters needed to join them.
4. **UI model?** Immediate-style Go authoring with deliberately retained internal state and event-driven updates; evaluate retained existing alternatives fairly.
5. **Rendering?** Shared-device ordered rendering over an existing native GPU abstraction; prototype wgpu-native against SDL3 GPU before committing.
6. **Text?** Go-text plus a carefully engineered editor/IME/cache integration; no new shaping engine.
7. **Platform integration?** Reused platform layer with narrow OS-specific service bridges, explicit capability reporting and real native tests.
8. **Concurrency advantage?** Yes, through scoped cancellable tasks and typed result commits; goroutines alone are not differentiation.
9. **AI controllability?** Commands, semantic projections and explicit domain queries make it practical, but applications still declare intent and constraints.
10. **Largest risks?** Internationalized editing, accessibility, GPU/native lifetimes, platform coverage and maintenance economics.
11. **Required prototypes?** P0–P11 in the plan, with native/text/accessibility gates ahead of broad widget development.
12. **Meaningfully better than existing Go options?** Only if it makes complex editor workflows, direct GPU canvases, native accessibility, async work and automated testing measurably easier together while meeting defined latency/deployment budgets.

Vellum should look like a coherent Go workbench platform with replaceable low-level foundations. The next investment is evidence that this architecture can be delivered and maintained—not a production widget catalog.

<!-- Source reference definitions -->

[CC-MOD]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/go.mod
[CC-WIDGET]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/widget.go
[EG-MEMORY]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/memory/mod.rs
[EG-WGPU]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui-wgpu/src/renderer.rs
[GIO-APP]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/app.go
[GIO-BUTTON]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/widget/button.go
[GIO-DRIVER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/internal/driver/driver.go
[GIO-GPU]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/gpu.go
[GIO-SEMANTIC]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/semantic/semantic.go
[GIO-WAYLAND]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_wayland.go
[PC-APP]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/lib.rs
[PC-COMMANDS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/commands.rs
[PC-CONTROL]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/control.rs
[PC-GPUCANVAS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/gpu_canvas.rs
[PC-JOBS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/jobs.rs
[QT-ACTION]: https://doc.qt.io/qt-6/qaction.html
[SL-MCP]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/testing/mcp_server.rs
[SL-PARTIAL]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/partial_renderer.rs
