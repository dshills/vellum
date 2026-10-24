# Mac personal-use widget checkpoints

Each source checkpoint uses Prism with Anthropic claude-sonnet-5-5, complete
staged-diff coverage, and focused race/routing tests. Native app observations are
recorded separately after packaging. No release/pacing/rendering policy changes.

## Color well

Root UI race suite and offscreen ColorWell/ColorPicker managed tests passed.
`color-well.json` reports a medium-confidence stale-session concern. Retained as
intentional: opening must advance the session to revoke closed-panel callbacks;
old button descriptions must not toggle a newer panel session. Owner dispatch
reconciles the view after activation, before further native editing. The managed
routing regression opens through Return, commits hex through the live field, and
closes through Escape; component tests cover old sessions and read-only state.
The suggested removal of the opening epoch would permit callbacks from a prior
closed description to reach a new editor. No source change made for this finding.

## Token field

Focused UI race tests, managed committed additions and keyboard chip removal, and
gallery model tests passed. Prism completed with zero findings. Completion menus
reuse ComboBox; no claim of per-keystroke filtering or Mail-style inline editing.

## Font picker

Focused UI race tests, managed Up-arrow size changes, committed family changes,
preview typography and gallery tests passed. Prism's catalog-iteration suggestion
was applied and retested. Both reviews also flag silent refusal of typeface events
from a prior family; this is deliberate stale-event protection, covered by UI
tests. Those actions must not edit a different family's typeface or emit user
validation errors. Accepted callbacks are followed by owner reconciliation, which
installs current-family controls. No remaining actionable finding.

## Path control

UI race tests, icon artwork tests, managed overflow-menu navigation and gallery
model tests passed. A focused managed test confirms disabled read-only rows
propagate Disabled to both icon and label; explicit child flags are redundant.
Prism then identified misleading read-only overflow help: corrected to expose
the omitted ancestor labels without an action hint, with a new UI regression.
Final review flags redundant navigation in a two-item path. Retained deliberately:
the current segment consistently offers its ancestor menu, even when that sole
ancestor also has a direct button. This is optional UX preference, not a failure
of controlled navigation or lifecycle. No remaining actionable finding.

## Secure field

UI race tests and vet passed. Focused managed/native checks cover accepted text,
visual reveal/remask, blank AX values/default names, protected secure subrole,
reflection and fast semantic JSON, manual snapshots, omitted text geometry,
Cocoa substring denial, native editing, refused switches, both accepted switch
directions, and inactive secure queries. Gallery model tests and driver vet pass.
Initial fixtures needed isolated manual lifetimes and a substring-only Cocoa
probe because inert views intentionally do not implement window geometry; fixed.

Prism prompted centralized semantic redaction, detached action filtering, explicit
nil-handler docs and protection before native unusable-state delegation. Applied
and retested. Final review repeats a protection-lag concern; rejected based on the
actual transaction invariant: native clientID, active state and secureClient change
together only in accepted publish, before replacing text. SetSecureClient changes
only a pending request. Refusal retains old client/text/protection together; no
incoming secure text exists in native state until acceptance. Attachment starts
with empty text and no active client. Pure Cocoa tests exercise refused and both
accepted switch directions and prove the old/new protected snapshots and substring
queries stay redacted. Immediate clearing on a refused public switch would break
this invariant. No remaining demonstrated privacy defect; no memory-erasure or OS
secure-event-input guarantee is made.

## Native packaging follow-up

The first packaged light check exposed a blank path strip inside its form group.
Exact Scroll/FormGroup reproduction confirmed the toolbar had zero content width.
PathControl now grows horizontally within the allocated group; the failing
regression passes alongside path routing/UI tests. Secure widget SemanticsProperty
metadata now also redacts Text/default Name, covered in masked/revealed raw-node
checks. Prism's remaining suggestion concerns Number/SelectedID on secure nodes
and Icon handling: Secure is valid only on TextField, whose accepted secret is
Text; unrelated descriptor fields do not derive from that value. An explicit
application label is not replaced by a password value. No demonstrated additional
secret exposure; the concrete text/label paths are protected and tested.

Native pre-rebuild observations: light ColorWell opens a readable compact panel,
recent-color selection updates RGB/hex, Escape restores its focus; native typed
token commit creates one chip and keyboard Return removes it. Secure Show changes
its action label while AX remains a secure field without a value. Final package
checks are kept with the local delivery receipt rather than presented as physical
pointer or human VoiceOver qualification.

## Dark surfaces

Native dark check found the upstream toolbar's light default background behind
path segments and token chips. Both composites now explicitly inherit their
containing surface. TokenField's root also fills its allocated width so its
overflow toolbar cannot collapse with the default entry. Race/routing tests pass;
a completed software frame verifies empty pixels in both toolbars equal the dark
form surface. Prism requests confirmation of GrowX1: intentional, included in the
review context and this rendered containment test. Caller Style still overrides
GrowX and max size. This is the same overflow-container sizing constraint proven
by the path regression, not an unrelated change.

## Overnight identity and queued catalog actions

New regressions failed on the preceding checkpoint: removed/disabled/renamed
TokenField suggestions were still delivered between owner data mutation and the
next View; an observed SecureField ID replacement inherited visual reveal state.
Token actions now match the current entry against the rendered snapshot; live
suggestions also retain the 64-entry bound. SecureField resets reveal and issues a
private identity lease at each observed ID replacement, revoking old input/reveal
callbacks even across an A→B→A render sequence. API and mounted-driver regressions
pass, as do the adjacent widget routing/Cocoa privacy suite and vet.

Prism's first identity-round-trip finding was reproduced and addressed by the
lease. Its final finding concerns changing ID away and back without a View. Such
writes do not create or replace a rendered control; direct public Go-field writes
cannot be observed retrospectively by a component, and a generation incremented
in View has the same observation boundary. Retained components belong to one
owner, and callbacks capture their original receiver; copying an immutable lease
does not redirect callbacks to a copy. No additional API/setter contract is added.

## Overnight color-editor lifecycle

Public regressions reproduced retained private open state after observed ID,
disabled and read-only transitions. These now cancel/revoke the private session.
Prism identified the invalid-description path; all View errors now close it too.
Tests cover open→empty/reserved ID/oversized recents/invalid style→repair with no
automatic reopen and no admitted old editor action. Identity is observed before
validation deliberately: a rejected replacement still invalidates the old private
panel. Session capture for returned callbacks occurs only after this observation.
Final review requests an empty-to-valid open-state test; that exact sequence is
already covered by TestColorWellInvalidDescriptionCancelsSession/empty-id.

Focused race tests pass, including actual managed replacement/disable/read-only
cancellation and 12 keyboard open/Escape cycles. Every cycle retains the keyed
editor, clears the transient dialog body/active state and restores anchor focus.
This is automated routing/lifecycle evidence, not a physical-pointer soak claim.

## Overnight Mac list commands

Reproduced Command-C erroneously toggling Charlie via typeahead in a multi-list.
Unsupported command letters now stay available to application routing. Command-A
proposes all enabled rows through the existing controlled selection callback,
without moving focus. The public Manual OS-service guard still refuses clipboard
shortcuts; Command-C isolation uses retained-row chord dispatch and is labeled as
such. Command-A/rejection/Shift-range/lock tests use normal Manual routing.

Prism's rejected-anchor concern was reproduced through Shift navigation and
fixed by leaving the prior anchor alone. The public Apply replacement contract is
explicitly tested with a different TargetID and with empty IDs. Source row/catalog
eligibility normally prevents empty select-all; a defensive empty-set no-op also
keeps it from clearing selection. Caps/Num flags are ignored at both outer choice
admission and inner Command-A matching, tested without changing OS lock settings.
Final review explicitly reports no functional issue/no change needed. Full list
race tests pass; native Command-A verification follows packaging.

## Flat Mac source-list rows

Native inspection found outlined segmented-button rows and checkmarks rather than
flat source-list membership. Multi-list rows now suppress those decorations,
retain the focused-row outline, and fill the allocated Mac form width. A focus
regression reproduced selected rows staying blue after Tab left the list; a
coalesced window-owner restyle updates all selected rows on departure/return.
No timers, unbounded queues or new public API are introduced. Retired records
remain protected by postControl's existing lifetime check.

Focused UI race tests, native list/choice race regressions and driver/gallery vet
pass. Completed light/dark software frames assert active and inactive row pixels,
form containment and focused versus unfocused borders. Mode transitions retain the
row and restore the original single-selection icon and border presentation.

Both Prism reviews (Anthropic claude-sonnet-5-5, complete coverage) assert that
SetType does not restore icons. This is contradicted by pinned core/switch.go:
SetType unconditionally assigns Check/None/None for SwitchSegmentedButton. The
passing retained-row transition test verifies these fields and configured icon
size; syncChoice also Updates children when Multiple changes. No redundant second
default-icon table was added. Raw findings are retained for traceability; this is
a documented false positive, not a claim that Prism returned zero findings.

## Inactive source-list label painting

A new completed-frame test reproduced darker label-sized rectangles inside the
inactive gray selection in both themes (light label 195 versus row 217; dark
label 113 versus row 83). The core text child's box was repainting the inherited
translucent row background. Only multiple-list text children now skip their own
box; glyph rendering and the parent selection/focus remain unchanged. No upstream
rendering patch is needed. Light/dark pixels now match at the label's empty corner
and row interior. Adjacent list race tests, retained mode tests and vet pass.

Prism Anthropic claude-sonnet-5-5 reviewed both versions. Its child timing concern
is not reproducible: NewSwitch registers Makers in Init and creates children on
later Update, after this hook is installed; newly created children invoke the
same hook. The prior hook is preserved. Stylers belong to individual child
lifetimes, so replacement does not accumulate them on an existing label. The final
RenderBox restoration concern is explicitly covered by retained single→multiple
mode assertions; core resetStyleWidget clears/defaults the entire Style on every
style pass. Both restored single-mode box and suppressed multi-mode box pass.

## Whole-demo reset

Reset fields previously reset only name/search/copies. The new failing-before test
shows edited selections, values, errors and panels survived. Reset now recreates
all sample components on the existing gallery owner, restoring the initial values,
masked secure entry and closed editors while preserving dark/disabled state. Old
components are permanently disabled before replacement. New closures are created
against the same receiver; no initialized model with foreign callbacks is copied.

The entire gallery race suite and vet pass, plus strengthened light/dark tests for
old secure/reveal/date/token/font/path/combo/reset action refusal, new callback
ownership, and programmatic reset of a disabled dark gallery.

Prism Anthropic claude-sonnet-5-5 reviewed initial and final patches. Its concern
about dropping non-demo handles/caches is hypothetical: the full struct contains
only sample values, component pointers, errors/toggles and the two explicitly
preserved flags. Clearing the rest is the reset contract. Its final named concern
about combo/date/search not being disabled is contradicted by the existing
setDisabled assignments, which include those three plus all five new components
and the notice. Passing stale-proposal tests cover these paths. No extra framework
generation API or state is introduced for this gallery-only action.

## Bounded viewport and logical-DPI matrix

All five public composites now have geometry coverage in twelve software cases:
logical widths 480/660/960 × scales 1/2 (160/320 logical DPI, proportionate pixel
viewports) × light/dark. Checks require nonempty contained form/control bounds,
interactive child containment and nonoverlapping typeface/size fields. At each
combination, actual target-resolved pointer events open the color editor; its
surface stays in the viewport, and Escape closes it and restores well focus.
All twelve pass with the race detector; vet passes. This adds test coverage only,
not a native Retina/display-change claim or a general minimum-width guarantee.

Prism Anthropic claude-sonnet-5-5 suggested refreshing the popup record after
closure; applied, with explicit retained-record checks. View errors are also
reported outside the view callback. Final review suggests additional nil guards
for every fixed fixture record. Missing controls would already fail the test via
a panic with a stack location; this affects diagnostics, not false acceptance or
production behavior. It is not a blocker for this passing coverage checkpoint.

## Long token labels

A valid 240-byte tag allocated a 1916-pixel chip inside a 384-pixel toolbar.
Clipped hit-test bounds alone hid that defect; the strengthened regression checks
full allocation too. Visible token labels now abbreviate centrally at 24 runes,
keeping the remove suffix visible under the compact Mac defaults. Accepted text,
duplicate checks, removal payloads and full accessible names are unchanged.
Short/long/Unicode value tests and actual mounted keyboard removal pass, as do UI
race tests, adjacent token/dark toolbar regressions and vet. Prism Anthropic
claude-sonnet-5-5 completed coverage with no findings.

## Standalone token-removal focus continuity

Native keyboard removal of a chip left no focused control reported by AX. The
standalone demo now uses the existing application-owned root FocusRequest to move
to tag entry after it accepts a removal. The ordinary gallery still embeds view
without a nested focus request; main selects windowView only for -mac-widgets.
Missing removals do not request focus. Reset clears the target while retaining the
sequence, so another removal after reset is a new request. This is example owner
policy; TokenField does not emit an invalid nested root focus request.

Focused gallery model race tests and vet pass. Prism's initial suggestion about
in-place token-slice mutation was addressed by constructing a fresh bounded slice,
with a regression retaining the prior accepted slice. Final Anthropic
claude-sonnet-5-5 review completed with no findings. Native focus verification
follows packaging and is kept in the private delivery receipt.

## Five-widget disabled/re-enabled routing

Mounted Manual tests cover secure entry, path navigation, font size, tag removal
and the color well. A disabled target is not enabled in inspection and refuses
actual targeted MouseDown with ErrManualDisabled; keyboard Return causes no
accepted edit. Re-enabling plus an explicit new focus request retains the native
widget and restores activation. Non-color cases propose exactly one edit; color
opens and Escape cancels without a color edit. Race tests and vet pass.

Prism Anthropic claude-sonnet-5-5 asked for stronger secure input semantics. The
test now checks zero commits after the typed chord, exactly one after Return,
and exact accepted synthetic text from insertion into the prior selection. The
fixture is ASCII, so its UTF-16 offsets equal byte offsets; Unicode native-input
semantics are covered separately. Final review's call-count concern is internally
contradictory (it acknowledges the non-color branch runs for secure): calls==1 is
asserted for secure there. Only non-secure fixtures have a nil secure-value getter.
No further change is warranted by that finding. This is software event-routing
coverage, not physical pointer or native text-input/IME qualification.

## Mac input focus surfaces

Six completed-frame repros failed before this checkpoint: a focused white input
painted 233-gray; a dark input (35/35/37) painted 54/54/56. Explicit field backgrounds
now remain flat, nonzero borders use full focus outlines, and transparent
borderless editors do not paint a private box/underline. Bordered rows containing
such an editor share an outer outline across the editor and accessory controls.
The maximum outline allowance is fixed, so focus does not change allocation.

A retained reparent/style-removal regression found stale parent eligibility;
shared state is now synchronized from the current descriptor and theme before
native updates. Reparenting retains the editor, leaving restores the old parent's
appearance, removing explicit style restores filled-field defaults, and edits
still commit once. Initial Prism reports prompted origin-ownership checks too.
The final implementation reuses accessibilityBindFocusInvalidation: it adds no
focus event handlers, owner callbacks, pending wake flags or property-map writes.
Only originating reconciliation updates a changed shared focus cue. Tests prove
a.Drain does not style the native parent, ordinary focus keeps its single existing
AX refresh, and a full model queue cannot fail this passive appearance work.

Focused race checks pass for these transitions, all twelve widget geometry cases,
adjacent combo/search/token/color/font/secure controls, native text composition/
retirement, pointer re-enable and secure selectors. The full gallery race suite
and driver/gallery vet pass. No rendering backend, pacing option, public API or
release gate changes.

All four Prism Anthropic claude-sonnet-5-5 reports are retained. Earlier callback,
property-map and pending-flag concerns refer to superseded implementations; those
paths have been removed. The final theme-order concern is contradicted by the
actual reconcile order: attach(next) installs every record's new theme before
syncRecord(next.r,n); syncInputOutline sees that installed theme and new n.Children.
Parent updates before descendant sync already belong to the native locked
reconciliation structure; no intermediate frame is published. The unit-width
observation also does not apply: public Border.Width is a logical scalar and
applyStyle has already assigned Dp units before this final styler.

## Numeric focus fill consistency

The native Copies stepper still showed a gray focused fill after the text-field
checkpoint. Four completed-frame regressions reproduced it in inline plus/minus
and compact-stepper modes, light and dark. NumberField now uses the same explicit
background/state-layer rule; it returns before managed text border handling, so
Spinner keeps its existing outlined border, arrow geometry and numeric behavior.

All four rendered fills now match the theme. Actual Up/Down commits remain exact,
allocation remains fixed, and the existing pointer half-step, held gesture,
clamping/rejection, re-enable and font-picker checks pass with the race detector.
Driver/gallery vet passes. Prism Anthropic claude-sonnet-5-5 completed coverage
with no findings. Native light/dark confirmation follows in the private receipt.

## Source-list focus ownership

The native-origin regression failed before this change: application Drain ran
one list style pass directly. Row focus hooks and their pending flag are removed.
The existing AX focus notification invalidates the window; originating syncChoice
compares selected rows' actual backgrounds against their active/inactive color
and updates the list if needed. This handles native styling of only the focused
row without treating that as proof that its siblings were styled too.

Focused race checks cover the new boundary, both-theme active/inactive pixels,
single/multiple mode restoration, ListBox lifetime/routing/queue cases and adjacent
input focus. Vet passes. Settled active and inactive lists explicitly report no
further appearance work.

Prism Anthropic claude-sonnet-5-5 reported a speculative perpetual-update concern.
Every valid SelectedIDs entry is selected by ui.ListOptionSelected, and the final
styleSourceRow always installs one of these uniform backgrounds for selected rows,
including disabled rows. No application-supplied image reaches this comparison.
Nil represents an unstyled row and must request the initial update, rather than
skip it. Settled-state tests demonstrate the comparison converges. No change is
needed for the proposed nonuniform-image or disabled-selection scenarios.
Native light/dark checks follow packaging in the private delivery receipt.

## Public consumer integration

The independent sdk-contract module composes all five widgets with no native
driver dependency. It validates the tree, checks exported focus helpers, invokes
public descriptor proposals and verifies owner acceptance, channel merging and
stable token removal. This establishes public API integration only; mounted/native
routing and visual evidence remain separate. Focused race checks and vet pass.
The linked guide covers retention, theme styles, owner dispatch, stable IDs,
acceptance, focus targeting and existing limits, with all five setup patterns.

Prism Anthropic claude-sonnet-5-5 completed coverage. Its sole finding explicitly
says no action is needed and suggests optional nil-callback diagnostics in the
test. These descriptors are required to expose the invoked callbacks; a missing
callback already fails the test at the exact invocation. Fatalf terminates a
missing-node lookup. No production or integration correctness issue was found.
No runtime changes or repeated native campaign are required for this checkpoint.

## Ordinary borderless field focus compatibility

The prior input-focus change also suppressed focus cues on ordinary TextFields
with zero resting borders. Corrected-theme Manual repros failed for inherited
and explicit opaque fills; an unthemed native-style repro failed too. The first
attempt used an invalid empty Manual theme and is not counted as reproduction.

An editor now suppresses its private box/underline only when it is explicitly
transparent and borderless, and its direct parent actually supplies the shared
outline. A row publishes its stable state once at mount; originating stylers read
that state through the current native parent, so reparenting does not retain an
old owner. No focus handlers, model callbacks, polling or backend policy changes
are introduced. Ordinary opaque, inherited-fill, transparent and unthemed fields
retain their native focus cue. Removing/restoring a shared parent's border moves
the cue to/from the same retained editor without losing focus.

Focused race tests cover these cases, fixed allocation across Tab, shared-input
reparent/style removal, queued refresh ownership, numeric fills and Mac field
geometry. Vet passes. Prism Anthropic claude-sonnet-5-5 completed coverage with no
findings. Native composite confirmation follows in the private package receipt.

## Native window activation for source lists

The native attachment already receives main-thread Focus(bool), but the driver
ignored it. Mac multi-selection lists now use that callback and an initial
main-thread GLFW focus query. Each bridge owns one bounded activation signal.
Callbacks record only the latest observation; at most one cancelable deferred
node reaches the originating native FIFO. Publication happens on that window's
render owner before one subtree restyle. All rows therefore see one published
state during the pass, even if a later native notification arrives meanwhile.

Inactive selections are gray with no active row outline. Keyboard focus and
accepted membership are preserved. The originating host holds the signal, so
new/reparented descendants inherit it without per-control registrations. Other
controls' native-window activation cues are unchanged. There is no foundation
patch, polling, new public API or rendering/pacing/release-policy change.

Race tests cover initial observation before binding, bounded bursts, unchanged
latest state, foreign source/wake rejection, exactly-once consumption, concurrent
close and bridge retirement. Mounted light/dark tests route the native FIFO
without draining a saturated application queue, check every selected row, retain
keyboard focus, mount a new list while inactive and reject foreign/closed signals.
Vet and adjacent origin routing, ListBox and attachment/window lifetime checks
are recorded separately. Native activation confirmation follows packaging.

Prism Anthropic claude-sonnet-5-5 completed coverage and reported a hypothetical
SendDeferred admission failure. In the pinned events/deque.go implementation,
SendDeferred only returns the error from its prepare callback. This callback
unconditionally returns a valid event and nil; there is no closed-deque failure
or capacity-refusal path. The pending marker therefore cannot be stranded by the
reported error. Queue cancellation and delivery/close races are exercised. No
production change is needed for this finding.

## Activation composition bounds and native observation

Rebinding an already-bound activation source is now a no-op. The mounted tests
also wrap the actual software compositor: one activation transition composes
exactly one changed frame; duplicate notifications and a coalesced return to the
published state perform no restyle or composition. Focused race checks and vet
pass. Prism Anthropic claude-sonnet-5-5 completed coverage with no findings.

Native validation at 610c629 used a temporary public-API two-window consumer and
an overlay that only printed observed/published activation states. CUA invoked
Open control window, then closed that second window. The first attachment's
journal includes active→inactive→active publication on its native render owner;
the second has its own distinct signal. The returned source screenshot/AX tree
retain both selected rows and Alpha keyboard focus. CUA quit finishes the process
with exit0 in 57.24 seconds, below its 60-second application deadline.

This is native callback/publication and retained-focus evidence. It is separate
from automated light/dark color and compositor checks. Targeted CUA images did not
provide an inactive-window observation; blue images are not used to claim that
visual state. An earlier app-switch attempt remained inconclusive, and its
bounded 120-second probe timed out while the desktop tool stalled. The successful
two-window check supersedes that incomplete native attempt. Diagnostic overlays,
helper apps and their logs stay private and are absent from shipped binaries.
No broad release qualification or performance campaign was run.
