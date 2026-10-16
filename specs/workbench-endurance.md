# Composed native workbench endurance

The separate [stored-workspace smoke runner](../examples/replay/cmd/workspace-smoke/README.md)
defines an 18-action native accessibility/owner scenario for 10k or one million
stored table rows and 100k outline nodes. It uses the independent public example,
commands/history, palette activation, measured cell selection and two native
windows. Each proposal is sent once; refused or uncertain outcomes fail. A
streaming source oracle and completed table/outline geometry are required. Each
step has a cooperative 30-second helper cap without extending caller deadlines.
Scenario completion is handed off after application closure returns. Final
window/service retirement and worker joining are checked separately from an
external process receipt. Renderer request flags and present-attempt counters do
not establish actual GPU coverage or OS presentation. Compilation/offscreen
checks do not qualify a native run. Physical input, the full P9 record, P11,
performance/memory budgets and overnight endurance retain their own gates.

The Mac workbench has a dedicated `-endurance` mode. It requires a new clean
absolute `-endurance-output` directory and a 10-second to 12-hour duration.
Ten seconds is a preflight; at least ten minutes of activity is required for a
steady composition trace. It refuses smoke, document-smoke, exit-after and
semantic-server modes in the same process. Window geometry persistence is
locally disabled. It changes no input source, VoiceOver or clipboard setting.

The existing inspector demonstrates controlled value TextField, package
RadioGroup, rotation SegmentedControl, disclosed FormGroup and StatusBar.
Valid property changes use the same sparse document journal as commands and
undo/redo. Integer values preserve the full uint32 range. Invalid requests
leave accepted document/history unchanged; callbacks use the current owner
selection. Disclosure expansion is transient UI state, outside document data.

The composed fixture contains the indexed outline, pooled table, docked panels,
command bar, GPU canvas, document/history and inspector controls. Every activity
step selects a component, checks property replay, moves the table viewport,
updates canvas fit/zoom and invalidates the native view. Periodic actual Cocoa
value requests count only after observing the accepted domain value. Requests
revoked by a newer semantic generation may retry within their fixed deadline.
This proves eventual controlled edit acceptance, not loss-free delivery of every
individual native request. Native resize requests and effective pixel/window
manager dimensions are distinct observations.

Allocating worker results use a captured revision, deterministic checksum and
bounded task slots. Both fresh owner completion and stale-result refusal must
occur. Worker termination and owner completion are joined before shutdown,
including failure cleanup. Timed-out queued owner callbacks are revoked; an
already executing callback is joined before closing its resources. Process
watchdogs bound a broken native loop. Geometry reads hold the native window
mutex and occur outside the application owner.

Private outputs retain raw observations, a bounded 256 MiB Go execution trace
and an 8 MiB final post-GC heap profile. Samples record Go memory/GC, goroutines,
mailbox/task counts and effective geometry. Observation/profile overhead is part
of this diagnostic run. Owner round-trip time is not frame UI work, presentation
latency or CPU execution time. RSS/private/native/GPU observations need separate
platform evidence; unavailable attribution must remain explicit.

The complete original P11 remains required: all accepted components in one
workspace, primary roughly 1,000 visible controls/text plus canvas at the stated
logical geometry, five measured runs, cold/warm distinction, pacing/input/frame
and native/GPU/resource budgets, 1,000 composed lifecycle cycles, overnight
idle/background soak and the physical display variants. A passing bounded
composition run alone does not satisfy that gate or the unchanged P6 targets.


## Immutable component worker snapshots

Component data uses immutable scalar pages: at most 64 branches × 64 leaves ×
256 records. The application/document count remains 100k..1m. Snapshot captures
share a root and copy only source descriptor/revision. Owner edits and history
replay copy one root/branch/leaf (3,072 array bytes on arm64) before publication;
workers never observe mutation of their frozen pages. Old roots are retained only
by their current source/snapshot holders; journal entries capture sparse values.

Save capture shares component data. Full record materialization/validation/JSON
encoding and decoded page construction run on the document worker. Load admission
transfers the prepared root; v1 JSON stays compatible. Geometry/reference capture,
initial pre-window construction and legacy table projection admission retain their
existing work and are not declared constant-time by this change. Indexed worker
traversal has a measured lookup tradeoff. The [snapshot checkpoint](../docs/reports/macos-part-snapshots/README.md)
records isolated operation costs and native save/load/calculation/editing checks;
it does not qualify whole-frame performance, memory or full P11.

## All-component native lifecycle fixture

The native window-cycle command's `-composed` mode uses fresh editor/history,
managed grid and indexed outline, docked GPU canvas, modal and every accepted
UI kind, plus 1000 dense short labels. It retains the survivor/context/completion/
late-admission/native teardown checks. Actual Cocoa text changes count only after
controlled model acceptance; model history replay and grid/tree geometry stay
separate observations. Modal proof is native AX target presence/absence.

Density reports defined labels and frame containment within the measured root
and scroll viewport after geometry publication. Requested1920×1080 dimensions
are not effective geometry or100%scale qualification. Software tests replace
Extension content with allocation placeholders and qualify only controls/labels.
Composed deadlines are 10 seconds..4 hours; the original mode remains 20 minutes.

Workbench disposal recursively retires private panel records before widget
teardown, including unadopted records after a failed shape rebuild. It revokes
interaction generation and refuses retired queue admission/delivery. Ordinary
window removal clears backing slots holding View/CloseRequest closures. Domain
models remain application-owned. The [checkpoint](../docs/reports/macos-composed-lifecycle/README.md)
records short native observations. The later [AX ownership checkpoint](../docs/reports/macos-ax-retention/README.md)
retains the failed 600-cycle campaign, corrects native teardown temporary retention
and passes 1000 composed cycles with all 1000 label frames contained per cycle.
Observed effective root geometry remains separate from the requested primary
physical/display workload. Full P11 still requires the five measured cold/warm
runs, sustained pacing/input/frame and resource budgets, overnight soak and
physical display variants.

The [completed-outline retention campaign](../docs/reports/macos-tree-retention/README.md)
repeats 1000 composed cycles on committed bf3b1b5 after retained-view publication
fixes. Each cycle additionally matches owner geometry to 18 completed native AX rows;
source, prepared inputs, producers, authorization and terminal cleanup qualify. Warm sampled
Go heap is 42.6–46.7 MiB and RSS 350.4–366.0 MiB. This establishes that sampled
lifecycle workload. Native standard-control callbacks, hierarchical tree actions,
modal data-view transfer, paced frame/input budgets and full P11 remain separate.

The [composed native action checkpoint](../docs/reports/macos-composed-actions/README.md)
adds 20 actual Cocoa-selector steps per fresh window: standard controlled values,
button, disclosure heading, tabs, text, modal buttons and hierarchical tree
selection/collapse/expansion. Owner callback/request acknowledgment precedes retries;
accepted asynchronous requests wait for worker/index/geometry and native publication.
Disclosure uses one window-scoped reserved task with off-owner bounded index work.
Scalar partial action records survive failures. Native command/fixture race and
offscreen layout checks pass; five plus twenty native cycles qualify 500 steps.
The tree now has four children beneath n0 and 64 stored nodes. This changed
interaction fixture has no new 1000-cycle qualification; the earlier retained
campaign remains versioned to bf3b1b5. Sustained P6/P11 budgets, modal data-view
transfer, physical input/display and human accessibility remain separate.

The current fixture also includes the controlled Breadcrumb composite alongside
the progress ring/gauge and color picker. Its two final Cocoa selector steps
require exactly one rejected root request (path unchanged) and one accepted
ancestor request (path truncated). Completion checks the ancestor's transition
from native button to current-location static text, retirement of the old leaf,
and positive native geometry. This raises the plan to 26 steps per fresh window;
earlier 24-step retention evidence remains bound to its original source.

The current action plan is version 2 with 33 steps. A Selections page adds the
bounded ListBox and named managed Canvas2D alongside the existing custom GPU
canvas. Three list proposals require exactly one callback each: accepting the
last ID, rejecting the first ID, then repeating the accepted last ID. All eight
native rows, the disabled row, unique identities and bounded clipped visible
rows must agree with the accepted selection. Host selection of an offscreen row
does not qualify native scrolling or focus reveal.

Two named Canvas2D presses must select the corresponding fresh per-window scene
object exactly once. The native target belongs to the intended canvas parent and
scope; host and both children must share one nonpending completed generation.
Native identities/incarnation persist across the selections while generation
advances. Positive finite clipped geometry, names, roles and selected flags are
required. Returning to Workspace restores the original density/teardown checks.
The catalog/layout/routing regressions include both added accepted kinds. Earlier
24/26-step campaigns remain historical; this 33-step mix needs its own complete
native campaign and does not close P11 or physical input acceptance.


## Idle and background soak mode

The native window-cycle fixture adds `-soak` as a distinct workload: use
`-composed -cycles=1 -sample-every=1`, a 40-second to 12-hour total duration,
and quiet plus settle strictly shorter than each phase, with a cooperative timeout 30 seconds to five minutes longer than that duration.
Four equal phases request a visible unfocused window, complete occlusion behind
its survivor, minimization, and visible unfocused background-worker completion.
The fixture verifies actual process-owned Cocoa visibility, key, miniaturization,
occlusion and screen-point frames before entering and around each quiet sample.
A requested window arrangement alone never qualifies its effective state.
Each phase transition records one bounded, main-thread snapshot of both owned
windows, their actual Cocoa outer frames, native ordering indices, opacity,
alpha and levels, and whether the cover's frame contains the target. Failed
transitions preserve the last observation and fail the run. Frame containment
and ordering never substitute for `OcclusionVisible=false` in the occluded phase.
The cover requests 32 window-manager points of margin on each side of the
unchanged composed target. Occluded-phase admission additionally requires the
observed cover to be visible, unminimized, opaque with alpha one, at the same
native level, ordered ahead of the target, and to contain its full Cocoa frame.
Requested margins are not assumed to survive window-manager constraints.
Fixture geometry reads share the backend window mutex with resize callbacks;
the fixture releases that mutex before dispatching any native operation.

`-soak-quiet` defaults to 60 seconds, `-soak-settle` to ten seconds, and
`-soak-interval` to five minutes. During each quiet interval the harness posts no
owner callbacks, reads no Cocoa state and collects no heap. Getrusage CPU deltas
measure this whole process as a percentage of one logical CPU, including both
windows and runtime/native threads. Primary-window paint-request wake deltas
are separate counters; they do not count every application wake or GPU present.
Endpoint state samples cannot prove continuous occlusion between samples.
An endpoint query error or rejected target state emits one bounded
`vellum.soak.interval-failure` record before returning the original failure.
The record preserves the actual start/end target and cover snapshots, query
status/error, capture times, quiet boundaries and individual state/containment
predicates. Both windows come from one bounded main-thread query per endpoint,
outside the quiet interval. Missing or ambiguous windows fail that query.
An unattempted endpoint is explicitly unsuccessful. No later query replaces a
failed observation. Text fields are limited to 512 bytes of valid UTF-8, numeric
frame/alpha strings preserve nonfinite diagnostic values, and the encoded record
is limited to 64 KiB. Output errors remain joined with the original failure.
Successful intervals emit no failure record; diagnostic cover predicates do not
replace strict native target occlusion or waive a failed run.
Nonfinite window frames and invalid/reversed CPU intervals are refused. Short,
overslept or wall/monotonic clock-divergent samples cannot qualify the 60-second
CPU interval, even if the bounded fixture completes. The original static-idle
target remains <0.5% of one CPU with <=2 app-originated wakeups/s and no repeated
presents; CPU/primary-paint samples alone cannot qualify that complete target.

Current diagnostic builds also retain versioned [native API-attempt
observations](native-activity.md), with raw process counters and separate read
brackets outside the quiet interval/native queries. The current validator must
match their PID to the actual owned process receipt and check exact counter
deltas and brackets. These observations count framework API attempts across
windows; OS wakeups and actual presentation remain unavailable. They do not
upgrade an old sample or waive any idle/overnight requirement.

Forced GC, Go memory/RSS samples, JSON output and one reserved 1 MiB worker
allocation/checksum/owner completion per background sample occur outside quiet
intervals. Both worker and owner termination are joined on error/cancellation.
The existing 512 MiB Go heap and 1 GiB RSS guards are process safety bounds, not
professional native/GPU allocation budgets. Short preflights exercise control
flow without certifying the 60-second CPU target or overnight behavior. An
eight-hour trace is intended for overnight coverage; actual runtime evidence,
app-wide wake/present observations, physical displays and full P11 remain open.


The cycle/soak fixture logs close-policy requests as owner observations and
attaches application/window cancellation plus domain-registration observations to
failed native actions. These reads are individually synchronized, not an atomic
lifecycle snapshot. Policy invocation does not identify a human, OS or application
actor, nor prove completed native closure. Forced teardown bypasses the policy.
An unavailable AX lookup after unregistration remains an error; diagnostics do
not retry it or turn a partial campaign into a pass.

The version-2 33-action [five-window preflight](../docs/reports/macos-composed-controls/README.md)
recorded 57.762 seconds through its fifth terminal sample. A simple 1000-window
projection is 11,552 seconds, already beyond the prior 10,800-second cap. The
composed diagnostic deadline is therefore bounded at four hours, with a separate
owned-process watchdog of four hours five minutes for that exact campaign. This
projection guides test scheduling and is not a sustained-performance result.
The 1000-cycle cap, 33-step plan, sample-every-25 cadence, 512 MiB sampled Go
heap and 1 GiB sampled RSS guards are unchanged; the bounded profile is still
8 MiB maximum. Complete current-source execution and cleanup remain required.


The [committed `6330b2d` campaign](../docs/reports/macos-parent-graph-endurance/README.md)
now completes the current 1,000-window/33,000-action mix under that unchanged
four-hour bound, with 41 forced-GC samples and a bounded final Go heap profile.
Exact raw order, action/owner/native completion and source/cleanup bindings pass
the strict post-terminal validator. Fifty-four extra queue attempts/admissions
remain separate from 33,000 completed actions; their causes were not observed.
The terminal snapshot retains one hidden inactive untitled Cocoa window and no
original-menu-restoration proof. Sample ranges and a Go profile do not establish
unbounded-retention, native/GPU/private allocation budgets or the empty-control
baseline. The original failed overnight remains failed. Physical/human, ordinary
latency, full idle/wake/present and P11/Mac requirements remain unchanged and open.
