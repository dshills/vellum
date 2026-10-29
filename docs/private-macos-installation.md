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

## Quit and reopening

Command-Q follows each window's save/close policy. Cancel preserves a dirty
window, but clean windows may already have closed; Quit is not atomic across
all documents. Last-window closure ends Run. Current Notes/Document apps restore the last successful external paths for their
startup windows, with recovery priority and native geometry reuse. They do not
restore an arbitrary unbounded window set. Finish current
file operations or decisions before retrying closure.
