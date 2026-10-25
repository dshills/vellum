# Public editor model

Status: implemented portable model. `editor` owns document bytes, directional
selection, composition, sparse undo/redo, and monotonic input epochs. The native
adapter binds `ui.MultilineText` to a model, converts platform offsets to Unicode
scalar offsets, and routes committed edits through this model. Live native
behavior is verified separately from portable model tests.

## Ownership and admission

Create a model with `editor.New(text, editor.Config{})`. Zero configuration uses
a 4 MiB document byte budget and 128 retained history changes. Negative bounds
are refused. The model and its sole `history.Journal` belong to one application
UI owner; worker IO uses detached snapshots. The current driver admits one live
native binding per model. Native undo is disabled so there is one document history.

A model must not be copied by value after creation: its journal callbacks belong
to the original model. Pass the model pointer, use detached `Snapshot` values, or
call `New` for an independent document. A copied model has no usable authority.

`Validate` safely refuses nil, zero, and copied models. `Snapshot` returns a zero snapshot
for those models; `Journal` returns nil and `MaxBytes` returns zero. Descriptors
must validate before mounting. `MaxBytes` exposes initialized document budgets
so native admission can refuse an over-budget mutation before platform editing.

Text remains exact UTF-8 bytes. CRLF, combining scalars, and final empty lines are
preserved without normalization. NUL, invalid UTF-8, and bytes above `MaxBytes`
are refused. `Range{Start,End}` uses a half-open document-wide Unicode scalar
range. `Selection{Anchor,Focus}` preserves direction and allows an empty caret.
CR and LF count separately; a Unicode scalar is not a grapheme cluster.

## Input transactions and clocks

`Snapshot` contains committed `Text`, preview-inclusive `VisibleText`, visible
`Selection`, `Composing`, `Marked`, and a detached `history.State` value. Three
monotonic clocks start at one:

| Field | Advances when |
| --- | --- |
| `Revision` | Committed content changes, undo, redo, or whole-document install |
| `VisibleEpoch` | Visible bytes change, including preview and cancellation |
| `SelectionEpoch` | The visible directional selection changes |

The journal's revision is a saved-state identity and can revisit an undo state.
These editor clocks never revisit an earlier value, preventing ABA acceptance
after undo, preview/cancel, or installing identical document bytes. Exhausting a
required clock refuses the complete operation without mutation.

`Edit` captures both `BaseRevision` and `BaseVisibleEpoch`, a scalar `Range`, its
exact `Before` bytes, replacement `Insert`, and the resulting `Selection`.
`Apply` validates both content clocks and the removed span before committing one
history change. A selection-only edit preserves undo/redo and dirty identity.
Selection requests use `Select`; they preserve content and history. Native
adapters use `SelectionEpoch` to recognize external selection requests while
preserving platform caret affinity and scroll for ordinary snapshot echoes.

`ErrStale`, `ErrRange`, `ErrInvalidText`, `ErrTooLarge`, and
`ErrRevisionExhausted` refusals preserve bytes, selection, composition, clocks,
and both history stacks. Invalid or busy history is refused through the
corresponding `history` error.

## Composition and sparse history

`Preview(Edit)` replaces a span of the current `VisibleText` and validates the
resulting `Marked` range. Both selection endpoints must lie within that marked
range, including when calling `Select` during composition; settle composition
before moving the selection outside it. Committed bytes and journal state remain unchanged.
Each preview captures the latest visible epoch; the committed revision remains
the baseline revision until commit. One bounded preview document is retained
alongside committed bytes; older previews are released.

`CommitComposition` commits the final preview once. Unmarking may leave visible
bytes unchanged while advancing committed revision and adding one history item.
A no-net composition clears preedit without adding history or changing committed
revision. Its final visible selection becomes the committed selection.
`CancelComposition` restores exact baseline bytes and directional selection with
no history change; visible/selection epochs advance if their values change.
Commit/cancel without active composition is a no-op. Committed edits and replay
refuse active composition with `ErrComposition`; callers must settle it first.

History closures retain cloned removed/inserted spans and before/after selections,
not the complete document for routine sparse edits. Composition commits and
`Replace(text, selection)` retain the smallest scalar-aligned contiguous changed
span. Entirely different replacements can still retain the full before/after
documents. The history bound counts entries, not bytes; worst-case full document
replacement history scales with both configured bounds. No byte-level memory
ceiling is claimed beyond the document bound and entry bound.

## Persistence

`Journal()` is supplied directly to `document.Session`. The application codec
captures committed `Snapshot().Text`, encodes exact bytes, and decodes without
newline normalization. It must settle native composition on the originating
window before capture, save, undo, redo, or close that requires committed input.
The native settle operation is asynchronous and acknowledges the originating
control's FIFO boundary; it is not a synchronous platform-thread wait.

The codec's install callback calls `Install(text, selection)` inside
`Journal.Replace`, as `document.Session.Open` already does. `Install` validates
the complete replacement before assignment, requires an executing journal
callback, and does not reenter the journal. Calling it without an executing
journal callback is refused. The caller is responsible for using it only for replace;
the journal's busy signal does not distinguish arbitrary callback kinds.
Successful installation advances editor revision even for identical bytes;
the session replaces history and saved identity exactly once. An installation
failure preserves both model and history. Active composition must first settle.

## Verification

Portable tests cover copied-model refusal without aliased history, exact Unicode/CRLF/trailing bytes, directional selection
replay, sparse ranges, stale content and preview ABA, bounded retention, atomic
invalid/budget/exhaustion refusals, multiple composition previews, unmark without
a visible change, no-net/canceled composition preserving redo, and actual
`document.Session` save/open using the same journal. The persistence test also
checks failed installation leaves document/history unchanged.


## Managed Mac adapter

`ui.Node{Kind: ui.MultilineText, Editor: model}` mounts an exact-text native
renderer backed by a private buffer with native undo and autosave disabled. The
model is the sole document authority. `TabInserts` enables literal Tab input;
otherwise Tab moves managed focus. `SoftWrap` enables visual wrapping without
normalizing or modifying bytes. A model has at most one live native binding,
including across windows; independent documents require independent models.

Cocoa updates reserve both payload credit and a position in the originating
event FIFO before platform mutation. Native admission has 256 slots, a 50 MiB
serialized-message limit and a 64 MiB aggregate serialized-payload budget,
including reservations and in-flight messages. Delta fragments support the
4 MiB native document bound. The native driver refuses a model configured above 4 MiB before mounting,
including an initially empty model; portable models can use larger bounds. A
smaller configured model bound is enforced before platform mutation. Model history, native shadows, geometry, temporary encoding
and allocator overhead are separate memory; these figures are not process-memory
or latency guarantees. Legacy editor changes publish at each accepted event
boundary, including events consumed by widget handlers, so a following Undo or
native action cannot cross an unpublished edit.
Ordinary wake-event compression preserves deferred publication and reserved
input nodes. Only adjacent ready deferred nodes may merge under their bounded
dequeue protocol; priority window events retain their existing precedence.

Initial publication waits for an actual platform-main callback to confirm queue
readiness. One readiness worker per window returns control immediately and
requests a bounded publication retry; it never waits while holding the UI owner.
Startup probes report unavailable without publishing the main queue. A shadow
is installed only after native acceptance.

Switching between a managed single-line field and a document editor first
quiesces the outgoing native client. An outgoing field's focus-loss change waits
for its admitted composition records and native unmark acknowledgment; see the
[controlled field contract](controls.md#numeric-and-text-editing). The adapter
then publishes the current focused client, including when focus returned to the
same field during the handoff.

Native clients carry a fresh mount identity and acknowledged shadow. Committed,
visible and selection clocks reject stale shadow input, including undo/install
ABA. Publication updates the shadow only after native acceptance. Disabled or
retired controls refuse input; remounts cannot inherit old-client messages.
An outgoing disabled editor can still publish inactive native state. An exact
match of binding/content/selection epochs, visible bytes and directional selection
must quiesce without dirtying the owner or changing model/history. Divergent
private text/selection or stale epochs request restoration, then quiesce after
originating reconciliation. Disabled-only selection restoration includes preedit
scalar positions. Authoritative model mark/unmark may temporarily bypass the
private widget's Disabled flag under originating ownership with syncing enabled;
the control admission remains disabled and the flag is restored before return.
This model restoration cannot admit user input or activate the native client.
Native UTF-16 ranges convert to document scalar offsets. Committed external
selection requests that fall inside a shaped grapheme normalize once to a
forward caret or outward selection boundaries, retaining selection direction;
the model records that normalized native selection. Preedit selections retain
exact scalar positions. Portable models do not impose grapheme boundaries. Shaped cells, caret
affinity and clipped viewport geometry support native substring, range and point
queries; missing geometry stays unavailable. In-process Cocoa/NSAccessibility
selectors do not establish external-client, VoiceOver or real IME acceptance.

Call `cogent.SettleEditor(ctx, application, windowID, model, callback)` from the
application owner before capture, persistence, replay or close. It asynchronously
unmarks the originating Cocoa client and acknowledges an ordered FIFO barrier.
Never wait synchronously on the UI owner or Cocoa main queue. One pending settle
request per window is admitted; another returns the bounded-mailbox full error
before launching work. Immediate refusal does not invoke the callback. Accepted
requests invoke it once on the owner while the application is open, including
cancellation, window closure or client retirement errors. Whole-application
closure cancels user dispatch. Settlement covers the named model's single live
binding; it is not a global multiwindow flush.


## Current renderer font family and gallery integration

The managed multiline document body requires its monospace rich-font family.
Fresh body rich/text styles normalize that family and clear the copied custom
font before Lines admission, character sizing and body/tab shaping. General UI
`Typography.Family` overrides style the editor's line-number gutter, while the
document body remains monospaced. Size, weight and slant reach both. Normalizing
the body must not mutate widget UI typography or earlier submitted text styles;
no repeated downstream family correction is needed. Full layout invalidates the
cached tab shape alongside line and gutter shapes, including after retained
font, scale or wrap changes. The ordinary
[gallery](../drivers/cogent/cmd/gallery/README.md) consumes public editor.Model;
full software routing and the independently validated failed native prefix in
[the text/modal result](../docs/reports/macos-text-modal/README.md) cover exact
Unicode edits, grapheme deletion and one Undo/Redo journal. The complete native
text/modal record remains open after actual drawable unavailability. Go FIFO
establishes no live IME, physical input or native pixels.

## Optional plain prose presentation

`ui.MultilineText` accepts `Prose: true` together with `SoftWrap: true`. This
explicit opt-in honors `Style.Typography` for proportional body text and removes
the line-number gutter. The zero value preserves existing code editors. Prose
on another kind, or without wrapping, is refused. This is plain text rather than
rich formatting. Code-specific fixed-column tab layout is not offered in prose.

Wrapping uses measured pixel widths and grapheme-safe shape boundaries. Native
caret/selection geometry and painting use the same proportional glyph advances.
Changing presentation, font or available width must preserve exact document
bytes, directional selection, composition, binding identity and application
history. Resize invalidates old visual-line affinity. Buffer views retain one
immutable wrap-input snapshot; teardown releases it with the view. No widget or
application callbacks run from the buffer wrapping callback. Native physical
input and human IME/VoiceOver acceptance remain separate from routed tests.
