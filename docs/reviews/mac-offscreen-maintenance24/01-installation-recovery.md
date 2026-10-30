# Checkpoint 1: interrupted private installation

Internal review, 30 October 2026. Features and desktop checks remain frozen/deferred.

Scope: installer failure coverage and recovery instructions; no installer runtime,
framework API, app storage or installed payload change.

Reviewed publication ordering, temporary-directory cleanup, version preservation,
lock lifetime and retry behavior. Six injected boundaries cover post-extraction,
byte verification, signature verification, receipt writing, staging rename and
final publication rename. Before and after retry, the entire prior synthetic
installation remains byte-identical. A separate bounded POSIX subprocess is killed
after extracting its fictional app; only hidden staging debris remains, and a new
install plus repeat verification succeeds. The killed process is test-owned.

Findings: ordinary exceptions already clean staging; abrupt termination intentionally
cannot guarantee cleanup. Document retained staging rather than claiming automatic
rollback or deleting potentially active work. POSIX-only kill test explicitly skips
other hosts. No runtime defect found; no production correction needed.

Checks: install-private-macos_test.py (7 tests, six fault subcases),
private-macos-bundle_test.py (9 tests), git diff --check. All passed. Signatures
are mocked for fictional binaries; no native app executed, notebook accessed,
power-loss durability claim, signing change or desktop interaction.
