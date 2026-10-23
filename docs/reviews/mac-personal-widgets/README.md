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
