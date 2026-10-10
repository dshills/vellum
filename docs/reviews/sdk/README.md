# Independent local SDK and reproducible native benchmarks

Prism reviewed the new SDK implementation, specifications and documentation using
Anthropic Claude Sonnet 5.5 (`claude-sonnet-5-5`) at a medium severity threshold.
The final report covers 29 selected files with complete coverage, no truncated
bytes or skipped chunks, and no medium or high findings. Earlier reports record
issues corrected in dependency rollback/cleanup, workflow configuration, CPU
accounting and the readiness pipe reader.

The seventeen historical patch files and six historical input manifests copied
into driver ownership were excluded from this SDK review using explicit patch
stage path globs. They are byte-identical to their existing prototype sources;
[provenance records](../../reports/sdk/final/historical-patch-provenance.json) and
[input provenance](../../reports/sdk/final/historical-input-provenance.json)
record those comparisons. The phase-four cumulative multiwindow patch was
reviewed separately in [its implementation review](../windows-services/README.md).
No new implementation or specifications were excluded. The final reviewer rules
explain the documented preparation-lock/manual-recovery contract and contain no
severity overrides or baseline suppressions.

[Final canonical evidence](../../reports/sdk/final/README.md) records fresh
driver-owned preparation, portable race tests, native tests/vet/build, twelve
upstream ownership race regressions, native driver race tests, standalone-module
race tests/vet/build, and a generated application outside this repository at a
path containing spaces. Eleven preparation safety tests and ten benchmark harness
tests pass. Three native processes each verify GPU pixels and record ten render/
readback samples, RSS and quiet CPU. Source hashes identify the checked files.
The workflow passes actionlint; no remote CI run is claimed.

Generated raw logs, measurement JSON and historical staging evidence are retained
as local validation records and were not sent as review inputs. Their summary
prose was included. Quiet CPU remains above the sub-1% target. This is a local SDK
using checkout replacements; public distribution/license audit, Windows/Linux
runtime qualification and complete native accessibility remain release gates.
No macOS input sources, accessibility settings or other OS preferences changed.
