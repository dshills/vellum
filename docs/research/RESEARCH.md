# Vellum: source-level research

Research date: 9 October 2026. Status: architecture research; no framework or prototype implemented.

## Finding

A professional Go desktop framework is technically feasible. Its strongest justification is an integrated editor platform: docking, large data views, commands, background tasks, accessible semantics, deterministic testing, and custom GPU surfaces. Native compilation and a GPU renderer alone do not justify a new framework: Gio, Fyne, Cogent Core, and newer GoGPU projects already occupy parts of that space.

The leading architecture is an **immediate-style Go API over deliberately retained internal state**, with an application-owned domain model, a command dispatcher, and replaceable rendering/platform adapters. A new implementation is only justified if early experiments show that this combination cannot be delivered more economically by extending an existing Go framework. The evidence favors reusing text shaping, GPU backends, windowing, and accessibility adapters rather than implementing those systems from scratch.

Read [RECOMMENDATION.md](RECOMMENDATION.md) for the decision, [ARCHITECTURE_OPTIONS.md](ARCHITECTURE_OPTIONS.md) for alternatives, [ARCHITECTURE_COMPARISON.md](ARCHITECTURE_COMPARISON.md) for the framework matrix, [GO_ECOSYSTEM.md](GO_ECOSYSTEM.md) for dependencies, and [PROTOTYPE_PLAN.md](../../specs/prototype-plan.md) for validation gates.

## Evidence and limits

The five required projects were downloaded from public source repositories and inspected at exact commits. Additional source inspection covered go-text, three Go WebGPU lines, SDL3 and its Go binding, AccessKit and its C bridge, Fyne, Ebitengine, Cogent Core, GoGPU UI, GLFW, and egui_dock. Repository identifiers and full revisions are recorded in [SOURCES.json](SOURCES.json). Source links below use those immutable revisions; symbols are named so readers can follow the paths.

| Core subject | Inspected revision | Scope caveat |
|---|---|---|
| PhotoCraft | `0d10f117333c` | Current application, not a reusable framework specification |
| egui | `f65f89597cc2` | Current development source; PhotoCraft declares egui/eframe 0.36, so its dependency need not equal this HEAD |
| Gio | `9393ca1c7322` | GitHub mirror of upstream; Go module declares Go 1.24 |
| Dear ImGui | `2859f6723c13` | Main development source identifies itself as 1.93.0 WIP |
| Dear ImGui docking branch | `778992face67` | Docking and platform viewport findings are branch-specific |
| Slint | `c44731f8e293` | Current development source; backend capabilities differ |

**Observation** means inspected implementation or an explicitly attributed upstream statement. **Interpretation** means architectural reasoning about that evidence; inferred designer intent is not presented as a maintainer quote. **Proposal** means a Vellum design choice requiring validation. Missing functionality is described narrowly: failing to find a bridge does not prove that no external extension exists. No application was built or run, no screen reader was exercised, and no comparative performance numbers were measured. Test source demonstrates test design, not that tests passed in this research session.

## 1. PhotoCraft: application design versus framework capability

### Entry, UI lifecycle, and dependency boundaries

**Observation.** The workspace separates document/raster/geometry data, editing engine, CPU composition, GPU composition, file formats, automation, and `ui-egui`. The UI crate depends on the engine and rendering crates; the engine does not need an egui widget tree to edit documents. The manifest enables eframe's wgpu and AccessKit features. This is a native self-drawn UI hosted by eframe, not a collection of native OS controls. [Workspace][PC-WORKSPACE], [UI dependencies][PC-DEPS]

The actual frame entry is `impl eframe::App for PhotocraftApp` in `crates/ui-egui/src/lib.rs`. In this revision it has separate `logic`, `raw_input_hook`, and `ui` methods. `logic` drains control requests and OS events, polls jobs/dialogs, processes shortcuts, and schedules repainting. `raw_input_hook` normalizes native-menu key equivalents and Windows pinch behavior and injects synthetic events. `ui` constructs panels, invokes `canvas::document_area`, and shows dialogs, command palette, task UI, and additional document windows. [Application lifecycle][PC-APP]

**Interpretation.** Separation of one-time event processing from UI construction matters because layout can involve additional passes. Go should preserve a comparable side-effect boundary, rather than making arbitrary work occur whenever a drawing closure is evaluated.

### Commands, application state, and undo

**Observation.** `Session` owns document states, active-document selection, job state, and editing operations. `CommandSpec` carries an ID, label, menu placement, shortcut, parameter description, enablement predicate, execution function, and journal flag. `commands::find` looks up the registry; `Session::execute` enters the dispatch path. Parameter descriptions are human/agent-readable strings, not proof of a complete machine-validated schema. [Command registry][PC-COMMANDS], [Session][PC-ENGINE]

The editing path is concrete: `Session::edit` saves the previous document `Arc`, clones the document structure, invokes an edit on the candidate, then publishes a new `Arc<Document>` only after success. It records or coalesces history, trims history, advances the revision, and resets damage information. `undo`/`redo` restore snapshots and layer targets and compute damage; they refuse to move history while the active document has a job. Raster tiles use `Arc::make_mut`, so document snapshots can share unmodified pixels. [Edit and history integration][PC-ENGINE], [Raster storage][PC-RASTER]

**Interpretation.** Undo is primarily a domain-model architecture, not a button toolkit feature. Go can use immutable snapshots, explicit copy-on-write chunks, or domain-specific inverse edits. It cannot copy `Arc::make_mut` semantics by shallow-copying a struct containing mutable slices and maps. Vellum should provide transaction/history interfaces and coalescing conventions, while applications decide what constitutes an undoable state.

### Docking, controls, and visual design

**Observation.** PhotoCraft's docking is custom. `dock.rs` defines named panel groups, tab state, fixed/preferred/minimum heights, resizing, collapse/reorder behavior, and serializable `DockLayout`. The inspected UI manifest does not use `egui_dock`. `panels.rs` integrates that layout into the editor chrome. This is evidence for a particular editor's panel system, not evidence that egui includes a general IDE docking manager. [Dock model][PC-DOCK], [Panel integration][PC-PANELS]

`theme.rs` implements color/shape tokens, font installation, typography helpers, and conversion into egui styling. Files such as `layer_tree_ui.rs`, `point_curve.rs`, `canvas.rs`, and `widgets.rs` implement application-specific interaction and drawing. Consequently, attractive spacing, panel hierarchy, layer controls, and editor affordances come substantially from PhotoCraft's design system and UI code. egui contributes layout, hit testing, input responses, text widgets, paint primitives, and integration. [Theme][PC-THEME], [Custom controls][PC-WIDGETS], [Layer tree][PC-LAYERS]

**Interpretation.** A Go equivalent needs a professional workbench library and design system. A renderer plus stock buttons would reproduce only a small part of the result. Dense desktop controls should have configurable density, keyboard behavior, validation, theme tokens, and focus visuals from the beginning.

### Canvas and GPU rendering: traced path

**Observation.** `canvas::document_area` reaches view rendering in `canvas.rs`; GPU-capable views use `GpuCanvas` in `gpu_canvas.rs`. `GpuCanvas::composite` invokes the separate `photocraft-gpu` compositor into display textures. Unsupported document features can fall back to CPU composition and upload. A custom `egui_wgpu::CallbackTrait` supplies the viewport painting path, while GPU resources live in the renderer's callback resource store. This uses egui's integration seam, not egui's ordinary image widget as an entire image-processing engine. [Canvas][PC-CANVAS], [GPU canvas][PC-GPUCANVAS], [GPU engine][PC-GPU]

The implementation handles damage rectangles, tiled textures, mip updates, device-health tracking, per-view parameters, and optional higher-precision canvas formats. It uses `Rgba16Float` for eligible higher-bit-depth documents subject to capabilities and an explicit pixel budget; this is not a promise of HDR display output. One file-header comment still describes a future GPU compositor, but the implemented `composite` method already calls it. This is a specific example of why source comments and READMEs must be checked against the code path. [GPU canvas][PC-GPUCANVAS]

**Interpretation.** Adopt damage-aware resource updates, shared-device composition, texture budgets, and a defined failure path. Do not equate GPU-backed widgets with a GPU document engine. CAD, waveforms, image editing, and 3D need application-owned rendering algorithms and data structures.

### Background work and stale-result protection

**Observation.** `jobs.rs` creates jobs with progress and cancellation state, executes native work on threads, and applies results through the session. Workers operate on snapshots. `apply_job` finds the target document and compares the original snapshot identity before applying; a closed or changed document rejects the result. Cancellation is cooperative. UI job handling polls completion and presents state; the application controls repaint scheduling. [Jobs][PC-JOBS], [Jobs UI][PC-JOBSUI]

A useful power-management detail appears in `PhotocraftApp::logic`: transport arrival wakes the UI, and screenshot polling occurs only while screenshots are pending. The source records that an earlier periodic poll caused idle rendering. Immediate-mode UI does not intrinsically require constant redraw. [Application lifecycle][PC-APP]

**Interpretation.** Go's meaningful advantage is making this worker/result/commit pattern convenient and typed. Merely replacing a Rust worker thread with a goroutine does not remove document races, cancellation latency, or stale results.

### Automation, platform work, and tests

**Observation.** `ui-egui/src/control.rs` handles inspection and synthetic interaction requests. The `automation` crate provides headless and bridged execution paths and MCP transport. Domain commands are useful outside the GUI, while UI inspection adds view-specific state. The application supplies native menu services, file-dialog integration, OS event handling, and a separate tablet crate. These features are not all supplied by egui. [UI control][PC-CONTROL], [Automation headless path][PC-HEADLESS], [Automation server][PC-AUTOMATION], [Native menu][PC-MENU], [Tablet adapters][PC-TABLET]

The UI manifest includes `egui_kittest` with wgpu/snapshot/eframe features. Source includes docking, shortcut, job UI, and layer-editing tests; engine tests cover domain behavior. This layered structure is a better test model than relying solely on screenshots. [UI dependencies][PC-DEPS], [Dock tests][PC-DOCKTEST]

**Conclusion for Go.** Reuse the architectural separation, command centricity, immutable worker inputs, custom render seam, and visual discipline. Rust contributes ownership checks, `Arc`, typed traits, and deterministic destruction patterns; Go must replace those with explicit mutation rules, resource closing, bounded caches, and race testing. PhotoCraft's appearance is not evidence of Rust-specific rendering magic or measured superiority to Go.

## 2. egui: immediate API, extensive retained machinery

### Frame and widget lifecycle

**Observation.** `Context` contains persistent memory, per-viewport previous/current pass data, fonts, paint state, and interaction state. The `run`/pass machinery begins input processing, invokes UI construction, ends the pass, and can repeat a discarded layout pass subject to `Options::max_passes`. `request_repaint` and delayed repaint callbacks are separate from continuous animation. Integration receives output including shapes, textures, platform effects, and viewport information. [Context][EG-CONTEXT]

A widget call allocates/layouts a region, registers identity and interaction, contributes paint, and returns a `Response`. There is no requirement for an application-owned widget object for each button, but there are retained structures behind the call. At pass start, the inspected code hit-tests `prev_pass.widgets`, orders layers, and computes interaction from current input and prior geometry. This resolves an immediate-mode ordering problem: input arrives before this pass has rebuilt all rectangles. [Interaction path][EG-INTERACTION], [UI construction][EG-UI], [Context][EG-CONTEXT]

**Go lesson.** Define when geometry is authoritative, how new widgets become interactive, and how events are consumed once. A multi-pass drawing API that launches tasks or writes files inline can duplicate side effects. Prefer processing accepted actions once, then evaluating layout/paint from a stable state snapshot.

### Identity and state

**Observation.** `Id` derives stable hashes from source values and parent salts; the current API distinguishes globally unique IDs and local salts. Stable IDs are needed when labels change or list order changes. `Memory::data` is a typed ID map for widget state; focus, popups, areas, caches, and viewport state have additional stores. `TextEditState` retains cursor/selection, undo state, horizontal offset, IME purpose, and interaction time, using explicit load/store against an ID. Application data is expected to live outside this memory. [IDs][EG-ID], [Memory][EG-MEMORY], [Text editor state][EG-EDITSTATE]

**Go lesson.** Use stable domain keys for rows, documents, tabs, and fields; avoid translated labels or slice indices as identities. Plain typed state structs work well in Go. A generic `map[ID]any` may be useful for extension widgets but should not become the only state architecture. Cache eviction, disappeared-widget cleanup, focus restoration, and durable layout migration need explicit policies.

### Layout, painting, and custom widgets

**Observation.** `Placer` and `Layout` maintain a cursor/region and place children according to direction, wrapping, alignment, and grid state. This is not a full browser layout engine. Some widgets can request an extra pass when size information changes. Painters emit shapes with clipping/layer context; epaint tessellates into clipped primitives. Renderer batching must respect clip, texture, and ordering constraints. Custom widgets can use the same allocation, response, and painter machinery. [Placement][EG-PLACER], [Layout][EG-LAYOUT], [Tessellation][EG-TESSELLATOR], [Painter][EG-PAINTER]

**Interpretation.** Ordinary control flow and local layout make tool UIs easy to build. Constraints become visible for intrinsic-size negotiation and complex resizable workspaces. Go should keep small layout primitives but permit retained measurement caches and a separately modeled docking tree. Arbitrary drawing must be able to register hit regions and semantics; pixels alone do not constitute a usable widget.

### Input, focus, keyboard, and dragging

**Observation.** The integration translates native events into `RawInput`; egui derives pointer/keyboard state, routes interactions, manages focus, and produces cursor/IME/clipboard output. Focus is retained per viewport. The drag-and-drop plugin stores a typed payload across frames and clears it on cancellation or release. Native file drops travel through platform integration, which is distinct from in-UI drag payloads. [Input state][EG-INPUT], [Winit integration][EG-WINIT], [Memory][EG-MEMORY], [Drag payloads][EG-DND]

**Go lesson.** Separate key-down shortcuts from committed text and IME composition. Give pointer capture, modal scopes, focus order, drag ownership, and shortcut priority explicit contracts. Application-local dragging, cross-window dragging, and OS file/data dragging are three different capabilities; do not advertise one based on another.

### Text, fonts, DPI, and internationalization

**Observation.** The inspected epaint source uses **HarfRust** shaping, font fallback runs, glyph allocation, and cached text layouts (“galleys”). Font caches are per viewport where appropriate; atlas changes invalidate dependent galleys. `text_layout.rs` still identifies script-aware segmentation/RTL-bidi work as incomplete. Older blanket claims that egui has no shaping are therefore incorrect for this revision; full internationalized editing is also not established. [Text shaping][EG-TEXT], [Font/cache management][EG-FONTS]

Text editing tracks selection and IME composition beyond the application string. UI coordinates are logical points; the rendering integration supplies physical screen dimensions and pixels per point. DPI changes must affect glyph rasterization and layout/cache keys, not merely scale a finished bitmap. [Editor state][EG-EDITSTATE], [GPU screen descriptor][EG-WGPU]

**Go lesson.** Reuse go-text for shaping/segmentation/fallback and test mixed-direction editing end to end. Font rasterization, shaping, line breaking, caret mapping, selection, and input methods are separate responsibilities. A successful Latin label does not validate a text subsystem.

### GPU, windows, accessibility, and docking

**Observation.** egui is renderer-independent; egui-wgpu and egui_glow provide different integrations. In egui-wgpu, a custom callback has `prepare`, `finish_prepare`, and `paint` phases. It can encode preparatory GPU work and then paint within the UI render pass, with renderer-owned callback resources. This is the concrete seam PhotoCraft uses. [GPU callback contract][EG-WGPU]

eframe/winit supplies the native window/event-loop layer. egui viewports model additional native windows when supported by the integration; an `egui::Window` alone is a floating UI region and should not be confused with an OS window. AccessKit output provides semantics and action integration, but custom controls still need meaningful roles, labels, values, and actions. [Viewport model][EG-VIEWPORT], [Winit integration][EG-WINIT], [Context accessibility][EG-CONTEXT]

Docking belongs to an ecosystem. `egui_dock::DockState` retains surfaces and trees containing tabs, with serialization support; those surfaces do not by themselves prove OS-level detachable windows. PhotoCraft uses its own model instead. [egui_dock state][EG-DOCK]

### Performance, allocations, and testing

**Observation.** egui retains font/layout caches and GPU buffers. Repeated UI evaluation still involves collection traversal, hash lookups, shape generation, and allocations in some paths; it is not an allocation-free guarantee. The ID source explicitly favors hashing tuples over formatting fresh strings. Virtualized row APIs in scroll areas reduce submitted widgets. [Fonts][EG-FONTS], [ID construction][EG-ID], [Scroll-area rows][EG-SCROLL]

`egui_kittest::Harness` drives passes with synthetic input and uses AccessKit-backed query state. Screenshot support is an additional renderer-dependent layer. Its settling behavior provides a useful precedent for testing scheduled work without arbitrary sleeps. [Harness][EG-TEST]

**Conclusion for Go.** Transfer stable IDs, explicit invalidation, cached text, ordered paint streams, semantic testing, and custom render phases. Do not port pervasive `Arc<Mutex<…>>`, Rust type maps, lifetime workarounds, or cloning conventions mechanically. Go can use a single mutable UI owner and immutable messages with smaller synchronization surfaces.

## 3. Gio: a strong Go-native baseline with consequential boundaries

### Frame, layout, state, and events

**Observation.** A window emits `app.FrameEvent`, containing time, metrics, dimensions, an input source, and a `Frame` callback. `app.NewContext` resets operations and constructs a constrained layout context. Widgets issue drawing/input/semantic operations; `Frame` submits the operation list. `app.Main` reserves the OS main-thread integration. `Window.Invalidate` is explicitly concurrency-safe. [App events][GIO-APP], [Window lifecycle][GIO-WINDOW]

`op.Ops` serializes operations into reusable storage to reduce garbage. Macro recording/replay is available. `layout.Flex`, `Stack`, and constraints provide compositional layout; `layout.List` traverses a visible subsection and retains a first-item/offset position. It does not require submitting every row of a large collection. [Operations][GIO-OPS], [Flex][GIO-FLEX], [List][GIO-LIST]

State is often an application-owned struct: `widget.Clickable` holds gestures and press history; `widget.Editor` holds editing state. Event tags connect handlers to the `input.Router`, whose state includes pointer/key routing, queued commands, focus, clipboard, and semantic data. Widget pointers used as tags must remain stable while the widget is active. [Clickable][GIO-BUTTON], [Editor][GIO-EDITOR], [Router][GIO-ROUTER]

**Interpretation.** Gio already demonstrates that immediate-style composition, Unicode text, GPU drawing, event-driven idle behavior, and Go concurrency are compatible. A new project must improve professional workflows or integration, not merely claim these capabilities as novel.

### Rendering and the package-extraction question

**Observation.** `gpu.GPU` exposes `Release`, `Clear`, and `Frame(*op.Ops, RenderTarget, viewport)`. The renderer collects Gio ops, builds paths, maintains texture/path caches, renders coverage/clipping and opacity layers, and draws the result. Backend implementation packages include OpenGL, Metal, Vulkan, and Direct3D **11**, not D3D12. The low-level device interface is in `gpu/internal/driver`. [Public GPU contract][GIO-GPU], [API aliases][GIO-GPUAPI], [Internal driver][GIO-DRIVER]

The package itself documents its role as rendering Gio drawing operations and integration with external windows. Public API aliases for native render targets do not turn it into a stable general-purpose compute/3D API. Go's `internal` import rules prevent an unrelated module from directly importing the driver. Extracting that layer requires upstream work or a maintained fork including its dependencies.

`app.CustomRenderer(true)` disables Gio's GPU-context creation and explicitly makes the caller responsible for backend initialization, swapping, and pacing. `ViewEvent` exposes native handles with validity/lifetime rules. This makes a Gio-shell experiment plausible; it does not make custom GPU rendering an automatic extension of its normal renderer. [Custom renderer and window handles][GIO-WINDOW], [View events][GIO-APP]

**Conclusion on the three reuse choices.** Building directly on Gio is the cheapest serious baseline. Reusing standalone `layout`, `unit`, geometry, or text-related code may help, but much of layout/widget code depends on Gio context/ops/input types. For text, use Gio's upstream dependency **go-text** directly. Extracting the internal GPU driver is poor initial leverage. A Gio platform shell with an independent renderer is possible but retains Gio's platform limitations while adding custom presentation work.

### Text and input integration

**Observation.** Gio's text implementation imports go-text shaping, font scanning, and font parsing alongside bidi processing. It stores logical runs and visual-order indices, exposes glyph/cluster information, and caches shaped layouts. The editor handles selection/cursor behavior and communicates input state to the platform. `Window.updateState` propagates text-input visibility, hints, editor state, clipboard requests, and wakeup time. [Text shaping][GIO-TEXT], [Shaper][GIO-SHAPER], [Editor][GIO-EDITOR], [Window state][GIO-WINDOW]

The `io/transfer` protocol supports MIME negotiation and transfer events between registered UI participants. That is not equivalent to complete OS drag/drop. In the inspected Wayland driver, data-device motion/leave/drop callbacks are empty; clipboard selection is handled separately. [Transfer protocol][GIO-TRANSFER], [Wayland driver][GIO-WAYLAND]

### Accessibility and desktop platform coverage

**Observation.** `io/semantic` offers labels, descriptions, selected/enabled state, and a small role set: button, checkbox, editor, radio button, and switch. `input.Router` constructs semantic nodes; `app.Window` retains/lazily updates trees and diffs. Searching the inspected native drivers for semantic lookup and accessibility bridge code located Android integration, but no equivalent Windows UI Automation, AppKit NSAccessibility, or Linux AT-SPI bridge. This source audit does not validate desktop screen-reader support. [Semantic operations][GIO-SEMANTIC], [Router][GIO-ROUTER], [Android bridge][GIO-ANDROID]

There are distinct Windows, AppKit, X11, and Wayland drivers. The macOS/Linux paths use CGO; the Windows installation guide states no special C compiler is needed for Gio's Windows support. Linux development requires native window/input/graphics packages. Native menus, rich clipboard formats, external drag/drop, tray integration, notifications, and professional screen-reader patterns need a capability-by-capability assessment or additional integration. [Driver sources][GIO-DARWIN], [Windows driver][GIO-WINDOWS], [X11 driver][GIO-X11], [Linux installation][GIO-INSTALL-LINUX], [Installation][GIO-INSTALL]

### Multi-window, custom drawing, performance, and tests

**Observation.** Window objects provide per-window state; operation-based custom 2D drawing is natural. The architecture supports multiple window instances, but coordinated cross-window docking and GPU-resource sharing are additional application/framework work. `gpu/headless` provides an offscreen rendering path that still depends on a working graphics backend. Router/layout/widget unit tests and text benchmarks can exercise behavior without all OS integration. [Window][GIO-WINDOW], [Headless renderer][GIO-HEADLESS], [Text benchmarks][GIO-TEXTBENCH]

**Interpretation.** Serialized ops and reusable slices are valuable under Go's GC, but do not establish a specific latency or memory advantage. Device setup, shaping misses, visible widget count, canvas workloads, and application allocation rate can dominate. A framework foundation decision should require both a custom-canvas experiment and an accessibility/IME experiment, not only a widget demo.

## 4. Dear ImGui: the professional-tools reference

### Frame and identity mechanisms

**Observation.** `ImGui::NewFrame` processes the event queue, updates viewport/font/texture state, and resets per-frame interaction data while retaining the context. Widget functions such as `ButtonEx` derive an ID, calculate size, call `ItemSize`/`ItemAdd`, evaluate `ButtonBehavior`, and append drawing. `Render` finalizes the frame and assembles window draw lists into `ImDrawData`; an external renderer consumes it. [Frame/context][IM-CORE], [Button implementation][IM-WIDGETS], [Draw-data types][IM-HEADER]

IDs hash a label, pointer, or integer with the current ID stack. Context/window storage retains active/hovered/focused IDs, editing state, table settings, window geometry, and other state. `PushID`-style scopes permit repeated labels. Application data remains separate; immediate mode removes some synchronization work, not all retained state. [ID methods][IM-CORE], [Internal state][IM-INTERNAL]

**Go lesson.** Use scoped IDs and reusable contiguous buffers. Avoid a global implicit context when an explicit `*UI` is easy to pass. C++ pointer identities, macros, manual vectors, and stack assumptions are not a Go API template.

### Layout, tables, trees, property editing, and custom drawing

**Observation.** ImGui's local cursor layout fits dense tools: same-line controls, groups, trees, sliders, numeric drag controls, and explicit sizing. The table implementation separates setup from first-row layout, tracks column visibility/widths, supports freezing and sorting state, and cooperates with `ImGuiListClipper`. The source notes that auto-sizing columns and clipping require care because unsubmitted content cannot contribute measurements. [Tables][IM-TABLES], [Widget implementations][IM-WIDGETS], [List clipper][IM-HEADER]

`ImDrawList` supports custom geometry, clip rectangles, texture references, and callbacks. Rendering and platform backends are separate. A custom waveform or node graph is easy to draw, but application code still owns hit testing, domain selection, undo, spatial indexing, and semantic meaning. [Draw lists][IM-DRAW], [Backend guide][IM-BACKENDS]

**Interpretation.** The productivity advantage comes from exposing practical tool-building primitives together. Vellum should include virtual tables/trees, editable numeric fields, scoped shortcuts, inspectors, and panel management as coherent components. It should not force application teams to rediscover every interaction convention.

### Docking and multiple OS windows

**Observation.** Docking is implemented in the separately inspected docking branch. `DockContext` processes queued dock/undock changes and retains a dock-node structure and settings. Enabling platform viewports requires backend support and calls such as `UpdatePlatformWindows` and `RenderPlatformWindowsDefault`; it is more than drawing additional floating panels. [Docking branch][IM-DOCK], [Docking header][IM-DOCKHEADER]

**Go lesson.** Model docking as serializable state with explicit pane IDs, split/tab nodes, transactions, and window ownership. Desktop integration differs on Wayland, where clients cannot assume control over global window positions. Start with in-window docking and validate native detachment independently.

### Allocation, text, and general-desktop limits

**Observation.** The source states a goal that a typical idle frame should not call `malloc/free`, and buffers retain capacity. This is an allocation design goal, not a measured whole-application guarantee. The current font/backend protocol supports incremental texture updates and dynamic font behavior introduced in the 1.92 generation; old descriptions of a single immutable startup font atlas are incomplete. [Core design notes][IM-CORE], [Texture protocol][IM-BACKENDS], [Fonts][IM-FONTS]

The project explicitly says full internationalization—RTL, bidi, shaping—and accessibility are not supported. Same-context concurrent API use is unsupported. Main-loop scheduling and OS-native facilities belong to the host. The separate Dear ImGui Test Engine is relevant for automation, but its licensing must be evaluated independently of ImGui's MIT license. [Project scope][IM-README], [Threading FAQ][IM-FAQ], [Test Engine][IM-TESTENGINE]

**Conclusion for Go.** ImGui is an excellent interaction and allocation reference for engineering tools. A Go binding is useful as a benchmark/control implementation, but inherits the text/accessibility gaps and a C++ boundary. The intended framework's inclusive desktop requirements justify going beyond the ImGui model rather than renaming it.

## 5. Slint: useful retained mechanisms without adopting its DSL

### Component and property architecture

**Observation.** Slint compiles declarative components into runtime item trees and related code. `ItemTreeVTable` exposes traversal, item lookup, subtrees, layout information, and accessibility behavior. Properties can hold a value or lazy binding; reading within a binding registers dependencies, and changed values mark dependents dirty. This allows demand-driven recomputation rather than re-evaluating all bindings each frame. [Item trees][SL-ITEM], [Properties][SL-PROP], [Compiler lowering][SL-COMPILER]

**Interpretation.** Retained identity and dependency information make incremental work and inspection natural. Vellum can adopt revisioned layout/text/paint caches without introducing a pervasive reactive language. The benefit comes from tracking what changed, not from requiring `.slint` files or generated application code.

### Layout and rendering

**Observation.** Layout functions solve box/grid/flex arrangements from constraints and measurements. The partial-renderer machinery tracks cached geometry and rendering-property dependencies to compute damaged regions. The platform interface separates window creation/event-loop services from renderer and window-adapter responsibilities. [Layout][SL-LAYOUT], [Partial rendering][SL-PARTIAL], [Platform][SL-PLATFORM]

The repository includes software, Skia, FemtoVG, and newer renderer integrations. GPU API availability and text behavior depend on the selected renderer/backend; “Slint uses one GPU API” would be inaccurate. Native custom-render notifications are exposed through window APIs, but renderer support and lifecycle restrictions still apply. [Renderer sources][SL-RENDERERS], [Rendering API][SL-API]

**Go lesson.** Separate measure, arrange, paint, and semantic invalidation. An animation that changes opacity should not reshape text or rebuild a table model. Retain expensive results, and make invalidation rules inspectable. A software reference renderer is useful for tests, but production GPU parity is a separate undertaking.

### Input, accessibility, styling, and native integration

**Observation.** Runtime input code performs event dispatch against retained items. The winit backend's AccessKit adapter translates nodes and supported actions, including focus and text-selection handling, with deferred actions where necessary. Component styling and standard widget definitions are separate from the platform interface. Public language surfaces in this repository include Rust, C++, JavaScript/Node, and Python; the inspected tree does not establish an official Go API. [Input][SL-INPUT], [AccessKit integration][SL-ACCESS], [Public APIs][SL-BINDINGS]

Slint's retained model therefore gives semantics a persistent home, but a custom control must still describe its behavior. A Go wrapper would inherit its runtime and binding/build model rather than produce the requested ordinary-Go framework.

### Testing, concurrency, performance, and licensing

**Observation.** The preliminary testing backend can substitute event-loop/time behavior; its documentation distinguishes tests that do not render pixels and use fixed text metrics from real rendering. The current tree also includes an embedded MCP inspection/testing server, so semantic automation is already a concrete toolkit pattern, not an unexplored invention. These internal testing APIs explicitly lack normal semver guarantees. [Testing backend][SL-TEST], [MCP testing implementation][SL-MCP]

The platform event-loop proxy is the route for externally initiated work; retaining an item tree does not make it freely mutable from any worker. Slint's framework is offered under GPLv3, a royalty-free license with conditions, or commercial terms; examples/documentation have different licensing. It is a source of ideas, not code assumed safe to copy into a permissive Go core. [Platform event proxy][SL-PLATFORM], [License terms][SL-LICENSE]

**Conclusion for Go.** Adopt durable semantic identity, selective invalidation, independently testable platform interfaces, and introspection. Avoid translating Rust's pinned property machinery or compiler-generated vtables into application-facing Go APIs. Retained and immediate approaches are not mutually exclusive implementation strategies.

## 6. Additional projects selected for specific lessons

| Project | Evidence and reason to study | Consequence for Vellum |
|---|---|---|
| **Cogent Core** | Actual retained Go widget tree with sizing phases, scene dirty flags, splits/tabs, and a GPU package using the maintained WebGPU fork. This is a closer existing competitor than a webview wrapper. [Widgets][CC-WIDGET], [Rendering][CC-RENDER], [Splits][CC-SPLITS], [GPU][CC-GPU] | Must be a baseline in the build-versus-extend experiment. Its embedding/tree/style/update conventions and generated support code differ from the brief's preferred API; do not reject it merely for retained mode. Screen-reader completeness was not established by this limited audit. |
| **Fyne** | Retained Go widgets, recycled table-cell templates, a central UI scheduling API, and desktop integration. Inspected accessibility code is gated by `accessibility` and includes Windows/macOS paths; the non-supported path is a no-op. [Table][FY-TABLE], [Thread dispatch][FY-THREAD], [Accessibility][FY-ACCESS], [Fallback][FY-ACCESSFALLBACK] | Reuse the cell-provider model and benchmark its actual table performance. Verify custom GPU access and Linux accessibility instead of assuming parity. |
| **Ebitengine** | Go-native image/shader rendering and go-text integration. `FPSModeVsyncOffMinimum` and `ScheduleFrame` support event-driven redraw. The public window API is package-global. [Scheduling][EB-RUN], [Text][EB-TEXT], [Window API][EB-WINDOW] | A strong 2D/rendering reference; not the obvious shell for independently managed document windows, rich native text services, and accessibility. Do not dismiss it as necessarily running a busy game loop. |
| **GoGPU UI** | Retained `Widget` layout/draw/event tree, a11y tree structures, GPU-view widgets, and testing packages. [Widget contract][GG-WIDGET], [A11y tree][GG-A11Y], [GPU view][GG-GPUVIEW] | Evaluate as an emerging alternative. ARIA-style role structures and README platform claims alone do not validate native screen-reader bridges or production maturity. |
| **Flutter** | Separate widget configuration, retained elements/render objects, and stage-specific dirty propagation. [Inside Flutter][FL-INTERNAL] | The public API and internal persistence can differ. Adopt selective work and lifecycle clarity without adopting Dart, a widget object for every expression, or Flutter's engine distribution. |
| **Qt** | QAction unifies user actions; QDockWidget supplies professional panel behavior; Qt Quick separates scene-graph rendering from item logic. [Actions][QT-ACTION], [Docking][QT-DOCK], [Scene graph][QT-SCENE] | Study command enablement/check state and complete docking UX. Reusing Qt itself would trade Go-native design and small deployment for a mature desktop stack. Qt Widgets and Qt Quick are distinct architectures. |
| **VS Code** | Commands plus contextual enablement/visibility drive menus, keybindings, and tools. [Command guide][VS-COMMAND], [Context keys][VS-CONTEXT] | Keep one command implementation and scoped availability. Prefer typed Go predicates to importing a string expression language. |
| **SDL3 / AccessKit** | Source demonstrates platform-specific input/dialog/tray code and cross-platform accessibility adapters. The C AccessKit SDL example uses SDL2 window APIs, so it is a design reference rather than a drop-in SDL3 integration. [SDL text input][SDL-KEYBOARD], [SDL dialog][SDL-DIALOG], [AccessKit C example][AK-SDL] | Reuse native machinery behind adapters; explicitly budget the Go bridge and SDL3 handle integration. |

GTK's accessibility/native-text and snapshot-rendering design, iced's message/update model, Makepad's drawing/layout integration, Nuklear's compact command buffers, and Skia/Vello's vector rendering remain targeted follow-up subjects if a prototype exposes a corresponding gap. They were not deeply audited here. Wails/Tauri remain useful delivery baselines but use webview UI, contrary to the selected native-rendering direction. Go ImGui bindings remain integration baselines, not substitutes for solving internationalized editing and accessibility. Game-editor UI systems are represented here by the deeper ImGui study. This prioritization avoids pretending every ecosystem received equivalent review.

## 7. Cross-cutting findings

### Immediate, retained, or hybrid

| Model | What it buys | What it costs in the target applications |
|---|---|---|
| Immediate API, recompute every requested frame | Ordinary control flow; direct application data access; easy custom drawing | Stable identity is still necessary; repeated layout/shape work; input geometry and side-effect timing need rules |
| Retained public object tree | Persistent identity, natural inspection, incremental layout and events | Lifecycle and synchronization complexity; stale derived state; potentially large object/GC graphs |
| Immediate API with internal retained records | Go control flow plus cached measurements, semantics, focus, docking, resources | Requires explicit cache keys, eviction, invalidation, and rules for widgets absent from a frame |

**Proposal.** Choose the third, and describe it precisely. Retain *UI records*, not application domain truth: widget/edit state, hit geometry, semantic nodes, layout results, animation state, font/glyph caches, dock/window state, and resource handles. Rebuild only visible declarations on an invalidated window. Version retained subtrees only when a measured need justifies the complexity. Never use raw object addresses as durable automation IDs.

### Text is a pipeline, not a font loader

**Observation.** go-text provides font parsing, font discovery/fallback, segmentation, a Go HarfBuzz port, shaping outputs with clusters, and line wrapping. `HarfbuzzShaper` reuses internal buffers/font objects but allocates returned glyph data; it should not be shared unsafely between concurrent calls. Gio and Ebitengine consume it in different rendering/editing architectures. [go-text shaping][GT-SHAPING], [Input segmentation][GT-INPUT], [Font discovery][GT-FONTS], [Wrapping][GT-WRAP]

**Proposal.** The pipeline should be UTF-8 model → grapheme/bidi/script/font runs → shaped glyph clusters → line layout → visual caret/selection mapping → raster/cache → GPU paint. Include language, direction, font collection revision, variation axes/features, size, wrap width, and relevant scale/raster mode in cache keys. Distinguish byte offsets, rune positions, grapheme boundaries, UTF-16 native indices, and glyph clusters with named conversion functions. Test combining marks, emoji sequences, Arabic/Hebrew mixed with numbers, Indic scripts, CJK fallback, and ligatures. No implementation of a new shaping engine is proposed.

Editable text adds undo grouping, composition ranges, clipboard normalization, navigation policy, and platform IME geometry. Rich text adds styled spans and inline objects; code editing adds scalable storage, line indexing, selections, syntax data, and background services. A robust basic editor belongs in the framework; a full IDE editor is a separately scoped component. Color glyph formats and fallback/raster support must be validated rather than inferred from Unicode shaping support.

### Native platform matrix

The following is an integration map, not a claim that every candidate implements every cell. SDL and Gio sources demonstrate that platform-specific code remains necessary. [SDL platform sources][SDL-VIDEO], [Gio platform sources][GIO-WINDOW]

| Concern | Windows | macOS | Linux: Wayland / X11 |
|---|---|---|---|
| Window and DPI | HWND/message pump, per-monitor DPI policy, resize events | NSApplication/NSWindow/NSView, backing scale, AppKit main thread | xdg-shell/configure and scale protocols; X11 windows/events/XRandR; logical vs buffer dimensions |
| Keyboard and IME | Key events distinct from text; TSF/IMM integration and composition geometry | NSTextInputClient marked/selected text and candidate rectangle | Wayland text-input protocol availability; X11 XIM/IBus/Fcitx paths; compositor/input-method variation |
| Clipboard | Unicode text and negotiated native formats | NSPasteboard types and ownership | Wayland data offers/selection lifetime; X11 selections/INCR; primary selection differs from clipboard |
| Drag/drop | OLE IDataObject/IDropTarget or platform-library bridge | NSDragging/NSPasteboard | Wayland data-device protocol; Xdnd on X11; inbound and outbound support tested separately |
| Accessibility | UI Automation providers and actions | NSAccessibility tree/actions | AT-SPI over session D-Bus on both display protocols; test actual desktop/session activation |
| Menus/dialogs | Native menu/dialog ownership and modality | Application menu roles, NSOpenPanel/NSSavePanel, responder integration | Toolkit/portal dialogs; ownership, asynchronous portal responses, sandbox file access |
| Tray/notifications | Shell notification-area icon; desktop notification APIs | NSStatusItem and user notification APIs | StatusNotifierItem/AppIndicator compatibility; freedesktop notifications; desktop-dependent support |
| Multiple windows | Focus/activation, monitor DPI, shared GPU device | Application lifecycle, window close/menu semantics | Wayland cannot assume arbitrary global positioning; X11 behavior differs |

Microsoft's DPI and TSF guidance and the XDG FileChooser protocol are primary references for these responsibilities. The inspected SDL3 header makes text-input operations main-thread calls, and its source includes separate Cocoa, Windows, and Wayland keyboard implementations. [Windows DPI][WIN-DPI], [TSF][WIN-TSF], [Portal file chooser][LINUX-PORTAL], [SDL keyboard contract][SDL-KEYBOARD]

### GPU architecture and escape hatches

**Proposal.** Keep two levels: high-level clipped 2D drawing and a versioned advanced GPU extension. The extension should expose the selected device/queue and an ordered preparation/render context, including target format, pixel size, scale, clip, color/alpha convention, and resource-generation ID. Applications should share the framework device, render to persistent offscreen targets where appropriate, and composite those targets without CPU readback. Applications should not present the framework's swapchain independently.

wgpu-native and SDL3 GPU are both credible prototype candidates. wgpu-native provides a broad graphics/compute model and WGSL workflow through a native library; SDL3 GPU combines explicit graphics/compute with its window layer and backend-specific shader formats. Both avoid writing Metal/Vulkan/D3D12 backends in Vellum. Neither makes GPU work costless or handles domain geometry, text editing, accessibility, or color management. [WebGPU binding implementation][WG-DEVICE], [SDL GPU implementation][SDL-GPU]

Raw Vulkan/Metal/D3D12 backends would multiply shader, synchronization, resource-lifetime, and driver-validation work. OpenGL remains a useful compatibility/baseline option, but should not be selected as the sole long-term desktop foundation without weighing macOS support direction and modern compute/custom-render needs. Backend-specific extension packages are preferable to promising that every native feature works identically across all APIs.

### Go concurrency and GC

**Proposal.** One UI owner serializes application-facing UI state changes. Workers receive immutable inputs, report bounded/coalesced progress, and post typed results. The UI owner checks document revision/window lifetime before committing. `context.Context` communicates cancellation/deadlines; it does not forcibly stop a goroutine. CPU pools are bounded independently from I/O tasks. The native main thread and UI owner may be the same execution lane initially; goroutines must never silently substitute for OS thread affinity.

Recover a panic only at an explicit task boundary and report it as failure; do not claim recovery makes a partially mutated domain model valid. Cancel tasks when their scope closes, keep completion delivery alive through shutdown, and never let a worker block forever trying to publish progress to a closed UI. Starting a task from a re-evaluated draw closure needs a command/event guard; a progress widget observes a task handle and does not spawn one.

Go GC makes ordinary owned state convenient but is not a real-time scheduler. Prefer reusable flat command buffers, typed state, bounded text/resource caches, and visible-data virtualization. Native/GPU allocations require explicit accounting and closing; they are not fully represented in Go heap metrics or reclaimed on a useful schedule merely because a Go wrapper becomes unreachable. [Go GC guide][GO-GC], [CGO runtime boundary][GO-CGO], [Gio operation storage][GIO-OPS]

### Commands, semantics, and AI control

**Proposal.** Commands own stable IDs, argument/result contracts, scoped enablement, execution, cancellation, and undo policy. Menus, toolbars, keybindings, command palette, tests, CLI, and optional automation adapters invoke the same dispatcher. Availability must be checked at execution time, not only while painting a disabled button. Distinguish visibility, enabled state, checked state, and authorization. For a long-running edit, identify the target document explicitly rather than whichever tab happens to be active when it finishes.

The semantic tree supplies role, label, value, selection, bounds, relations, and actions. The command registry supplies domain verbs such as `component.rotate`; a query provider exposes application-specific objects and revisions. These are complementary: accessibility cannot infer a CAD model or safe domain command parameters from pixels or generic widgets. PhotoCraft and Slint already demonstrate that command/UI introspection can support agent transports. [PhotoCraft control][PC-CONTROL], [Command registry][PC-COMMANDS], [Slint MCP][SL-MCP]

An optional local IPC/CLI/MCP package should translate into this semantic/command layer, never sit in rendering. Prefer OS-restricted Unix sockets or named pipes; an HTTP loopback option requires authentication, bounded requests, and an explicit opt-in. Export only selected commands/data, redact sensitive fields, attach expected revisions and request IDs, and keep audit/result records. A million-row table should expose virtual ranges and selected items, not serialize a million nodes each frame. Equivalent test and accessibility paths should remain available without running an AI service.

### Testing and performance evidence

**Proposal.** Test the domain model and command dispatcher without a window; test widget behavior with synthetic input, a deterministic clock, stable IDs, and semantics; test layout and text with fixed font fixtures; test actual GPU output with pinned drivers or tolerances; and test IME/accessibility/clipboard/DnD against real OS services. A headless GPU renderer is not the same as a platform-free logic harness, and semantic command invocation is not a replacement for input-routing tests.

No framework receives a measured speed ranking in this report. [PROTOTYPE_PLAN.md](../../specs/prototype-plan.md) defines provisional numeric budgets, workloads, measurement boundaries, and go/no-go conditions for startup, idle, 60/120/144 Hz, tables, trees, canvases, resize, text, memory, and task responsiveness. Counts of widgets, tests, or GitHub stars are not evidence of those budgets being met.

## 8. Research conclusion

The defensible opportunity is a Go editor/workbench framework whose state, commands, custom rendering, native services, semantics, and tests fit together. The highest risks are internationalized editing, desktop accessibility, native surface/resource lifetime across windows, platform completeness, and ongoing maintenance—not producing the first triangle or button. Proceed to the planned experiments before selecting a permanent platform/GPU stack or promising a general professional toolkit.

<!-- Source reference definitions -->

[AK-SDL]: https://github.com/AccessKit/accesskit-c/blob/8b6ed37c20ed4c59390e253407983333053662ba/examples/sdl/hello_world.c
[CC-GPU]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/gpu/gpu.go
[CC-RENDER]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/render.go
[CC-SPLITS]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/splits.go
[CC-WIDGET]: https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/widget.go
[EB-RUN]: https://github.com/hajimehoshi/ebiten/blob/dcb93b02c135c20ea835952e54e34f889e97de4d/run.go
[EB-TEXT]: https://github.com/hajimehoshi/ebiten/blob/dcb93b02c135c20ea835952e54e34f889e97de4d/text/v2/gotext.go
[EB-WINDOW]: https://github.com/hajimehoshi/ebiten/blob/dcb93b02c135c20ea835952e54e34f889e97de4d/window.go
[EG-CONTEXT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/context.rs
[EG-DND]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/drag_and_drop.rs
[EG-DOCK]: https://github.com/Adanos020/egui_dock/blob/cdcdb359af4b74f8cc7921091db7217f0759e4f7/src/dock_state/mod.rs
[EG-EDITSTATE]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/widgets/text_edit/state.rs
[EG-FONTS]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/epaint/src/text/fonts.rs
[EG-ID]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/id.rs
[EG-INPUT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/input_state/mod.rs
[EG-INTERACTION]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/interaction.rs
[EG-LAYOUT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/layout.rs
[EG-MEMORY]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/memory/mod.rs
[EG-PAINTER]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/painter.rs
[EG-PLACER]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/placer.rs
[EG-SCROLL]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/containers/scroll_area.rs
[EG-TESSELLATOR]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/epaint/src/tessellator.rs
[EG-TEST]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui_kittest/src/lib.rs
[EG-TEXT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/epaint/src/text/text_layout.rs
[EG-UI]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/ui.rs
[EG-VIEWPORT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui/src/viewport.rs
[EG-WGPU]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui-wgpu/src/renderer.rs
[EG-WINIT]: https://github.com/emilk/egui/blob/f65f89597cc29867f87ebc49dcbf5dd2ef4ff20c/crates/egui-winit/src/lib.rs
[FL-INTERNAL]: https://docs.flutter.dev/resources/inside-flutter
[FY-ACCESS]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/internal/driver/glfw/accessibility_windows.go
[FY-ACCESSFALLBACK]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/internal/driver/glfw/accessibility_notdarwin.go
[FY-TABLE]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/widget/table.go
[FY-THREAD]: https://github.com/fyne-io/fyne/blob/0a528b3a387bff10487deb866dfb40c8e90fe1cf/thread.go
[GG-A11Y]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/a11y/tree.go
[GG-GPUVIEW]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/core/gpuview/widget.go
[GG-WIDGET]: https://github.com/gogpu/ui/blob/4d78f0d03a87c5c48bb66d3dc34f625de43067d0/widget/widget.go
[GIO-ANDROID]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_android.go
[GIO-APP]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/app.go
[GIO-BUTTON]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/widget/button.go
[GIO-DARWIN]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_darwin.go
[GIO-DRIVER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/internal/driver/driver.go
[GIO-EDITOR]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/widget/editor.go
[GIO-FLEX]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/layout/flex.go
[GIO-GPU]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/gpu.go
[GIO-GPUAPI]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/api.go
[GIO-HEADLESS]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/headless/headless.go
[GIO-INSTALL]: https://gioui.org/doc/install
[GIO-INSTALL-LINUX]: https://gioui.org/doc/install/linux
[GIO-LIST]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/layout/list.go
[GIO-OPS]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/op/op.go
[GIO-ROUTER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/input/router.go
[GIO-SEMANTIC]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/semantic/semantic.go
[GIO-SHAPER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/text/shaper.go
[GIO-TEXT]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/text/gotext.go
[GIO-TEXTBENCH]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/widget/text_bench_test.go
[GIO-TRANSFER]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/io/transfer/transfer.go
[GIO-WAYLAND]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_wayland.go
[GIO-WINDOW]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/window.go
[GIO-WINDOWS]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_windows.go
[GIO-X11]: https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/app/os_x11.go
[GO-CGO]: https://go.dev/src/runtime/cgocall.go
[GO-GC]: https://go.dev/doc/gc-guide
[GT-FONTS]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/fontscan/fontmap.go
[GT-INPUT]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/shaping/input.go
[GT-SHAPING]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/shaping/shaping.go
[GT-WRAP]: https://github.com/go-text/typesetting/blob/e49e2bdf1e4b82a816d663c5861f830800e2ae3b/shaping/wrapping.go
[IM-BACKENDS]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/docs/BACKENDS.md
[IM-CORE]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui.cpp
[IM-DOCK]: https://github.com/ocornut/imgui/blob/778992face67ecefaa5c2932db4b2101545ea4df/imgui.cpp
[IM-DOCKHEADER]: https://github.com/ocornut/imgui/blob/778992face67ecefaa5c2932db4b2101545ea4df/imgui.h
[IM-DRAW]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui_draw.cpp
[IM-FAQ]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/docs/FAQ.md
[IM-FONTS]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/docs/FONTS.md
[IM-HEADER]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui.h
[IM-INTERNAL]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui_internal.h
[IM-README]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/docs/README.md
[IM-TABLES]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui_tables.cpp
[IM-TESTENGINE]: https://github.com/ocornut/imgui_test_engine
[IM-WIDGETS]: https://github.com/ocornut/imgui/blob/2859f6723c13de3a9661b8087fb7919fa3796813/imgui_widgets.cpp
[LINUX-PORTAL]: https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.FileChooser.html
[PC-APP]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/lib.rs
[PC-AUTOMATION]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/automation/src/server.rs
[PC-CANVAS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/canvas.rs
[PC-COMMANDS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/commands.rs
[PC-CONTROL]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/control.rs
[PC-DEPS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/Cargo.toml
[PC-DOCK]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/dock.rs
[PC-DOCKTEST]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/dock_tests.rs
[PC-ENGINE]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/lib.rs
[PC-GPU]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/gpu/src/lib.rs
[PC-GPUCANVAS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/gpu_canvas.rs
[PC-HEADLESS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/automation/src/headless.rs
[PC-JOBS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/engine/src/jobs.rs
[PC-JOBSUI]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/jobs_ui.rs
[PC-LAYERS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/layer_tree_ui.rs
[PC-MENU]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/native_menu.rs
[PC-PANELS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/panels.rs
[PC-RASTER]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/raster/src/lib.rs
[PC-TABLET]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/tablet/src/lib.rs
[PC-THEME]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/theme.rs
[PC-WIDGETS]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/crates/ui-egui/src/widgets.rs
[PC-WORKSPACE]: https://github.com/storytold/photocraft/blob/0d10f117333c95b11803c13725279f68462408be/Cargo.toml
[QT-ACTION]: https://doc.qt.io/qt-6/qaction.html
[QT-DOCK]: https://doc.qt.io/qt-6/qdockwidget.html
[QT-SCENE]: https://doc.qt.io/qt-6/qtquick-visualcanvas-scenegraph.html
[SDL-DIALOG]: https://github.com/libsdl-org/SDL/blob/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/src/dialog/SDL_dialog.c
[SDL-GPU]: https://github.com/libsdl-org/SDL/blob/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/src/gpu/SDL_gpu.c
[SDL-KEYBOARD]: https://github.com/libsdl-org/SDL/blob/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/include/SDL3/SDL_keyboard.h
[SDL-VIDEO]: https://github.com/libsdl-org/SDL/tree/bb93c303acc0a1cb9b3fc45936e53b0f0a98c20f/src/video
[SL-ACCESS]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/winit/accesskit.rs
[SL-API]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/api.rs
[SL-BINDINGS]: https://github.com/slint-ui/slint/tree/c44731f8e2939eac66502fc1a521a839922d46be/api
[SL-COMPILER]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/compiler/llr/lower_to_item_tree.rs
[SL-INPUT]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/input.rs
[SL-ITEM]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/item_tree.rs
[SL-LAYOUT]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/layout.rs
[SL-LICENSE]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/LICENSE.md
[SL-MCP]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/testing/mcp_server.rs
[SL-PARTIAL]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/partial_renderer.rs
[SL-PLATFORM]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/platform.rs
[SL-PROP]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/core/properties.rs
[SL-RENDERERS]: https://github.com/slint-ui/slint/tree/c44731f8e2939eac66502fc1a521a839922d46be/internal/renderers
[SL-TEST]: https://github.com/slint-ui/slint/blob/c44731f8e2939eac66502fc1a521a839922d46be/internal/backends/testing/README.md
[VS-COMMAND]: https://code.visualstudio.com/api/extension-guides/command
[VS-CONTEXT]: https://code.visualstudio.com/api/references/when-clause-contexts
[WG-DEVICE]: https://github.com/oliverbestmann/webgpu/blob/de8bdd5d5976db28883db41211e049446078f8fd/wgpu/device.go
[WIN-DPI]: https://learn.microsoft.com/en-us/windows/win32/hidpi/high-dpi-desktop-application-development-on-windows
[WIN-TSF]: https://learn.microsoft.com/en-us/windows/win32/tsf/text-services-framework
