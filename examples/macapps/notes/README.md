# Notes

A native Go example consuming Vellum's public `ui`, `editor`, `command`,
`history`, `document` and application lifecycle APIs. It has a folder sidebar, note list, title field and plain-text
editor. Folder/note selection, title/body search, new/edit/delete, pinning and
Undo/Redo start from fictional sample notes and can be saved as a local notebook. No Apple Notes data,
accounts, cloud sync, clipboard fixtures or network services are accessed. Only this
app's private notebook/recovery files and explicitly selected notebook files
are read or written.
The portable in-memory view retains fixed sample dates. The durable executable
uses “Local note” labels instead of presenting fictional timestamps as real.

`New`, `Model.View`, `Theme` and `Model.Commands` are the consumer interface.
Use `app.ChromeStandard` on the native window so macOS owns the draggable title
bar and window buttons. The view begins directly below native chrome. Install the returned
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

The source's light three-column hierarchy, subtle folder selection with amber text, pale amber note selection, grouped note
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
Physical body-click focus and proportional selection passed in the later v1 RC; see the [source-bound acceptance](../../../docs/reports/macos-v1/README.md).
The writing area now opts into `Prose: true` and `SoftWrap: true` on the same
public multiline editor, with proportional typography and no line-number gutter.
Native rendering and pointer verification are
coordinated separately by the main development agent; this package alone makes
no pixel-identical or native-acceptance claim.

## Durable executable

Run `go run ./cmd/notes` from `examples/macapps`. The default private directory is
`~/Library/Application Support/Vellum Notes` on macOS. `-data-dir /absolute/private/path`
selects a separate notebook/recovery slot for testing. The directory must be
private (0700); one process holds its permanent lock inode. Invalid files or
unavailable storage stop startup without overwriting those files.

Command-S saves to `Notebook.vellumnotes` initially; Command-Shift-S chooses a new
file destination; Command-O opens a notebook. Matching accessible toolbar buttons
use the same commands. Save As switches the current destination; a later launch
starts with the default local notebook, so use Open to return to an external file.
Saving a new fictional sample notebook is explicit; a clean sample journal is
not labeled as an already saved disk file. The whole notebook includes stable
note IDs, next ID, folders/selection and exact plain UTF-8 body bytes. Undo/Redo
remains available after Save and tracks the saved revision; history is not
serialized across launches. Open replaces the complete notebook and clears
session history only after validated decoding and stale-revision checks.

Input settles through the existing search/title/body boundary before file or
close actions. Unsaved Open/Close offers Save, Discard Changes or Cancel.
Failed/canceled saves keep the notebook dirty and do not close it. Atomic writes
use `document.FileStore`; an error after replacement is reported and never
misrepresented as acknowledged success. New note/delete/pin/content history
continues to use the same application journal.

Every two seconds, an idle owner may capture a changed committed notebook to a
separate recovery session. Native composition, pending input, file IO and a
recovery decision postpone capture. Recovery never marks the main notebook saved.
After interruption, startup explicitly offers Restore, Discard or Close without
changing files. Restored content is dirty with fresh session history and saves
to the default local notebook. Successful save/open or explicit discard writes
an atomic clean marker. Recovery is the last completed snapshot, not a guarantee
that the newest keystroke survives force quit, power loss or blocked IO.

Regression tests cover exact Unicode/CRLF persistence, identity and Undo/Redo,
malformed/duplicate/oversized documents, atomic replacement and failures,
stale-save refusal, native input settlement, composition-preserving recovery,
restore/discard/close and owned worker retirement. Native UI observations and
physical pointer evidence are recorded separately. No human IME or VoiceOver
qualification is implied.

If the main save succeeds but writing the recovery clean marker fails, saved
bytes remain intact and the app shows a Recovery error while keeping the window
open. Retry close after storage is repaired. This avoids silently acknowledging
a recovery-file operation that failed. Scheduler and file-worker shutdown have
a three-second retirement bound; a failure retains the repository lock until
process exit.

## Visual correction from the owner comparison

The app uses shared Mac control metrics, compact folder rows with a trailing count
column, a 28 Dp search field, denser note cards, one-pixel dividers and a smaller
editable heading. The former blank 52 Dp client header is removed with the switch
to standard native chrome. Saved-path detail is available in the status
accessibility help rather than filling the writing area's footer. These layout changes are
application-owned. The separately approved prose option extends the existing SDK
control presentation; the document and history paths are unchanged.
The approved prose presentation uses the retained multiline editor with
proportional text and no gutter. The font stack explicitly includes emoji fallback for titles, rows and prose.
A frameless body uses caret/selection focus feedback. The later app-restricted
native helper verified physical body-click focus and proportional selection; see
the source-bound v1 RC report. The offscreen layout test
checks actual geometry at 1100×740 and 1380×900; it does not prove native fidelity.

## Native v1 candidate shell

The executable now uses the per-window `Driver.TitlebarToolbars` factory with
original-color sample icons. New/Open/Save/Undo/Redo/Delete/Pin invoke the same
registry and input-settlement path as keyboard shortcuts and the native File,
Edit and Note menus. Availability refreshes after invalidation. The portable
model view keeps its client toolbar unless its host calls `SetNativeToolbar(true)`.
No duplicate client toolbar is mounted in the native executable. Customize the
native toolbar using its standard context menu; layout persistence is AppKit-owned.
The body, file format and history limits are unchanged.

## macOS v1 delivery

Current v1 packages pair `Mac examples source.zip` with
`vellum-v1.0.0-private-sdk.tar.gz`. Follow the [independent build guide](../../../docs/building-macos-apps.md)
and [v1 support/API scope](../../../docs/macos-v1.md). The app's native toolbar
uses full-color original artwork and the same command/history/file settlement
paths. Native Open cancellation returns keyboard input to the live parent.
Source compatibility applies to the framework's named stable APIs; this example
model is application code, outside the framework compatibility promise. Human
IME/VoiceOver remain unqualified; final package results identify their exact source.

## Find in the current note

Command-F opens Find; Option-Command-F also opens replacement. Command-G and
Shift-Command-G navigate matches and wrap. Escape/Done closes the bar and
returns focus to prose. Find searches the current note body (the sidebar search
still filters titles and bodies across the current folder). Literal Unicode
matching supports match case and whole words. Query/replacement are bounded,
session-local controlled fields; no regex or cross-note replacement is provided.

Replace and Replace All use the existing editor and notebook history. Replace
All is one document Undo/Redo step. Host settlement admits search, find query,
replacement, title and body before an action; native composition must settle.
Recovery/dirty dialogs retain priority, and rejected/stale operations preserve
contents. The generated Document Starter and A/B Documents share this recipe.
