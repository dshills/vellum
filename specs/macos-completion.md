# macOS framework completion

The active completion target is a professional Go-native UI framework on the local
macOS development machine. Windows/Linux qualification is deferred. Existing
portable packages and target adapter sources are retained, but cross-platform
acceptance does not define this milestone's success.

Completion requires evidence for the intended professional application class:
application ownership and cancellation; standard controls and keyboard workflows;
commands; document edits and undo/redo; docking and independent windows; scalable
data views; custom GPU drawing; native clipboard, menus, choosers and file drops;
high-DPI geometry; functioning native accessibility/input; deterministic tests;
measured sustained performance; independently consumable SDK and Mac packaging.
The research/prototype plan and detailed subsystem contracts remain authoritative
for their acceptance workloads. Historical claims are not current qualification.

Managed-grid accessibility now has an explicit per-grid pending/completed
transaction: stale actions are revoked immediately, last-completed cells remain
readable, and matching semantics/frames publish atomically with stable native
identity. Replacement and real removal retire old objects. The separate stored
400-cell Cocoa acceptance mode and scoped lifecycle/race fixtures qualify this
behavior. Full-frame latency targets and human screen-reader checks remain open.

## Current implementation and remaining work

Application ownership, cancellable tasks, commands, controlled standard controls,
keyed reconciliation, layout/theme, docking/window sessions, virtual tables/trees,
shared-device rectangle canvas, Cocoa clipboard/menus/file panels, demand painting,
resource lifecycle checks and private SDK/bundle tooling exist.

Native dispatch now has explicit synchronous failure and desktop wake termination
semantics. See [contract](native-dispatch.md) and its evidence; broad panic resource
recovery remains outside that contract.

Document history adds bounded sparse undo/redo, revision-aware dirty/save handling
and focused native text precedence. [Document persistence](document.md) adds asynchronous save/open and native
unsaved-close prompts with a working component-document consumer. Its bounded
acceptance evidence remains separate from actual file-panel selection.

External file-drop callbacks and Mac file-URL routing are implemented; physical
Finder workflow acceptance remains separate. See [contract](file-drop.md).

[Retained GPU drawing](drawing.md) adds arbitrary indexed 2D meshes, strokes,
waveform envelopes and rectangle spatial culling, with an independent consumer.
[Retained images](images.md) add a real 4096² native texture workload with changed
tile uploads, alpha-safe filtering and independent-window retention evidence.
[Advanced GPU views](advanced-gpu.md) add custom shaders and a retained 3D
viewport with optional depth and native fault/resize/closure evidence.
An independent imaging consumer adds ordered compute/float-target preparation
and bounded asynchronous captures. P4 still needs broader performance qualification. The public [multiline editor](editor.md) now provides a supported model, managed
control and independent document application beyond the historical prototypes;
its native input/geometry evidence and separate read-only external AX check pass
against frozen sources. Human IME and VoiceOver acceptance remain separate. Semantic command/UI inspection exists through APIs
and native accessibility. Portable execution now has injected real/manual
workers and clocks, reserved owner results, bounded drain batches and separate
worker/delivery accounting.

The [managed replay checkpoint](replay.md) is internally qualified on the
offscreen software backend with supplied, digest-checked fonts and a declared
theme. It uses actual retained-widget pointer/key routing, focus/modal behavior,
measured geometry and completed software frames. One fixed 51-operation record
with nine captures matches exact model, semantics, focus, UTF-16 selection,
geometry and image identities at batches 1, 7 and 64. Explicit worker selection,
clock advancement and originating timer dispatch preserve observable pending
work; future-only timers do not force an implicit advance. The scope is one
active managed application per process, with multiple windows sharing its clock.
See the [checkpoint report](../docs/reports/macos-replay/README.md).

The [indexed grid foundation](grid.md) now provides bounded fixed or indexed variable-height,
two-axis portable geometry with frozen headers/columns and stable cell identity.
Its 10k and million-row procedural fixtures have identical visible callback work.
That proves the projection cost. The [managed adapter](grid-native.md) adds
bounded native cell pools, clipped measured geometry, routed pointer/keyboard
navigation and visible-cell accessibility. Retained dataset memory, frame
performance and the complete P6 table/tree
workload remain open.

[Captured row selection](grid-selection.md) adds range/All membership without an
owner scan, bounded key toggles and immutable worker snapshots. Mac input routes
Shift ranges, Command toggles/All and Escape clear; accepted membership drives
visible rendering and AX attributes separately from the active cursor. Captured
counts describe identities in their captured scope, independently of later
filtering. This does not complete loaded projection integration or
the sustained P6 latency/memory workloads.

[Versioned cell drafts](grid-editing.md) add application validation, explicit
conflict recovery and atomic immutable-source admission. The Mac adapter uses a
single clipped native field and originating input settlement before commit;
the independent grid consumer exercises sparse versioned edits. Source reordering
and unrelated updates retain drafts, while edited-cell updates refuse silent
overwrite. This checkpoint remains separate from loaded-data workers, sustained
P6 performance, large-tree acceptance and human IME/external AX workflows.

[Worker-built grid views](grid-projection.md) add reusable bounded immutable
sort/filter/height indexes, cancellation and opaque latest-result owner admission.
The independent loaded-data consumer stores 10k/million records ×20 numeric
fields and uses separate domain/view revisions with worker-built edit candidates.
This advances loaded projection integration; sustained latency/memory, large-tree
workloads and broader Mac acceptance remain open.

The additive [large-tree contract](tree.md) moves stored-data construction,
expansion and filtering onto bounded workers and admits sealed indexes through
`Outline`. `ui.DataTree` provides a bounded managed native row pool, opaque
captured input and completed-frame inspection. Its stored 100k-node consumer
exercises wide branches and deep paths; tree latency, native memory and human
accessibility acceptance remain separate.

[Stored-table timing](grid-performance.md) links protected actual scroll input
to owner work and completed composition, including intermediate post-input
native paints. The unchanged warm targets require actual release measurements;
the existence of instrumentation and CPU tests does not establish a target pass.

The [single host-update checkpoint](../docs/reports/macos-grid-update/README.md)
removes duplicate pooled cell/header updater/style passes with matched
offscreen software pixels. Three interleaved fixed-10k release processes per
build observe p95 19.55→17.07ms and p99 20.41→18.00ms. The unchanged targets
remain failed; the complete multi-case P6 campaign remains open.

The [unchanged-refresh checkpoint](../docs/reports/macos-grid-refresh/README.md)
reuses complete, validated grid projections and fixes embedded native-parent
identity comparisons. Same-generation refresh retains private editing, styles,
disabled/help semantics and canonical scrollbars without Cell reads or spurious
host/cell updates. Built-in projections also refuse borrowed input-stamp drift.
Original-overlay regressions fail; the final native adapter and portable virtual
race suites pass, with matching software grid pixels. Three interleaved fixed-10k
release processes per build observe p95 17.233→14.624 ms and p99
18.954→15.566 ms. Both 4/6 ms targets still fail. Raster/upload cost and the
complete multi-case P6 campaign remain open.

The [empty-sprite checkpoint](../docs/reports/macos-empty-sprites/README.md)
removes the transparent full-window source on fresh composition without a
drawable sprite. Original overlays fail meaningful assertions; core/manual and
managed retirement/cursor/dialog/popup/drag/resize regressions pass under race,
as do the full offscreen native adapter and affected consumers. Software grid
pixels remain matched. Three interleaved fixed-10k release processes per build
observe p95 14.178→13.090 ms and p99 14.891→13.294 ms. Both targets still fail;
source-level raster/upload attribution and full P6 remain open. No physical GPU
upload-count, partial-upload, memory-reclamation or new scheduling claim follows.

The [source-cost checkpoint](../docs/reports/macos-source-costs/README.md) adds an
explicit detailed recorder with at most 32 scalar records across a complete
ticket and separate 4096-ticket capacity. Default rendering retains its call
path. Race, pixel/copy parity, bounded/unwind/mutation-oracle and full adapter
checks pass. Three optimized fixed-10k diagnostic processes preserve the
400-cell geometry; CPU scene rasterization is about 79% of source Draw time.
The remaining roughly 21% is primarily the image API's elapsed work. These are
nested CPU-side diagnostics; upload failures, physical transfers, GPU execution
and presentation remain unavailable. This guides a bounded repaint investigation
without qualifying an ordinary latency improvement or closing P6.

The [row-range checkpoint](../docs/reports/macos-raster-rows/README.md) bounds
per-path scanner row work for exclusively owned native raster and glyph-cache
callers. General scanners retain direct-cell and complete-row behavior. Exact
span/state differentials and small-path benchmarks qualify that bounded change.
The ordinary four-case comparison uses five release processes per build and
4,800 validated samples. Native latency differences are small and mixed, with
every run still failing the unchanged targets; no general frame improvement or
full P6/Mac completion is inferred.

Full P9 professional workload and large-data replay acceptance, external
application automation transport, native services/input/accessibility, GPU
presentation and sustained performance qualification remain open. This bounded
software fixture does not complete the macOS milestone.

Remaining Mac acceptance includes broader external AX and human
screen-reader workflows; candidate-panel placement through scrolling/focus/scale;
cross-display DPI and monitor changes; physical GPU loss policy acceptance;
representative sustained frame/input latency, idle CPU and bounded font/native/GPU
memory; actual chooser selection, native menu activation, clipboard roundtrips and
external file-drop workflow. Automated selector/fixture tests cannot substitute
for broader user-facing acceptance. OS preference changes require specific user
authorization and restoration; routine checks do not change those preferences.

Public distribution also requires the third-party/native dependency audit.
The selected namespace and MIT license resolve owner choices but do not certify
that audit, signing/notarization or production qualification. Existing version is
v0.1.0-alpha.1 and macOS remains experimental until acceptance is complete.

## Review and stop points

Use internal agent review during implementation. Run Prism on significant completed
batches using the user's selected Anthropic Claude Sonnet 5.5, resolve findings,
run affected tests and commit stable checkpoints. Avoid repeated unchanged Prism
runs. Native acceptance must record source identity, command and exit status;
qualification remains open wherever evidence is absent.

The ordered GPU scene candidate now compiles detached paint into exact integer
compute on the originating window device, preserving the scene's existing source
ordinal and reconstructing partial work for CPU fallback/snapshots. It remains
opt-in through the native driver. Shader, real outline text/path and shared-drawer
pixel checks exercise the native GPU backend; the actual stored-table source
route is separately observed through diagnostic Cocoa frames. These bounded
checks establish their tested behavior; representative ordinary latency, broader
source eligibility, native/GPU memory and full P4/P6/Mac acceptance remain open.
See the [rendering contract](rendering.md).

The [ordinary GPU scene comparison](../docs/reports/macos-gpu-performance/README.md)
uses one optimized binary for forty interleaved CPU/GPU-option processes across
fixed/variable stored 10k/million cases. All 4,800 observations retain matched
geometry and unchanged warm targets; both modes fail those targets, and GPU tails
are mixed. Separate diagnostic processes establish the main-scene GPU route in
all four numeric-table cases. GPU mode adds roughly 11 MB of process-wide Go
allocation per observation; profiling and source inspection guide a bounded
span-storage investigation rather than establish allocation attribution. This
does not qualify broader scene support, native/GPU memory, tree/workbench or full
Mac/P4/P6 performance. The renderer remains opt-in.

The [sparse span-band checkpoint](../docs/reports/macos-span-bands/README.md)
replaces full-height construction row arrays with charged geometric bands and
keeps the existing sealed packet/shader format. Reference bytes/pixels, explicit
native parity and forty matched stored-table processes qualify that bounded
change. Observed process-wide Go allocation falls about 5.7–5.9 MB per observation;
both builds still miss unchanged warm targets. This closes neither sustained
performance nor native/GPU memory acceptance, and the renderer stays opt-in.

## Current allocation and native-service evidence

[Separate allocation diagnostics](../docs/reports/macos-allocation/README.md)
identify repeated static glyph-outline construction, scene-packet allocation and
accessibility publication as measured candidates for further Mac performance
work. Their process-wide sampled figures include measurement helpers and do not
qualify ordinary latency or retained/native/GPU memory. The next optimization
must retain bounded font admission and exact geometry/pixel/fallback behavior.

A read-only native-service inventory confirms that panel cancellation/installation
and injected document chooser smokes do not establish actual Open/Save path
selection or installed menu activation. Physical Finder drag acceptance remains
open. The text-only clipboard restoration flag is not preservation-safe for an
ordinary general pasteboard; a full roundtrip requires bounded flavor preservation
and interference handling. No gate closes from this inventory alone.

## Actual document chooser workflow

[Native chooser acceptance](../docs/reports/macos-native-choosers/README.md) now
verifies real Open and Save As selections, native cancellation for both modes,
exact temporary-file persistence, restored controls, an unchanged surviving
window and normal native closure of both editor windows. The ordinary consumer
uses real FileStore/ChooseFile services, without injected paths or chooser results.
Unique directory/file/content identities and exact output bytes qualify the
observed workflow; private session path/revision counters remain unobserved.
Earlier failed automation attempts are historical, not acceptance. This closes
this editor's selected-panel workflow, while native menus, general clipboard,
physical Finder drag and broader Mac/performance acceptance remain open.

The current `live-v20` qualification has 47 command receipts, 11 stages, exact
original/input/saved byte identity, and normal exit in 8.49 seconds. Internal
source review corrected the external acceptance harness; eight fault probes
exercise its failure records and bounds. Native focused-window observations
match both shortcuts, while foreground checking and global posting remain
non-atomic. Earlier full workflows remain historical observations, and later
failed attempts remain failed. These harness changes are not production fixes.

## Static outline allocation work

Renderer-owned outline reuse addresses the repeated construction identified by
the allocation diagnostic. It requires independent original geometry/pixels,
private ownership, cache retention limits, aggregate cold/warm GPU admission and
retirement checks, plus matched actual native CPU/GPU measurements. Cache hits
cannot relax the source support, per-glyph or logical work bounds. New cache
payload/reference limits are per renderer and do not qualify total font/native
memory. Unchanged latency targets and broader Mac acceptance remain open until
representative evidence passes them.
