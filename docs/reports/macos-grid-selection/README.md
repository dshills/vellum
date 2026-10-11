# Captured grid row-selection checkpoint

The portable Grid adds opt-in captured row membership: Replace/Extend/Toggle,
All/Clear, stable-key anchors and immutable worker snapshots. One captured row
interval plus bounded detached key exceptions represents a million-row selection
without an object or widget per row. Same-incarnation sorting, filtering and
height changes retain captured identities; a new incarnation clears membership.
CapturedRows describes that captured scope, not a later filtered row count.

Mac input captures modifiers at MouseDown: plain replaces, Shift extends and
Command toggles. Plain arrow/Page replaces, Shift navigation extends, Command+A
captures All and Escape clears outside existing modal precedence. Accepted
membership drives visible cell backgrounds and AX row/cell attributes, with a
separate active cursor. Native SelectedCells includes only visible materialized
cells. Default callers retain their existing active-cell behavior.

The independent Grid application adds `-row-selection`, per-window All/Clear
controls and captured-count status. It combines with `-variable-rows`. Its
immutable procedural selection index contains detached identity and scalar
projection parameters, with no mutable owner/counter backpointer. These fixtures
do not measure stored-million-row domain or index memory.

Internal review reproduced a same-stamp captured interior-row disagreement and
a missing final live-source guard. Both exact original regressions pass after
bounded row-index agreement checks and final stamp verification. Consumer tests
also caught the native cursor border expanding cell geometry by two pixels.
Cell/header placement now fixes the indexed exterior rectangle and derives
nonnegative content inside it. The unchanged geometry oracle and new four-scale,
one-pixel-row regression qualify this correction. Historical failures remain
separate from final source-stable qualification.

Core tests cover random set operations, retained concurrent snapshots, added and
filtered keys, hidden anchors, payload/capacity/count overflow, cancellation,
atomic source drift and generation exhaustion. In the common 800×140 viewport,
SelectAll makes exactly 10 live RowKey and 10 IndexOf calls, one capture, zero Cell
reads and 19 combined fixture-index RowKey/IndexOf reads each at both 10k and 1M
rows. This proves bounded callback work, without a latency or allocation claim.

Native tests use actual offscreen primitive pointer/key routing and production
Cocoa selector fixtures. Consumer tests preserve existing fixed/variable suites
and test selection at drain batches 1/7/64. Actual Cocoa smoke checks completed
frames, owner-issued All1M minus two, captured membership across filtering,
Clear with an older snapshot retained, and independent window lifetimes in fixed
and variable modes. Smoke does not synthesize physical desktop input or certify
the compositor's teardown. Process exits and completed stages are separate data.

[qualification.json](qualification.json) binds final test inputs, commands and
raw logs. Native and consumer versions retain their original before/after input
maps. Archived producer scripts and `.go.txt` files are inert historical evidence.
The current core/desktop wrappers refuse existing evidence, own timeout process
groups and record launch/timeout failures. Supply a fresh output directory when
reproducing; the desktop wrapper optionally takes a consumer qualification path
as its second argument. Injected failure checks and the existing process helper
tests separately cover that wrapper behavior.

The [review record](../../reviews/macos-grid-selection/README.md) preserves
internal findings, their resolutions and the single significant Prism batch
using the authorized Anthropic Claude Sonnet 5.5 provider.

This advances P6. Editable cell conflicts, loaded worker projection integration,
large-tree acceptance and sustained latency/frame/memory workloads remain open.
External AX and human IME/VoiceOver acceptance are separate. No input sources,
VoiceOver settings or macOS preferences changed, and no public release, signing
or notarization ran.
