# CI integration: portable Finder queue fixtures — internal review

During the main integration run, review found the dependency-free native callback
queue tests used literal `/tmp/...` paths. Windows filepath.IsAbs requires a drive
or UNC volume, so those positive fixtures would be rejected before reaching the
queue behavior they intended to test. The platform workflow runs these tests.

Build all positive paths from t.TempDir and filepath.Join, and retain the exact
original path for the detached-copy assertion. Keep relative-path rejection;
append NUL to an otherwise absolute native path so that check is no longer
masked by the same Windows absolute-path failure. No runtime validation, queue
bounds, platform claim, API, dependency, desktop policy or release gate changed.

Internal review checked every enqueue path and preserved readiness, aliasing,
capacity, retirement and invalid-path assertions. Focused Go 1.25.6 race tests
passed against byte-identical queue/test sources in an isolated portable module
using the real root app package, without importing or initializing AppKit/GPU.
The same source cross-compiled for Windows; this is compilation evidence only.
Hosted platform execution and the complete Framework run remain the final checks.
The initial main commit's stable API workflow passed on all three hosts.
No external code review service or desktop checks were used.
