# Durable installation — 2026-10-29

Added a private bundle installer using immutable versioned directories under
~/Applications/Vellum. Source archive bytes are snapshotted before hashing and
extraction; only bounded, app-rooted regular ZIP members/directories are admitted.
Extracted bytes and strict signatures are verified before publication. A repeat
install verifies the identical receipt and app bytes; prior versions are never
replaced. No application data or security/input settings are changed.

## Internal review and verification

Reviewed ZIP containment, executable modes, owned snapshot/hash binding, failed
publication cleanup, repeat-install tamper refusal, symlink/synced destinations,
exclusive installation locking and preservation of old versions/user data.
Review corrected an existing-version symlink ambiguity before commit. Temporary
test roots are canonicalized because macOS /var is a system symlink; the intended
user Applications destination has no redirected parent.

Five focused installer tests passed. The real build20 SDK-paired four-app bundle
was installed under build20-62e953cbfb2b and the identical install command passed
again with extracted-byte and strict signature verification. No framework API or
feature change; no external review. Running older apps are left intact until
normal closure for the native lifecycle checks.
