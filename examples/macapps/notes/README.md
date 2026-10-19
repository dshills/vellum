# Notes

A native Go example consuming only Vellum's public `ui`, `editor`, `command`
and `history` APIs. It has a folder sidebar, note list, title field and plain-text
editor. Folder/note selection, title/body search, new/edit/delete, pinning and
Undo/Redo work against fictional in-memory sample notes. No Apple Notes data,
accounts, cloud sync, files, clipboard fixtures or network services are accessed.
The displayed dates are fixed sample dates, not the computer's clock.

`New`, `Model.View`, `Theme` and `Model.Commands` are the consumer interface.
Use `app.ChromeUnified` on the native window; the view reserves the measured
52 Dp native toolbar before its first interactive content. Install the returned
registry on the window. Its window-scope Meta+Z and
Meta+Shift+Z shortcuts intentionally handle the unified app history before the
focused text control's own fallback. Do not replace them with a second Undo
route. Native hosts must also install `SetSettlement` and route app actions
through `Model.Request`. The included executable owns its app/window and chains
public `cogent.SettleTextField` for the always-mounted search field, then for the
current title, then `cogent.SettleEditor` for the current body, using that window's
lifetime. Search settles even when no note is selected; the title/body IDs are
captured after its Change is admitted. All acknowledgments run on
the UI owner; the host invalidates after completion and never waits on Cocoa
main. `CurrentInput` supplies the exact stable title ID and public editor model.

The request gate retains at most one action, disables other structural controls,
and reports "Finishing text input". Native title/body Change callbacks remain
admissible during that interval. Completion executes the requested action once;
admission/cancellation/settlement errors execute no action and release the gate.
Duplicate completions are ignored. Undo enters settlement even with no admitted
history, because a native title edit may still be waiting to create its entry.
With no boundary installed, `Request` and the model's direct mutators retain
portable synchronous behavior. Direct mutators are useful for model consumers;
native UI consumers must use the request path before structural/history actions.

Body changes are admitted once from the public editor's committed Change
callback; toolbar actions also flush earlier committed model changes after
settlement. Replay uses public editor replacements and reveals the affected
note. Creating, editing and deleting a note are separate history steps. Redo
restores the same stable note identity and retained editor. Body-history replay
places the caret at the beginning; title-only edits preserve the body selection.

The model belongs to one application UI owner. It retains at most 64 live notes,
32 KiB per note body, 256 UTF-8 bytes per title/search and 32 app history entries.
Each note's public editor retains one internal text history entry because the
app registry owns unified Undo/Redo. Detached note snapshots cannot mutate the
model. A composition that remains active after settlement refuses selection,
new/delete and history actions, preserving preedit. A committed title can still
be admitted while the independent body is composing. The portable direct model
path refuses destructive actions until composition is committed or canceled.

The visual sources are Apple's [macOS Tahoe 26 Notes getting-started guide](https://support.apple.com/guide/notes/get-started-apd441e46563/4.13/mac/26)
and [folder guide](https://support.apple.com/guide/notes/add-and-remove-folders-apd558a85438/4.13/mac/26).
The selected Tahoe version and the real Apple guide screenshots were inspected
on 19 October 2026. Reference captures remain in the development artifact area;
no Apple screenshots, personal Notes contents, icons or product imagery are
redistributed by this example. The app uses Vellum's bundled symbols.

The source's light three-column hierarchy, amber folder selection, grouped note
list, fine dividers and bold editable title are represented. Fidelity gaps are
explicit: plain text has a separate title control; rich inline formatting,
attachments/thumbnails, drawing, tables, math, collaboration, account controls,
native translucent materials and Apple's exact toolbar icons are omitted. The
note title button selects a list card; its preview/date region is display-only.
Folders are fixed fixtures and cannot be created, moved or deleted here. Dates
and list order are fixed aside from pinning and adding notes.

Regression tests cover ordered create/edit/delete replay, identity preservation,
native editor callback deduplication, committed-input ordering, IME refusal,
folder/search/pin behavior, budgets, app command routing, detached snapshots and
accessible validated trees including empty states. Fake asynchronous boundary
tests cover pending guards, duplicate completion, errors, every structural UI
route, pending search admission before a query-clearing structural action, input
admission before Undo and the empty-history case. These tests do
not prove physical typing, native IME behavior or visual fidelity. A first live
run exposed an unadmitted title paste before window Undo; the explicit native
settlement wiring addresses that ordering. A renewed native check verified
pending title admission before Undo and the matching Redo; a separate native
accessibility body edit also survived app Undo/Redo with the same note identity.
Physical body-click focus remains unqualified. The editor's monospace body font
and line-number gutter also remain a material visual mismatch without a public
editor-presentation API. Native rendering and pointer verification are
coordinated separately by the main development agent; this package alone makes
no pixel-identical or native-acceptance claim.
