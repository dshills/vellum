# Checkpoint 2: completion requires inspected linkage and explicit expectations

Internal review, 30 October 2026. Refactored the existing inspector's per-app
check for producer reuse; installed-receipt inspection retains the same path.
Every ZIP-extracted signed app is now inspected before completion. Actual minimum
must equal the required --expected-minimum-macos and SDK must match preflight.
This flag is an assertion, never a deployment override or support-policy change.

Completion publication independently requires complete five-app linkage coverage,
matching app identities, ARM64/system-library-only route and OS/SDK expectations.
It reserves BUNDLE_MANIFEST exclusively only after these checks. Missing evidence,
wrong architecture, temp-library/rpath or deployment/SDK drift leaves no completion
or automatic delivery. Recorded per-app evidence binds observed signed executable
hashes. Old completed bundles/installations remain compatible and unchanged.

Reviewed ordering through ZIP extraction, signature verification, inspection,
source/toolchain recheck and completion-before-install. Failure-injection tests
exercise absent evidence and every mismatch before marker publication, plus
repeat publication refusal. No source API/backend/gate/audit/qualification change.

Checks: producer tests (12), linkage tests (6), dispatcher (2), whitespace, and
actual static inspection of retained build24 through the refactored path. All pass.
No app built/launched/reinstalled and no notebook or desktop access. New guards
apply to future packaging; historical receipts were not rewritten.
