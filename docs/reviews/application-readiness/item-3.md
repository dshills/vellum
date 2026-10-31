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
