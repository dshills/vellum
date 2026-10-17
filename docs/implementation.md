# Framework implementation phases

This implementation proceeds alongside the unresolved foundation checks. Those
checks still prevent a production-ready claim. They do not prevent building and
assessing the workbench API. Each phase receives Prism review, relevant tests,
and a separate commit.

1. Application lifecycle and keyed UI API — implemented; native driver and executable example.
2. Commands and cancellable tasks — implemented; shared registry, menus, toolbar, palette, shortcuts and bounded owner-thread completions.
3. Docking, document tabs, inspector and layout persistence — implemented; menu-driven move/split, resizable native panes, transactional versioned layouts.
4. Virtualized tables and trees — implemented; 1,000,000 real records, 100,101 tree nodes, 20/24 native row pools, cancellable sorting and stable domain selection.
5. Shared-device GPU canvas with navigation and selection — implemented; 96 retained GPU objects, pointer-anchored zoom, pan, selection and inspector integration.

The native driver exposes multiple desktop windows in one process main loop. Root UI ownership and
logic tests are portable; native runtime support is not claimed across platforms.

Phase 5 has passing model/native regression tests and real GPU pixel verification.
The live pointer automation check remains unverified because the computer-use
window coordinate mapping targeted a different region; see the phase 5 review
evidence. This limits runtime certification, rather than the implemented API scope.

## Framework expansion

1. Keyed reconciliation — implemented; native identity and retained state survive
   insert/remove/reorder/reparent and docking moves. See the
   [contract](../specs/keyed-reconciliation.md). Portable/native checks pass;
   live focus/accessibility workflows remain acceptance checks.
2. Layout and theming — implemented; optional value styles, application-local
   themes, native layout/appearance mapping and reset semantics. See the
   [contract](../specs/layout-theme.md).
3. Standard controls and interaction — implemented; controlled values, queued events,
   keyed tabs/scroll/modal children, focus requests and semantic metadata. See
   the [contract](../specs/controls.md); platform screen-reader publication remains
   an acceptance gate.
4. Multiple windows and platform services — implemented; scoped callbacks,
   independent close lifetimes and macOS clipboard/file panels/application menus.
   See the [contract](../specs/windows-services.md).
5. Independently consumable SDK and benchmarks — implemented; driver-owned pinned
   bootstrap, standalone consumer/generator and repeatable host measurements. See
   the [contract](../specs/sdk.md). Publication and production performance remain
   release gates.

## Framework completion work

1. Workbench interaction — implemented; pointer docking/reordering, independent
   floating windows, redocking and transactional session persistence. Domain data
   survives native adapter remounts. See [contract](../specs/workbench-interaction.md)
   and [evidence](reports/workbench-interaction/README.md).
2. Native accessibility and keyboard workflows — implemented for macOS managed
   controls, active dock panels, modal scopes and bounded virtual views. Native
   selector probes and keyboard regression tests are separate from external
   screen-reader/IME acceptance. See [contract](../specs/accessibility.md).
3. Rendering hardening — implemented; demand painting, synchronized composition,
   live resource rebuilds and device-loss error/shutdown. Local race checks cover
   resize/minimize/restore and teardown. Automatic physical GPU recreation,
   cross-display DPI and sustained workload budgets remain open. See
   [contract](../specs/rendering.md) and [evidence](reports/rendering/README.md).
4. Windows/Linux services and qualification tooling — implemented; native-main
   clipboard, OS chooser transports, menu widgets, target smoke/CI harness and
   explicit strict qualification gates. This macOS host cannot establish target
   runtime, TSF/IBus/Fcitx, UIA/AT-SPI or native Wayland support. See
   [contract](../specs/platform-support.md) and [evidence](reports/platform-support/README.md).
5. Versioned private SDK — implemented; alpha version API, independent consumer
   scenarios, integrity-checked relocation, tutorial and private application
   packaging. The owner selected `github.com/dshills/vellum` and MIT for Vellum's
   own code. Distribution audit and actual target acceptance remain public-release
   gates. See [contract](../specs/release.md)
   [tutorial](tutorials/building-an-app.md) and [evidence](reports/sdk-release/README.md).

Per-phase Prism reviews and verification records accompany these implementations.
The private SDK evidence records a tested precommit snapshot; its clean committed
archive is generated after committing, so its future commit identity cannot be
embedded in the earlier evidence. These
implementations extend the experimental framework; they do not certify a public
or production-ready cross-platform release.

## macOS completion milestone

The current goal focuses framework completion and acceptance on the local Mac;
Windows/Linux runtime work is deferred. See the [macOS completion checklist](../specs/macos-completion.md)
for the complete remaining scope.

Document history is implemented with bounded sparse changes, undo/redo, save
revision validation, dirty status and focused native keyboard precedence. The
workbench records component edits and updates source revisions on replay. See
[document history](../specs/history.md) and [document persistence](../specs/document.md).
The portable session now saves/loads on a worker and guards unsaved native closes;
the workbench persists component data, geometry and ordered file references.
Actual file-panel selection remains an acceptance gate. Native command phases also run in focus-free windows;
startup shutdown and late native creation use stable stop/admission handling.
See [local evidence](reports/macos-history/README.md).

External file drops are implemented through `ui.Node.FileDrop`, bounded absolute
path payloads, synchronous owner admission, clipped target geometry, active
tab/workbench eligibility and modal host fallback. The workbench references
dropped files with undo/redo and a bounded status display. Portable/native race
checks and a synthetic cursor/Cocoa receive fixture pass; the physical Finder
drag attempt did not produce delivery and remains unqualified. See
[contract](../specs/file-drop.md) and [evidence](reports/macos-file-drop/README.md).

Native dispatch now acknowledges failure, preserves panic identity, closes
admission and avoids a terminal crash-dialog wait. Desktop wakes serialize with
GLFW termination. Reproducible patch22, twenty repeated race runs and real native
loop/windowed checks support the bounded contract. See
[dispatch](../specs/native-dispatch.md) and
[evidence](reports/macos-native-dispatch/README.md); broad panic resource recovery
and remaining macOS qualification stay open.

Retained custom drawing now adds immutable indexed 2D meshes, world strokes,
million-sample envelope reduction and paint-ordered rectangle spatial culling.
The independent drawing application exercises a flat shared-device native
renderer, explicit resource accounting, retention and snapshot color handling.
See [contract](../specs/drawing.md) and [bounded evidence](reports/macos-drawing/README.md);
the broader P4 workloads and Mac qualification gates remain open.

Retained image drawing adds immutable 128×128 RGBA tiles and source patches,
image-only or image-below-mesh canvases, nearest and alpha-safe linear filtering.
The native consumer verifies a real 4096² source, one-tile 64 KiB updates,
independent upload baselines, retention, removal, rebuild and resize. See
[image contract](../specs/images.md) and [evidence](reports/macos-images/README.md).
The source patch benchmark measures portable allocation/work only; sustained
GPU/frame/input and broader Mac/P4 acceptance remain open.

Advanced GPU views expose the window's shared device/queue, ordered preparation
and custom draw callbacks, optional four-sample depth and scoped tracked resource
helpers. The independent 3D viewport provides orbit/dolly controls and native
depth, retention, fault repair, resize and closure evidence. The shutdown check
exposed and fixed concurrent removal during backend quit through reproducible
patch 23. See [contract](../specs/advanced-gpu.md) and
[evidence](reports/macos-advanced-gpu/README.md). The imaging consumer adds actual
compute and float-color render preparation with bounded asynchronous captures.
See [evidence](reports/macos-imaging/README.md). Broader P4/Mac qualification remains open.


Public multiline editing now has an exact UTF-8 portable model, directional
scalar selections, transient composition, monotonic stale-input guards and one
sparse undo journal shared with document persistence. The managed Mac control
reuses the prepared text renderer and originating input FIFO; asynchronous
settlement precedes Save/Open/Close/Undo/Redo. The independent editor consumer
uses two separate document models. See [editor contract](../specs/editor.md).
Two native race-binary runs and a separate read-only external AX process pass
against frozen source inputs. See [editor acceptance](reports/macos-editor/README.md).
Human IME candidate panels and screen-reader workflows remain acceptance gates.

## Explicit execution foundation

The portable execution package provides bounded manually selected workers and
manual/real timers with cancellation-visible running work. App completion
reservations merge actual result arrivals with ordinary callbacks; DrainN uses
an initial arrival cutoff. Tasks and document persistence accept Executor
injection and retain separate worker/owner accounting through termination.
Race tests cover saturation, canceled accepted work, panic/Goexit, closure,
cleanup failure and refusal rollback. See [contract](../specs/execution.md).

The managed [replay checkpoint](../specs/replay.md) is now internally qualified
with supplied, digest-checked fonts and a declared theme. The offscreen adapter
routes real pointer/key events through retained widgets, focus and modal hit
testing, and captures measured semantics/layout and completed software frames.
One fixed 51-operation record with nine captures produces identical model state,
focus, UTF-16 selection, geometry and image identities at batches 1, 7 and 64;
event/frame counts are reported separately. Workers, clock advancement and
originating widget timers require explicit steps. Future timers remain visible
without forcing time forward. This checkpoint permits one active managed
application per process, with multiple windows sharing its supplied clock.
See the [checkpoint report](reports/macos-replay/README.md).

The new opt-in [Mac semantic endpoint](../specs/automation.md) supplies a private
Unix socket, typed queries/actions, revision checks and bounded owner dispatch.
The workbench shares its selection, command and history validation with this
transport and publishes explicit redacted domain DTOs. Headless transport and
domain tests are distinct from a live professional automation campaign.

The full P9 professional workloads, large-data replay, external automation
workflows, native services/IME/accessibility, GPU presentation and sustained
performance acceptance remain open. Software replay does not certify those gates.

## Indexed grid foundation

`virtual.Grid` adds a separate indexed-source model for fixed or variable-height two-axis
virtualization, frozen headers/leading columns, bounded detached snapshots and
stale press identities. Active keys and top-row anchors survive indexed source
reordering, subject to viewport extent clamping, without a full-row owner scan. The procedural 10k and million-row ×20
fixtures expose 400 visible cells with identical bounded source-call counts.
See [contract](../specs/grid.md) and [portable evidence](reports/macos-grid-model/README.md).
The [managed adapter](../specs/grid-native.md) exposes `ui.DataGrid` with bounded
native cell/header pools, a single keyboard focus target, routed pointer/wheel/
arrow/page navigation, completed-frame geometry inspection and visible-cell
accessibility. Optional `GridRowMetrics` supports bounded indexed geometry and
pixel paging; `RowHeights` offers cancellable prefix preparation outside the UI
owner. The independent Grid example opts in with `-variable-rows`, retaining
stable-key heights through reverse/filter and source-revision toggles. See the
[variable-row report](reports/macos-grid-variable/README.md).
Optional [row selection](../specs/grid-selection.md) captures immutable ranges or
All with bounded stable-key exceptions, and permits cancellable enumeration on
workers. The managed grid routes Shift/Command selection and exposes visible
membership independently of its cursor. `examples/grid -row-selection` adds
All/Clear controls and combines with `-variable-rows`.
Optional [cell drafts](../specs/grid-editing.md) add per-cell versions, bounded
recoverable drafts and atomic immutable-source completion after application
validation. The Mac adapter uses one native inline field and settles its input
before commit. `examples/grid -editable-cells` demonstrates sparse copy-on-write
edits, external-cell conflicts and cancellation alongside the existing modes.
Reusable [worker-built views](../specs/grid-projection.md) add immutable
sort/filter/variable-height indexes with explicit array/string budgets and latest
owner-result admission. The independent `examples/loadedgrid` consumer uses real
stored records, separate domain/view revisions and cancellable tasks.
Its Apply action uses ordinary text-field settlement to read the final controlled
filter value after all preceding admitted input. Repeated unchanged settlement
does not emit duplicate changes before redraw.
Native sustained performance and broader Mac acceptance remain subsequent P6 work.

## Indexed tree and native timing foundation

`virtual.TreeData`, `TreeIndex` and `Outline` add sealed stored-tree data and
worker-built expansion/filter indexes. The owner navigates a bounded viewport
and admits request-bound results while keeping the last accepted view usable.
`ui.DataTree` adds managed native rows, clipped completed-frame geometry,
keyboard/pointer navigation and visible-row accessibility. The independent
`examples/loadedtree` consumer holds separate 10k and 100k stored domains.
Portable and native race checks pass, including actual offscreen input at
owner batches 1/7/64. A Cocoa race binary passes the bounded owner-command and
geometry workflow. See the [tree contract](../specs/tree.md),
[native contract](../specs/tree-native.md) and
[large-tree evidence](reports/macos-large-tree/README.md). Human accessibility,
500-visible-row latency and sustained retention remain acceptance work.

Opt-in native frame tracing records each initialized grid scroll through its
owner transition, reconciliation and completed native composition. The same
ticket includes intermediate post-input frame work; queue and surface waits
are reported separately. The elapsed work total includes GC and excludes
actual backend acquisition/present calls. It does not measure physical display
latency. The stored-grid release runner enforces at least 400 clipped cells,
paired 10k/million-row geometry and five runs per case for budget qualification.
The first single-process exploratory probe fails the unchanged 4ms p95/ 6 ms p99
targets at 20.30ms/22.77ms. See the [timing contract](../specs/grid-performance.md)
and [timing evidence](reports/macos-grid-performance/README.md). P6 remains open.
The [single host-update optimization](reports/macos-grid-update/README.md)
removes duplicate pooled updater/style work. Its interleaved fixed-10k release
comparison observes p95 19.55→17.07ms and p99 20.41→18.00ms with matching
geometry and software-frame pixels. This remains an exploratory comparison.
The [unchanged-refresh checkpoint](reports/macos-grid-refresh/README.md) now
reuses validated projections, fixes embedded native-parent identity comparisons
and checks borrowed source health without rematerializing unchanged cells. Actual
private editing and style/help/disabled/scrollbar transitions remain current.
Original-overlay regressions fail; final native adapter/portable virtual race
suites and native default/offscreen vet pass, with eight matching software frames.
Its three-process/build fixed-10k release comparison observes p95
17.233→14.624ms and p99 18.954→15.566ms; the 4/ 6 ms targets still fail.
Raster/upload cost is the next investigation, with full P6 and broader Mac
acceptance still open.

The [empty-sprite checkpoint](reports/macos-empty-sprites/README.md) now omits
the scene-sized transparent sprite source when no active sprite can draw, before
allocation or clear. Active callbacks retain tier order and every-composition
blink behavior; event-only sprites keep Active-only deadlines. Original overlays
fail and final core/managed/full adapter race checks pass, including software
cursor, overlays, drag/drop, resize/reactivation and matched grid pixels. Three
interleaved optimized fixed-10k processes per build observe pooled p95
14.178→13.090 ms and p99 14.891→13.294 ms. The unchanged 4/6 ms targets still fail.
The full P6 campaign and broader Mac acceptance remain open; further raster/upload
attribution and any bounded damage policy require separate work and evidence.

The [source-cost checkpoint](reports/macos-source-costs/README.md) now exposes
explicit scalar source diagnostics with shared ticket bounds and deep snapshots.
Default ComposeMeasured and normal Draw bodies retain their calls; opt-in costs
remain nested, with unavailable GPU success/failure stated explicitly. Final
race/vet/packaging and meaningful mutation regressions pass. Three optimized
fixed-10k detailed processes identify CPU scene rasterization as about 79% of
source Draw time, compared with about 21% in Copy/image API work. The next
performance experiment should reduce unnecessary raster work while preserving
paint/composition correctness. Detailed timings cannot qualify warm targets;
full P6 and broader Mac acceptance remain open.

The pinned native raster scanner now offers explicit row tracking for renderer-owned
accumulation. Scene and private glyph-mask renderers select it; the general scanner
keeps its full-row behavior. Pending cells, clip changes, clearing and resize
initialization preserve the existing span stream. See [rendering](../specs/rendering.md).

Managed grids publish one copied completed accessibility subtree per visible
binding scope. Pending scroll/layout revokes actions while retaining read-only
last-completed records; ready semantics and copied frames apply atomically in
Cocoa. Stable overlapping objects survive that transition, while replacement,
scope removal and closure retire them. See [the publication contract](../specs/accessibility.md#managed-grid-publication).
This addresses accessibility continuity and allocation churn; it does not claim
that the 4/6 ms UI-work targets have been met.

Stored-table qualification now refuses a warmup that fails to return to its
starting viewport or has extra callbacks. Optional bounded input diagnostics
capture admission/delivery scale, signed wheel remainder and owner transitions,
including untraced input, scrollbar and layout changes. Exclusive trace files
preserve failed runs after native retirement. Diagnostics remain separate from
ordinary latency qualification; the earlier offset mismatch is still unproven.
See [the measurement contract](../specs/grid-performance.md).

## Covered uniform raster fills

Stage 38 omits only adjacent, fully covered uniform Src clears during raster
replay. Dirty-buffer and actual widget pixels/call behavior match the original
dispatch; 12 preparation/race/vet checks, three meaningful faulty-overlay
proofs, eight retained-grid image comparisons and two 400-cell Cocoa checks pass.
The isolated 500-cell replay median improves by 43.3% with zero allocations.
Fresh matched ordinary baseline/candidate processes retain 4,800 samples;
pooled candidate p95 UI work is 6.0–6.7% lower across the four cases on this host.
40 of 40 processes still fail 4 ms p95 / 6 ms p99. This does not complete Mac/P6
acceptance or establish GPU upload/native memory gains. See
[the report](reports/macos-covered-fills/README.md) for all distributions, source
bindings and reviews.

## Native accessibility publication work

Native render callbacks omit managed Grid legacy rectangles already ignored by
Cocoa; full inspection, atomic scoped semantic/frame publication and editor text
geometry remain intact. Native commit reuses the validated node map, and async
payload producers/consumers have explicit Cocoa pool lifetimes. Eleven
preparation/race/vet/build checks and three live Cocoa processes pass. Pooled p95 changes are mixed across these four matched cases; the distributions below are the observation, without a general speedup claim.
All 40 ordinary processes still fail both unchanged warm UI-work targets. The
bounded payload fixture does not qualify native/GPU memory or complete Mac/P6.
See [the report](reports/macos-ax-publication/README.md) for full distributions,
source/artifact bindings, reviews and coverage limits.

## Ordered scroll-surface experiment

The pinned raster backend adds exact actual-paint guards and bounded toroidal
CPU pane caches, retaining original painter order and full GPU upload. Current
stored-table experiments regress p95, so managed rendering keeps ordinary replay
by default. Driver.ExperimentalGridScrollReuse and its manual counterpart opt in
for main scenes; modal/popup scenes remain ordinary. Curve subdivision stabilizes
integer translation and intentionally changes some prior antialias edge pixels.
Pixel/race/native checks and all 9,600 experimental observations are retained in
[the report](reports/macos-scroll-surfaces/README.md). The 4/6 ms targets and full
Mac acceptance remain open; no default performance gain is claimed.

## Keyed Grid and fixed-allocation layout

Native cells now retain overlapping row/column identity and shaped text in their
bounded current projection. A verified fixed-allocation boundary sizes changed
text and repositions descendants, then paints the complete scene in original
order. Full-layout fences, callback requeue and private preview/tree completion
remain coherent and quiescent. Final preparation/race/vet/build and three Cocoa
checks pass. The ordinary matched comparison and open Mac/P6 targets are in
[the checkpoint report](reports/macos-keyed-grid-layout/README.md). The default
renderer still uses full raster replay and image upload; ordered GPU scene-source
rendering remains the next architecture step.

## Inline color component

The public ui.ColorPicker composes managed RGBA sliders, a hex field and an
immutable checker preview. Callers retain accepted straight-alpha color and can
normalize/refuse edits, show validation and target generated field IDs for focus.
The gallery demonstrates reset/disable/focus. The checkpoint also fixes Slider
minimum sizing overridden by backend final styling; removing overrides restores
native defaults. See [the report](reports/macos-color-picker/README.md). Broader
Mac performance, human accessibility and release gates remain open.

## Color-inclusive lifecycle validation

The composed fixture exercises 24 Cocoa selector steps per window, including
accepted RGBA/hex publication. Five fresh windows pass 120 steps. The requested
1000-window campaign fails after 343 completed windows when the next window's
tree-row request observes its accessibility bridge retired. Cleanup completes;
the triggering retirement cause and a passing complete campaign remain open.
Four completed actions retry queue admission without duplicate accepted callback
counts. Live testing is paused while the human uses the machine. See
[the preserved failure report](reports/macos-interacting-retention/README.md).

Late accessibility refreshes now refuse to republish retired bridge registry
entries. Requests canceled by retirement expose a copied first-close path through
typed errors and the optional native fixture record. Targeted/full offscreen race
tests and native build checks pass; the 343-window failure's triggering event and
fresh endurance acceptance remain open. See
[the retirement checkpoint](reports/macos-ax-retirement/README.md).

## Typed palette placement

Public `placement.Controller[T]`, managed Button sources and `ui.Canvas2D`
provide same-window capture, threshold previews, target acceptance/snapping and
exactly-once application command delivery. The independent `examples/placement`
module demonstrates three types, keyboard placement, selection, pan/zoom, grid
snapping and bounded Undo/Redo. See the [contract](../specs/placement.md) and
[checkpoint](reports/macos-palette-placement/README.md) for automated routing
evidence and the outstanding physical Mac checks. Full Mac/P9/P11 gates remain open.

## Bounded list selection

Public `ui.ListBox` adds an always-visible single-selection control with stable
IDs, scrolling, controlled selection/rejection, clamped arrows, Home/End,
measured page navigation and initial-letter cycling. The gallery's Lists page
demonstrates it; the independent SDK consumer uses only exported APIs. Mac
semantics expose list/row roles, selected and visible rows, clipped geometry and
origin-scoped actions. See the [contract](../specs/list-box.md). Automated routing
and inert Cocoa evidence do not qualify physical input or human VoiceOver.
The [checkpoint report](reports/macos-list-box/README.md) binds the exact scoped
checks and outstanding acceptance gates.

## Placement desktop acceptance attempt

The [native placement probe](reports/macos-placement-desktop/README.md) adds a
dedicated opt-in Mac executable with bounded read-only observations and diagnostic
removal controls. Writer/observer race controls, the independent placement
offscreen suite and native vet/build pass. The actual desktop session reached
its ten-minute deadline while external app binding remained pending; the request
later returned a timeout. No pointer/key actions were requested. Its group was reaped and absent. Native placement and
external accessibility workflow acceptance remain open; the timeout cause is
unassigned. This does not change the feature API or close broader Mac gates.

## Pending native accessibility publication

The [pending-publication checkpoint](reports/macos-pending-publication/README.md)
adds private reuse of existing Cocoa node dictionaries after an exact prepared
wire comparison. Native scope/epoch preflight retains full JSON fallback; pending
and full commits share ancillary geometry cleanup. Bounded detached baselines and
permanent exhaustion revocation preserve the proof across lifecycle boundaries.
The final offscreen/native checks pass. Separate ordinary grid cohorts establish
neither consistent speedup nor the required 4/6 ms targets; physical/human and
full Mac qualification remain open. Public application APIs are unchanged.

## Current external macOS accessibility

The [current read-only AX check](reports/macos-current-external-ax/README.md)
reaches the native editor text/ranges and the placement palette/canvas through
separate already-trusted clients. Original source, diagnostic overlays, raw
process output and corrected collector interpretation are retained. No input
workflow or CUA retry is performed; physical/human and full Mac gates remain open.

## Initial native window sizing checkpoint

Explicit native Window dimensions now request initial Dp client size before Show,
without becoming body minima or invoking screen-fraction sizing. Saved geometry
and OS limits retain precedence. Two measured native size/shrink checks and
read-only external outer-frame observations pass on the current display;
fractional-scale/boundary/restore tests are automated. See the
[report](reports/macos-initial-window-size/README.md). Existing performance and
physical/cross-display/complete macOS gates remain open.

## Current native table viewport checkpoint

The stored-grid performance fixture now validates native client/render/screen
geometry outside measured frames and joins worker teardown. Its consumer refuses
partial or drifting endpoints, clipping outside render and scale mismatch.
[Forty current native runs](reports/macos-current-grid-viewport/README.md)
completed with matched row-count geometry; all 40 missed the unchanged 4 ms p95
target and 32 missed 6 ms p99. The actual logical grid was 767×624 after rounding.
This fixture checkpoint does not establish an algorithmic speedup, GPU execution,
physical input or full Mac qualification.

## Adaptive scene tile bin counting

Owned stage 0075 replaces dense per-reference count loops with four rectangle
difference corners and an in-place prefix, while sparse batches count directly.
Original admission limits, serialized bytes, painter order and ownership remain
unchanged. [The checkpoint](reports/macos-adaptive-tile-binning/README.md)
includes frozen byte/admission oracles, 24 retained-grid software frame pairs,
six explicit native GPU readback roots and 40 matched ordinary grid processes.
The complete table/tree, native/GPU memory and full Mac gates remain open.

## Current private SDK consumption

[Current SDK evidence](reports/macos-current-sdk/README.md) binds clean source
238b3b6 to a private archive with 7,411 hashed files, including owned stages
0074/0075. A relocated generated consumer and public API contracts pass their
race/vet/build scope. Nine demo applications and four replay/workspace commands
compile with replacements confined to the extracted SDK. A subsequent bounded
native run of that SDK hello executable verifies its rectangle center-pixel
readback, three further snapshots and owned shutdown. Full demo runtime, physical
input, performance/memory and public distribution remain open.


## Packaged public SDK software replay

The [SDK replay follow-up](reports/macos-current-sdk/README.md)
executes both already compiled offscreen replay commands at batches 1, 7 and 64,
including the professional workspace with 10k/1M stored rows and a 100k-node
outline. All nine cases complete 1,449 operations and 189 captures; 162 workspace
captures match unchanged frozen model/layout/semantics/text/image records.
Independent data, history, cancellation, placement and window-lifetime checks
pass. Runtime source remains SDK `238b3b6`, distinct from current documentation
endpoint `d4b07d0`. This advances packaged software routing evidence; native and
physical workflows, timing/memory budgets, full P9/P11 and Mac release remain open.


## Current native grid profile attribution

The [separate CPU/allocation diagnostics](reports/macos-current-grid-profiles/README.md) complete two native
million-row GPU-requested cases, 1,024 samples and 512 full geometry pairs on
unchanged SDK `238b3b6` source. Offline profiles suggest accessibility publication
and repeated dirty-cell text sizing as investigation targets. They do not prove
a causal bottleneck or savings; the Reconcile region includes accessibility
work beyond `reconcileScoped`. Existing ordinary 4/6 ms failures and full Mac,
P6/P11, physical/human, native/GPU memory and release gates remain open.


## Detached accessibility storage and native evidence

The [copy-storage checkpoint](reports/macos-ax-copy-storage/README.md) retains source `3a14e53`, exact legacy/
ownership/allocation regressions, one Anthropic Prism review and actual Cocoa
grid identity/action/closure acceptance. Fresh native runs complete 4,800
ordinary samples and 2,400 full geometry pairs. All 20 candidate p95 and 19 p99
processes miss 4/6 ms; fixed-row CPU p95 is higher in all five pairs. One paired
allocation diagnostic estimates fewer Go objects, but establishes no native
latency gain, broad no-regression or native/retained memory acceptance. P6/P11
and full Mac, physical/human and distribution/release gates remain open.

Owned stage 0076 adds exact-location reuse for small glyph scene coverage after
the existing bitmap return and original work charges. Private epoch/generation
identities preserve the 64-bit outline-entry size; a separate 8192-entry metadata
index validates current coverage identity, clip, target and scanner state.
Saturated misses use a bounded pause before retry. Warm grid packet/pixel parity,
exact logical charges, FIFO/collision/exhaustion and cooldown routing have scoped
regressions. Warm software gains and cold/bitmap costs remain separate from the
open native frame budgets and full macOS qualification. The
[owned contract](../drivers/cogent/native/patches/glyphplacement/README.md) records
the bounds and fallback requirements.


The [exact-location glyph coverage checkpoint](reports/macos-glyph-placement/README.md)
records source `2be8e79`, its single Anthropic Prism review, final automated
checks and separate five-process current-source native sanity. Cocoa identity,
action/closure and all 480 measured callbacks pass; all four ordinary cases
still miss 4/6 ms targets. DPI96 offscreen capture medians improve, while bitmap
and cold/miss costs remain explicit. No native speedup, broad no-regression,
actual GPU execution or full Mac qualification follows.


## Intrinsic grid text at fractional display scale

Passive managed grid labels now retain completed intrinsic shaping across
rounded physical edge changes when natural dimensions and fixed text/style
inputs remain unchanged. Unsupported sizing units/parts and all content, state,
pending and full-layout paths keep ordinary sizing. No extra cache is added.
Software full-layout/frozen-prior pixels, public/private geometry, dependent
styles and custom-shaper pending regressions qualify this source change.
An isolated owner-sync plus local-layout kernel improves from about 819 to 275 µs
for fixed rows and 707 to 280 µs for variable rows at DPI162, including predicate
costs. These are offscreen kernel observations, separate from painting, native
frames and the still-open 4/6 ms and full Mac gates. See the
[contract](../specs/grid-native.md#reusing-intrinsic-grid-text-across-rounded-cell-edges).


## Intrinsic grid sizing and matched native frames

The [intrinsic-sizing checkpoint](reports/macos-grid-intrinsic-sizing/README.md) records guarded source
`599442a`, mixed-text full-layout/prior-rule geometry and pixel oracles,
dependent-style and pending-work controls, full installed automated checks
and one Anthropic Prism review. Forty matched native ordinary processes
complete 4,800 samples and 2,400 full geometry pairs. Candidate TotalUIWork
p95 is lower in all twenty pairs, with median paired reductions of about
10–14% across the four groups. All twenty candidate processes still miss
4 ms p95; eight meet 6 ms p99. These are scoped workload observations;
P4/P6/P9/P11, broader performance/memory/endurance, physical/human, current
SDK closure and release/distribution remain open.

## Current private SDK and relocated consumers

The [current private SDK refresh](reports/macos-sdk-refresh/README.md) binds clean `2010a33` to
7,421 hashed files including stage 0076 and current intrinsic grid sizing.
The relocated generated consumer passes race/vet/build and public API contracts
pass race/vet. Nine application roots and four actual replay/workspace command
entrypoints compile from a second SDK extraction, with 65 direct Go commands.
One separately admitted SDK hello run verifies its initial rectangle pixel,
three further readbacks and owned shutdown. Other demo runtimes, ordinary frame
and memory/endurance budgets, physical/human workflows, native component/rights
closure and public distribution remain open.

### External workbench automation consumer

Added `examples/automation`, an independent public-Go client of the existing
semantic API. Its dedicated workbench diagnostic waits for a native window and
nonempty public GPU snapshot, disables geometry persistence within that process,
and joins readiness retirement plus endpoint transport on closure. Descriptor
write failures close the endpoint before the executable's exit helper.

The client runs 40 serial queries/actions against 100k or million-row workbenches,
checks typed refusals, stable virtual identities, redaction, command availability,
rotate/raise, sparse Undo/Redo and edit/selection ABA rejection. Raw native output
is discarded, credentials stay in memory, invoke transport failures are not
retried, and bounded child cleanup preserves failure observations. A successful
run also requires matching terminal audit metadata, normal native exit, endpoint
removal and a refused post-close Call. The example is included in the non-GUI
example check suite. The [native campaign](reports/macos-automation-native/README.md)
passes 40 semantic requests each at 100k and one million actual records from clean
781f847. Evidence is the source-reviewed client assertions and saved scalar
summaries; raw per-call DTOs/native streams are deliberately discarded. Both
workbenches exit normally and their inherited owned groups are observed absent.
Physical input, AX, presentation, frame, memory, endurance, release and full macOS
acceptance remain open.

The [current grid profile diagnostics](reports/macos-grid-profile-diagnostics/README.md)
complete three CPU-profiled native apps and 3,072 callbacks on clean `30025cc`,
plus one exact-PID native stack sample. All owned groups close and are observed
absent. Go `cgocall` totals do not establish native function cost; the native
sample includes substantial waiting and AX work. Profiled phase distributions
remain diagnostic. Ordinary frame targets, memory/endurance, physical/human
workflows and full macOS acceptance remain open. No runtime source changes here.

The [pinned-source GPU consumer checkpoint](reports/macos-source-built-gpu/README.md)
builds the exact WGPU source/header commits with recorded Rust/compiler/features,
then passes private static-only Go consumers and nine headless GPU roots with
23 subtests. Two Rust stripping warnings and the failed initial consumer attempt
remain preserved. The candidate is not adopted into installed libraries or the
SDK; complete component/rights closure, broader native UI/presentation, older
macOS runtime, performance/memory/endurance and full Mac release remain open.

The [source-built GPU workbench check](reports/macos-source-gpu-framework/README.md)
links the private pinned-source archive into clean `7829cfc` and passes public-Go
semantic clients at 100k and one million records: 40 requests each, native window
and nonempty GPU snapshot readiness, edits/Undo/Redo, typed and stale-revision
refusals, joined transport, owned endpoint removal and post-close refusal. Both
owned groups retire. This is source-reviewed assertion/scalar-summary evidence;
raw DTOs/native streams are not retained. The original warning-policy failure
and omitted-empty-setting parser failure remain preserved. Installed libraries,
runtime source and SDK are unchanged; physical input, presentation, ordinary
performance/memory/endurance, rights closure and full Mac acceptance remain open.

### Private source-built GPU SDK selection

SDK packaging now supports an explicit private ARM64 static GPU candidate with
an expected digest and separately retained unverified provenance. Generated
applications and native examples resolve it locally, with exact capsule and
manifest checks plus a required relocated Go module origin. The framework root
and portable examples retain their existing module graphs. Public packaging
refuses the candidate override and the native/Rust/standard-library component
and rights audit remains pending. See the [SDK contract](../specs/sdk.md#private-arm64-mac-gpu-candidate).
