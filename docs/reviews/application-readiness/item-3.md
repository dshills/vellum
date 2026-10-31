# Item 3: local signed-app staging

The v1.1.2 build into Documents failed because its file provider attached
com.apple.FinderInfo to the signed app directories. The previous retry into
/private/tmp passed, but did not prevent recurrence for a chosen synced output.

The builder now keeps all app bundles, binaries and verification extraction in
fresh private local staging, regardless of delivery location. Only the ZIP and
completion receipt enter delivery output. ZIP creation strips resource forks and
extended attributes, copy hashes are compared, and verification extracts the
actual delivered ZIP into local staging. The existing canonical launcher checks
signatures and restores the exact selected archive if staging expires.

Default Prism completed full coverage. Its medium failure-cleanup concern is
addressed: a context manager removes its own newly created staging on exception,
while retaining successful app files for launch. It never removes prior outputs.
The low cross-platform test finding is fixed by marking native staging tests as
macOS-only; actual packaging already rejects unsupported hosts. Eight Python
checks pass, including injected failure cleanup and delivery-path isolation.
The full build from committed source follows this checkpoint.

Full build5 from clean 04f1174f883469f6f62ce0975f1349127d17db9f passed all 19 strict signed bundles and ZIP-extraction checks while delivering into the synced workspace. The output contains exactly the archive and receipt, no app bundles. The original working suite was moved aside, not deleted. Canonical launch --verify-only restored the selected ZIP into fresh local staging and verified the same source/hash identity; launch then opened the v1.1.2 hub visibly. Producer native_runtime remains its original not-executed fact, with this subsequent hub startup recorded separately in item-3-result.json. The compatibility Launch.command still delegates to the canonical launcher.
