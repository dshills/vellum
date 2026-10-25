# Mac preview and editing widgets

## Quick Look

`Application.QuickLook()` is optional. On a worker call `PreviewFiles(ctx, window,
QuickLookOptions{Items, SelectedID})`. Items contain stable ID, absolute local Path
and Title, 1–16 regular files. The service checks paths outside the UI owner; the
OS reads previews. Keep files available until the call completes. Escape/Space or
close returns nil. Caller/window/application cancellation revokes the session.
Unavailable platforms return `ErrUnavailable`. Foreign preexisting Quick Look
panels refuse rather than taking over another responder's state. One native
auxiliary session is active at a time; Colors/Fonts and Quick Look don't overlap.
The actual AppKit panel provides its system navigation and supported formats.
Full-screen preview and external preview-provider behavior are OS-owned; no
cross-window handoff, editor, file chooser or file-writing API is added.

Set `DocumentList.Preview` or `OutlineView.Preview` for a Space-key proposal.
Selection and Return activation retain their existing behavior. Call the service
on a worker, using a reserved window completion for result publication, as the
gallery does. Quick Look page provides generated text and original-color PNG
samples through both a document list and outline; temporary files are removed
when the session completes. Native/physical verification is separate from
request, stale-event, parser and lifetime regression tests.

## Find / Replace

`ui.FindBar` retains controlled Query/Replacement/Open/options and takes an existing
`editor.Model` for counts. Change edits one field; Request supplies `FindRequest`.
Call `request.Apply(model)` on the owner, or handle the request in your search
engine. FindNext/Previous wrap; FindReplace replaces a selected match or selects
the next; FindReplaceAll is one undoable journal edit and maps the existing selection through replacements. IME composition refuses.
Search is literal Unicode simple folding with scalar offsets and optional word
boundaries (letters/digits/underscore), bounded to 4 MiB documents, 1024 query
scalars, 10000 matches and 64 KiB replacement. Excessive queries refuse instead of
partially replacing. Matches are nonoverlapping and cached by model identity, content revision, query and options; selection-only navigation reuses them. No regex, language-specific expansion or all-match highlighting.
Use Small for 24pt or regular 30pt controls. Host command routing owns Cmd-F/G and
Escape; the demo shows focused query entry, editor focus return and Undo/Redo.
