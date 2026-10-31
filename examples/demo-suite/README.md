# Vellum demo suite

A repository-owned hub for the maintained feature demos. It includes Forge,
Documents, Notes, Reminders, Forma, Luma, Pulse, palette-to-canvas placement,
controls, workbench, editor, grids/tree, GPU demos and deterministic replay.

## Build and launch

From a clean committed Apple Silicon macOS checkout, prepare dependencies using
the root README. Install Python 3 and the Go version declared in the root go.mod.
Then run from this directory:

```sh
./build.sh --build-version 1
./Launch.command --verify-only
./Launch.command
```

Builds default to a new private /private/tmp directory. To choose one, use
`./build.sh --output /private/tmp/my-new-suite --build-version 2`; an existing
output is refused. Source/Go version comes from release/version.json, vellum.go
and go.mod. All module replacements are checkout-relative. No user-specific path,
release tag or Go cache directory is embedded in these build scripts.

The build compiles all catalog entries from that checkout, packages 19 private
ad-hoc signed apps plus replay, verifies source cleanliness again, then verifies
signed executables and app/source metadata both before and after ZIP extraction.
BUILD.json records exact source and executable hashes; DELIVERY.json binds the
completed archive. Only successful builds replace build/latest-suite.json.
Version strings identify the source API release; the recorded commit identifies
the actual development checkpoint, which may be later than its release tag.

The launcher checks the selected archive hash, manifest, executable hashes,
metadata and strict app signatures before opening the hub. If its temporary app
folder has expired, it restores that same checked archive into a fresh directory
and remembers the restored location. `--verify-only` performs these checks and may
restore missing app files, but does not launch any app. If both app files and the
archive have expired, rebuild; no different version is substituted automatically.
Checks bind the locally selected receipt; they are not independent publisher
identity or notarization. No public assets are uploaded by this workflow.

For direct source development, `./run.sh` runs the hub. Build first to populate
its focused app launchers; notifications require the packaged hub.

## Try the demos

Start with **What’s new** for Forge, Documents and palette-to-canvas shortcuts.
**All feature demos** lists the full catalog. Controls gallery opens Mac widgets
with native titlebar toolbar and system appearance. Palette placement is confined
to one native window. Hold a tool, drag onto the canvas, try rejection, Escape,
Undo/Redo and snapping; native checks remain separate from software routing.

Documents and Notes use separate persistent repositories under
`~/Library/Application Support/Vellum Demo Suite/<source version>/`. Earlier suite
versions and the dedicated daily-evaluation workspace are preserved. Quit an
existing instance normally before reopening the same notebook repository. Nothing
imports personal notebooks automatically. Other demos retain their own documented
storage behavior; closing the hub does not close focused applications.

The **macOS Services** tab exposes explicit clipboard, file-dialog, notification
and Finder-drop actions. Clipboard Copy replaces clipboard contents; Save sample
writes only to the chosen path. Notification permission is requested only on an
explicit click. No permission requests occur automatically at startup.

The workbench's local automation endpoint is separately opt-in. Its private log
contains a capability; share it only with a client you intend to control that
workbench. Ordinary launches keep automation disabled.

## Checks and scope

```sh
python3 -B bundle_tools_test.py
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
```

Tests cover hub UI/model contracts, build/launcher catalog agreement, remembered
restoration without launch, version
mismatch, output preservation/alias refusal, changed archive refusal and changed
executable refusal. The normal example/offscreen check scripts include this module.
Build/signature checks do not establish human IME/VoiceOver, every physical input,
other hardware/display configurations, sustained performance or public release
qualification. Default rendering/pacing and native distribution policy are unchanged.
