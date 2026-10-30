# Checkpoint 4: execute shipped maintenance suites from the verified SDK

Internal review, 30 October 2026. Added an explicit five-suite declaration to new
SDK manifests, with exporter coverage refusal. Checker validates that declaration
and all suite paths before executing any, then runs the extracted copies with
60-second child deadlines, outside-SDK cwd, no PYTHONPATH/PYTHONHOME and disabled
user site imports. Existing all-file archive verification and generated-Go checks
are retained. No framework/driver dependency, native runtime or feature change.

Tests cover suite coverage, missing-script pre-admission, execution paths/cwd/env
and legacy SDKs that must report unexecuted rather than passed maintenance checks.
Historical archives without the new field remain readable. Scope is the named
verifier/linkage/metadata/preflight/native-input suites, not all Python tooling.

Checks passed: SDK checker (26), SDK exporter (43), SDK CI wrapper (5), whitespace,
and an actual private precommit SDK export/all-file verification/extraction into a
path with spaces followed by all five maintenance suites. The snapshot truthfully
records source clean=false at 82403f3d; it is not a release or a new app candidate.
An initial manual import driver lacked its local scripts import path; corrected
that orchestration only, then the extracted child suites passed without checkout
import overrides. Snapshot digest and results remain in private artifacts.

No app compiled/launched/installed, public asset, original SDK replacement,
desktop, notebook or global-setting operation. Build24 stays installed unchanged.
