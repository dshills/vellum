# Generated document first-launch storage — 2026-10-29

The document starter now selects an app-owned private notebook/recovery slot
under os.UserConfigDir. On macOS this is Application Support followed by
vellum.document. and the first 128 bits of SHA256 of the application's Go module.
The generated README lists the exact directory. The applicationStorageID is
app-owned source and must remain stable across app/module renames and upgrades.
Different apps must use different module identities. The minimal template and
A/B Documents diagnostic retain their existing behavior.

An explicit absolute -data-dir continues to win, including when OS configuration
is unavailable. Relative overrides, unavailable configuration and nonabsolute
configuration fail before repository access. There is no Vellum Notes migration
or fallback. Existing OpenRepository privacy/symlink/lock guards are reused;
existing folder permissions are never silently changed. No framework API change.

## Internal review and checks

Reviewed deterministic identity generation, path joining/containment, error
propagation, explicit-override precedence, copied application imports and tests,
repository permission/lock admission, and unchanged worker shutdown ordering.
The resolver itself does no IO. Runtime construction remains on the startup
path before the UI owner starts. No unresolved finding in this checkpoint.

Passed the generator identity/import test (same module stable, different module
separate); all 43 existing SDK exporter/starter tests; and race-enabled offscreen
tests plus go vet for an independently generated example.com/closeout-document
consumer. Its three generated storage tests cover default identity, override
precedence, rejected relative paths and configuration errors; copied repository
tests cover concurrent/malformed/symlink storage. Native no-argument launch is
verified separately against the final package. No external review service used.
