# Large-data foundation review

One Prism batch used Anthropic Claude Sonnet 5.5 against the frozen staged
implementation, tests, specifications and selected guides. Coverage is complete:
65 files in eight chunks, no truncation/skips, and one excluded lockfile
(`examples/loadedtree/go.sum`). The input ledger also records that lockfile among
the 66 staged files. Original source snapshots, provider output and process exit
remain unchanged.

Prism returned one medium finding: querying `ps` on the ready record could compete
with the first measured scrolls. The correction moves all external process
observation to the completed end record. Baseline/warm RSS and process CPU deltas
are now explicitly unavailable; native staged Go-memory records remain. A
deterministic regression fails against the original runner, and all 14 protocol
tests plus the two original launch-cleanup regressions pass after correction.
Independent review approves the exact four-file delta. Seven unchanged Go files
and their five qualification/build checks are inherited by exact source and
prepared fingerprints. This correction receives no second full Prism batch.

A discarded nil render-pass allegation was also inspected. The prepared
`drawAllMeasured` checks `rp == nil` and returns before pipeline operations or
`rp.End()`. Submission remains false, so the frame cannot qualify. The original
default path has the same guard. No production change is warranted for that
allegation; physical GPU-fault qualification remains open.

Internal reviews independently cover the portable tree, managed native adapter,
consumer task identity/cancellation and frame tracing/tooling. Raw certificates,
original failures and final checks are preserved in the
[tree archive](../../reports/macos-large-tree/archive.json) and
[timing archive](../../reports/macos-grid-performance/archive.json).
`audit.json` binds the final source, prepared bytes and archives to those records.

The native 100k tree functional workflow passes. The grid latency target remains
unmet: the single unprofiled exploratory p95/p99 is 20.30/22.77ms versus unchanged
4/6ms targets. The diagnostic CPU profile identifies upload and runtime wake
costs; cumulative percentages are not summed or equated with frame wall-time
phases. Older native runs retain their original ready/setup process sampling and
are not observer-free acceptance records. Full P6, memory/endurance, physical
input, human accessibility and production release remain open.
