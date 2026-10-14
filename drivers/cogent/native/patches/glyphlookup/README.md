# Direct exact glyph lookup

Owned patch 0059 replaces callback traversal with direct loops for normalized
path hashing and exact cache-hit comparison. Full geometry validation still
precedes hashing; every command and normalized coordinate remains part of the
key. A hash hit still requires exact path equality. Cold copies, font ownership,
cache limits, admission and renderer state are unchanged.

A deterministic 2,000-path oracle compares the pinned original walker and corrupts
each command and coordinate to check exact comparison. Preparation verifies all
input/output hashes and the patch SHA256. New tests are MIT; existing upstream
notices are preserved. Performance evidence and its scope are in
[the report](../../../../../docs/reports/macos-glyph-lookup/README.md).
