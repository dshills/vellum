# Dependency applicability records

The latest recorded inspection is the [31 October application-readiness triage](../reviews/application-readiness/item-5-advisory.json).
GitHub alert 4 remains open and low severity. It references rand 0.7 under the
preserved itertools notice manifest's development dependencies. Rand is absent
from the recorded compiled-source superset for the bound native artifact.

This is an artifact-scoped applicability finding, not a claim that all upstream
benchmark builds or future native binaries are unaffected. Do not change original
notice manifests to make a dependency scanner quiet or dismiss alerts as a side
effect of documentation cleanup. Re-evaluate against changed compiled inputs.

The [25 October record](dependency-alerts-2026-10-25.md) and
[30 October recheck](dependency-alerts-2026-10-30.md) retain their original dates,
source assumptions and disposition. The [current status](../status.md) separates
source releases from native artifact identities. Follow
[native dependency maintenance](../decisions/0012-native-dependency-maintenance.md)
when changing pins or patches.
