# Offscreen SDK consumer CI — 2026-10-29

Added a separate macos-15 Framework workflow job for clean committed private SDK
export and relocation. It verifies minimal and document consumer race/vet/build,
public API scenarios, member integrity and exact replacement origins under a
path containing spaces. No window or benchmark diagnostic is launched. The job
retains bounded logs and the successful JSON report, not an SDK release asset.

The orchestration binds the verifier to the SHA256 of its own fresh export and
to the source commit/clean identity. This is producer-output binding, not
independent third-party authenticity certification. A dirty checkout, output
inside the checkout, mismatched report or failed child cannot leave a successful
relocation report. Each child has a deadline and bounded logs; output is exclusive.

## Internal review and checks

Reviewed workflow isolation/deadlines/read-only permissions, archive trust and
module-origin checks, failure reporting and child environment filtering. No API,
audit, release gate or runtime change. The existing verifier now records the
document consumer replacement map separately and names both consumer checks.
Passed five new orchestration tests and all 24 SDK integrity/origin/timeout tests.
The full command passed from committed source 9ef6485 in 73.2 seconds: fresh
private export and independent minimal/document consumer race/vet/build plus
public API scenarios. See 03-sdk-relocation.json for the exact SDK/source binding.
This validates the CI command locally; execution of the new hosted job requires
a future push. No external review service used.
