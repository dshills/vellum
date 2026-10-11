# Multiline document editor

This independent Go module opens two document windows using the public
`ui.MultilineText` control. Each document has one `editor.Model`, one sole
`model.Journal()` for Undo/Redo, and one `document.Session[string]`. Models,
selection, history and session decisions belong to the application's UI owner;
workers receive detached immutable text for chooser and file IO.

Open, Save, Save As, Undo, Redo and Close first use
`cogent.SettleEditor(windowLifetime, application, windowID, model, callback)`.
The asynchronous acknowledgment passes earlier admitted native input and commits
preedit in its originating native window before the consumer captures bytes or
moves history. It does not block the native main queue or the UI owner. This
consumer uses one live native binding per model and two independent documents.

The codec preserves exact UTF-8 bytes, including CRLF and final blank lines. The
normal document limit is 4 MiB. `document.FileStore` replaces a regular file via
a same-directory temporary file; `document.Session` prevents a stale worker
completion from overwriting newer owner edits or marking them saved. A save
which wrote an older revision leaves the newer revision dirty. Save and Close
uses a second originating input acknowledgment after IO, so preedit or native
edits admitted during the write are settled and checked before closing. Dirty Open and
Close require an explicit decision; canceled or failed Save and Close keeps the
window and document. The document view enables `SoftWrap`; the control also supports unwrapped code
views. Cancel cancels the per-operation acknowledgment context and prevents the
pending action from starting. An operation can be canceled while input acknowledgment or
worker IO is pending. Native chooser and file IO run on a bounded session worker.

From the repository root, prepare the pinned native driver once, then run:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. ./drivers/cogent/env.sh
cd examples/editor
go run .
```

The native dependency preparation procedure is in
[the development guide](../../docs/development.md). The module uses local
replacements for the root library, native adapter and prepared dependency trees.
The consumer does not contain a fork of the model or adapter.

Checks:

```sh
go test -race -count=1 ./...
go vet ./...
go build -race -o /tmp/vellum-editor .
/tmp/vellum-editor -smoke
```

The portable consumer regressions exercise delayed native acknowledgment before
capture, exact byte round trips, atomic failed/stale Open, stale Save and Close,
chooser cancellation, Cancel through a full owner mailbox, explicit close
choices, sole undo journals and independent window lifetime cancellation. The controlled Find text field keeps its search state separate from document
history and reports exact substring counts when Enter or focus loss commits the
controlled value. Native
acceptance additionally exercises real Cocoa text-protocol methods through
explicit deterministic probes, native accessibility button methods, composition
commit/cancel, controlled-field focus handoff, UTF-16 ranges, bidi and wrapped screen rectangles, model replacement
and control remount, configured native edit refusal, settled save/open/close and
native teardown. The second smoke document has a 64-byte test limit.

The Cocoa input fixture does not operate a human IME, change input sources or OS
preferences, qualify VoiceOver, or establish performance results or platform
coverage beyond the recorded host. Stale native
shadow interleaving and queued barrier retirement are deterministic adapter
regressions; the consumer live fixture observes programmatic owner replacement,
fresh client identity after remount and correct subsequent native publication.
Acceptance evidence will record the exact reviewed build and limitations.

The separate `-ax-fixture` mode opens one document for 60 seconds with known
Unicode/CRLF text, terminal blank lines, soft wrapping and a selected emoji. It
prints the process ID, exact text and expected scalar/UTF-16 selection as JSON.
It makes no accessibility permission request and changes no OS settings.

```sh
/tmp/vellum-editor -ax-fixture
```

On a macOS host where the inspector is already trusted, the repository's
optional external wrapper launches this mode, compares the complete UTF-8 value
and tests selected/substring/line ranges, screen bounds, point lookup and RTL
cell order through a separate read-only AX client. From the repository root:

```sh
python3 scripts/check-editor-accessibility.py --binary /tmp/vellum-editor --output /tmp/editor-external-ax
```

External AX qualification and the deterministic Cocoa input smoke are separate
evidence. Recorded runs and exact source/binary identities belong in
[the macOS editor report](../../docs/reports/macos-editor/README.md); the wrapper
must be run against the final binary before its result qualifies that build.
A passing read-only inspector does not qualify human VoiceOver navigation or
human IME input.
