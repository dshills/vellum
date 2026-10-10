# Document history and macOS keyboard precedence

The dependency-free `history` package implements bounded application document
undo/redo. `history.New(limit)` creates a clean journal with a positive change
count limit. A Journal and every callback belong to the application's serialized
UI owner. There are no worker goroutines, implicit document snapshots or locks.

## Changes and retention

`Change{Label, Apply, Revert}` captures immutable before/after domain data.
`Journal.Apply` runs Apply once, records the successful change and discards the
redo branch. Undo runs Revert; Redo runs Apply again. The journal keeps at most
limit entries across both stacks, evicts the oldest applied transitions and
releases discarded callback captures. `Clear` releases both stacks and preserves
dirty status. The count bound does not bound the bytes captured by an application.
Applications with large datasets should capture sparse deltas or immutable roots.
A callback can assign several related properties as one undoable change.

Each callback must either apply its complete model change or return an error
without modifying the model. Validate before assigning. Failed validation,
failed Apply, failed Undo and failed Redo preserve both history stacks. This is
an application callback contract, not a rollback engine for partially mutated
models, IO or external services. An operation refuses reentrant journal changes
with ErrBusy. Callback panics propagate and invalidate the journal, release its
captures and report dirty state. Further replay/save/clear refuses ErrInvalid;
Reset explicitly represents a recovered/replaced model.

## Dirty state and persistence

Snapshot returns immutable labels, counts, current Revision, Dirty, Busy and
Invalid. Revisions identify states independently of undo depth. Returning to the
saved revision through Undo or Redo is clean; editing a new branch at the same
depth remains dirty. Eviction cannot turn a different state clean.

MarkSaved acknowledges successful persistence of the current owner-held model.
For asynchronous saves, capture Snapshot.Revision with the model snapshot and
use MarkSavedRevision on completion. A changed revision returns ErrStale without
clearing dirty state. Reset(saved) represents a replaced document, invalidates
old save completions and releases history; it does not modify the document model.
Dirty false describes journal-tracked state only. Applications must route all
undoable document mutations through their journal and explicitly Reset for
untracked model replacement. Views, selection, docking, pan/zoom and background
calculations need not be document edits.

## Command routing

Command.ShortcutScope defaults to WindowShortcuts, preserving priority before
focused native widgets. AfterFocusShortcuts run only after a focused widget has
had an opportunity to handle the chord. Shortcut uniqueness spans both scopes.
Registry.ShortcutInScope implements native phase dispatch; explicit Invoke and
Shortcut retain normal availability/context checks independently of focus.

The native driver registers first/final Scene key handlers. Owned patch 0020
also delivers those phases when a Scene has no focused widget. Native text fields
and editors consume their own Undo/Redo before a document fallback. Workbench
edit.undo uses Meta+Z and edit.redo uses Meta+Shift+Z with AfterFocusShortcuts;
toolbar and command palette invoke the same available document commands directly.
A focused modal owns its separate Scene and does not leak document fallback to
its underlying window. Native menus retain their separate explicit command IDs.

## Workbench and acceptance

The example keeps 256 sparse component edits. Rotation captures exact before/after
canvas geometry and source data; value edits reject overflow. Undo targets the
captured component even when selection changes. Every accepted edit/Undo/Redo
advances the source revision and clears table projection state, so a background
result from an older model cannot become valid after returning to older content.
Selection, worker snapshots and layout state remain independent. The status
shows unsaved document changes. The [document session](document.md) supplies
asynchronous persistence and unsaved-close handling.

Portable race tests cover branches, clean revisions, eviction, failure admission,
reentrancy, panic invalidation and stale saves. Adapter tests exercise real native
text Undo consumption against the first/final handlers. Model tests cover captured
identity, exact canvas geometry and stale worker projections. The independent SDK
consumer tests the public history and command APIs.

`go run -race ./cmd/history-smoke` from the native module creates an actual macOS
window. It injects ordered native key chords, waits for AppKit focus publication,
and verifies text Undo/Redo, document fallback, dirty status and disabled refusal.
It does not toggle input sources or VoiceOver. This check is not human
screen-reader or live IME acceptance.

`Replace(saved, install)` prevalidates initialization, busy state, nil callback and
revision exhaustion before model installation. Atomic installation failure preserves
both stacks/revision identities. Success releases captures, advances identity and
sets dirty according to saved. Panic invalidates history; explicit replacement can
recover an invalid journal. Install must not reenter the journal.
