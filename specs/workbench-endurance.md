# Composed native workbench endurance

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
Composed deadlines are10seconds..45minutes; the original mode remains20minutes.

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
