# Architecture comparison

Research date: 9 October 2026. See [SOURCES.json](SOURCES.json) for exact source revisions and [RESEARCH.md](RESEARCH.md) for traced implementations. Ratings describe architectural fit, not measured speed or product-quality certification.

## Core framework matrix

PhotoCraft is included as an application reference; its engine capabilities must not be attributed to egui. Dear ImGui docking/viewports refer to the separately pinned docking branch. Slint's renderer and platform choices change its behavior.

| Dimension | PhotoCraft | egui | Gio | Dear ImGui | Slint |
|---|---|---|---|---|---|
| **UI model** | Editor UI over egui; independent engine | Immediate developer API with retained memory/pass state | Immediate composition with explicit state structs and ops | Immediate calls into retained context/window state | Declarative components, retained item trees and properties |
| **Rendering** | egui chrome plus custom wgpu canvas/compositor; CPU fallback | Epaint shapes → tessellation → renderer integration | Gio ops → path/coverage/texture renderer; Metal/Vulkan/GL/D3D11 | Draw lists → caller-supplied backend | Multiple renderers, including software, Skia and FemtoVG |
| **Layout** | Custom editor chrome and fixed/grouped panel model | Local cursor/directional/grid layout; occasional extra pass | Constraints, Flex/Stack, virtual List | Local cursor layout; tables and explicit sizing | Retained box/grid/flex solving and invalidation |
| **State** | Session/document snapshots; separate UI/view state | Application data separate from ID-keyed widget state | Application-owned Clickable/Editor/List state; router retention | Application data plus IDs, storage, table/window state | Item/property/model state and dependency tracking |
| **Input** | Egui plus shortcut, tool, synthetic-input and native-service code | RawInput → interaction/focus; integration emits native effects | Tagged filters/router, commands, pointer and key gestures | Backend-fed event queue, capture flags, active/nav IDs | Retained item dispatch and backend translation |
| **Text** | Egui UI text plus domain text engine | Current source has HarfRust shaping; full bidi remains a gap | Go-text shaping/font scanning, bidi, editor machinery | Fonts/UTF-8 editing; upstream declares full shaping/bidi unsupported | Renderer-dependent text stack and retained text items |
| **Docking** | Custom serializable right-panel groups | Ecosystem/custom implementation, not core docking | Build above primitives | Dedicated docking branch with dock tree/settings | No general IDE docking solution established by this audit |
| **Custom rendering** | Direct example of custom GPU passes/resources | Strong egui-wgpu callback seam | Rich 2D ops; generic GPU/compute not public; CustomRenderer replaces ownership | Draw lists, textures, callbacks; host owns GPU | Rendering hooks/custom items subject to backend constraints |
| **Accessibility** | Enables AccessKit through eframe; custom controls need audit | AccessKit output/integration; custom semantics required | Limited semantic vocabulary; inspected OS bridge found only on Android | Upstream explicitly lists accessibility unsupported | AccessKit translation in winit backend; custom roles/actions required |
| **Platform abstraction** | Eframe/winit plus app-specific native services/tablet code | Integrations such as winit/eframe | In-tree Windows/AppKit/Wayland/X11 drivers | Platform backend split from renderer; host supplies shell | Platform/WindowAdapter/Renderer separation |
| **Testing** | Engine tests, synthetic controls, kittest and snapshots | AccessKit-based kittest harness plus GPU screenshots | Layout/router/widget tests and GPU headless path | Core hooks and separately licensed Test Engine | Preliminary testing backend; mock time; current MCP inspection support |
| **Concurrency** | Worker snapshots → session commit, cooperative cancellation | App-managed workers and repaint callback; synchronized context internals | Goroutines and concurrent-safe invalidation; widget mutation still owned | Same-context calls cannot run concurrently | Event-loop proxy; UI state remains thread-constrained |
| **Performance mechanisms** | Damage regions, shared tiles, caches, GPU resources | Caches, visible rows, batched paints, event-driven repaint | Reusable serialized ops, cached paths/text, virtual lists | Capacity reuse, clipper, table visibility, low-allocation design | Dirty properties/geometry, cached rendering, backend-specific partial updates |
| **Deployment** | Native Rust app plus platform packaging | Rust/native dependencies chosen by integration | `go build` with target native dependencies; CGO varies by OS | C++ plus chosen platform/renderer dependencies | Runtime/compiler/bindings and selected native backend |

Evidence by column: PhotoCraft [frame][PC-APP], [GPU][PC-GPUCANVAS], [jobs][PC-JOBS], [dock][PC-DOCK]; egui [context][EG-CONTEXT], [text][EG-TEXT], [GPU callbacks][EG-WGPU], [tests][EG-TEST]; Gio [app][GIO-APP], [GPU][GIO-GPU], [semantics][GIO-SEMANTIC], [platform source][GIO-ANDROID]; Dear ImGui [core][IM-CORE], [tables][IM-TABLES], [docking][IM-DOCK], [scope][IM-README]; Slint [properties][SL-PROP], [platform][SL-PLATFORM], [accessibility][SL-ACCESS], [testing][SL-TEST].

## Strengths, weaknesses, and lessons

| Subject | Strongest contribution | Principal mismatch with this brief | Architectural lesson |
|---|---|---|---|
| PhotoCraft | Complete editor composition, commands, engine/UI split, custom GPU integration | Application-specific panel/tools/undo design; not a general toolkit; no runtime validation here | Professional quality requires domain, workbench, design system, and rendering working together |
| egui | Small immediate API, extensible painting, useful test/semantic integration | Current internationalized text limits; docking and shell services outside core | Preserve immediate authoring while retaining identity/state/caches internally |
| Gio | Existing Go-native foundation, text, operations, events and platform drivers | Public GPU extension boundary and unvalidated desktop accessibility/OS DnD | Do not rebuild its strengths casually; test its hard boundaries first |
| Dear ImGui | Dense tool controls, docking, tables, simple renderer integration | Declared accessibility/internationalization gaps; host owns native application integration | Virtualize aggressively and make editor primitives first-class |
| Slint | Retained invalidation, components, renderer separation, accessibility/testing | DSL/code-generation and binding model; conditional licensing; no established official Go API | Keep persistent semantic/layout records without importing the whole reactive/compiler architecture |

## Fit against the professional application requirements

“Available mechanism” does not imply that an application satisfies the requirement. In particular, an accessibility tree requires platform bridges and meaningful widget metadata; a GPU renderer requires an extension contract for an application's renderer.

| Requirement cluster | Existing evidence | What Vellum must integrate or prove |
|---|---|---|
| Dockable panels, persistent layouts, multiple windows | PhotoCraft custom groups; ImGui docking; egui_dock trees; Gio window instances | Split/tab graph, versioned serialization, keyboard pane navigation, detach/reattach, monitor loss and fractional DPI |
| Toolbars, menus, palette, keyboard workflows | PhotoCraft CommandSpec; Qt QAction; VS Code contexts | One dispatcher, scoped shortcuts/chords, native menu roles, text-focus precedence, checked/disabled reasons |
| Trees, large tables, property inspectors | Gio List; ImGui tables/clipper; Fyne table reuse | Stable row keys; horizontal and vertical virtualization; selection independent of materialized rows; edit validation |
| Custom 2D/3D canvases and large datasets | PhotoCraft GPU callback; egui-wgpu; WebGPU/SDL3 GPU | Shared device; persistent targets; damage updates; culling/LOD; clipping/color contracts; bounded uploads |
| Background work and responsive interaction | PhotoCraft snapshot jobs; Gio invalidation | Explicit task scope, cancellation, progress coalescing, bounded workers, revision-checked commit |
| Undo/redo | PhotoCraft domain snapshots and coalescing | Pluggable application transaction/history model, not generic widget-state snapshots |
| Text, IME, high DPI | Gio/go-text; platform-specific input source | Full composition/caret/selection behavior across scripts, scales and monitor transitions |
| Clipboard, DnD, dialogs, tray, notifications | Separate SDL/Gio/native service implementations | Capability matrix per target; rich formats and outbound dragging; portal and sandbox behavior |
| Accessibility, semantic automation, testing | egui kittest/AccessKit; Slint AccessKit and MCP; PhotoCraft commands | Stable semantic IDs, virtualized relations, native adapters, shared action path, restricted optional automation |

## Avoid false equivalences

1. **Immediate mode does not mean stateless.** Every inspected immediate system retains interaction state; some retain previous geometry and large caches.
2. **Retained mode does not require a browser or DSL.** Cogent Core and Fyne expose ordinary Go retained APIs. API taste alone is insufficient justification for a new engine.
3. **GPU accelerated does not mean custom GPU extensible.** Gio's operation renderer and egui-wgpu's callback contract expose different application capabilities.
4. **CGO-free does not mean native-library-free.** The inspected go-webgpu binding dynamically loads wgpu-native; purego-sdl3 loads SDL3. A Go source package can still ship foreign machine code.
5. **Headless is ambiguous.** Command tests, synthetic widget tests, offscreen GPU tests, and real OS integration tests retire different risks.
6. **Cross-platform is a matrix, not a checkbox.** Linux Wayland and X11, screen readers, IMEs, mixed DPI and GPU drivers require distinct validation.

## Decision criteria

Use hard gates before preference scoring: custom canvas interoperability; internationalized editing; native screen-reader behavior on all three target OS families; multiple windows and correct lifetime; idle behavior; reproducible packaging; and maintainable dependency terms. Failure of a gate cannot be compensated by a high widget count.

For candidates that pass, compare application code complexity, total integration work, frame-time distributions, allocations/retained memory, ease of custom controls, dependency ownership, and upgrade cost. [PROTOTYPE_PLAN.md](PROTOTYPE_PLAN.md) defines the same workloads and pass/fail evidence for each contender. No weighted numerical ranking is offered before those measurements exist.

<!-- Source reference definitions -->

[EG-CONTEXT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/context.rs
[EG-TEST]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui_kittest/src/lib.rs
[EG-TEXT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/epaint/src/text/text_layout.rs
[EG-WGPU]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui-wgpu/src/renderer.rs
[GIO-ANDROID]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_android.go
[GIO-APP]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/app.go
[GIO-GPU]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/gpu.go
[GIO-SEMANTIC]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/semantic/semantic.go
[IM-CORE]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui.cpp
[IM-DOCK]: https://github.com/ocornut/imgui/blob/778992face67ecefaa5c2932db4b2101545ea4df/imgui.cpp
[IM-README]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/docs/README.md
[IM-TABLES]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui_tables.cpp
[PC-APP]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/lib.rs
[PC-DOCK]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/dock.rs
[PC-GPUCANVAS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/gpu_canvas.rs
[PC-JOBS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/jobs.rs
[SL-ACCESS]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/winit/accesskit.rs
[SL-PLATFORM]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/platform.rs
[SL-PROP]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/properties.rs
[SL-TEST]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/testing/README.md
