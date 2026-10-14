# Managed indexed tree checkpoint

`ui.DataTree` is an additive leaf backed by one owner-held `virtual.Outline`.
One live binding per model is supported. Copies, a foreign owner, two simultaneous
windows using the same model, closed originating lifetime, and incompatible native
budgets refuse. Existing `virtual.Tree` and `cogent.VirtualTree` remain compatible.

`TreeChange` reports an already accepted selection or viewport transition. It does
not request replay of that mutation. `TreeDisclosure` is an expansion/collapse
intent. Its captured row reference, visible index, model incarnation, current
scope, and originating lifetime are checked before the callback. The application
admits a task, calls `Outline.Prepare`, builds the sealed index off-owner, and
admits it through `Outline.Complete`. No tree scan or builder allocation occurs
in the native input, owner synchronization, layout, render, or accessibility paths.
The old index remains interactive while an index request is pending.

The control supports primitive pointer row selection/disclosure, logical-pixel
wheel scrolling, a normalized vertical scrollbar, Up/Down, Page Up/Down, Home/End,
Left collapse-or-parent, Right expand-or-first-child, and Return toggle. Boundary
navigation is a consumed no-op. Unsupported modifier combinations do not mutate
the tree. Only the host is a tab target. Focus, disabled ancestors, active modal
scope and native tab-page ancestry are checked at delivery. A down/up pair must
match the generation-bound row ref, host/row bounds, display scale, and native
incarnation; moved or reused row slots refuse the press.

The native pool contains at most 500 rows, each with one disclosure glyph and
one label. It shrinks on smaller accepted views. Row height/viewport are limited
to 65536 logical pixels and bounded native coordinate extents. Indentation clamps
to 24 levels rather than allocating nested widgets for deep chains. Admission
checks each materialized key/parent and descriptor ID at 4096 bytes, each label at
64KiB, and aggregate logical payload at 8MiB before pool growth. Aggregate payload
counts raw keys, worst-case HTML-escaped labels, generated row/disclosure targets,
and encoded AX row IDs. This budget excludes sealed data/index arrays, sharing,
allocator/shaping overhead, transient AX JSON serialization and process heap; it is not an RSS limit.

`InspectTree` is owner-held inspection of a completed native frame. It returns
detached visible row/disclosure/page bounds in scene pixels, selected and expanded
states, accepted stamp/generation, retained native row count, and native display
scale. Pending geometry or stale accepted model state returns `ErrTreeLayout`. A failed
projection withdraws measurement/AX actions, and later paints cannot republish
the retained old rows until successful bounded synchronization.
Normal inspection never scans immutable stored nodes. Previous/next AX page
buttons reveal a viewport, retaining selected identity. AX tree-row identities
use generated `tree:` namespace, expose current materialized rows only, and retire
when the accepted projection or binding changes. The existing Cocoa bridge maps
these semantic roles; human VoiceOver/external AX certification remains separate.

Manual input uses the actual Core hit-test, focus, modal, key, scroll and originating
queue router. Frames are explicit software frames with supplied fonts/theme; no
implicit clock/worker advance is added. `Manual.Snapshot` detaches nested Grid and
Tree geometry arrays as well as ordinary element arrays.

The independent `examples/loadedtree` workload stores 10k/100k immutable nodes with
a wide branch and a deep chain. Generation, validation, key/parent indexing and
visible-index construction run on bounded workers. Actual routed CPU checks are
separate from Cocoa owner-command/measurement smoke. No desktop physical-input,
human IME/VoiceOver, full P6, native services/GPU, or frame-latency claim follows
from these correctness checks.


The opt-in [warm tree recorder](tree-performance.md) links reserved actual scroll
input to the owner transition and matching completed native frame. Pending/stale
layout and unchanged/refused delivery cannot qualify a successful timing sample.
The native row sync installs all row/footer state before one host updater pass;
separate per-row passes are redundant. Tests compare retained updater/style counts
and final software pixels across four DPI settings. The ordinary input contract
and existing ownership, scope, geometry and payload limits remain unchanged.
