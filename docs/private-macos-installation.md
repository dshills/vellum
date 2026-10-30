# Private Mac installation and updates

From a prepared checkout or matching SDK, install a completed private producer
bundle (not a partial build directory):

```sh
python3 -B scripts/install-private-macos.py --bundle /absolute/private-bundle
```

The default destination is `~/Applications/Vellum/build<build>-<source-prefix>`.
Open apps inside its `apps` directory. The installer prints the exact installed
version path and records source, manifest and named app ZIP identities in
INSTALLATION.json. It verifies owned snapshots of the app ZIPs, bounded ZIP
members, extracted file bytes and strict macOS signatures. A repeated install
of the same identity rechecks bytes/signatures; corruption refuses reuse.

Updates create a separate version. Existing versions and application documents,
preferences and recovery slots are preserved. Quit an older version normally
before launching a newer copy with the same bundle identifier; finish its save
or cancellation decision first. Default storage identities remain stable across
versions. Do not run two versions against the same notebook directory. Keep
prior versions until the new version passes your normal workflow.

A destination must be nonsynced, without symlink parents. Documents/Desktop and
known iCloud/CloudStorage locations are refused: File Provider added FinderInfo
to an earlier extracted app and invalidated its strict signature check. The
installer does not strip file attributes, quarantine, or modify security settings.
A different nonsynced absolute directory may be supplied with --destination.
macOS system aliases such as /var should be canonicalized for this option.

Forge, Measure Pad and Document Starter launch normally. Documents is an A/B
diagnostic and requires `--args -data-dir /absolute/private/workspace`. The
starter accepts the same override for isolated testing; its ordinary launch uses
its own private Application Support directory. Notes uses its established Notes
storage; all development checks must supply a disposable override.

The installation is private, ad-hoc signed and unnotarized. This is byte/identity
and signature verification of the explicitly selected local producer payload,
not independent publisher authenticity, notarization or a new distribution audit.

For a fresh SDK/app build and this installation in one offscreen command, see
[build, verify and install](private-consumer-bundle.md#build-verify-and-install-without-desktop-checks).
The optional producer DELIVERY.json binds this installation receipt; it makes
no native runtime/launch claim. The standalone installer remains available for
retrying already completed bundles.

## Interrupted installation

Ordinary extraction, verification, receipt and publication failures remove their
staging directory and publish no new version. An abruptly killed installer may
leave a hidden `.install-*` staging directory; its presence is not a completed
installation. Retry the same completed bundle: a new staging directory is used,
and the lock is released by the operating system when the old process exits.
Existing installed versions are not overwritten. Select a prior version by its
exact versioned directory if needed; there is no automatic application-data
rollback. Preserve unsaved work before launching a different version. Do not
delete staging directories while an installer may still be active.

These failure paths have synthetic fault-injection and subprocess-termination
coverage. They do not establish power-loss durability or native app behavior.

## Quit and reopening

Notes, Document Starter and A/B Documents opt into coordinated Command-Q. Every
participating document settles input and resolves its save decision before any
window closes. Cancel preserves every participating window; successful saves
made before a later Cancel remain saved. Ordinary red-button close remains a
single-window action. Forge and Measure Pad retain their existing single-window
save/close policies. Last-window closure ends Run.

Notes/Document apps restore the last successful external paths for their fixed
startup windows, with recovery priority and existing native geometry reuse.
They do not restore an arbitrary unbounded window set. Finish current file
operations or decisions before retrying closure. Finder Open With accepts one
notebook or study at a time; registration never changes system defaults. See
[the app-building guide](building-macos-apps.md#finder-document-opening).

## Current private build23

The final verified installed version is
`~/Applications/Vellum/build23-e0e011c605f8/apps`, paired with an SDK from source
`e0e011c605f8ee6ea84827414a1636bcca74c97e`. No app window was launched for these
installation checks. Earlier source-df07cd4 build23 and build22 directories remain
retained; use the e0e011c version for this candidate. See the
[current handoff](mac-document-workflows-handoff.md#source-and-package-identity)
for checks, exact SDK identity and desktop work still deferred by the owner.
