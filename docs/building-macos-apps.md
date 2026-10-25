# Building a personal macOS application

Start from an identified personal SDK and an independent application module. The
[small Mac starter](../examples/personal-mac/README.md) shows one controlled
window and a full-color native toolbar. The [Notes consumer](../examples/macapps/notes/README.md)
adds real local files, shared history and recovery using the same public APIs.
No gallery internals or private driver packages are required. The original frozen
personal SDK still supports the small starter; current Notes needs the private
prose snapshot because its multiline node explicitly selects `Prose: true`.

## Run the independent Notes source

The private Notes delivery pairs `Notes source.zip` with the exact
`vellum-notes-prose-private-sdk.tar.gz`. This private snapshot adds the explicitly
approved prose presentation to the personal-v0.1 base; it is not a published
SDK release. Its manifest records the exact source commit and archive digest.
Extract both next to each
other. Module replacements in `notes-source/go.mod` point into that SDK, rather
than back to a development checkout. Use Go 1.25.6 and macOS development tools:

```sh
tar -xzf vellum-notes-prose-private-sdk.tar.gz
unzip 'Notes source.zip'
. ./vellum-v0.1.0-personal.1/drivers/cogent/env.sh
cd notes-source
go test -race ./...
go vet ./...
go run ./cmd/notes -data-dir /absolute/private/notes
```

The default executable directory is Application Support/Vellum Notes on macOS.
The explicit test directory must be private; missing directories are created
with mode 0700. An already broad or symlink directory is refused. The app does
not tighten permissions on unrelated existing folders. A permanent lock inode
protects the private notebook/recovery slot from concurrent instances.

For a new application, run `scripts/new-app.py` from the extracted SDK with a new
absolute directory and your Go module name. Its module replacements must stay
inside that SDK. `go list -m -json all` makes the actual origins inspectable.
Keep the complete generated replacement set, including the SDK-owned WebGPU
linker compatibility modules. Go does not inherit replacement directives from
a dependency module; losing those entries can select a different native library.
Verify all replacement origins before building. Root framework tests do not
cross module boundaries: test your own app module.

## Model, view and commands

Keep mutable application state on one application UI owner. Return fresh `ui.Node`
trees from View and retain stable editor models and IDs. Controlled callbacks
receive proposals; accept them into the model, report validation errors and
invalidate the window. Do not share editor/models with workers, nest `Do` or
`Drain` inside callbacks, or perform filesystem IO on the owner.

Notes implements that contract in [model.go](../examples/macapps/notes/model.go)
and [view.go](../examples/macapps/notes/view.go). New/delete/pin/title/body actions
share one `history.Journal`. Window-level commands enter the same request path
as buttons. One application journal owns Undo/Redo; the retained text editors
keep only their bounded internal fallback history, avoiding a second application
Undo route. Saving marks a revision; Undo becomes dirty and Redo to that exact
saved revision becomes clean again.

Before structural, history or file actions, the Notes executable settles its
native search field, then the current title and body through public driver APIs.
The acknowledgment runs on the owner and rechecks the original `WindowRef`.
It does not block Cocoa main. `Model.Request` is an example-owned gate, not a new
framework API: it retains at most one action and ignores duplicate completion.
A composition that cannot settle refuses the action and preserves preedit.

## Files and recovery

[notebook.go](../examples/macapps/notes/notebook.go) captures a detached versioned
snapshot with bounded note count/text size, stable identities and the next ID.
Decoding rejects malformed/unknown/duplicate fields and invalid identities
before installation. Allocate replacement editors before assigning live state.

[files.go](../examples/macapps/notes/files.go) configures `document.Session` with
the app journal and `document.FileStore`. Capture and Install run on the owner;
Encode, Decode, chooser and IO run through the existing worker/completion path.
Open refuses dirty replacement until an explicit decision and checks revision
identity before installing a decoded result. Invalid data and stale results
preserve the live model and history. The `Written` result and the acknowledged
error are distinct: a file can have been replaced before a later sync/stale error.

Notes initially saves its private notebook; Save As switches the current
file destination. Restart opens the private default, so external notebooks are
reopened explicitly. There is no external-file merge/watch or persistent Undo
log. Keep those limits visible in your application instructions.

Recovery uses a separate session and journal so a recovery write cannot mark the
main notebook saved. A two-second timer posts owner work; full queues, composition,
prompts and busy IO can delay a capture. Only a completed snapshot advances the
recovery revision. Startup offers Restore/Discard/Close explicitly; restored
content is dirty and history is fresh. Successful file actions or explicit
discard atomically write a clean marker. Recovery is the last completed copy,
not a guarantee for the latest keystroke. A recovery-marker failure after a
successful main save preserves saved bytes, reports the error and keeps the
window open for retry.

Use `CloseRequest` to refuse native close until settlement and the save decision
finish. Then close the exact original `WindowRef`. On shutdown cancel the
scheduler, join owned IO and release the directory lock only after retirement.
The Notes executable retains the lock until process exit if the three-second
retirement bound fails. Window lifetime cancellation prevents late installation.

## Mac presentation and verification

Use standard native chrome for ordinary applications. The Notes unified-chrome
example reserves a measured 52 Dp header; that measurement is specific to its
configuration. The supported `TitlebarToolbar` entry point has its documented
single-window/customization limits. Retain original-color `IconImage` assets;
use the Mac theme and system preferences where appropriate. See
[Mac themes](mac-theming.md) and [native toolbar](native-titlebar-toolbar-prototype.md).
The current Notes body font/gutter and rich formatting limits remain documented;
this stabilization adds no editor-presentation API.

Test model/history, malformed input, file round trips, failed writes, stale
results and closed-window cleanup without desktop control. Then check the actual
packaged app's native input, chooser, focus and shutdown separately. A routed
regression or AX tree is not physical-pointer or human IME/VoiceOver evidence.

Use the existing `scripts/package-app.py` with a notice bundle bound to the actual
binary. Record the consumer commit, frozen SDK hash, native capsule, executable
and final ZIP hashes. Verify signatures after extracting the exact ZIP you will
use. Personal apps are ad-hoc signed and unnotarized; private use does not turn a
changed payload's pending public-distribution audit into approval.

## Notes presentation

Use `app.ChromeStandard` to keep native window buttons and dragging. The client
view starts below that title bar; do not reserve another title-bar band inside
this view. Shared `ui.MacTheme` supplies field/button metrics, with application
styles controlling compact column/card spacing. Style size constraints describe
content boxes; padding and borders add to the allocated bounds. Test rendered
geometry rather than assuming a nominal field height equals its outer height.

`ui.MultilineText` with `Prose: true, SoftWrap: true` retains the same editor model
and native settlement while selecting proportional typography and no gutter.
The default remains the code editor. Prose wrapping uses pixel-measured shaping;
it is not a rich-text document or a fixed-column tab editor. Font fallback and
unsupported bidi-isolate interaction remain explicit limits. The body title is
a separate controlled field. Offscreen layout and geometry do not prove native
pointer acceptance or human IME/VoiceOver quality.
