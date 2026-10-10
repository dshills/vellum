# Prototype plan and performance acceptance criteria

Research date: 9 October 2026. **Latest implementation: [foundation integration results](FOUNDATION_INTEGRATION_RESULTS.md) and [ADR 0007](decisions/0007-foundation-integration.md). Native preflight/event pumping, lifetime/GPU cleanup, deferred font resources and macOS packaging are now implemented and checked locally. G1 remains open; older results below preserve their original experimental status.** **P0 has now been implemented and run on local macOS; see [P0_RESULTS.md](P0_RESULTS.md) and [ADR 0001](decisions/0001-p0-foundation.md). A bounded macOS P1–P3 feasibility trial has also run; see [FOUNDATION_RESULTS.md](FOUNDATION_RESULTS.md) and [ADR 0002](decisions/0002-foundation-gaps.md). A bounded follow-up locally resolves the representative single-line grapheme failure and demonstrates native selection, lazy spelling and opt-in demand painting; see [REMEDIATION_RESULTS.md](REMEDIATION_RESULTS.md) and [ADR 0003](decisions/0003-remediation-feasibility.md). P1–P3 remain partial, with candidate-panel placement, multiline/bidi editing, full screen-reader workflows, font-memory bounds and the platform matrix still open. A single-field native composition client is now implemented; see [NATIVE_INPUT_RESULTS.md](NATIVE_INPUT_RESULTS.md) and [ADR 0004](decisions/0004-native-input-prototype.md). Live Japanese/Chinese conversion, Undo/Redo and cancellation pass; visual candidate-panel placement and full VoiceOver workflows remain unverified. Original input, Dictation and VoiceOver settings were restored after approved tests. Remaining local foundation checks now have isolated multiline-grapheme/RTL-cluster fixes and lifetime, font-memory, queue and package evidence; see [FOUNDATION_CHECKS_RESULTS.md](FOUNDATION_CHECKS_RESULTS.md) and [ADR 0005](decisions/0005-foundation-checks.md). A separate [foundation contracts follow-up](FOUNDATION_CONTRACTS_RESULTS.md) now demonstrates opt-in exact multiline text and a bounded ordered mailbox; see [ADR 0006](decisions/0006-foundation-contracts.md). The mailbox was not connected to native preflight/event pumping in that experiment; ADR 0007 records the subsequent integration and remaining global memory/platform/live accessibility/distribution checks. G1 has not passed. Workbench expansion remains on hold. P4–P11 remain designed experiments.** Budgets are provisional engineering acceptance targets, not measured results, competitor claims, or promises. Revise a budget only with recorded workload/hardware evidence.

## Sequence and decision gates

Run P0 first. Then retire native surface/packaging, text and accessibility risks in P1–P3 before expanding widgets. P4–P8 test the core contracts. P9–P11 test integration and sustained behavior. Each prototype should remain small and disposable: one architectural question, fixed input data, explicit metrics, a brief decision record and a reproducible runner. Do not turn the first successful window demo into the permanent framework by accident.

| Gate | Experiments | Exit decision |
|---|---|---|
| **G0: build versus extend** | P0 | Select A/D if it meets requirements with less ownership; otherwise justify testing B and optionally C |
| **G1: viable desktop foundation** | P1, P2, P3 | Confirm native windows/GPU/builds, multilingual editing and screen-reader route across the platform matrix |
| **G2: viable core and workbench** | P4, P5, P6, P7, P8 | Confirm GPU contract, event/state ownership, virtualization, docking/windows and worker lifecycle |
| **G3: defensible production commitment** | P9, P10, P11 | Confirm testability, semantic transport, native services, packaging, memory and latency under load |

A failed gate can trigger a targeted fix and rerun or a change of foundation. If both A/D and B fail the accessibility/text/native-services gates within a bounded investigation, narrow the supported application class or defer the framework; do not hide the gap behind a roadmap checkbox.

## P0. Compare existing Go foundations

**Question:** Is the missing product primarily a workbench package rather than a new UI core?

**Candidates:** Gio full stack (A), Cogent Core extension (D), and a small independent shell/render spike (B only after baseline inspection). Include Fyne for data-widget ergonomics and GoGPU UI as an emerging custom-canvas challenger where useful. Limit C to its shell-specific question.

**Fixture:** Identical mock engineering workspace: document tabs, property inspector, filtered tree, sortable virtual table, command palette, background calculation, and an embedded canvas. The domain model and generated dataset are identical across candidates. Implement only enough behavior to expose extension boundaries, not a polished product.

**Record:** Application code complexity; custom patches and private imports; work needed for shared-device rendering; native semantic/IME access; clean-build dependencies; maintainability of commands/tasks. Keep framework-specific prototypes separate so one wrapper does not handicap a contender.

**Pass/decision:** Prefer A/D if all hard gates are achievable through supported APIs or small maintainable upstream changes. Choose B for subsequent work only when concrete gaps outweigh its extra core ownership. A wrapper replacing most of another toolkit is a failed reuse strategy. Record the decision in an ADR with source/patch evidence.

## P1. Window, GPU, idle scheduling, and clean build

**Question:** Can the candidate create and present a native surface on every target, sleep when idle, and produce a reproducible distributable?

**Compare:** SDL3 + maintained wgpu-native Go binding versus SDL3 GPU; Gio normal renderer as control. For C, use Gio CustomRenderer/native handles without private imports. Render a small clipped triangle/rectangle scene with one texture and a resize indicator.

**Exercise:** Windows D3D12, macOS Metal, Linux Vulkan on Wayland and X11; minimize/restore, occlusion, maximize, repeated resize, DPI change, unsupported adapter, surface recreation and shutdown. Capture foreign callbacks/thread affinity, errors and cleanup. Inspect CGO/FFI counts and binary dependencies. Build on clean target CI runners; run packages on machines without the development tree.

**Measure:** Time to first *interactive presented* frame, shader/device setup, idle CPU/wakeups, frame pacing, process memory, native artifacts and clean build steps.

**Pass:** Meets shell startup/idle budgets below; correct scale/surface lifecycle; actionable error on missing GPU/library; no dependency on files left in a developer's checkout. **Stop/change:** A candidate requiring a new native backend or long-lived unsupported fork fails the reuse premise. Missing legacy-GPU fallback is an explicit support decision, not silent success.

## P2. Text rendering and real IME editing

**Question:** Does the selected text/platform combination support professional multilingual input and stable measurement?

**Fixture:** One multiline editor plus a property-field row using go-text; a fixed licensed font corpus and controlled system-fallback variants. Strings cover Latin kerning/ligatures, Arabic/Hebrew with numbers and punctuation, Devanagari, CJK, combining accents, emoji ZWJ sequences, mixed fonts and long wrapped lines.

**Exercise:** Logical/visual caret navigation; grapheme deletion; selection and clipboard round trip; undo groups; wrap changes; font fallback; 100/125/150/200% scale; cross-monitor movement. Enter and cancel Japanese/Chinese composition using real platform IMEs. Verify candidate-window position and composition selection after scroll, scale and focus changes. Convert UTF-8/UTF-16/cluster positions explicitly. Test font discovery with a cold cache and unavailable preferred fonts.

**Measure:** Cold discovery time, shaping cache hit/miss cost, glyph upload volume, cache size, edit-to-present latency and text correctness. Compare shaping outputs to appropriate upstream/reference fixtures; compare caret behavior with native editors where semantics are ambiguous.

**Pass:** Exact text preservation, grapheme-correct editing, correct mixed-direction selection, visible/correct IME composition and bounded caches. Meets warm-text/edit budgets. Color emoji and rich text are recorded as separately supported capabilities. **Stop/change:** Shaping success without correct native editing does not pass.

## P3. Desktop accessibility bridge

**Question:** Can the semantic model drive real native assistive technology, including editing and virtualized data?

**Fixture:** Button, checkbox, menu, tree, virtual table, error-bearing property editor, multiline text and a canvas exposing two selected domain objects. Use AccessKit C through a minimal Go bridge for B/C, or the candidate toolkit's supported bridge for A/D.

**Exercise:** VoiceOver on macOS, NVDA on Windows, Orca in a Linux AT-SPI session. Test focus navigation, roles/names/values, expanded state, selection, text ranges/caret, invoke/set-value actions, announcements and window transitions. Query while the renderer is idle or the window is occluded. Scroll/remove rows and ensure stable identity and correct accessibility traversal.

**Measure:** Tree size/update payload, action round-trip latency, callback thread/lifetime behavior, and update cost with accessibility active and inactive.

**Pass:** A user can complete the fixture using keyboard and screen reader; actions reach the same command/input validation path; virtual rows are discoverable without materializing the full dataset; no stale/freed-node access. **Stop/change:** An inspectable JSON tree without functioning OS adapters fails. This gate precedes marketing any accessibility claim.

## P4. Custom GPU canvas on the UI device

**Question:** Can application rendering coexist with UI clipping, text and multiple views without CPU readback or unsafe resource ownership?

**Fixture:** A simple 3D mesh viewport, a 4096×4096 RGBA image with changed tiles, a waveform with one million source samples reduced to visible resolution, and a schematic canvas with 100,000 indexed primitives and culling. Distinguish total dataset size from submitted geometry.

**Exercise:** Shared device/queue; persistent offscreen texture; render and compute preparation; clipped/scaled composition; MSAA resolve; sRGB/linear and premultiplied-alpha checks; floating-point targets; async screenshot readback; repeated create/destroy; resizing and simulated device-generation invalidation. Use actual device-loss mechanisms where feasible and a controlled adapter fault path otherwise; label which was tested.

**Measure:** UI CPU, GPU timestamps where available, upload bytes, draw/pass counts, frames in flight, foreign calls, resource accounting, readback stalls and end-to-end pacing.

**Pass:** Canvas satisfies its budget while chrome remains responsive; static canvas reuses resources; no per-frame full-image upload or CPU readback in normal composition; no stale target use. **Stop/change:** If advanced rendering forces an independent incompatible device, redesign the seam before adding widgets.

## P5. Minimal widget identity, actions, and lifecycle

**Question:** Can the API remain simple Go while preserving focus/state and consuming side effects exactly once?

**Fixture:** Button, numeric field, text field and a keyed list that can reorder/filter/delete entries. Use a single UI declaration pass and internal measure/arrange phases. Bind controls to a small command dispatcher and persistent typed state records.

**Exercise:** Duplicate labels, translated labels, reorder while focused, row removal during pointer capture, disable after input but before dispatch, modal shortcut capture, rapid press/release in one frame, virtualized disappearance, double-click, and internal layout retries. A “start task” command increments a counter so duplicate execution is observable.

**Measure:** Allocations after warmup, retained record count after churn, build/layout/paint cost, hit-geometry consistency and event consumption.

**Pass:** Stable IDs preserve the correct entity state; invalid/cancelled actions do not execute; one accepted click starts exactly one task; disappearing widgets cancel capture appropriately; no state growth proportional to historical IDs. **Stop/change:** Fix ownership/API semantics now if callers need hidden global state or unsafe mutation.

## P6. Large table and tree virtualization

**Question:** Does interactive cost scale with visible content, while preserving selection, search and semantics over large datasets?

**Fixture:** One million table rows × 20 columns, with about 50 visible rows × 8 visible columns; 100,000-node tree with controlled expansion and at most 500 visible nodes. Use stable keys, fixed-height and variable-height variants, frozen headers/columns, selection and an editable cell. Generate a deterministic seed and keep data allocation outside UI-overhead accounting.

**Exercise:** Scroll, horizontal pan, jump-to-key, keyboard page movement, sort/filter in a cancellable worker, updates arriving while editing, all-rows selection without storing one widget per row, and accessibility virtual ranges. Include long strings and cold text caches as a separate workload.

**Measure:** Visible callbacks, layout time, frame-time percentiles, allocations, UI-specific retained bytes, sorting time, and cache behavior. Compare 10,000 versus one million rows at the same visible count to detect hidden O(total rows) frame work.

**Pass:** Meets table/tree budgets; smooth warm scrolling; no per-frame full-model scan; stable selection through sort/filter; memory proportional to data plus bounded UI/index state. Document the intentional cost of variable-height indexing. **Stop/change:** Do not mask O(N) layout by testing only small models.

## P7. Docking, multiple windows, and persistent layout

**Question:** Can a retained docking graph safely manage panes across windows and monitor changes?

**Fixture:** Six panes with split/tab nodes and versioned persistence; two native windows sharing resources. Keyboard commands move/switch panes; pointer dragging resizes, reorders and detaches where supported.

**Exercise:** Drag/drop cancellation, minimum sizes, tab reorder, detach/reattach, focused text during moves, closing a source window mid-drag, monitor removal, mixed DPI, restart restore, corrupted/old layout data, and hidden panels. Wayland must use permitted window behavior rather than assume global positioning.

**Measure:** Resize latency, layout/semantic update cost, resource retention after window closure, focus routing and persistence correctness.

**Pass:** No content/state loss, correct focus/native window ownership, layout restores into available work areas, and shared GPU resources survive until their last user closes. **Stop/change:** If native detachment is not viable on a platform, explicitly narrow the feature or change the shell; in-window docking alone does not pass the multi-window requirement.

## P8. Async Go tasks and transactional commit

**Question:** Can concurrency improve application code without hiding races or lifecycle behavior?

**Fixture:** Cancellable CPU export/calculation, an I/O task and a streaming instrument source. Each task has progress, result/error, target revision and scope; UI exposes task status and cancellation.

**Exercise:** Progress floods, saturation of a bounded worker pool, task failure/panic, cancellation at each stage, document edits/undo while work runs, document/window closure, application shutdown, and completion racing cancellation. A deliberately non-cooperative worker tests the policy for abandoning results and reporting delayed shutdown.

**Measure:** Progress display lag, cancellation acknowledgement/worker exit separately, UI responsiveness, goroutine count, queue size, allocations and race-detector results. Do not use race-instrumented binaries for timing comparisons.

**Pass:** No worker mutates UI state directly; stale results never commit; progress cannot block workers indefinitely; no leaked goroutines or retained documents after cooperative completion; one worker panic becomes a failed task without publishing partial state. Cancellation cannot promise forced goroutine termination. **Stop/change:** Replace implicit scheduling with an explicit API if lifecycle cannot be explained and tested.

## P9. Headless behavior, rendering, and replay

**Question:** Can ordinary Go tests cover UI behavior independently of a desktop session, while retaining meaningful GPU/OS coverage?

**Fixture:** The P5/P6 workspace with injected clock, event queue, deterministic fonts, task executor and ID lookup. Record a sequence opening a project, editing a property, sorting a table, cancelling a task and undoing.

**Exercise:** Command tests without UI; synthetic input tests through real routing; layout and semantic assertions; software/reference output for simple primitives if available; GPU screenshots on a pinned backend; replay under different event batching; explicit “run until idle” with a maximum step count.

**Measure:** Repeatability, test startup/runtime, screenshot variance, replay divergence and ease of diagnosing failure.

**Pass:** Deterministic logic/layout snapshots for fixed inputs; no sleeps in logic tests; failures identify widget/command/revision; GPU goldens use exact comparison only on controlled configurations and documented tolerances otherwise. **Stop/change:** If test-only code bypasses essential input or command validation, redesign the harness.

## P10. Semantic control, native services, and distribution

**Question:** Can external control reuse safe application semantics while the native app behaves correctly outside its development environment?

**Fixture:** An inspector showing component R17, value 10k, package 0603 and rotation 90; commands to select, rotate, change value and save. Expose a versioned local socket/pipe API first; optionally map it to MCP after the underlying contract works. Use a packaged desktop build with native menus/dialogs.

**Exercise:** Inspect/query visible and virtual objects; invoke typed commands; reject unknown/malformed/disabled/stale requests; preserve request IDs and audit results; handle client disconnect and cancellation; redact a sensitive field. Verify test, menu, shortcut and external execution reach identical domain validation. Test clipboard text/image/MIME variants, inbound/outbound DnD, file dialogs, tray and notifications with capability reporting.

**Distribution checks:** Windows clean user machine; macOS app bundle with signing/notarization workflow and included native libraries; Linux minimum supported distribution under Wayland and X11, including a portal/sandbox package. Record every runtime library/driver requirement, build artifact hash, license notice and packaging step. Confirm operation without developer environment variables.

**Pass:** Structured automation works without screenshots; local endpoint is opt-in and access-controlled; native services meet declared support; reproducible package works on clean systems. **Stop/change:** Do not turn unvalidated transport access into an implicit application control surface.

## P11. Integrated latency, allocation, and endurance

**Question:** Do the architectural benefits survive real workload composition, long sessions and Go GC activity?

**Fixture:** All accepted components in one workspace. Repeat scroll/edit/resize/canvas activity while workers allocate and complete. Run steady-state traces for at least 10 minutes, a memory/lifecycle loop for 1,000 open/close cycles, and an overnight idle/background soak before a production commitment.

**Measure:** All budgets below, plus Go heap/GC metrics, RSS/private memory, native allocations, GPU resource bytes, goroutines, queues, dropped frames and input-event backlog. Separate cold-cache and warm-cache runs. Save traces, profiles and machine-readable samples, not just averages.

**Pass:** No unbounded growth after caches reach their stated budgets; no lost actions or stale commits; no unexplained latency spikes hidden by average FPS; regression thresholds met. **Stop/change:** Identify workload-specific bottlenecks and retest only affected gates; do not broaden scope to mobile/browser or another GPU backend to avoid fixing desktop failures.

## Measurement contract

Use a recorded baseline with at least a four-core CPU, integrated GPU and 8 GiB RAM; include an Apple Silicon Mac, Windows machine and Linux machine. Prefer the same x86 hardware for Windows/Linux where practical, but do not present cross-OS numbers as hardware-controlled if it differs. Record exact CPU/GPU/driver, RAM, OS/compositor, display refresh, scale, power mode, toolchain, dependency SHAs and build flags.

Primary view: 1920×1080 logical work area at 100% scale, roughly 1,000 visible controls/text runs and one active canvas. Also run 2560×1440 at 144 Hz and 3840×2160 at 200% scale. Clearly distinguish logical dimensions from physical pixels. Use release builds, fixed data seeds and locally stored assets. Run warmup, then at least five measured runs; publish p50/p95/p99, maximum, sample count and confidence/variation across runs.

Frame intervals are **16.67 ms at 60 Hz, 8.33 ms at 120 Hz and 6.94 ms at 144 Hz**. CPU submission, GPU execution and compositor presentation can overlap; report them separately and measure the actual critical path rather than summing unrelated percentiles. Present timestamps are not photon timestamps. Hardware camera/photodiode validation is a separate end-to-end latency test.

| Metric / workload | Provisional target | Measurement boundary and interpretation |
|---|---|---|
| Warm startup, empty workbench | p95 ≤300 ms | Process launch → first interactive presented frame; excludes OS first-run trust prompts |
| Cold startup | p95 ≤1,000 ms | Cold app/font/shader cache, local SSD; report each setup stage; do not hide first-run font scanning |
| Static idle | Mean <0.5% of one logical CPU; ≤2 app-originated wakeups/s after settling | 60-second interval, no animation/caret/tasks; no repeated presents; also measure occluded/minimized |
| Go heap / process memory, empty workbench | Live Go heap ≤40 MiB; process private/resident footprint target ≤150 MiB | Record native/driver contribution separately; platform memory definitions differ; baseline includes 1,000 controls and fonts, no domain dataset |
| Framework-controlled GPU memory | ≤96 MiB at 1080p empty workbench | Include tracked textures/atlases/buffers/swapchain estimate; driver-private allocation reported as unknown if unavailable |
| UI CPU work at 60 / 120 / 144 Hz | p95 ≤4 / 3 / 2.5 ms | Input processing + build/layout + encoding, excluding intentional application work and present wait |
| UI GPU work | p95 ≤2 ms at 1080p | Chrome/text only; timestamp queries if supported, otherwise mark unavailable |
| Integrated 60 / 120 / 144 Hz pacing | ≥99% of intervals within 1.1× target period; report every >2-period stall | Warm interaction trace; separate refresh/compositor quantization and startup stalls |
| Input delivery → render submission | p95 ≤5 ms at 60 Hz; ≤3 ms at 120/144 Hz | Monotonic native-event receipt to relevant GPU submission; also measure dispatch-only cost |
| Input delivery → actual presentation | p95 ≤33.3 ms at 60 Hz; ≤20 ms at 120/144 Hz | Native present feedback where available; otherwise label estimate; no claimed physical input-to-photon measurement |
| Warm layout, 1,000 visible controls | p95 ≤1 ms | Layout only; cache misses recorded separately; test resize invalidation |
| Warm cached text, 1,000 short runs | p95 ≤1 ms CPU, no glyph upload when unchanged | Layout/cache lookup + paint generation; render GPU time separate |
| New text paragraph, 2,000 code points | p95 ≤4 ms shaping/layout | Fixed font fixture, multilingual corpus, warm font data but uncached paragraph; distribution by script |
| Table scroll, P6 | Total UI CPU p95 ≤4 ms, p99 ≤6 ms | One million rows, ~400 visible cells; compare same viewport at 10k rows; warm text cache |
| Tree interaction, P6 | Total UI CPU p95 ≤4 ms | ≤500 visible nodes, 100k model nodes; expanding huge branches indexes incrementally |
| Canvas, P4 | Combined CPU critical work p95 ≤4 ms and GPU scene+UI p95 ≤5 ms for 60 Hz profile | Report each scene separately; 120/144 Hz is a separate reduced/optimized scene tier, not automatic |
| Resize | No application-caused stall >50 ms; UI work p95 ≤8 ms during trace | Continuous resize, mixed DPI and window recreation; distinguish OS modal resize behavior |
| Warm UI allocation | Target ≤64 KiB/frame and ≤100 allocations/frame for baseline controls | Measure framework-attributed work; caches/resource creation excluded only when separately reported; zero idle frame work |
| Task progress / cancellation | Progress visible ≤100 ms; cooperative cancellation acknowledgement ≤100 ms | Worker exit measured separately; algorithms must provide bounded cancellation checkpoints |
| Lifecycle retention | No positive unbounded trend after warmup and 1,000 cycles | Track Go, native and GPU bytes and goroutines; retained caches must fit explicit budgets |

The large-canvas texture budget is workload-specific: a 4096² RGBA8 level is 64 MiB before mipmaps, MSAA, staging and additional views; a full mip chain is about one third more. Do not judge an image editor against the empty-shell GPU budget. For P4, begin with a 512 MiB framework/application-tracked GPU budget and demonstrate eviction or graceful refusal, rather than allowing unlimited caching. These figures are calculated storage sizes and proposed limits, not measurements.

## Reporting and final commitment

Each experiment delivers: exact revision/build instructions; platform/hardware manifest; deterministic fixtures; raw timing/memory/error data; relevant screenshots/traces; pass/fail by requirement; observed limitations; and a short keep/change/reject decision. Keep failures in the report.

At G3, choose one platform layer and one renderer, record an API/ownership ADR, define supported OS/GPU/IME/screen-reader versions, identify maintainers, and estimate work from the prototypes. Only then begin production implementation. If an existing Go framework meets the same gates with smaller maintenance exposure, make Vellum a workbench extension rather than a new core.
