# Owned native dependency inputs

`dependencies.json` pins upstream module versions, Go module checksums, ordered
patch stages, and individual patch SHA-256 digests. `prepare.py` verifies each
stage's input-file hashes before applying that stage. Upstream module source and
its original licenses are retained intact in ignored `../build/` copies.

These patch bytes are copied without modification from the research foundation:
Cogent stages `nativeinput`, `foundationchecks`, `foundationcontracts`, then
`foundationintegration`; go-text stages deferred sbix then deferred static outlines.
The manifest records their original locations for provenance. Historical prototype
sources and evidence stay in the repository, but preparation never reads them.
No AccessKit prototype C/Rust bridge is included in the framework driver.

The copied patch set carries forward experimental fixes; a published fork/release,
upstream disposition and full third-party distribution/license audit remain gates
before distributing framework binaries or a public SDK release.

Patch 0016 fixes native multiwindow ownership races discovered by live race
instrumented handoff checks: composer publication/withdrawal and lock rechecking,
queued-close loop termination after teardown, protected fast resize readback, and
scene title routing to each scene's own window. Its stage verifies original hashes
after patches 0001–0015 and expected final source hashes, including regression tests.

Patch 0016 also gives new windows independent sizing contexts, snapshots shared
screen metadata, and synchronizes startup main-queue publication. The manifest
and regression list identify the complete cumulative safety change.

Patch 0017 adds shared device-loss notification, bounds recoverable surface
reconfiguration, and holds composition/resource ownership through GPU submission.
Surface textures remain retained through presentation. Its input/output hashes
and backend regression tests make these adapter-owned fixes reproducible.

Patch 0020 delivers window shortcut first/final phases even when a Scene has no
focused widget. Focused widgets retain their existing precedence. The owned
regression verifies document fallback, consumed window commands and unregistered
chords, with stage input/output hashes for independent SDK preparation.

Patch 0021 makes main-loop stop a closed, idempotent signal that also works
before queue publication. It releases startup/unaccepted delivery waiters while
preserving synchronous output for accepted callbacks. Both loops acknowledge
refused deliveries; native window creation is one admitted callback and returns
an error on refusal. Six owned race regressions cover these transitions.

Patch 0022 acknowledges admitted callback failure before replaying the original
panic, stops further admissions and avoids a crash-dialog wait on a terminal main
loop. A process-wide desktop barrier serializes native empty posts and GLFW
termination; all five input-event posts use it. Its race tests cover queued/inline
panic, nil panic, incomplete Goexit, pending callers and event lifecycle ordering.
The barrier does not guard every GLFW operation or standalone GPU helper lifetimes.
See [dispatch contract](../../../specs/native-dispatch.md) for supported outcomes.
