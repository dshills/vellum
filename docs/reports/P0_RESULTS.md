# P0 results: extend before committing to a new core

Recorded 9 October 2026. **P0's comparison is complete for the local macOS scope. Prioritize a bounded Cogent Core extension trial (D) for the next foundation gates. Keep Gio (A) as the immediate-style control and Architecture B as the fallback. Do not commit to a new UI core yet.**

Both native workspaces support the common command/task fixture without toolkit forks or private imports. Cogent additionally rendered a scene through its public shared-device GPU integration. These observations weaken the argument that Vellum needs an independent core merely to obtain engineering widgets and a GPU canvas. They do not establish that Cogent meets the full Vellum requirements: neither prototype exposed workbench controls to native accessibility, and multilingual input, multi-platform packaging and demanding GPU lifecycle behavior remain unvalidated.

The decision is recorded in [ADR 0001](../decisions/0001-p0-foundation.md). Source and reproduction instructions are in [prototypes/p0](../../prototypes/p0/README.md).

## What was built and run

The Gio and Cogent applications share a toolkit-independent Go model with two documents, 100,000 generated component rows per document, stable component IDs, editable inspector values, rotation, filtering, sorting, six commands and a cancellable checksum task. Each workspace includes document tabs, a bounded expandable navigation tree, a table, an inline searchable command palette and a 96-component canvas. The fixture deliberately limits tree/canvas content so the experiment stays about integration boundaries.

The worker copies input, captures document identity and revision, coalesces progress notifications, and sends one result through a bounded channel. The UI owner commits results only if the original document revision still matches. Task commands are validated in the model; duplicate starts are rejected. Domain rows are separate from the sortable table projection.

Both projects built and ran natively on **macOS 26.6.2 / darwin arm64, Go 1.27.1**, using the existing Xcode toolchain. Local app bundles were used for desktop inspection. The Gio workspace was approximately 1280 pixels wide; Cogent expanded to approximately 3072 pixels. Their visible content and startup instrumentation are consequently unsuitable for speed comparisons. This was a developer-host build, not a clean-machine deployment test.

Dependencies are pinned to the research snapshots: Gio `9393ca1c7322260c50d0023a9c49643c0c89a1c0`, Cogent `5cc519b24044b34414e8dcfc0d6a33106c368f3c`; the Cogent graph uses maintained WebGPU binding `de8bdd5d5976db28883db41211e049446078f8fd`. Exact pseudo-versions and transitive modules are recorded in each `go.mod`/`go.sum` and the evidence module lists. These pins make the experiment reviewable; they are not release recommendations.

P0 did not implement B or C. Its plan made an independent shell spike conditional on finding gaps that justify additional ownership. Fyne and GoGPU remain research comparisons; they were not additional executable baselines in this bounded trial.

## Observed behavior

| Check | Gio | Cogent Core | Evidence / limit |
|---|---|---|---|
| Native workspace | Rendered | Rendered | Screenshots, native logs |
| Inspector edit and rotate | Verified | Verified by keyboard | R17 set to 22k and rotated; revision advanced |
| Filter and sort | Verified | Verified by keyboard | NET_16 reduced 100,000 to 3,125 rows; selected ID remained 17 |
| Palette search/invoke | Verified | Verified by keyboard | Cogent filtered to rotate, then Tab/Return dispatched once |
| Document edit isolation | Verified | Verified by keyboard | Separate revisions/values; final Cogent log shows one rotation on each document |
| Stale task rejection | Verified | Verified by keyboard | Status: Discarded stale calculation |
| Cancellation | Verified | Verified by keyboard | Status: Calculation: context canceled |
| Lazy table construction | Structural observation | Structural observation | Limited visible row count, not a latency measurement |
| Table scrolling | Verified by native automation | Not verified with real pointer input | Cogent automation cursor limitation described below |
| Public shared GPU integration | Ordinary Gio paint ops only | Supplemental scene rendered | Public xyzcore path plus pinned source inspection |
| Native control accessibility | Controls absent in queried AX tree | Controls absent in queried AX tree | Window chrome present; neither passes P3 |
| Timed shutdown | Exited | Exited after quit-route correction | Final local smoke runs exited with code 0 |

The model's five race-detector tests passed: projection/selection isolation, command validation, stale-result rejection, completion after switching documents, and cancellation/duplicate-start rejection. `go vet` passed in all three modules; both candidate builds passed. The Cogent linker emitted a duplicate `-lobjc` warning but produced a working executable. [Final checks](../../prototypes/p0/evidence/final-checks.txt) and [uncached model test output](../../prototypes/p0/evidence/model-tests.txt) are retained.

Gio's short startup report counted at most **17 row layout calls per frame and 17 retained row states** for 100,000 rows. This report covers a separate short startup smoke run; its three frames do not represent the longer edit/scroll session. The scrolled screenshot independently shows traversal beyond the initial rows. Cogent's final palette/tab run reported **12 visible rows**; its top-level table child count of two is container structure, not two row widgets. The source implements visible row sizing inside its list/table machinery. Neither result proves P6 performance, column virtualization, variable-height behavior or memory bounds during prolonged churn. [Gio counters](../../prototypes/p0/evidence/gio-smoke.json), [Cogent counters](../../prototypes/p0/evidence/cogent-final.json).

## Public GPU boundary

The main Gio canvas uses ordinary `clip`/`paint` operations. It proves that custom application geometry can coexist with Gio widgets, but it does not prove a public arbitrary shared-device render-pass API. Gio's renderer interface accepts Gio operations while its backend driver remains internal; an independent GPU integration would need a specifically demonstrated supported boundary. [Pinned Gio GPU API](https://github.com/gioui/gio/blob/9393ca1c7322260c50d0023a9c49643c0c89a1c0/gpu/gpu.go).

Cogent's main 2D canvas uses its painter API. A separate `-gpu-probe` renders a lit cube using public `xyz`/`xyzcore` APIs. Source inspection of `xyzcore.configFrame` shows it retrieves the window's GPU surface, then `ConfigOffscreenFromSurface` takes that surface's GPU/device. Its render source gives the texture to the window composer. The screenshot demonstrates that this path works on the tested Mac; the shared-device conclusion comes from the source path rather than timing or device instrumentation. [Pinned native integration](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/xyz/xyzcore/render_notjs.go), [scene configuration](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/xyz/render.go), [GPU screenshot](../../prototypes/p0/evidence/cogent-gpu.png).

This is useful reuse evidence for D. It is not a substitute for a Vellum custom pass with application-owned textures, clipping, resize, deferred resource disposal, multiple windows and device recovery. Those remain P1/P4/P7 work.

## Complexity and maintenance observations

The formatted Gio application has 379 lines; Cogent has 395 including the supplementary GPU probe. The shared model has 247 lines. These small prototypes do not show a decisive application-code-size advantage. Their responsibilities differ: Gio explicitly constructs rows and manages keyed interaction state; Cogent supplies Table/Tree/Tabs/Splits but needs careful event, focus, visibility and document synchronization. Counts include comments/blank lines and are not effort or maintainability scores.

Both applications use public packages, ordinary Go and the same model; there are **zero private imports, upstream patches or application code-generation steps**. Cogent's built-in table derives columns from struct fields and tags through toolkit reflection. Thus its default table does not demonstrate the proposed no-reflection-binding consumer contract. A D extension would need to prove a typed workbench facade or a modest supported custom table. Replacing most of the toolkit to achieve that contract would invalidate the reuse choice.

The native imported package closures contain 186 packages for Gio and 462 for Cogent, including the standard library and the Cogent GPU probe. CGO package lists contain 5 versus 11 entries including `runtime/cgo`; Cogent's list includes platform library carrier packages, so this is not a count of eleven live native boundaries. Unstripped recorded binaries are approximately 10.3 MiB and 50.8 MiB. These are footprint/dependency observations only, not an optimized deployment comparison. The broader Cogent closure is a real maintenance consideration; it must be weighed against the widgets and GPU functionality being reused. [Package and native dependency evidence](../../prototypes/p0/evidence/manifest.json).

The async adapters keep worker state out of toolkit widgets: Gio wakes the window with `Invalidate`; Cogent sends a coalesced public custom event and applies updates in the scene event handler. No background worker acquires a widget lock. Cooperative cancellation and revision validation are demonstrated; task pools, panic injection, non-cooperative workers, close/completion races and long-running UI race tests remain P8 work.

## Integration issues and test limitations

Several bugs belonged to the prototype and were corrected without modifying either framework:

- Gio needed to consume pending editor changes before executing a button command, so the command commits the displayed value.
- Cogent's shortcuts are window-wide. Registering them in both tabs let a hidden document overwrite a visible document's shortcut. The final version registers each shortcut once and dispatches against `CurrentTab`.
- Cogent click handling assigns button focus after an ordinary callback. Focus transfer therefore uses the public deferred callback. Dynamic palette visibility also explicitly clears `states.Invisible`; merely changing DisplayNone back to Flex was insufficient.
- The prototype initially called Cogent quit from inside `RunOnMain`, which blocked shutdown. The final version follows its upstream menu's asynchronous quit route. Both final modes exited normally. [Public focus/defer behavior](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/widgetevents.go), [quit menu](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/core/menu.go).

Native automation also had a distinct pointer limitation. Cogent's GLFW mouse-button callback queries `GetCursorPos`; the pinned Cocoa GLFW implementation obtains `mouseLocationOutsideOfEventStream`. Injected move/hover positions followed the requested target, while injected clicks used the physical cursor location and missed the widget. Keyboard tests therefore verified Cogent's commands. This is evidence about this automation route, not proof of a defect for a user moving a real mouse. Real pointer selection, scrolling, splitter dragging and resizing remain manual validation items. [Cogent native callback](https://github.com/cogentcore/core/blob/5cc519b24044b34414e8dcfc0d6a33106c368f3c/system/driver/desktop/event.go), [pinned GLFW Cocoa cursor query](https://github.com/go-gl/glfw/blob/0bd588dc30cf/v3.4/glfw/glfw/src/cocoa_window.m).

AX snapshots of both candidates exposed Cocoa window controls and menus, but no workbench buttons, fields, tree or table. Keyboard usability does not imply VoiceOver usability. P3 must produce an actual native semantic route, including editing actions and virtualized rows. A source audit or local automation endpoint cannot count as passing that gate. [Gio AX snapshot](../../prototypes/p0/evidence/gio-ax.txt), [Cogent AX snapshot](../../prototypes/p0/evidence/cogent-ax.txt).

No Windows, Linux, real IME, screen-reader completion, clean-machine distribution, FPS/p95 latency, idle CPU, heap/GPU accounting or endurance results were obtained. Startup logs use different start boundaries and are intentionally excluded from ranking. Initial diagnostic screenshots/logs predate some wrapper fixes; the manifest distinguishes them from final evidence.

## Next decision gate

Give D the first bounded foundation trial: native lifecycle/builds (P1), real multilingual editing (P2), and a native accessibility bridge (P3). Bring the existing Gio control through the same fixtures where useful. Start with P3 adapter feasibility early, because both baselines currently fail to expose controls.

Continue D only if those capabilities can be reached with public APIs and small maintainable changes, and a typed facade can retain the desired Go application contract. Move to a minimal B foundation experiment if the accessibility/native/custom-GPU route requires an unsupported fork, replacing substantial toolkit internals, or adopting unacceptable consumer binding machinery. B would then have a concrete gap to solve; it would still need to pass the same gates.

P0 provides no basis for accepting the production foundation or changing performance budgets. P1–P3 remain required before expanding widgets or committing to an independent core.
