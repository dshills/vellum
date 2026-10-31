# Updating private macOS applications and preserving data

This guide covers Vellum's existing private ad-hoc-signed application route.
It does not establish a general file-format migration promise, notarization or
quarantine acceptance. Use the [status page](status.md) to distinguish source tags,
private candidate builds, demo versions and older public downloads.

## Before switching versions

1. Identify the exact app, build and data directory. The demo suite uses versioned
   repositories under Application Support/Vellum Demo Suite; the daily Documents
   launcher uses its separately pinned versioned evaluation profile. Standalone
   Notes and generated applications have their own directories. Never assume that
   two windows or identically named apps use the same data.
2. Save work and quit normally, resolving every participating document's decision.
   Cancel preserves windows and edits. Do not force-quit to make an upgrade easier,
   and do not run two versions against the same notebook repository.
3. After the app exits, copy the complete application data directory to a separate
   private local backup. Keep the original. Preserve permissions and use mode 0700
   for private directories; do not put notebook data in a public shared folder.
   Include notebook, recovery and window/session metadata, not just the main file.
4. Retain the old app and its verified archives/receipts. Verify the selected new
   packet and installation with [the private installation tools](private-macos-installation.md)
   before launch. A checksum proves selected bytes, not publisher authenticity.

A source tag does not automatically update an installed SDK, demo suite or app.
The demo launcher may restore its same selected archive after temporary staging
expires; that is restoration, not downloading a newer release. Rebuild deliberately
from the chosen source to update it.

## First run and rollback

For acceptance, give Notes, Documents and generated Document Starter an explicit
fresh absolute private `-data-dir`, following their app guides. The daily launcher
selects its own qualified candidate/profile and refuses candidate drift. It does
not migrate the old profile automatically. Keep generated apps' module-derived
storage identities stable across upgrades.

Check edit, history, Save/Open and normal restart using disposable data first.
Only after those checks should you deliberately select a backed-up everyday
notebook. Historical fixtures demonstrate named compatibility cases, not support
for arbitrary future JSON schemas or third-party edits.

To roll back, quit the new app normally and select the exact retained old app.
Restore a backup into a separate private directory for inspection; retain both
original and changed copies. Do not assume an old binary can read a file rewritten
by a newer schema. Do not overwrite a live repository or transplant one app's
session metadata into another app's storage.

## Refusals and recovery

- **Already running / locked repository:** find and quit the owning app normally.
  A lock file can remain after clean exit; its existence alone does not prove a
  stale process. Do not delete lock files or defeat repository locking.
- **Permissions or symlink refusal:** choose a new private directory or restore a
  private backup. The application intentionally refuses unsafe existing locations;
  it does not repair permissions on unrelated folders.
- **Recovery offered or startup refused:** preserve the notebook, recovery file
  and metadata first. Follow the app's explicit recovery decision. Do not delete
  a recovery file to suppress a warning or edit notebook JSON while an app is open.
- **Missing temporary demo files:** run the canonical demo Launch.command. It
  verifies and restores the selected archive. If the archive has also expired,
  rebuild from the intended source; no newer version is substituted automatically.
- **Signature or deployment failure:** stop launch and verify the exact selected
  packet. A host below a binary's recorded minimum cannot be qualified by changing
  a plist. Do not remove quarantine or weaken system protections as a workaround.

Record reproducible data-loss, save, history or startup failures in the
[daily-use issue log](daily-use/ISSUE-LOG.md), including candidate, host, exact steps
and save state, without copying confidential notebook content.
