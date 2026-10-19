# Reminders example

A native Go consumer of Vellum's public `ui`, `command`, and `history` packages.
The light sidebar, smart-list tiles, list colors, grocery sections, completion
circles, and spacious list layout follow Apple's
[Reminders guide for macOS Tahoe 26](https://support.apple.com/guide/reminders/get-started-remne4b02adc/7.0/mac/26).
No Apple artwork, application data, or private framework types are embedded.

`New()` constructs the UI-owner model; `View()` returns a fresh `ui.Node` tree;
`Theme()` returns its app-local appearance. Pass `Commands()` to the native host
to enable Command-Z and Shift-Command-Z after the focused editor declines them.
Recommended starting content size is 1000 × 740 device-independent pixels.

The native runner installs `SetInputBoundary(window.Context(), boundary)` with
the public `cogent.SettleTextField` API. Structural UI actions settle search and
composer, then the title/notes/date fields if a details dialog is open, one FIFO
barrier at a time. Controlled Change callbacks remain admitted during that
sequence; only one structural action is retained. Add and Done validate after
settlement, so their enabled state does not rely on an older model snapshot.
Cancel settles before discarding its draft. Document Undo/Redo settle first.
Undo admits an open valid details draft as one history change before replay, so
it removes that edit instead of an earlier reminder change. Redo refuses a
changed or invalid open draft without changing the collection, history branch,
or draft; choose Done or Cancel before retrying. An unchanged draft permits
existing Redo. Focused text editors
retain first refusal of their normal keyboard Undo/Redo shortcuts.

Consumers calling synchronous `Add`, `Select`, `SetCompleted`, `SetFlagged`, or
`Update` directly supply already-committed model values. A native consumer wraps
structural actions in `AfterInput(func() error { ... })` and reads field values
inside that callback. `Undo`/`Redo` already use this boundary. Boundary admission
returns immediately; deferred failures appear in the view status. The boundary
must publish Change before completing, invoke completion once on the UI owner,
and never complete an immediate refusal. Call `Close` after host dispatch stops
to release the pending action; canceled window lifetimes and late callbacks
cannot mutate the closed model. Reminders contains no multiline editor, so it
needs no `SettleEditor` binding.

## Working interactions

- Select lists or Today, Scheduled, All, Flagged, Completed, and Assigned filters.
- Search titles and notes within the selected list. The search is case insensitive.
- Type a title and choose Add Reminder; New Reminder focuses that input.
- Complete or reopen a reminder with the circle action. Show/Hide includes or
  excludes completed reminders; the Completed filter always shows them.
- Choose a reminder title or information action to edit its title, notes, due
  date, priority, and flag. Done creates one atomic history change; Cancel and
  native dialog dismissal discard the draft.
- Undo/Redo creation, completion/reopening, flag changes, and detail edits. Redo
  restores the same stable object identity. All actions are keyboard focusable;
  completion actions expose their state and proposed action in accessible names.

The demonstration holds at most 256 reminders and 64 history changes. Titles
are limited to 512 UTF-8 bytes, notes to 4096 bytes, and searches to 512 bytes.
Failed validation leaves the collection and history unchanged. History captures
individual before/after reminders instead of whole collection snapshots.

## Deliberate limits

This is an independent demonstration, not Apple's Reminders application. All
data is fictional and in memory; closing the example discards it. Today is fixed
at 2026-10-19 so tests and screenshots are reproducible. There is no iCloud,
EventKit, notification scheduling, location access, persistence, collaboration,
recurrence, attachments, list creation, or natural-language date parsing. Due
dates use an explicit `YYYY-MM-DD` field and notes use a single-line editor.

Vellum's bundled symbols are used in place of Apple's SF Symbols. Smart-list
tile labels provide the hit target; their decorative count/icon area does not.
Completion controls are accessible buttons with state in their names rather
than platform checkbox controls. Exact Tahoe glass materials, native sidebar
vibrancy, and Apple's app-specific layout behavior are outside this example.

The unified-chrome demo reserves 52 device-independent pixels for the native
header measured on macOS Tahoe 26.6.2. Sidebar controls and content actions start
below that band while their backgrounds continue across the full window
surface. The reserve belongs to this tested demo configuration; it is not an
adaptive framework safe-area API or a guarantee for other chrome, toolbar,
accessibility, or OS configurations.

## Verification

From the parent `examples/macapps` module, run `go test ./reminders`.
Regression coverage exercises stable-identity Undo/Redo, no-op/failed edits,
capacity and history bounds, detached snapshots, smart-list/search behavior,
controlled UI callbacks, modal draft cancellation, invalid dates, keyboard
history precedence, accessible completion states, and descriptor/theme validity.
Asynchronous fake-boundary tests cover settlement ordering, rejected and duplicate
actions, field validation, canceled lifetimes, and late completion after closure.
These are automated model/routing checks. Physical pointer, native rendering,
and platform accessibility verification must be reported separately by the
demo-suite host.
