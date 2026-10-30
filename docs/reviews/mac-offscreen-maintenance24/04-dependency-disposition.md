# Checkpoint 4: dependency alert applicability and drift guard

Internal review, 30 October 2026. Re-read GitHub alert 4 and primary upstream
advisory/fix, then inspected the hash-bound original dev-dependency declaration,
82-component notice inventory, existing actual-compilation analysis and Cargo.lock.
The full disposition and limitations are in
[the security recheck](../../security/dependency-alerts-2026-10-30.md).

Reviewed applicability instead of changing an archived manifest to quiet GitHub.
No rand is listed in the source-built archive's compiled packages or lock, so the
notice reference does not justify a runtime pin/archive change. Alert remains
open. No remote dismissal, scan suppression or audit decision was made.

Important scope finding: build23 uses the ordinary pinned prebuilt libs-darwin,
not that source-built SDK overlay. The source-built inventory cannot certify the
prebuilt archive's complete dependency graph or safety. Documented explicitly;
no generalized non-reachability or native security qualification claim remains.

Added a focused CI guard for the existing disposition: exact inventory/declaration
hashes, artifact identity, dev-only declaration and no rand component in the
recorded inventory. Simulated drift requires reassessment. Reuses the existing
all-file notice verifier; no original notice or audited release record changed.

Review correction: kept the historical archive recheck separate from CI so the
SDK-contained tooling test does not depend on reports excluded from SDK exports.
The archive's artifact_sha256 field maps paths to hashes rather than a scalar;
corrected the manual evidence probe to select the one static libwgpu_native.a.
This was a probe-shape error, not changed native evidence.

Checks: dependency-disposition_test.py (3 tests), check-stages_test.py (2),
all 408 notice hashes via existing native_gpu.component_notices, bounded in-memory
recheck of historical receipt/analysis/lock and static-archive identity,
git diff --check. All pass. Tests observe applicability evidence only; no new
binary scanner, dependency feature, native run or desktop activity.
