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

## Verify an existing delivery without reinstalling

From a checkout containing the verifier, on macOS:

```sh
python3 -B scripts/verify-private-macos.py \
  --bundle /absolute/completed-private-delivery \
  --installation /absolute/Applications/Vellum/build24-sourceprefix \
  --manifest-sha256 EXPECTED_BUNDLE_MANIFEST_SHA256
```

Use the complete five-app producer delivery, including SDK, named ZIPs,
SHA256SUMS, BUNDLE_MANIFEST.json and DELIVERY.json. The selected installation must
have its bound INSTALLATION.json. If the manifest records a preflight digest,
PREFLIGHT.json is also required. A copied INSTALLATION.json, when present in the
delivery, must match the installed receipt exactly. Older completed deliveries
without a preflight field remain supported.

The verifier hashes owned temporary archive snapshots, reuses the bounded SDK
all-file/source checker, verifies all receipt links, compares installed bytes to
the app ZIPs, and checks app metadata plus strict signatures for extracted and
installed apps. It never builds, reinstalls, executes an app, reads notebooks or
writes to the delivery/installation. Temporary extraction is removed afterward.
A nonzero exit means verification failed. JSON output records the exact identity
and explicitly says native runtime was not executed.

The optional manifest digest should come from a separately trusted record.
Omitting it checks local self-consistency only. Neither mode establishes publisher
authenticity, current compiler identity, notarization, a new distribution approval,
power-loss durability or native interaction. Do not modify the selected delivery
or installation concurrently with verification.

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

## Current private build24

The current verified installed version is
`~/Applications/Vellum/build24-b9212489afc9/apps`, paired with an SDK from green
source `b9212489afc921171fc5abbd5275d8919ae930dd`. It includes e589f36e's retired-window
action correction. Packaging, independent integrity/signature verification and
static linkage inspection passed; native launch/input checks remain deferred.

All five executables declare minimum macOS 27.0; the inspection host is 27.2.
This private package cannot run on the historically qualified 26.6.2 host and
extends no framework qualification to 27.2. Build23 and prior directories remain
retained. See the [current handoff](mac-document-workflows-handoff.md#source-and-package-identity)
and [packet summary](reports/mac-build24/README.md) for exact identities and limits.

## Static native linkage inspection

After byte/signature verification, inspect the current private five-app
system-library-only route from a checkout containing the checker:

```sh
python3 -B scripts/check-private-macos-linkage.py \
  --installation /absolute/Applications/Vellum/build24-sourceprefix \
  --installation-sha256 EXPECTED_INSTALLATION_JSON_SHA256
```

The required receipt digest pins the explicitly selected installation. The checker
uses bounded-duration lipo/otool calls and reads file metadata; it does not execute
apps or the dynamic loader. It requires one thin ARM64 executable per app, macOS
load metadata, system-library paths only, the system dyld, no embedded dyld
environment overrides and no rpaths. Extra Mach-O binaries, symlinks, malformed
metadata and a deployment minimum above the current inspection host are refused.
This deliberately checks the existing packaging route, not an arbitrary dylib,
plugin or framework dependency resolver. Run the byte/signature verifier first;
the receipt digest alone does not authenticate installed executable contents.
Do not modify the installation concurrently with either check.

Build24's five binaries declare minimum macOS **27.0** and SDK 27.0; static
inspection was on macOS 27.2. They cannot be used on the historically qualified
26.6.2 host. This private candidate retains its actual requirements; neither
static inspection nor prior native evidence extends the framework support matrix
to 27.2 or qualifies launch there. Desktop checks remain deferred. No deployment
target was patched, no minimum support promise changed and no runtime compatibility
with older systems was established. See the [current handoff](mac-document-workflows-handoff.md).
