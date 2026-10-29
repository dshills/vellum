# Building a personal macOS application

Start from an identified personal SDK and an independent application module. The
[small Mac starter](../examples/personal-mac/README.md) shows one controlled
window and a full-color native toolbar. The [Notes consumer](../examples/macapps/notes/README.md)
adds real local files, shared history and recovery using the same public APIs.
No gallery internals or private driver packages are required. The matching private v1 SDK includes prose presentation, the native app shell
and the explicitly scoped [stable API contract](../specs/stable-api.md). Older
SDKs retain their historical capabilities; use the paired v1 source and SDK.

## Run the independent Notes source

The private v1 delivery pairs `Mac examples source.zip` with
`vellum-v1.0.0-private-sdk.tar.gz`. Its manifest pins the source commit and SDK
hash. Extract both beside each other. Replacements in
`macapps-source/go.mod` point into that SDK, never a development checkout.
Use Go1.25.6 and macOS development tools:

```sh
tar -xzf vellum-v1.0.0-private-sdk.tar.gz
unzip 'Mac examples source.zip'
. ./vellum-v1.0.0/drivers/cogent/env.sh
cd macapps-source
go test -race ./...
go vet ./...
go run ./cmd/notes -data-dir /absolute/private/notes
# The same source archive also contains the independent two-window shell.
# go run ./cmd/macshell
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
file destination. The Notes, Documents and generated document executables opt into bounded
WindowSession metadata so restart reopens the last successful external notebook.
Recovery is offered first; invalid/missing external files preserve the live
private default and their metadata for explicit repair. There is no external-file merge/watch or persistent Undo
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

## Application icons

The macOS packager accepts an optional `--icon /absolute/path/AppIcon.icns` or
`--icon /absolute/path/AppIcon.png`. A PNG must be exactly 1024×1024; full-color
RGBA transparency is preserved. It is resampled into the standard 16, 32, 128,
256 and 512 point sizes at 1× and 2× using macOS `sips`, then compiled with
`iconutil`. An existing `.icns` is decoded for validation and copied unchanged,
so it can contain individually designed small-size artwork. Sources must be
regular, nonsymlink files no larger than 16 MiB. Invalid artwork refuses the
package without publishing a partial app.

```sh
python3 scripts/package-app.py --binary /absolute/build/my-app \
  --output '/absolute/build/My App.app' --name 'My App' \
  --identifier com.example.myapp --version 1.0.0 --build-version 1 \
  --icon /absolute/assets/AppIcon.png --sign-identity -
```

The packager embeds `Contents/Resources/AppIcon.icns`, sets `CFBundleIconFile`
and records source/output artwork hashes in `APP_MANIFEST.json`. Icon bytes and
metadata are verified before signing, after signing and after publication.
The existing binary-bound notices, audit and exclusive-output rules still apply;
include the appropriate `--notice-bundle` for a distributed application.
No icon argument preserves the existing bundle behavior. Non-macOS targets reject
this option explicitly. A [complete starter example](../examples/personal-mac/README.md#application-icon)
includes original full-color artwork.

The icon is application identity shared by all windows, rather than a titlebar
or toolbar image. Launch the packaged `.app`; `go run` and standalone binaries
have no bundle icon. Finder/Dock may cache artwork: rebuild to a new bundle path
and relaunch instead of modifying a signed app in place. This supports classic
ICNS artwork, without Icon Composer `.icon`, asset-catalog appearance variants,
or runtime Dock badges/icon replacement. Packaging checks do not establish
that Finder/Dock have refreshed their cached pixels. If codesign reports Finder
information/resource-fork detritus, use a local build/extraction directory outside
synchronized Documents storage. This host injected that metadata there during
the example check; local temporary packaging/extraction passed. The packager
does not strip metadata from user-owned files.

The bundle key and Retina conventions follow Apple's
[CFBundleIconFile reference](https://developer.apple.com/documentation/BundleResources/Information-Property-List/CFBundleIconFile)
and [high-resolution icon guidance](https://developer.apple.com/library/archive/documentation/GraphicsAnimation/Conceptual/HighResolutionOSX/Optimizing/Optimizing.html).

## Complete document starter

Generate an independent application with
`python3 scripts/new-app.py /absolute/my-doc-app --module example.com/my-doc-app --template document`.
It copies the tested Notes app recipe into app-owned internal packages, including
prose editing, native toolbar/menu actions, input settlement, Save/Open/Save As,
dirty-close decisions, recovery and Undo/Redo. Full-color artwork stays local.

Run `go test -race -tags offscreen ./...`, then `go run .`. The packaged app
also starts normally from Finder. On macOS, its default notebook/recovery slot
is `~/Library/Application Support/vellum.document.<module-hash>`; the generated
README lists the exact directory. The 128-bit hash derives from your Go module,
so different application modules have different storage identities. Retain the
generated `applicationStorageID` across upgrades, even if the module is renamed.
Generating multiple apps with the same module intentionally shares that identity.

For tests or an existing private notebook, use
`go run . -data-dir /absolute/private/my-doc-data`. Explicit absolute overrides
take precedence without consulting the configuration directory. Missing or
nonabsolute OS configuration refuses startup instead of falling back to the
working directory or another app. The existing repository checks create a
private directory and refuse symlinks, broad permissions and concurrent access;
they do not tighten an existing folder's permissions. No Vellum Notes default
is reused or migrated. Use a separate directory per application. Other hosts
use `os.UserConfigDir` and remain unqualified. Customize the copied internal
model and views. The default minimal template remains available for utilities.

## Restoring documents and windows

`Files.ConfigureRestoration` is app-owned startup wiring. It reads a bounded
16 KiB WindowSession.json beside the private notebook, and opens the last external
notebook through the existing document worker. Writes use a separate document
session, never filesystem IO in View/owner callbacks. Active-window timestamps
order A/B startup creation so the last active window is created last. Existing
native window geometry persistence restores positions/sizes for stable titles.
Successful save/open records the path; missing/invalid metadata or an external
file never replaces the live default. Successful explicit file action may replace
a previously invalid session record. Recovery takes precedence and retains the
recorded external destination after explicit Restore. Quit waits for metadata IO.
Forge similarly retains its last successful study in its existing workspace
settings and reopens through its document session; interrupted recovery wins.

### Finder document opening

Set `cogent.Driver.OnOpenFiles` to receive a bounded detached `OpenFilesEvent` on
its UI owner. Do not nest `Application.Do` in the callback. Route paths through
existing validation and save decisions. `Initial` marks capture before native
input/window binding; warm requests wait for native readiness. Notes and
Document Starter accept one `.vellumnotes` file, Forge one `.vforge` study.
Multiple files are refused with status feedback; A/B Documents routes to the
active document. Recovery remains higher priority than opening another file.

For packaged cold-launch delivery, pass `--document-types /absolute/types.json`
to `scripts/package-app.py`. The bounded JSON array contains exact `extension`,
`uti` and `name` strings. For example:

```json
[{"extension":"vellumnotes","uti":"com.dshills.vellum.vellumnotes","name":"Vellum notebook"}]
```

Packaging verifies matching signed Info.plist declarations. Handler rank is
Alternate: use Finder's Open With when another app owns the type. Registration
does not change system defaults. The private bundle registers Notes, Document
Starter and Forge; Documents remains an explicit-data-directory diagnostic app.
