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

## Latest unattended evidence

The [current private SDK checkpoint](../docs/reports/macos-current-sdk/README.md)
verifies 7,411 hashed files from clean `238b3b6`, including owned stages 0074/0075.
The relocated generated consumer passes race/vet/build checks, and the public API
contract module passes race/vet. Nine demo applications and four replay/workspace
commands compile from that SDK using only its local replacements. These are
consumption/compilation checks, with no native application execution or public
distribution. Native packaging/release gates and full Mac qualification remain open.

The [unique lifecycle, grid diagnostics and workspace recovery checkpoint](../docs/reports/macos-workspace-recovery/README.md)
completes 1,000 unique-name cycles and 33,000 native/owner actions on `ddd23c6`.
Eight grid diagnostic configurations on `7e97a1a` complete 960 measured samples;
CPU p95 is 7.95–8.66 ms and GPU-route p95 is 6.12–6.45 ms. These diagnostic
measurements do not qualify the ordinary 4/6 ms targets or an optimization gain.

The corrected `eab01c9` professional workspace smoke passes all four default/GPU
requested × 10k/1M configurations: 72 single-admission native accessibility
actions, same-object placement Undo/Redo, stored table/outline operations, modal
and tab actions, independent inspector survival, worker joining and tracked
native cleanup. Its earlier overscan-target failure remains failed. These
selector/owner checks do not qualify physical dragging, GPU coverage, hidden
helper deallocation, original-menu restoration, the 216-operation replay or
full P9/P11. Professional memory, sustained/ordinary performance, full idle and
overnight, physical displays/input, human accessibility and distribution remain
open. The checkpoint retains raw results, builds, validators, reviews and failure
records with separate source identities.

The [16 October runtime checkpoints](../docs/reports/macos-runtime-checkpoints/README.md)
complete 1,000 same-name windows and 33,000 native/owner actions on `ddd23c6`,
including 999 sampled replacements. Separate `97aee13` diagnostics pass five
actual GPU scene fixtures, native composer-ticket nesting and two composed
windows with 66 actions. The corrected input fixture reaches native checked-write
refusal, restoration and readback; its original failure remains retained.

Four eligible 60-second intervals in the eight-minute trace measure
0.0736–0.0786% of one CPU, with zero recorded primary paint wakes, framework
empty-event-post attempts and surface-present attempts. These observations meet
the sampled CPU component only; actual OS wakeups/presentations, continuous state,
full idle and overnight qualification remain open. The same source's private SDK
passes 7,381-file integrity and relocated consumer race/vet/build checks without
native runtime execution. Saved release-byte/source-declaration comparisons
advance dependency provenance without completing build reproduction, component
closure or rights. Unique-name endurance and later grid diagnostics are excluded
from this completed checkpoint; full Mac/P11 and ordinary frame targets remain open.

The [native GPU release identity checkpoint](../docs/reports/macos-native-release-identity/README.md)
authenticates the retained WGPU macOS ARM64 ZIP, its repository/tag/numeric IDs
and commit URI through the official GitHub release certificate policy and normal
TUF verification. The corrected verifier passes all 27 negative controls; Prism
returns three medium findings, with two fixed and one refuted for this scope.
This advances release identity only. Executed compiler/features, source-built
correspondence, component/rights closure and public Mac distribution remain open.

The human authorized unattended native and performance testing on 15 October
2026. Prior paused-test notes below describe historical checkpoints. A return
message stops live testing immediately.

The [33-action composed campaign](../docs/reports/macos-composed-campaign/README.md)
completed 1,000 windows and 33,000 owner/native action checks on `c078eb8` with
6,227 prepared source files. Its sampled safety guards and tracked teardown
passed. The next soak preflight and a diagnostic reproduction failed the strict
native occlusion predicate. The corrected fixture uses a larger opaque cover,
observes actual containment/order and still requires native occlusion; its
40-second preflight passes on the separately recorded source vector. Those
one-second quiet samples do not qualify the 60-second idle target or overnight
behavior. Full P11, professional/native/GPU memory, physical workflows and Mac
release acceptance remain open.

The subsequent eight-hour attempt failed after 9,084 seconds with
`soak occluded state drift`: its visible phase and 30 eligible quiet samples
completed, but minimized/background phases and final cycle completion did not.
That failure remains failed. The [native-loop follow-up and soak diagnostics](../docs/reports/macos-loop-native/README.md)
preserve its partial trace and cleanup separately from five successful native
cases built from clean `82c0c5e`/6,229 leaves: four queue modes and five composed
windows with 165 native action checks. A separate eight-minute four-state run
passes with one eligible 60-second interval per state, sampled CPU 0.0214–0.0672%
of one CPU and no primary paint wakes. Application-wide wakeups/presents, causal
CPU improvement, full idle and overnight qualification remain open. The new soak
failure record preserves endpoint states for future diagnosis; it does not
establish or fix the unknown cause of the original state drift.

Two separate fixed-million-row CPU-renderer diagnostic processes on that same
clean source capture a whole-process CPU profile and nested source timings.
Source raster work is about 68.1% of source Draw elapsed time; image-copy work is
about 31.9%. Their profiled/detailed UI-work p95 values are 7.209/7.153 ms and
cannot qualify ordinary timing targets or an optimization gain. Process-wide
`cgocall` stacks do not identify native CPU work or a causal frame bottleneck.
The retained profile guides further raster/copy/accessibility investigation.

The [current native grid measurements](../docs/reports/macos-post-glyph-native/README.md)
complete 40 fresh ordinary processes and 4,800 samples on `77ca764`/6,227 prepared
files, after the source-bound endurance and corrected-preflight gates. All eight
cases miss the unchanged 4 ms p95 / 6 ms p99 targets: pooled p95 is 7.49–7.89 ms
for CPU raster and 5.70–6.60 ms for the GPU scene route. This configuration run
does not attribute improvement to prior optimizations or complete P6.

The [source-bound native follow-ups](../docs/reports/macos-native-followups/README.md)
include a complete 1,000-window/24,000-step composed run on `4cf9e36`/6,209 leaves
with final cleanup, and the completed stage-59/stage-60 four-case native renderer
comparisons. They retain their exact endpoints. Sampled memory safety guards pass;
full memory budgets and native/GPU allocation remain unqualified. All renderer
comparison arms miss unchanged 4 ms p95 / 6 ms p99 targets. A new pass does not
establish the cause of the prior 343-window failure.

The [native tab checkpoint](../docs/reports/macos-tab-close-native/README.md)
passes eight actual Cocoa stages on `4c0df93`/6,214 leaves; two meaningful fault
controls retain their earlier `4cf9e36`/6,209 scope. These selector workflows
remain separate from physical input and human VoiceOver. The checked RGBA row
upload correction and new progress controls do not imply whole-frame gains.
Current-source composed endurance, complete P4/P6/P11, overnight soak, physical
input/displays, human accessibility, native/GPU memory, service interoperability
and Mac distribution remain required.

## Current implementation and remaining work

Application ownership, cancellable tasks, commands, controlled standard controls,
keyed reconciliation, layout/theme, docking/window sessions, virtual tables/trees,
shared-device rectangle canvas, Cocoa clipboard/menus/file panels, demand painting,
resource lifecycle checks and private SDK/bundle tooling exist.

Native dispatch now has explicit synchronous failure and desktop wake termination
semantics. See [contract](native-dispatch.md) and its evidence; broad panic resource
recovery remains outside that contract.

The optional [clipboard extension](clipboard.md) implements bounded text, PNG,
HTML and custom MIME representations, first-item consistency and capacity-aware
cancellation/shutdown. Its native parser and dispatch policy can be checked
without accessing the clipboard. A separate guarded Cocoa acceptance now checks
general-pasteboard text/PNG/HTML/custom data, legacy text, native alias refusal,
controlled first-item reads and independent-process verification after Go exits.
Complete bounded raw items/flavors are restored after normal completion, an
intentional fixture failure and forced fixture termination. Private-board tests
exercise both journal gaps with foreign writers and preserve the derived legacy
string alias. Third-party app compatibility, broader live providers and concurrent
writer atomicity remain separate. See the [clipboard contract](clipboard.md).

The optional [Mac tray service](tray.md) adds a retained native menu-bar item,
named/template PNG icons and owner-validated command menus. Bounded shared Cocoa
admission and a main-loop cleanup hook retire actions and native resources before
termination. Offscreen parser/bitmap/ownership and injected cleanup tests remain
separate from actual tray interaction, VoiceOver and native loop teardown.

The optional [notification service](notifications.md) implements explicit
permission snapshots/requests, immediate local delivery and scoped ID withdrawal
in Mac app bundles. Offscreen content/settings fixtures remain separate from
OS prompts, presentation, removal and restart. Notify never requests permission.

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

The opt-in [semantic automation endpoint](automation.md) implements a Mac-local
capability-protected socket with typed DTO queries/actions, revision checks and
bounded owner admission. The workbench exposes selected/stable virtual component
inspection and a narrow shared-domain action allowlist. Headless transport/domain
checks are separate from a live professional automation campaign.

Full P9 professional workload and large-data replay acceptance, external
application automation workflows, native services/input/accessibility, GPU
presentation and sustained performance qualification remain open. This bounded
software fixture does not complete the macOS milestone.

Remaining Mac acceptance includes broader external AX and human
screen-reader workflows; candidate-panel placement through scrolling/focus/scale;
cross-display DPI and monitor changes; physical GPU loss policy acceptance;
representative sustained frame/input latency, idle CPU and bounded font/native/GPU
memory; broader chooser scenarios, physical menu interaction, third-party
clipboard compatibility and external file-drop workflow. Automated selector/fixture tests cannot substitute
for broader user-facing acceptance. OS preference changes require specific user
authorization and restoration; routine checks do not change those preferences.

Public distribution also requires the third-party/native dependency audit.
The selected namespace and MIT license resolve owner choices but do not certify
that audit, signing/notarization or production qualification. Existing version is
v0.1.0-alpha.2 and macOS remains experimental until acceptance is complete.

The Mac distribution input collector adds size/hash identities for the complete
SDK source payload, resolved workbench embed assets, listed cgo sources and
module-owned native linker-search candidates. It retains actual target/Go
configuration, notice candidates and source/pin identity without running an
application. See [release contract](release.md). This prepares review of actual
fonts/native inputs; it does not approve their rights, reconstruct a static
archive's embedded component closure or close the distribution/runtime gates.

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

The separate [native file-panel lifecycle checkpoint](../docs/reports/macos-file-interaction/README.md)
passes 200 actual alternating Open/Save caller-cancellation cycles and normal
shutdown with active panels of both modes. It requires visible tracked native
objects before cancellation, empty native/Go inventories before next admission,
joined application-owned callers and terminal retirement. Counts are limited to
tracked objects and are separate non-atomic samples. This does not qualify native
heap reclamation, exceptional/concurrent-panel cleanup or physical Cancel clicks.
The physical Finder attempt remains unqualified: its UI app-selection tool
stalled before a drag was issued while the bounded fixture exited on time.

## Static outline allocation work

Renderer-owned outline reuse addresses the repeated construction identified by
the allocation diagnostic. It requires independent original geometry/pixels,
private ownership, cache retention limits, aggregate cold/warm GPU admission and
retirement checks, plus matched actual native CPU/GPU measurements. Cache hits
cannot relax the source support, per-glyph or logical work bounds. New cache
payload/reference limits are per renderer and do not qualify total font/native
memory. Unchanged latency targets and broader Mac acceptance remain open until
representative evidence passes them.

## Accessibility publication allocation work

The owner may consume its newly constructed semantic snapshot containers and
group grid records by indices to reduce intermediate copying. Last-completed
cache records remain detached, and borrowed internal entries retain their prior
container ownership. The complete native JSON envelope must remain byte-identical
to the validated stdlib envelope. Encoding refusal, pending/completed scope
transactions, callback replacement, stale queued actions and retirement require
their existing checks plus owned/borrowed parity. No publication may be skipped
solely because a portable grid generation is unchanged. Actual Cocoa and matched
ordinary CPU/GPU measurements remain separate from local lookup benchmarks and
from broader Mac/performance/native-memory acceptance.

## Scene construction allocation work

Fixed operation staging blocks and reusable tile-count scratch must preserve
the preceding encoded bytes, painter order, independent-reset detection,
operation/payload refusal boundaries and source ownership. Checks include
original encoding digests, an independent tile-bin oracle, block boundaries
through 65,536 operations, exact retained capacities, and successful/refused
builder retirement. Offscreen allocation benchmarks describe those local
fixtures only. Live native GPU parity, full-frame allocation/latency and broader
Mac acceptance remain separate gates.

Live windows, input and performance checks are deferred while the user is using
the computer. They resume only after the user explicitly says it is free.
Code, preparation, offscreen checks, reviews and local commits may continue.

## Modal geometry and authorized native checks

[Modal geometry and native checks](../docs/reports/macos-modal-geometry/README.md)
fix a reproduced background accessibility-frame loss after a label-only modal
closes. Scope transitions request one main-host render; unchanged scopes keep
demand painting. Final native/offscreen regressions and the bounded fixture
shutdown correction pass, with one Prism finding preserved and resolved through
internal review and tests.

During the explicitly authorized fifteen-hour window, both CPU and opt-in GPU
grid campaigns completed five matched runs per fixed/variable 10k/million-row
case. Both still fail unchanged warm latency targets. Three small-fixture idle
CPU measurements pass their 0.5% one-core target over 63-second samples. A
separate bounded read-only external AX query responds promptly in both windows;
physical Finder drag remains unqualified after CUA app-selection stalls. No OS
preference or general clipboard write occurred. Human IME/VoiceOver, native
service interoperation, physical display/GPU cases, integrated performance and
endurance, distribution rights and full Mac completion remain open.

The [GPU frame ownership contract](gpu-frame-lifecycle.md) gives ordinary and
measured paths shared checked failure cleanup. Failed preparation/submission
aborts surface ownership and prevents presentation or image capture. This code
checkpoint remains separate from physical backend faults, device recreation and
actual native graphics-memory/performance acceptance.

Texture creation and synchronous staging now have complete-pair publication,
explicit replacement retirement, device-change cleanup and checked layout/copy
refusal in the [GPU ownership contract](gpu-frame-lifecycle.md). Counted ownership
and wrapper-only regressions remain separate from physical GPU readback/faults
and the outstanding native graphics-memory/performance gates.

Private compute-system retirement now releases its owned device/queue, pending
recording and child resources while preserving the borrowed adapter. Checked
admission and complete recording cleanup extend the [GPU ownership contract](gpu-frame-lifecycle.md).
Offscreen failure fixtures and tiny headless shader execution qualify only their
observed scope; physical memory, sustained performance and window checks remain
open. Storage-value staging/map batch integrity is a separate follow-up audit.

Storage-value readback now checks shape/usage/source/staging identity and performs
complete batch preflight, synchronized callback delivery and rollback/quarantine.
Automatic byte uploads repair staging; exact-length reads support short-target
retry. Vars.Config's temporary validation layouts retire explicitly. These
[GPU ownership contracts](gpu-frame-lifecycle.md) retain serialized, initially
unmapped ownership; physical memory, sustained performance and live-window
qualification remain open.

Shader reload and bind-group replacement now have transactional owned-reference
publication in the [GPU ownership contract](gpu-frame-lifecycle.md). Refusal keeps
the current cache generation and retirement queue intact; successful reload
explicitly retires the prior module. Internal fault kernels and tiny headless
layout-refusal/retry checks remain separate from native/GPU memory, sustained
performance, physical backend faults and live Mac qualification.

## UI element expansion

[Controlled breadcrumb navigation](breadcrumb.md) adds a reusable owner-held path
component to the public SDK and gallery. Managed input, keyed retirement, four-DPI
software frames and inert Cocoa roles remain separate from physical interaction,
human VoiceOver, the active source-bound endurance campaign and full Mac gates.


The public control set includes a controlled `Disclosure` container for retained
collapsible inspector/settings content. Its heading input, visible focus/modal
scope, child admission and separate Cocoa disclosure triangle follow the
[controls contract](controls.md). Deterministic routing and native selectors
remain separate from physical Mac input, human VoiceOver and the existing
memory/performance/endurance gates.

The public control vocabulary adds controlled switches and toggle buttons plus
passive progress bars and oriented decorative separators. The native adapter,
semantic actions and gallery follow the [controls contract](controls.md).
Offscreen lifecycle/input and Mac native-object checks do not substitute for
human VoiceOver, physical keyboard/pointer interaction or the remaining native
memory and sustained-performance gates. Further UI vocabulary is part of the
framework completion work, alongside the existing ownership/qualification effort.

[Radio groups and segmented choices](../docs/reports/macos-choice-controls/README.md) extend the public vocabulary
with bounded keyed options, exclusive controlled selection, enabled keyboard
navigation and scoped native accessibility. The [controls contract](controls.md)
and public catalog specify their API. Internal/native-object/software-frame
checks remain separate from physical Mac input, VoiceOver and the outstanding
memory/performance/endurance gates.


[Toolbar, status-bar and form-group containers](../docs/reports/macos-ui-containers/README.md)
extend reusable application structure. Native scrolling overflow retains controls,
managed enabled focus navigation reveals tools, and private headings/status captions
preserve child identity and editing drafts. Offscreen input/raster and inert Cocoa
role checks advance implementation; physical Mac input/VoiceOver and the existing
native-memory/performance/endurance gates remain unqualified.

The UI vocabulary also includes a controlled two-child SplitPane with retained
independently scrolling pane hosts, owner-accepted fractional sizing, pointer and
keyboard divider input, visual Tab order and Mac split-group/divider semantics.
Offscreen input/geometry/rendering and inert Cocoa roles are scoped evidence;
physical Mac input, human VoiceOver and the remaining memory/performance gates
remain open. Nested SplitPane descriptors support larger adjustable layouts.

The control vocabulary adds a passive Icon with a bounded portable symbol catalog,
icon-bearing buttons/toggles and primary/secondary/quiet/toolbar button appearance.
Icon-only actions require accessible labels; decorative symbols are omitted from
Mac accessibility. Controlled identity and owner-window activation retirement
apply across presentation changes. The [controls contract](controls.md) records
the API; software raster/input and inert Cocoa image roles remain separate from
physical interaction, human VoiceOver and the existing memory/performance gates.

The standard control set adds a managed Image over the existing immutable tiled
source, centered contain/cover and stretch fitting, visible-content clipping,
alternate text and bounded worker PNG/JPEG decoding. Independent preview/source
replacement tests advance application use without conflating ordinary scene
painting with MeshCanvas's direct GPU image workload. Physical Mac/VoiceOver,
native memory and sustained performance remain open.

## Native service and accessibility allocation checks

The [service/allocation checkpoint](../docs/reports/macos-service-performance/README.md) passes the real Cocoa menu/tray fixture with 100 additional tray create/clear cycles and final service cleanup, plus the candidate stored 400-cell Cocoa accessibility fixture. Two small adapter changes reduce process-wide allocation in a matched 4,800-sample table campaign; frame timing is mixed and unchanged 4/6 ms targets remain failed. Native selectors do not qualify physical menu/tray interaction or human VoiceOver. This advances scoped implementation and acceptance without completing P6, P11/endurance, notification delivery or full Mac/release qualification.


The [identity outline checkpoint](../docs/reports/macos-outline-extrema/README.md)
adds hash-pinned static-outline extrema under the existing renderer ownership
limits. A 40-process, 4,800-sample interleaved opt-in GPU grid comparison has exact
paired geometry and 17.2–20.0% lower pooled p95 UI work. Both binaries still fail
the unchanged 4 ms p95 / 6 ms p99 targets. This reduces measured host UI work
without qualifying presentation, default CPU performance, physical GPU recovery,
complete P6, integrated P11/endurance or Mac release acceptance.


The [composed workbench checkpoint](../docs/reports/macos-workbench-endurance/README.md)
adds editable inspector controls and a bounded native endurance mode. A single
profiled ten-minute process with 1m records/100k outline nodes completes 2,384
owner activity steps, 596 confirmed Cocoa edits and all 1,787 worker completions.
Effective geometry alternates between 1280×900 and 1100×780. This is diagnostic
composition evidence, not all-component/1,000-visible-control primary geometry,
five ordinary latency runs, individual native request delivery, integrated 1,000
lifecycle cycles or overnight/physical/presentation/full P11/Mac acceptance.
The original P6 targets remain failed. See the [fixture contract](workbench-endurance.md).


The [bounded scene scanner checkpoint](../docs/reports/macos-scene-scratch/README.md)
removes eight allocations per local capture fixture while retaining at most
64 KiB of idle scanner backing per renderer. A 40-process/4,800-sample opt-in GPU
comparison passes exact geometry and ownership checks but has mixed frame timing
and no demonstrated process-wide allocation reduction. Both 4/6 ms targets still
fail. Complete P6, composed P11/lifecycle/idle soak and Mac release remain open.


The [settled static-idle checkpoint](../docs/reports/macos-static-idle/README.md)
records five independent 100-second native hello processes, each with a settled
60-second process CPU/RSS sample. Observed CPU is 0.23–0.37% of one core; within-run
RSS changes by at most 0.079 MiB. The small counter/canvas fixture is not the
primary 1,000-control workbench. RSS is roughly 227–237 MiB, above the provisional
150 MiB reference, and is not private/native allocation attribution. Wakeup/
present counts, minimized/occluded/overnight cases, full P11 and memory acceptance
remain open; the original P6 frame targets still fail.


The [workbench snapshot checkpoint](../docs/reports/macos-part-snapshots/README.md)
replaces complete component-array copies during worker startup with immutable
shared snapshots, and bounds edit path copies independently of document size.
Persistence materializes component records on its worker. v1 JSON, history,
concurrent reader isolation and actual 1m-record native save/load/calculation/
short Cocoa editing checks qualify this composition behavior. Worker traversal
is slower in the local fixture; no whole-frame/memory acceptance is inferred.
Other owner scans, complete P6/P11 and Mac release gates remain open.

The [MenuButton](menu-button.md) adds a bounded flat action menu to the public
control catalog. The [checkpoint](../docs/reports/macos-menu-button/README.md)
covers actual popup keyboard/pointer routing, peer-owner drain retirement and
20 Cocoa selection cycles with real item geometry and modal wake progress.
This component checkpoint does not close frame-performance, composed lifecycle,
human accessibility, overnight or distribution gates.

The [upload pilots](../docs/reports/macos-upload-pilots/README.md) tested packed
inputs and unchanged-byte upload copies with actual GPU parity and paired native
release runs. Neither showed a useful pilot gain; both prototypes were rejected
and no renderer patch was added. Warm table layout/paint and reconciliation are
the next profiling priority. The unchanged 4/6ms targets remain failed.

The [passive grid text checkpoint](../docs/reports/macos-local-text/README.md)
removes the updater shape that an admitted local sizing pass immediately replaces.
Interactive text keeps its ordinary updater path; canceled local requests restore
shaping before paint. Internal review and real-shaper/grid race checks qualify the
bounded behavior. A 40-process/4,800-sample native comparison shows roughly 3%
lower mean UI work and fewer allocations in all four fixed/variable 10k/1m cases.
The unchanged 4/6 ms targets still fail; complete P6, composed P11/endurance and
Mac release acceptance remain open.

The [composed lifecycle foundation](../docs/reports/macos-composed-lifecycle/README.md)
adds an all-kind workspace with1000 dense labels to the native window-cycle
runner. It fixes nested dock record retention, failed rebuild cleanup, retired
interaction delivery and a closed Window.View backing-slot reference. Original
regressions fail; final race/vet checks and five native composed preflight cycles
pass. Native frame containment is separate from pixel/physical/primary geometry.
The1000-cycle composed campaign, fullP6/P11/memory/overnight and Mac release gates
remain required.


The [native AX ownership checkpoint](../docs/reports/macos-ax-retention/README.md)
adds local pools to attachment/teardown after the full composed campaign failed
its RSS guard at cycle 600. Object-count diagnostics and an original-code failing
before-outer-pool regression isolate retained accessibility temporaries. The fixed
1000-cycle campaign passes interaction, density and final cleanup; sampled warm
Go heap is 45.3–46.6 MiB and RSS 350.4–369.1 MiB. Raw wrapper status-format failure
is retained alongside exact code/prepared endpoint equality. This closes that
specific composed lifecycle campaign, while full P6/P11, physical native/GPU
budgets, overnight soak, human accessibility and Mac distribution remain open.


The [stored-tree frame checkpoint](../docs/reports/macos-tree-performance/README.md)
adds actual native input/owner/completed-frame timing for sealed 10k/100k wide/deep
trees and removes duplicate retained row updater/style work. Forty matched native
processes produce 9600 validated warm samples with identical phase geometry and
71 clipped rows; mean per-run p95 improves 8.8–10.9% and process Go allocation per
observed interval falls about 21–23%. Final p95 8.45–8.56 ms still fails the unchanged
4 ms tree target. Original updater regressions fail, race checks pass and twelve
software pixel records match. Full tree interaction/cold-cache/maximum-visible,
P6/P11, physical native/GPU budgets, overnight and Mac release remain open.


The [tree route investigation](../docs/reports/macos-tree-gpu/README.md) adds
explicit consumer GPU/source-detail modes with stamped event identity and early
diagnostic-capacity refusal. Separate diagnostics observe full CPU scene uploads
and the optional GPU route. A 20-process/4800-sample GPU-mode comparison rejects
packed scanline compaction: object counts fall 12–13%, bytes rise slightly and deep
p95 is 3.1% worse. No renderer/prepared change was adopted. Tree 4 ms, complete P6/P11,
physical/native memory, human accessibility and Mac distribution remain open.


The [tree/retained-view publication checkpoint](../docs/reports/macos-tree-publication/README.md)
refuses stale tree geometry and queued AX actions during same-generation layout,
retains read-only pending accessibility rows and publishes ready data/frames with
stable compatible Cocoa identities. It withdraws descendant table/tree geometry
before modal/ancestor/workbench transfers, including before panel.View. Original
regressions fail; final verification scopes remain documented separately from
human accessibility, full P6/P11, native/GPU memory and Mac distribution.

The [current completed-outline campaign](../docs/reports/macos-tree-retention/README.md)
passes 1000 composed native cycles on committed bf3b1b5 with exact current tree AX
completion per cycle and final main-loop cleanup. Warm sampled heap/RSS remain bounded
by unchanged guards; two owned native allocation snapshots are preserved. This updates
lifecycle/retention evidence without qualifying full P6/P11, physical display/input,
human accessibility, overnight background soak or public distribution.

The [composed native action checkpoint](../docs/reports/macos-composed-actions/README.md)
qualifies 500 Cocoa-selector steps across 25 fresh windows. It verifies native
requests, acknowledged owner callbacks and final native publication for standard
controls, modal buttons and hierarchical tree selection/disclosure. Tree indexing
reserves bounded window-scoped worker/completion credit and runs off-owner. This
broadens interaction evidence; it does not certify physical input, human VoiceOver,
menus/services, sustained P6/P11 or a new 1000-cycle retention campaign.

The [inline color-picker checkpoint](../docs/reports/macos-color-picker/README.md)
adds an owner-held RGBA/hex component built from managed controls, immutable
checker pixels and ordinary focus/AX semantics. Slider explicit/theme minimum
sizes now survive the backend final style pass. The scoped routed-input/Cocoa
checks do not close P6/P11, human accessibility or distribution gates.

The current interacting lifecycle fixture includes the owner-held color picker
in a Colors tab. Each fresh workspace has 24 Cocoa action steps, including red
and full-hex proposals with all five native values and preview publication
checked before proceeding. Its per-action elapsed time includes inspection,
owner admission and native publication waits; it is diagnostic wall time, not
UI CPU/GPU or input latency. Prior 20-step/1000-cycle results retain their exact
source/workload scope.

The composed runner admits a cooperative deadline up to three hours; ordinary
runs retain twenty minutes. This accommodates the measured roughly eight seconds
per prior 20-step window plus four new steps, without raising the 1000-cycle cap,
512 MiB sampled Go-heap guard or 1 GiB sampled RSS guard. The owned campaign
wrapper grants a three-hour-five-minute watchdog only to its exact 1000-cycle,
GC-every-25, three-hour composed command with a new bounded diagnostic profile.
Other wrapper processes retain fifty minutes maximum. A complete current
campaign, its terminal cleanup, profiles and exact source/prepared bindings must
be validated before qualifying this lifecycle workload. Full P6/P11, unsampled
native/GPU allocation, sustained latency/pacing, physical input/displays, human
accessibility, overnight background and distribution gates remain separate.

The [interacting lifecycle checkpoint](../docs/reports/macos-interacting-retention/README.md)
passes a five-window, 120-step preflight but fails the requested 1000-window
campaign after 343 completed windows. The next window's tree-row press reports
an attachment closed before owner/native completion; its triggering close path
is not established. Partial memory samples and clean final teardown do not
qualify the failed campaign. Four completed actions retried queue admissions;
the exact owner callback counters still require one accepted effect. An original
single-admission validator assumption is preserved and explicitly corrected.
Live checks are deferred while the human uses the machine. Investigation and a
fresh complete campaign remain required.

The [accessibility retirement checkpoint](../docs/reports/macos-ax-retirement/README.md)
fixes a reproduced late-refresh registry resurrection and adds copied first-close
request diagnostics. Original-bind oracles fail; current targeted and full
offscreen race checks pass. This correction requires internal review between Prism checkpoints.
It does not establish the prior endurance failure's triggering close event or
qualify a new live campaign. Live testing remains deferred while the human uses
the machine.


The [small glyph coverage checkpoint](../docs/reports/macos-glyph-coverage/README.md)
adds bounded per-renderer exact scanner-span reuse to CPU paint and private GPU
capture. Software grid composition matches across 48 frames at four DPI settings;
refusal, packet immutability, eviction and ownership checks preserve original
behavior. This source optimization requires internal and Anthropic Prism review.
It does not establish a native frame-rate gain or close P6/P11. GPU scene rendering
remains opt-in. Live tests remain deferred while the human uses the machine;
current complete lifecycle, physical/human checks and distribution remain open.


The [menu check-action checkpoint](../docs/reports/macos-menu-checks/README.md)
adds controlled boolean menu actions and passive separators to the desktop
control catalog. Offscreen real-popup pointer/keyboard, rejected/stale proposal,
visual and inert Cocoa selector checks qualify the tested source behavior.
Internal review and the selected Anthropic Prism checkpoint remain required.
Human screen-reader, physical menu/input and full Mac acceptance remain open;
live tests remain paused while the human uses the machine.


The owned stage-57 GPU encoder removes validated covered uniform clears while
preserving original admission and packet reconstruction. A 48-frame actual-driver
software matrix matches every full-composition pixel and removes 193–477 encoded
operations per frame. Original limit boundaries, frozen Encode digests, serialized
tile order, detached results and three faulty variants are checked. Gated native
integer/text tests now exercise the compact route, with a dedicated partial-frame
regression; they compile but have not run while the human uses the machine.
This does not qualify native GPU pixels, timing or memory. GPU rendering remains
opt-in, original P6 frame targets still fail, and full P6/P11, fresh complete
retention, physical input/displays, human accessibility and distribution remain
open. [Scoped evidence](../docs/reports/macos-gpu-covered/README.md).

Controlled document-tab closing adds owner-admitted close buttons and focused-tab
Delete/Backspace requests, preserving rejected pages, reparented drafts and the
original unmanaged native path. Private tab/artwork identity and atomic modal
lifetime capabilities prevent stale requests from reaching replacement content.
The public contract is [tab closing](tab-close.md). Actual software input,
serialized/native-object semantics and lifetime fixtures are distinct from live
Mac attachment, physical input, human VoiceOver and full P6/P11/endurance targets.


The [current native evidence checkpoint](../docs/reports/macos-native-current/README.md)
records the completed `e7f8e30`/6,214 composed 1,000-window/24,000-step lifecycle
and the separate `87e7de1`/6,216 and `8c1709e`/6,216 ordinary renderer comparisons.
The 40-process span-buffer campaign and 80-process ordinary-geometry campaign
retain all paired geometries, per-run/pooled variability and allocation observations.
Timing changes are mixed; no run passes both unchanged 4/6 ms thresholds. Sampled
retention guards and tracked main-thread native cleanup pass on `e7f8e30`.
Comparison exits and launcher cleanup do not provide equivalent native-resource
inventories. The
survivor profile precedes global application closure, and the 60 retries have no
observed cause. These scoped results do not qualify later-source lifecycle,
professional/native/GPU memory budgets, complete P4/P6/P11, overnight soak,
physical input/displays, human accessibility, interoperability or Mac distribution.

The [uniform background checkpoint](../docs/reports/macos-uniform-crop/README.md)
reduces redundant large CPU fills while retaining original widget painting and
render items. Offscreen routed-grid geometry/pixel parity, original-dispatch
fallback tests and fixed replay benchmarks are separate from native acceptance.
Live checks remain paused while the human uses the machine. Current full
lifecycle, P4/P6/P9/P11, idle/overnight, physical input/displays, human accessibility,
professional memory and distribution gates remain open.

The [managed workbench checkpoint](../docs/reports/macos-managed-workbench/README.md)
admits the concrete built-in Workbench to the explicit-clock software adapter.
It fixes native dock-tab click synchronization with the owner model and revokes
queued gestures across disable/hide/modal lifetimes. Actual software routing,
retained drafts, split/tab drags, a stored 10,000-by-20 grid and independent
floating-window closure are checked. This enables broader P9 composition; it
does not complete the versioned professional workspace record, worker-built
views/cancellation, canvas/GPU observations or native performance/retention gates.
Live tests stay paused while the human uses the machine.


## Integrated software workspace checkpoint

The independent [workspace consumer](../examples/replay/README.md#professional-workspace)
now supplies frozen version-2 216-operation/27-observation records for 10,000 and
1,000,000 stored rows × twenty fields with 100,000 stored outline nodes. Public
managed workbench, property/editor, canvas placement, commands/history, injected
workers and an independent inspector compose through actual offscreen routing.
Exact software observations and independent content/transition oracles pass at
batches 1/7/64. Protocol v1 compatibility remains checked. See the
[checkpoint evidence](../docs/reports/macos-workspace-replay/README.md).

This closes the absence of an integrated versioned software record. The broader
P9 GPU/OS workload, current-source native/endurance/performance and physical or
human checks listed above remain open. A runnable native build is not a live
qualification result. No native application is launched while the human is online.

## Named Canvas2D object checkpoint

Opt-in object labels, visible completed clipped bounds, stable-ID selection and
Option-key traversal extend the retained Canvas2D through public APIs. Pending
publication, origin guards, capture loss and retained subtree transfer checks
are specified in [canvas accessibility](canvas-accessibility.md). Offscreen software
routing and inert Cocoa fixtures qualify their tested paths; live attachment,
physical input/display changes and human VoiceOver remain separate P3 gates.
This checkpoint does not complete P3, P4/P6/P9/P11 or Mac release acceptance.

## Span header construction checkpoint

The [bounded header patch](../docs/reports/macos-span-headers/README.md) removes
per-stream operation allocations and reuses at most 511 temporary row headers
within one scene builder. Five actual software grid-capture pairs preserve packet
accounting/encodings/pixels while reducing allocated objects and bytes. Exact
admission, no retained span references, packet independence and both retirement
paths remain required. No native frame-latency or retained/native/GPU memory gain
is qualified; the failed 4/6 ms targets and full P4/P6/P11/Mac gates remain open.

## Packed glyph coverage checkpoint

The [packed coverage path](../docs/reports/macos-packed-coverage/README.md) transfers
an owned exact span backing from bounded cached glyph replay into scene packets.
It preserves the original sequential admission, operation boundaries and bytes,
including fallback for clipped/unsupported/budget-limited input. Later ordinary
spans and both retirement paths retain explicit ownership. Exact software
comparison and allocation checks advance the measured scene-construction work;
they do not qualify the failed native latency budgets or close P4/P6/P11/Mac.

## Owned Mac build checkpoint

The [private build producer](../docs/reports/macos-owned-build/README.md) binds a
workbench executable and native linker map to one observed successful compiler
invocation, with sampled input/tool/environment stability and bounded diagnostics.
The existing collector's arbitrary-pair label stays unverified; only the new
receipt carries the observed pairing. This advances distribution evidence without
claiming complete embedded component provenance or approving rights. Public audit,
native latency/memory/endurance, physical input/displays, human accessibility and
full Mac completion remain open. No native application is launched for this
checkpoint while the human is online.

## One-pass glyph fingerprint checkpoint

The [fingerprint checkpoint](../docs/reports/macos-glyph-fingerprint/README.md)
removes a repeated path traversal from the measured scene-capture lookup path.
Validation and internal candidate hashing share one walk; minimum-origin replay,
exact path matching, clipping and existing cache/work bounds remain unchanged.
Software grid parity and shared-computer elapsed observations advance source
work only. The failed native 4/6 ms targets, full P4/P6/P11, native/GPU memory,
endurance, physical/human checks and distribution rights remain open.

## ListBox control checkpoint

The [ListBox contract](list-box.md) adds bounded always-visible single selection,
stable item identities, controlled rejection/read-only modes, scrolling and
keyboard/page navigation. The gallery and public SDK consumer exercise the API;
Mac semantics expose list/row roles, clipped visible rows and scoped owner
selection. See the [checkpoint](../docs/reports/macos-list-box/README.md) for
automated routing and inert Cocoa scope. Physical input, human VoiceOver,
native/GPU memory, sustained performance, endurance and distribution acceptance
remain required before full Mac completion.

The ListBox native metadata follow-up corrects read-only `AXSelectedRows`
writability using a list-specific Cocoa compatibility class. Same-ID class
transitions retire old objects while preserving ordinary cached bounds. See
the [follow-up evidence](../docs/reports/macos-listbox-metadata/README.md) for the
external native query/selection checks and their exact scope. Native pointer,
keyboard and focus-reveal qualification remain open.

The [composed control coverage follow-up](../docs/reports/macos-composed-controls/README.md)
adds ListBox and named managed Canvas2D to the native all-kind lifecycle fixture.
Its version-2 plan requires 33 acknowledged owner/native actions per fresh window,
including rejected/repeated list proposals and current canvas publication.
Fresh native evidence is required; earlier 24/26-step results do not qualify this
mix, physical input, sustained performance or the remaining full Mac gates.

The 33-step composed preflight projects beyond the old three-hour diagnostic
deadline. Its bounded campaign deadline is now four hours; the ordinary mode
remains capped at twenty minutes. The [endurance contract](workbench-endurance.md)
retains the original workload, sampling and memory limits. This scheduling
adjustment does not qualify a completed 1000-cycle campaign or close full Mac
performance, memory, physical/human or distribution acceptance.


## Native accessibility parent graph checkpoint

The [parent graph checkpoint](../docs/reports/macos-ax-parent-graph/README.md)
replaces repeated full ancestor walks with two bounded snapshot-local sets.
Independent original-rule oracles agree with both baseline and revised native
Foundation validation across 17,066 parent/scope cases, shuffled deep trees and
atomic refusals. The complete offscreen race suite and vet pass; an actual Cocoa
smoke exercises native semantics/actions, geometry, focus, virtual paging, modal
routing and the workbench bridge on the revised source.

Five alternating optimized benchmark pairs retain large deep-chain reductions
without the first candidate's wide-tree regression. Whole-validation observations
include JSON parsing; the wide-case ranges overlap, and the 451-node fixture
shows a modest local reduction. This work is outside exact TotalUIWork timing
and does not qualify the failed 4/6 ms ordinary targets, native/GPU memory,
current-source endurance, successful overnight, physical input, human IME/
VoiceOver, distribution or full Mac completion. The earlier failed soak remains
failed and retained.


## Current-source composed lifecycle result

The [current-source lifecycle check](../docs/reports/macos-parent-graph-endurance/README.md)
completes 1,000 fresh windows and all 33,000 version-2 native/owner actions on
committed `6330b2d`, including the current ListBox and named Canvas2D paths.
Exact event order, completion predicates, density, sampled memory, source/build
bindings, terminal cleanup and fresh owned-process absence are checked. Fifty-four
actions retry once; the additional queue admissions have no observed cause and
are not accepted callback counts. One hidden inactive untitled Cocoa window and
false original-menu restoration remain explicit pre-exit limitations.

This closes that source's current 33-step execution gate, with no qualification
for native/GPU/private memory or unbounded retention, ordinary frame/input latency,
physical display/pointer or human accessibility. The failed overnight stays failed;
full idle, P4/P6/P9/P11, interoperability, distribution and Mac acceptance remain
open. Later source changes need their own appropriate runtime evidence.

## Pending native accessibility publication checkpoint

The [checkpoint](../docs/reports/macos-pending-publication/README.md) implements
private pending-only publication under the [accessibility contract](accessibility.md#pending-only-native-publication).
Exact detached wire comparison, native epoch/scope preflight and full fallback
preserve completed frames and identity. Final cache/epoch corrections pass the
affected-package offscreen race suite, vet, native semantics/grid checks and one
fresh ordinary pair per rendering mode. The earlier five-pair cohort remains
separate. Neither cohort qualifies the 4/6 ms targets or broad performance
regression freedom. Physical input, human accessibility, native/GPU memory,
P4/P6/P9/P11, overnight, interoperability and distribution remain open.

## Current external accessibility checkpoint

The [read-only current-source check](../docs/reports/macos-current-external-ax/README.md)
verifies external editor Unicode/ranges and native palette/canvas properties.
Two explicit fixture overlays disable editor geometry persistence and shorten
the placement deadline; they do not change framework implementation. The raw
placement collector failure remains preserved, with a separately reviewed
file-identity/native-ID correction over the same runtime data. This advances
scoped external AX evidence without qualifying physical placement/key events,
human accessibility, the earlier CUA timeout cause, performance or full Mac
completion.

## Initial native window sizing checkpoint

[Current source and native evidence](../docs/reports/macos-initial-window-size/README.md)
qualifies two small Dp initial client requests and later programmatic shrinking
on the current display, with separate read-only AX outer frames. Owned stage0074
removes persistent initial body minima and screen-fraction override, converts at
creation with one chosen screen snapshot, preserves saved geometry precedence,
and bounds final native dimensions. Automated tests cover scale/zoom, synthetic
secondary screen, tiny fractional and extreme sizes plus restore options.

This does not close physical resize/drag/cross-display, live saved restoration,
loaded/native performance or full Mac qualification. Historical 4/6ms failures
retain their original workload and failure status. New performance comparisons
must report actual initial/client/render geometry and screen scale.

## Current native table viewport measurements

[Current measurements](../docs/reports/macos-current-grid-viewport/README.md)
record native endpoints and display scale for 40 ordinary processes, 4,800 frames
and 2,400 complete paired row-count geometry comparisons. The corrected initial
client is 1519×1266 pixels at DPI162/DPR1; logical grid rounding yields 767×624.
All runs meet the 400-cell/50-row/eight-column workload minimum, but all miss
4 ms p95 and 32 miss 6 ms p99. P6/P11 remain open. Historical failures retain
their workload/status; GPU execution, native/GPU memory, physical interaction,
continuous extent/host stability and complete Mac qualification are unproven.

## Adaptive scene tile bins — source and native checkpoint

[Stage 0075 evidence](../docs/reports/macos-adaptive-tile-binning/README.md)
binds clean source `29ec8c9`, unchanged admission/serialized bytes and 24 actual
retained-grid software frame pairs. Six explicit native GPU readback roots pass.
Forty matched ordinary baseline `77da1dd` / candidate `29ec8c9` processes cover CPU/GPU
requested modes and fixed/variable million-row grids at identical geometry.
Candidate p95 fails 4 ms in 20/20 processes and p99 fails 6 ms in
18/20. P6/P11 and complete Mac qualification remain open.
Ordinary tickets do not prove GPU execution or native/GPU memory, and programmed
event routing does not qualify physical interaction.
