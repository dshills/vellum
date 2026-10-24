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

## Standalone tag entry and validation recovery

Public Manual keyboard repros showed successful Return moved focus to the
suggestion button, interrupting repeated entry. At the token bound it could leave
that now-disabled menu focused. A separate same-batch removal→Disable repro
failed with an actual disabled focus-target error during reconciliation.

The standalone demo now requests entry focus after acceptance, entry rejection
or removal. At 32 tags it advances to the enabled color well. Disabling cancels
the pending target while retaining the sequence; re-enabling does not replay it.
Nested gallery view still emits no root focus request. This is private demo
application policy using existing APIs; reusable TokenField and generic Return
semantics are unchanged.

Focused race tests type two successive tags through public Manual events, reject
a duplicate without changing accepted tags, correct it, cross the 31→32 boundary,
and exercise same-batch removal/disable followed by re-enable. Existing reset,
stale-callback, theme, standalone/nested focus and descriptor-error tests pass.
Gallery vet passes. Native repeated-entry confirmation follows packaging.

Prism Anthropic claude-sonnet-5-5 reported a hypothetical Invalid callback at
capacity or outside entry. TokenField invokes Invalid only from its entry Change
path, and that path returns before validation whenever len(Tokens)==MaxTokens.
View configuration errors are returned, not sent to Invalid. Empty/overlong and
duplicate text are entry rejections, for which retaining entry focus is intended.
The persistent root request is sequence-based: applyAccessibleFocus ignores an
already-consumed request; disable explicitly cancels it. No source change is
needed for the reported scenarios.

## Visible, user-worded inline feedback

Light/dark duplicate-entry repros found token-feedback had empty clipped bounds
while entry remained focused. The label now appears before the tag component,
so the focused viewport includes it. Completed-frame Manual checks require a
nonempty readable-height feedback box and entry inside the captured viewport.
Repeated entry/correction, capacity, same-batch disable, reset and stale-editor
checks continue to pass under race; gallery vet passes.

The demo's font/tag/color/date Invalid callbacks remove only the leading ui:
package prefix. Public component errors and diagnostic/configuration error paths
are unchanged. An integration test triggers all four real invalid-input paths
and inspects their displayed feedback. No public API or rendering policy change.

Prism Anthropic claude-sonnet-5-5 completed coverage and suggested extending the
formatter to View errors or narrowing the wording. The spec and review context
already explicitly distinguish input validation from unchanged configuration
and diagnostic errors. View failures describe invalid developer descriptors,
not rejected user edits; retaining their source detail is intentional for this
internal demo. No additional production change is warranted by that finding.
Native light/dark feedback visibility confirmation follows packaging.

## Widget discovery and published-preview status

README and the control index now link the newer personal-use widget demo and
public integration guide. The index includes existing managed controls and
composites, with source-checked public names. Composites are explicitly separated
from node kinds and AppKit controls.

GitHub release metadata and the published release body confirm alpha.3 was
published as a prerelease on 23 October 2026 at
320376ec8573a8603f75faaa3b3a7ea8a8ea0366. The documentation now states that fact,
retains ad-hoc signing/notarization and qualification limits, and distinguishes
the newer source widgets from published alpha.3 assets. Current status notes mark
the implementation log and preview ledger as historical checkpoint snapshots;
those original reports and ledger entries remain intact. The exact-payload audit
and separate publication approval remain required for new public artifacts.

All 146 local links in the four edited documents resolve; all 49 referenced public
identifiers exist in ui source. git diff --check passes. No runtime inputs changed,
so the tested 07d38ba app package remains current. Prism Anthropic
claude-sonnet-5-5 completed coverage of all four documents with no findings.

## Full tag-toolbar keyboard access

A new routed Manual regression fills all 32 tokens with long mixed-script labels,
uses End to reveal/focus the last chip, requires its whole box inside the toolbar,
and removes it with one Return. It checks the full accessible name, one full-value
removal event, the caller's focus request reaching the re-enabled entry, and no
redelivery after settling. Light/dark and matched DPI160/320 software windows pass.
Existing token/toolbar tests pass under race and driver vet passes. No runtime
source changed; this adds focused interaction evidence.

The initial theme fixture lacked the explicit Manual default background and was
rejected before mounting. After that correction, the fixed 400px window at DPI320
exposed a real narrow-layout limitation (roughly 200 logical pixels): a long chip
can have clipped bounds. The matched-scale regression scales the software window
with DPI to retain the same 320-logical-pixel field. It does not dismiss the narrow
finding; that reproduction and the fixed 24-rune abbreviation limit are now in
the spec/integration guide. Native display scaling is not claimed.

Prism Anthropic claude-sonnet-5-5 suggested cloning the test callback's token slice
before removal. This scenario removes the last element: append(prefix, empty...)
only changes the slice length and does not shift or overwrite the backing array.
View captures each removal target as a Token value (shown := t), and fullLast is
also a value. No rendered callback retains a slice header to these items. The
owner callback runs serially; no data race or aliased shifted view occurs. Keep
the ordinary owner-controlled removal pattern; no production fix is warranted by
this hypothetical finding. All reviewed files have complete coverage.

## Oversized toolbar focus recovery

The width200 token repro left the last focused chip completely offscreen in both
themes at DPI160/320. Pinned native ScrollToThis intentionally refuses an oversized
target to the right/bottom. A native final Focus listener now uses the existing,
current toolbar-scope property and ScrollDimToStart fallback only for oversized
frames. Native focus's later pass retains its normal leading margin. No queued
owner callback, polling, public API or renderer/pacing change is introduced.

Eight full-token cases cover two logical widths, themes and matched DPI values.
They require visible oversized leading geometry, unclipped raw ordinary-chip
geometry (not merely clipped published bounds), full accessible names, exactly
one full-value Return removal, consumer focus reaching the newly enabled entry,
an independent managed FocusRequest and native Shift-Tab traversal. The vertical
oversized action checks End visibility, one activation and Home return. Existing
nested/reparent/disabled-toolbar and token cases pass under race; driver vet and
native gallery build pass. The first trailing-edge candidate was superseded:
native focus's later scroll pass favors the leading edge. Narrow labels and the
trailing remove mark may still clip; this is explicitly retained as a limit.

Prism Anthropic claude-sonnet-5-5 completed coverage and raised two hypothetical
concerns. chrome.host cannot be nil in a published toolbar: mountChrome constructs
outer, initializes host to outer, then publishes chrome; no code clears host.
Retirement is guarded before using it. The focus handler tests states.Focused,
not ContainsFocus; nonfocusable layout/toolbar frames do not acquire that state
from normal native Focus (the pinned handler checks abilities.Focusable). The
current scope is read at event time, and the existing binder clears moved-out
frames and chooses the nearest nested toolbar. Passing reparent/nested tests
exercise that lifetime behavior. No stale-host or focus-within scenario described
by the review is reachable under these source invariants. Keep the bounded
fallback rather than adding speculative state or redundant ancestor walks.
Native app confirmation follows packaging; software DPI evidence is not a
physical display-scaling claim.

## Conditional gallery feedback rows

Native light/dark images suggested unnecessary blank space above tags and below
the font preview. Empty text bounds alone did not establish a layout defect: an
initial test only found the empty descriptors, so it was followed by a real
completed-frame comparison. Omitting those descriptors reduced the font group
from 250 to 222 pixels and the tag group from 163 to 140 at the pinned software
font/DPI160 in both themes. The comparison helper/source and its log remain local
execution evidence; the shipped code has no test transform or runtime option.

The private gallery now emits font/tag/date feedback labels only when nonempty.
Tag feedback remains before the editor. Light/dark mounted checks require no
blank feedback nodes, readable duplicate feedback, accepted correction, retained
focus and the same native entry incarnation across insertion/removal. Model tests
also reject and correct font/date values and require their rows to disappear.
Existing reset, stale callbacks, capacity, disable and descriptor-error cases
pass under race; gallery vet and native build pass. Measurements describe this
fixture, not a public fixed-height contract. Public components, shared styles,
rendering/pacing and release policy are unchanged.

Prism Anthropic claude-sonnet-5-5 completed coverage of all five changed files
with no findings. Private-package light/dark visual confirmation follows.

## Retained field visibility after validation layout

Native duplicate-tag images showed the lower entry outline at the scroll edge.
The gallery Manual helper used 760px height while the executable starts at 740.
Matching that size and comparing the unchanged input row's visible height
reproduced clipping from 38 to 24 pixels in both themes. The former in-frame
published-bounds assertion alone was insufficient because those bounds are
already clipped; the regression now compares the unchanged row height too.

Accepted main-scene root focus requests retain one immutable atomic visibility
intent per native window. The existing accessibility host consumes it before
child painting with completed layout geometry, using ordinary ScrollToThis.
Identity, current scope/enablement, focus, scene, request and lifetime checks
cancel obsolete targets. Native retirement clears the pointer. Focus-only
requests use normal NeedsRender; no owner task, deferred callback queue, polling,
backend option, pacing or release policy change. Dialog-scene policy is unchanged.

Focused race checks preserve text/selection and observe no added Focus/Lost
notifications. A wheel-away remains away after unrelated invalidation; a new
sequence reveals the field again without descriptor changes. Removal and native
retirement release pending references. Closing the last Manual application
window requires explicit Manual.Close, so the per-window retirement fixture keeps
a second window alive and uses public Step to retire its origin. Matched gallery
feedback/recovery, popup, nested/reparented toolbars, window replacement and origin
wake checks pass. Driver/gallery vet and final native build pass.

Prism Anthropic claude-sonnet-5-5 first requested a clear locking contract. Driver
refresh and Manual refresh already hold rootMu; Manual OnShow now does too, and
the helper documents that requirement. Targeted race/startup checks pass after
that correction. The final review asked whether OnShow could re-enter the mutex.
The actual Manual syncTopology call invokes RunManual only after refresh returns
and releases rootMu. Pinned core stages.runDeferred sends initial Shown after
rendering; renderwindow.go explicitly runs that phase outside render locks. The
host's pre-paint mutex is released before child rendering. Tabs defer Shown on
their page subtree rather than the root Body. Driver Run refresh visibly acquires
rootMu before applyWindowFocus. These paths do not synchronously re-enter the
mutex as hypothesized; no additional source change is warranted. Both complete
review outputs are retained. Native same-size confirmation follows packaging.

## Focus visibility cancellation, coalescing and origin isolation

Test/spec-only checkpoint; runtime source and the tested 87440e0 packages are
unchanged. Six pending-intent transitions exercise disabled ancestry, disclosure
collapse, same-ID native replacement, an empty request, moved native focus and
invalid-request validation. Content expands before paint so an unwanted reveal
would change the viewport. The replacement is explicitly focused; it remains
outside the viewport rather than inheriting the retired editor's intent. Each
obsolete intent is released and leaves the viewport at its prior scroll value.

The first invalid-case fixture used Manual refresh and tried to recover after its
terminal error, which is unsupported. The corrected case calls the real locked
request validator on its originating owner, verifies refusal and cleared intent,
and preserves the runner's failure policy. No product failure or recovery change
is claimed from that fixture correction.

Three requests before paint replace the one pending value; the latest target
finishes focused and fully visible using raw geometry. A second software window
reuses the same local IDs. Its explicit completed frame must leave the origin's
pending pointer and the second window's scroll unchanged; the origin's later
frame consumes its own intent and reveals its field. No physical pointer, human
accessibility or native two-window qualification is inferred from software scenes.

Focused race tests and driver vet pass. Prism Anthropic claude-sonnet-5-5 completed
coverage of both changed files with no findings. No rebuild or repeated native
campaign is required for this test-only change.

## Inactive-window managed input borders

The regression reproduced active focus borders after an inactive publication for
combo editor/accessory focus, secure text, ordinary bordered/borderless text,
compact numeric stepping and buttons in both light and dark themes. The adapter
now uses the existing published native-window activation to restore explicitly
managed resting borders. Reserved MaxBorder geometry and logical focus remain.

The final focused race run covers those fourteen cases plus adjacent input focus,
source-list activation, activation lifecycle/forwarding and focus visibility;
driver vet and the native gallery build pass. Synthetic editing selection is
settled before activation frame counting. Pixel assertions compare border/margin
areas and exclude editable contents to avoid independently blinking carets. Each
changed publication composes once, duplicate activation composes zero times,
model mailbox work remains pending, selection and geometry survive, and the
original border pixels return on reactivation. Early test-fixture revisions
counted a pending selection frame or compared blinking editor content; these
were corrected without changing runtime behavior to satisfy the fixture.

Prism Anthropic claude-sonnet-5-5 completed coverage of all six changed files.
Its one medium, confidence-0.5 finding hypothesized that alternate button mount
or update paths could omit the final styler. Source inspection disproves that
premise: mountScoped delegates to mountShallowScoped; reconciliation mounts new
or replaced records through mountShallowScoped as well. Reused records preserve
the existing widget and registered final styler, whose closure reads the current
record node/theme. There is one registration site and no second condition or
alternate registration/removal path to factor. The location-corrected finding
points at the activation lookup rather than a demonstrated alternate mount.
No additional product change is warranted; the raw review is retained.

Native activation delivery evidence from the prior two-window checkpoint is
reused. This checkpoint's inactive border pixels are automated software evidence,
not a new live inactive-window capture. Other control classes, unstyled native
borders and caret presentation are outside this bounded change. The user returned
and requested completion of the current task followed by stopping; no further
widget or qualification campaign is included.

## Next widgets, phase 1: Sidebar

Added owner-held Sidebar, bounded controlled sections/navigation, source-list
icons/badges, and separate Sidebar/Inputs pages in the private Mac gallery.
Mounted light/dark checks found a zero-height list under a disclosure; bounded
minimum height fixes it. Final race tests prove trailing badge layout, keyboard
navigation, in-place icon/badge removal and collapsed input cancellation. UI
package race tests, existing list/activation regressions and gallery tests pass.

Prism Anthropic claude-sonnet-5-5 reviewed the phase twice. Gallery validation now
renders an error node and adornment documentation matches Multiple ListBox-only
validation. Its speculative concern about Option comparability is inapplicable:
Option has only comparable value fields, the code compiles, and full snapshot
comparison intentionally revokes callbacks after item replacement. Final review
has no retained findings; both raw outputs are retained. Physical pointer and
human VoiceOver remain separate from automated routed evidence. No render,
pacing or release-policy changes.

## Next widgets, phase 2: Sheet

Added owner-held ui.Sheet on the existing modal Dialog infrastructure, explicit
Cancel/primary actions, fallback Return, exact modal-scope guard, opener focus
restoration and nonresizable Mac card styling. Presentation-kind changes retire
the old Dialog lifetime. Native test proves modal/scrim flags, no title/resize/
extra Close chrome, parent focus exclusion, one Escape cancellation, one accepted
confirmation, scope retirement and safe shutdown in both themes. UI race tests,
adjacent Dialog/Popover tests and standalone gallery tests pass. A native build
seam that accidentally used a test-only lookup was caught and replaced with the
production scoped record lookup before commit.

Prism Anthropic claude-sonnet-5-5 reviewed twice. Owner-only mutation/View and
inert disabled defaults are now explicit. The final validator also accepts a
Button default without an action (with a focused regression), matching the safe
runtime nil/disabled checks. No concurrency guarantee or implicit default action
is claimed. Both reviews and their dispositions are retained. Packaged native
interaction confirmation follows final integration; no AppKit sheet attachment
or animation is claimed.

## Next widgets, phase 3: CollectionView

Added bounded controlled icon/thumbnail collections on retained Multiple ListBox
routing, responsive native wrapping, measured grid arrows, Command/Shift ranges,
Return and generated exactly-once double-click activation. Source identity renews
row/catalog leases across component replacement and collection→ordinary→collection
ABA. Same-owner label/thumbnail/selection refreshes retain eligibility and release
old catalog thumbnail references. Disabled/removed/hidden/modal/closed guards reuse
the established list infrastructure. Each thumbnail remains immutable shared data.

Focused portable/native race tests cover responsive columns, actual routed
pointer modifiers and generated double-click, selection rejection, source/AX
retirement, nil presentation transitions and thumbnail-only source replacement.
Gallery checks and driver vet pass. Parent reran mounted checks with an explicit
Max.Y210 assertion and Shift+CapsLock+NumLock selection to address final review.

Prism Anthropic claude-sonnet-5-5 reviewed twice. Vertical disabled-cell skipping
now stays in its column and stays put when a final-row column is missing. The
thumbnail-only test confirms the pre-update snapshot comparison works. Final
size-order concern is disproved: styleCollectionHost is registered during mount,
then driver.go registers applyStyle; prepared core runStylers executes each list
in registration order, so user dimensions override defaults, as asserted by the
mounted Max.Y210 regression. Lock modifiers now pass the filtered value (the
existing selection reducer already reads only Shift/Meta). The catalog pointer
concern is also disproved: syncListBoxEpochs computes sourceChanged and renews
row/catalog leases BEFORE assigning the new compatible snapshot. Replacement
and same-owner ABA tests prove old queued selection, activation and AX handlers
are rejected. Keeping only the current snapshot prevents old thumbnails from
being retained for the control lifetime. Both raw reviews are retained with these
grounded dispositions; no retained issue requires another product change.

These are routed software checks. Packaged native appearance/input observations
follow final integration; physical input and human VoiceOver are not inferred.

## Next widgets, phase 4: ShortcutRecorder

Added owner-held controlled one-chord recording with Mac modifier glyphs,
canonical command-registry spelling, conflict validation, clearing and cancellation.
The native adapter publishes immutable atomic scopes, intercepts before either
command phase, bounds focus/cancellation references, and rebinds retained buttons
when a modal transfers scenes. Window deactivation cancels main/modal recorders.
Queued Begin/keys cannot survive focus loss, Tab/Escape, hiding, disablement,
replacement, modal retirement or closure. Queue refusal retains bounded cleanup
without clearing the existing refusal/failure policy.

Focused public/native race tests, UI/native vet and gallery checks pass. Actual
software FIFO tests cover Return→recorded chord, Tab traversal and Delete clearing.
A production-accessibility-handler modal fixture proves recording Escape cancels
without dismissing, Tab cancels/moves, and idle Escape restores modal dismissal.
A queue-full fixture uses a direct owner/adapter and checks refusal identity; it
does not attempt recovery from a terminal Manual failure. Initial fixture compile
and deferred-Shown setup mistakes were corrected without changing failure policy.

Prism Anthropic claude-sonnet-5-5 reviewed all13 phase files. Its one concern was
stale demo feedback after cancel/retry. The gallery now clears the prior error
when a new recording begins (Reset already cleared both errors), so a subsequent
cancel does not leave an obsolete error. Public error policy remains caller-owned.
The raw review is retained. Native input is local to a focused active window;
OS-intercepted keys, global registration, non-ASCII/physical-layout-independent
bindings and multi-stroke sequences are outside this version. No OS settings,
render backend, pacing or release-policy changes.

## Next widgets, phase 5: LevelIndicator and final integration

Added passive continuous/discrete capacity and controlled star ratings, threshold
colors plus readable value/status, explicit zero versus Unrated, light/dark and
disabled presentation, bounded immutable preview caching and callback-generation
revocation. Rating uses ordinary accessible buttons and existing keyboard routing;
no new renderer, driver kind or pacing policy. The gallery now has six distinct
pages (five new widgets plus Inputs), with a public control reference, usage/limit
documentation and an independent driver-free consumer contract for all five.

The complete UI package race tests and vet pass. Focused final native race tests
cover Sidebar/Sheet/Collection/Shortcut plus existing managed inactive borders;
all-page descriptor integration, native/gallery vet and gallery build pass.
The initial aggregate runner's SDK filter matched no test; that result is not
consumer evidence. The exact public consumer test was separately run under race
with its actual name and passes (next-widgets-sdk-final). Level source-pixel tests
cover fractions, thresholds, gaps, palette, immutable reuse and extreme ranges;
ratings cover controlled/rejected/disabled/stale/invalid and zero/unrated behavior.

Prism Anthropic claude-sonnet-5-5 reviewed twice. Preview construction now returns
an error through View rather than panicking. Nil receiver validation was already
safe (early return preceded defer and validate was nil-safe); the final path now
returns an explicit error for clarity as suggested. Raw reviews are retained.
Packaged native observations and their limits are recorded separately in the
private delivery verification receipt. No public asset publication is included.

## Packaged live corrections for the next widget batch

Live light-package checks found the Sheet at the upper-left with an oversized
outer surface/native-title side effect, collection labels horizontally offset
from their artwork, and accessible recorder Press starting a session without
transferring focus. Keyboard-focused recorder Return→CommandShiftK already
worked; rating edits, capacity threshold updates and clear worked through native
accessibility. These observations justified this bounded followup.

Sheet now centers measured retained content, removes generic dialog minimum and
outer scroll allocation, keeps the parent title and aligns actions trailing.
With no leading editor, its primary action receives initial focus; routed Return
confirms exactly once, Escape restores the opener. The regression caught an
extra10px outer scrollbar reservation; disabling outer Body/Scene automatic
scrolling removes it, while long content remains caller-owned bounded Scroll.
Collection uses AllocationLayout plus intrinsic label width; actual660px light/
dark tests check artwork, paragraph-line and completed glyph-pixel centers.
Accessible recorder Press uses exact source/session/window/modal guard and the
pending-focus bridge; owner delivery never writes native focus. Tests reject
captured actions after same-ID component replacement or actual modal retirement.

Focused race tests and native vet pass. Prism Anthropic claude-sonnet-5-5 reviewed
all12 changed files. Its nil-guard concern requires an unmounted/incomplete record
(the normal mount path installs control state), but the handler now also fails
closed if a required activation guard is absent. This preserves secure refusal
without a possible nil call. The raw review is retained. Updated private packages
and final light/dark native observations are recorded in the delivery receipt.

## Five-widget gallery appearance correction

Final native dark inspection reproduced light ink over the upstream white surface
and a non-growing sheet page. The gallery now supplies opaque fixed-appearance
window/page surfaces and fills page width, preserving explicit page styles.
Generated unselected tab labels honor explicit strip ink; selected tabs retain
the upstream paired selection colors. No new API or system appearance change.

Both Prism suggestions from the first pass were applied: preserve page-provided
fields and use direct explicit ink rather than repeated full style resolution.
The final stale-record concern is disproven by `syncRecordVisible`: it assigns
`r.node = n` before syncing controls and updates on Style changes. Native selection
changes also update widget state. The mounted race regression verifies explicit
ink application and withdrawal, with selection colors retained; controlled actual
keyboard tab navigation and both-theme gallery composition regressions pass.
The native app is rebuilt from the tested source; live appearance observations
are recorded in the local delivery receipt.

## Mac toolbar catalogue

Three size recipes and three display modes use existing public Toolbar/Button
nodes. Mixed, vertical and narrow examples share owner-held count/status.
Focused gallery race tests, public Manual geometry/identity/overflow routing,
gallery composition and vet pass; native packaging uses the tested build.

Applied the initial stable-item suggestion: mixed presentation now matches IDs,
not array positions. The final zero-MaxSize question is answered by ui.Size's
explicit contract: a zero maximum dimension is unbounded. Both-theme mounted
geometry and overflowing focus tests pass with that value. Narrow applies only
to horizontal samples, as documented; the vertical bar deliberately keeps its
72x154 viewport. Mixed and vertical bars intentionally retain their own display
choices, while the top selector controls the three size comparisons. Delete is
deliberately disabled, documented in the page, guide and spec, and skipped by
the actual keyboard regression. No remaining demonstrated correctness issue.

The native toolbar catalogue follow-up places mixed and vertical examples beside
one another and reduces page gaps. Both-theme public Manual geometry confirms
the vertical toolbar fits the standard 660x780 scene. Focused race tests and
rebuild pass; Prism viewport review reports zero findings. Native screenshots
are inspected again after packaging.

## Toolbar lock-key navigation

A native Down-arrow check reproduced focus entering disabled Delete. The desktop
backend explicitly sets NumLock on every macOS key; toolbar navigation rejected
all nonzero modifiers, allowing generic Frame navigation to handle these keys.
The regression failed before the fix for both orientations with NumLock, CapsLock
and both. Masking only lock bits restores enabled traversal, Home/End and wrapping
while preserving actual chord modifiers and existing child-editor precedence.
Focused chrome race/routing tests, gallery tests, driver/gallery vet and build pass.

Prism's shift-type concern is disproven by the pinned key.Modifiers definition
(events/key/modifiers.go): CapsLock and NumLock are iota bit positions on its
int64 bitflag enum. SetFlag in the macOS desktop path uses those positions. Go
contextually types the constant mask for e.Modifiers; compilation and the six
failing-before/passing-after cases verify it. The same convention is already
used by other managed control key handlers. No extra state or mask API needed.
Native post-package traversal is recorded separately in the local receipt.

## Full-color toolbar icons

Public IconImage accepts immutable full-color pixels on Button/MenuButton. A
private managedImage uses the existing contain/alpha painter inside the normal
font-relative icon slot. Six original 96px sample assets are generated once and
shared; the gallery starts with color and can compare monochrome symbols.

Prism's ordering suggestions were applied using the supported FirstMaker hook,
which precedes native label/indicator makers without mutating the plan slice.
Retired Maker/Updater guards were added and exercised after disposal. The final
MenuButton uncertainty is disproven by the shared mountControl core.Button path
and the actual false-menu-button/true-menu-button pixel/identity/lifetime tests.
These verify the leading private slot, full source colors despite magenta ink,
replacement, disabled alpha, semantic identity, symbol restoration and cleanup.

The final disposal-concurrency concern does not match this adapter's ownership:
reconciliation, native widget updating and disposeControl are serialized by the
existing owner path. Disposal clears r.node.IconImage and private source, just as
the existing Image cleanup clears r.node.Image; no update interleaves before the
retired flag is set. Later private updates reject retirement and cannot revive
pixels. Avoiding descriptor clearing would retain the image through a disposed
record. Race regressions and explicit UpdateWidget-after-dispose pass.

Portable UI race tests, focused native image/icon/menu/toolbar routing and pixel
checks, both-theme gallery/public Manual checks, independent public SDK consumer,
vet and native build pass. Live packaged appearance observations are separate
in the local delivery receipt; no physical-pointer or human VoiceOver claim.

## Icon-above-text toolbar layout

Public Button.IconPlacement selects leading/default or centered IconAboveText
content with Symbol or immutable full-color pixels. It retains one native action,
caption, focus target and semantic action. Gallery Small/Regular/Large stacked
heights are 52/60/72 pt. Other toolbar routing and lifecycle remain unchanged.

Prism's gap-ordering concern is disproven by mountShallowScoped: managed applyStyle
is installed after mountControl's content styler. The added mounted regression
sets Gap to explicit zero, checks that it overrides the stacked default, then
omits it and checks restoration to four. Local padding is checked too.
The native spacer dependency is explicitly documented: the pinned Button adds
its named `space` part for inline symbol/text. Regression requires that exact
part on inline restoration and its absence while stacked; a dependency change
therefore fails the native contract test. FinalMaker changes the declared private
plan only, never reparenting managed actions. Text-only stacked controls are
rejected by portable validation, so the concern about stacking those is inapplicable.

Focused race tests cover both themes, both icon sources and all three sizes:
nonoverlapping centered bounds inside the button, full-color completed pixels,
icon/caption pointer activation once each, Return once with macOS NumLock,
retained caption/button identity, native icon sizing and leading/default reset.
Gallery public Manual checks cover above-mode narrow overflow, End/Home, disabled
skipping and no disabled activation. Portable UI race suite, public SDK test,
vet and native build pass. Packaged native observations remain in the local
delivery receipt, separate from physical-pointer or human VoiceOver qualification.


## Width-responsive toolbar overflow

WithToolbarOverflow adds an ordinary managed MenuButton to a flat horizontal
action toolbar. Native layout retains all action controls, fits a prefix with
room for the trailing chevron, and excludes hidden originals from input/focus/AX.
The menu uses stable action IDs, original callbacks, labels, immutable full-color
artwork or symbols, and current disabled state. Vertical/editor toolbars retain
scrolling. The gallery provides an Overflow menu comparison switch.

Prism reviewed the complete changes three times with Anthropic Claude Sonnet 5.5
(18 files, complete coverage, no high findings). Concrete feedback addressed:
the concrete native frame is returned directly; focus handoff uses the pinned
WidgetBase.Defer API after Scene update/render, with current visibility/focus
checks; hidden actions are matched by ID with a shape guard. Filtered menu
availability drives both focus and AX. The final owner-binding concern was
addressed by binding the More button before initial layout, so its menu is empty
and unavailable until measured visibility exists. Disposal clears owner/hidden
properties. A direct pre-layout regression and native lifetime checks cover this.

Other speculative findings are closed by the existing ownership and allocation
contract: reconcile allocates every child before attach/order/sync and layout;
failed prepare returns without publishing partial records. No application owner
mutation can occur between synchronous SizeUp and SizeDown. Clearing visibility
properties halfway through measurement would expose a transient, incorrect tree.
Menu rows are freshly tree.New-allocated for every popup, never reused; regression
checks different button identities and empty indicators after checked-to-unchecked
reopening for both color and symbol artwork. Plain unchecked menus retain their
original empty icon field (nativeSymbol(empty) is the nonempty sentinel `none`).

Focused race checks cover both appearances, completed full-color menu pixels,
icon/label spacing, exactly-once pointer activation, disabled items and a suffix
containing only disabled actions, macOS lock-modifier keyboard navigation,
width changes with retained controls, stale actions, popup cancellation on resize,
and removal clearing image sources. Existing menu/context-menu/check routing and
scrolling toolbar tests pass. Gallery public Manual tests compare both modes and
both appearances; portable UI, independent SDK consumer, vet and native build
pass. Packaged native observations are recorded separately in the local delivery
receipt, without claiming human VoiceOver or physical-pointer qualification.


## Reordering and hiding toolbar actions

ToolbarLayout.Apply filters/reorders a complete flat horizontal catalogue before
overflow. ToolbarCustomizer is an owner-held controlled content sheet built from
existing Dialog/Sheet, scrolling, image, checkbox and button APIs. It retains
full-color artwork and action IDs/callbacks; customization never invokes commands.
The gallery supplies separate Small/Regular/Large/Mixed values with external
openers, live edits, Restore Defaults, Done/Escape and session-only state.

Prism reviewed all nine changed implementation/test/guide/spec files with
Anthropic Claude Sonnet 5.5 twice, complete coverage and no high findings. The
initial target-capture finding was addressed: the gallery saves through a closure
bound to the View target, checking current target/catalogue/Open before assigning.
A regression invokes an old checkbox precisely after a new opener changes target
and before its next View, then again after replacement View; neither toolbar is
modified. No queued target transition can cross-save configuration.

Final findings have source-backed dispositions. customizeTarget is private, set
only by the four internally generated opener closures or the Regular constructor
default. All four targets are inserted in catalogues on every View. It has no
external/persisted input and vertical toolbars have no opener. An unknown target
would represent internal state corruption; the existing visible error fails
closed rather than silently accepting a different toolbar. Every target is
exercised by the independent-per-bar regression, including all-hidden/restore.

The claimed spurious Hidden reordering does not occur: emit resolves the accepted
value before editing and resolves the candidate afterward. resolveToolbarLayout
serializes Hidden in resolved Order, then slices.Equal compares the normalized
lists. The portable regression hides two items and unchecks the first again;
the Change count remains unchanged. Rejected proposals, queued accepted edits,
invalid replacements, close/reopen, disabled sessions and removed IDs are covered.

Focused portable UI race tests and public SDK consumer checks pass. Native Manual
pointer routing covers both themes, retained action and artwork identity on reorder,
removal from toolbar/overflow on hiding, all-hidden, Restore Defaults, modal
focus/AX scope, stale input, disabling and window-close cleanup. Gallery tests
cover per-toolbar independence and existing sizes/modes/overflow behavior. Vet and
native build pass. Packaged app observations are recorded in the local delivery
receipt separately from automated routing. No drag customization, disk settings,
NSToolbar system palette, physical-pointer or human VoiceOver qualification is
claimed by this checkpoint.


### Packaged customization-sheet alignment correction

The first packaged screenshot exposed centered checkbox contents and a narrower
separator row caused by its empty native Text placeholder. Visibility controls
now justify their contents at the leading edge, the separator has the same fixed
artwork slot with a decorative minus symbol, and compact arrow buttons leave all
seven catalogue rows inside the 300 pt scrolling viewport. No editing/state or
native driver policy changed. Added native geometry assertions for both themes:
checkbox indicators start at their content edge, every arrow column has the same
X position, and all seven rows fit without clipping. Prism reviewed the two-file
correction with Anthropic Claude Sonnet 5.5, complete coverage, zero findings.
Focused layout/routing tests, gallery customization tests and rebuilt native app
pass. Final packaged observations remain separate in the local delivery receipt.

## Toolbar customization dragging

Prism with Anthropic claude-sonnet-5-5 reviewed all 11 changed files with complete
coverage. Its sole medium finding proposes subtracting the modal origin from
`e.Pos()`. No coordinate change was made: pinned core `routeCapturedPointer`
already calls `e.SetLocalOff(scene.SceneGeom.Pos)` before invoking the capture
handler, so `e.Pos()` is scene-local and matches row/viewport bounds. Only the
main-stage insertion sprite adds the origin. Subtracting twice would break modal
hit testing. A code comment now states this boundary, and the routing fixture
asserts a nonzero modal origin at both tested display scales.

Focused UI race tests, public SDK toolbar contracts, native customization routing
in light/dark at 160/320 logical DPI, gallery toolbar regressions, package vet and
native build passed. Routing covers hidden-action/full-color retention, threshold,
exactly-once commit/duplicate release, rejected/no-op/outside drops, release-click
suppression, Escape, wheel scroll with a clipped source, capture loss, order/source
replacement, sheet disable/close and window teardown. Existing arrow/visibility
regressions remain in place. Packaged native pointer observations are separate
from deterministic routing; no broader performance or release qualification is
claimed. No remaining actionable review finding.

## Toolbar preferences

Anthropic claude-sonnet-5-5 reviewed the full persistence checkpoint. Applied the
file-preservation finding: every failed load except a missing file disables writes
for that session, retaining the original invalid/newer file. Focused regressions
confirm this for malformed, oversized and unsupported-version files. Round trips,
independent bars, catalogue migration, Restore Defaults, save failure retaining
memory and temporary cleanup pass with the race detector; gallery toolbar checks
and vet also passed. The second finding calls CreateTemp's 0600 permission an
accident; it is the Go API's specified mode, also asserted by the file-mode test.
Deferred close/remove after rename are harmless cleanup. Atomic replacement is
claimed; power-loss durability of the containing directory is not. No remaining
actionable review finding.

## Fixed and flexible toolbar spaces

Prism Anthropic claude-sonnet-5-5 reviewed the full spacer diff. Replaced the
flagged nested slice insertion with insertion while constructing the catalogue,
and removed the second overflow-filter map. Fixed-width override and expanding/
contracting flexible-space behavior are now measured by native geometry tests;
node styles run after the default Styler, and native Max.X zero is the existing
unbounded convention. Spaces never enter the menu-eligible ID map; ordinary menu
options retain explicit catalogue alignment. UI race tests, native sizing,
overflow/customization routing, gallery toolbar regressions, independent SDK
contracts and focused vet passed. No remaining actionable finding.

## Toolbar context menu and display preferences

Prism Anthropic claude-sonnet-5-5 reviewed the full checkpoint. The reported
partial-load concern had no later failing check or intermediate owner observation;
nonetheless layouts, overrides and common display now publish together. Applied
the valid legacy-reload concern: missing/invalid display resets to the default
rather than retaining an earlier loaded mode, with regression coverage. Display
buttons also ignore stale modal callbacks and avoid unchanged saves. Native routed
context tests cover action/space forwarding, child-menu priority, lock-key-safe
Shift-F10, exactly-once menu selection, focused-child restoration, disable/stale
callbacks and shutdown. Gallery per-bar modes/customization and preference round
trips, legacy reload and focused vet passed. No remaining actionable finding.

## Toolbar interaction polish

Prism Anthropic claude-sonnet-5-5 reviewed the original and corrected full
checkpoints. Applied explicit terminal hasLine/stage guards in addition to
synchronous animation removal; a late callback regression confirms it cannot
recreate the sprite. Applied the final suggestion to compute immutable paint
coordinates before publishing the closure. Animations skip zero elapsed time so
an explicit clock cannot trigger a redraw loop; native scene time interpolation,
retarget alpha continuity, exactly-once commit and all cancellation tests pass.
Six-dot Grip artwork maps to the pinned bundled icon, with native grab cursor.

Focus cues now combine actual native focus with an atomic current input-scope
publication. Focus events explicitly restyle retained controls even when their
Focused bit was already set. Light/dark tests open/close the panel repeatedly,
check background suppression, stale flag handling and stable geometry. UI race
tests, scaled customization routing, context/space tests, adjacent input activation/
focus regressions, gallery tests, SDK contracts and vet passed. No backend policy
or release gate changed. No remaining actionable Prism finding.

## Native verification followup: held help and gap hits

The packaged held-pointer check exposed a target-row tooltip opening and retiring
capture. The shared list guard now covers every managed widget mount, including
non-control rows/labels, and checks the current scoped pointer token. Prism
Anthropic claude-sonnet-5-5 reviewed both focused checkpoints. Applied explicit
typed-nil-safe parent traversal and moved guard installation to the common mount
path; target grip/checkbox/arrow and non-control row regressions pass. A replaced
or destroyed list cannot leave an admitted capability: source disposal/loss ends
capture and clears the exact old list property; a mismatched property belongs to
a different gesture and is never deleted. The suggested race between assigning
b.active and the property is not possible under originating render ownership;
there is no callback between those assignments. Retain post-admission publication.

Native toolbar gap hits also require a passive Clickable container so right mouse
release generates ContextMenu. This adds no action/focus stop; the bar's state
layer remains neutral. Regression now routes actual Right input through the owned
core window, with explicit visible geometry; it also preserves Control on the
Right input produced by pinned desktop MouseButtonEvent's Control+Left mapping.
Raw offscreen Source input does not perform that desktop conversion. Child menu
priority, stale retirement, held capture cleanup, repeated focus tests and vet
pass; native build succeeded. No remaining actionable Prism finding.

## Visual corrections: trailing actions and modal marker pixels

Visual inspection caught Mixed's trailing actions displaced by flex space. The
shrinkable overflow frame advertised a small bottom-up minimum and incorrectly
used it as content width during growth. SizeDownSetAllocs now supplies the sum of
visible intrinsic widths plus gaps before native growth. Spacing uses the
backend's nonpainting Space/Stretch leaves. Tests assert trailing geometry and
sample the original full-color icon pixels, and the public gallery checks every
Mixed action with overflow both off and on.

A composed-pixel check also ensures the insertion marker is on the dialog's own
main stage rather than the window stage below it. Prism Anthropic claude-sonnet-5-5
reviewed all nine changed files; applied explicit Window/Dialog versus named popup
stage selection instead of an enum-order comparison. Both-scale/light/dark marker
pixels, capture/animation cleanup, responsive overflow, space/context regressions,
gallery geometry, vet and native build passed. No backend rendering or release
gate changes. No remaining actionable review finding.
