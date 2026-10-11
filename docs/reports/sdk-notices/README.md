# SDK notice inventory correction

SDK notice collection previously listed the full module graph without downloading
all source directories. Cold-cache modules could silently lose copied declarations;
comparing the first retained-image private SDK with the previous drawing SDK
exposed 33 absent notice files. Both archives passed source/consumer verification,
but their notice inventory coverage differed. Those historical artifacts remain
unchanged; a corrected committed SDK is validated separately.

The exporter now works from a temporary copy of the driver's module files, resolves
the framework/Core/typesetting replacements absolutely, downloads the external
module graph, verifies actual cached module contents and then lists source paths.
Missing source directories fail inventory creation. Canonical module files remain
unchanged; valid sources without an own license retain manual-review status and
public release remains gated.

All 28 exporter, 20 SDK checker and 16 bundle regressions pass. A real local Go
proxy test begins with an empty cache, compares cold and warm notice bytes and
checks canonical module-file immutability without external network traffic.
Two actual full-graph inventories on the corrected source match: 89 modules,
98 notice files, and 84 own module declarations. Five native-library modules
remain manual-review-required. Internal review finds no actionable defect.
Exact source, commands, results and hashes are in
[source-and-check-proof.json](source-and-check-proof.json).

[Release contract](../../../specs/release.md) preserves public distribution gates.
Collecting declared module notices does not qualify embedded native/Rust/font
redistribution, signing/notarization or the broader Mac framework. This correction
changes only exporter, tests and the release specification; native image acceptance
remains covered by its separate committed checkpoint and relocated consumers.
