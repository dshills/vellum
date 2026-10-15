# Duplicate ready accessibility cache reuse

The owner bridge keeps an already detached completed grid cache when a fresh
owned ready snapshot encodes identically. New callbacks still replace the old
callbacks, and admitted actions keep their semantic generation. The optimization
removes duplicate cache cloning after full JSON encoding; it does not change the
wire payload, native validation or native notification behavior.

Reuse requires builtin scalar values, matching scoped node order/count/IDs,
Go-only focus metadata and exact cached scalar types/values. Pending, custom or
mutable values and borrowed updates retain ordinary preparation. The completed
identity is installed under the same mutex as its cache. Direct preparation
invalidates that identity, and final publication cannot certify an intervening
preparation. Close clears the identity and retained cache.

## Offscreen allocation comparison

Original base: `c11858e4598362fae8b6e7e2a6c3e5f4e8245dee`. The original two
implementation files come from exact Git blobs through a Go overlay. Both routes
run the same benchmark body with fresh transferred node and callback containers
on each owned call. Initial semantic equality is seeded without native sends;
zero generation advancement is asserted after each benchmark. The original
overlay omits only the new implementation-specific regression functions.

Five samples per route use `-tags=offscreen -run=^$`
`-bench=^BenchmarkAccessibilityDuplicatePublication$ -benchmem -benchtime=300ms`
`-count=5`. Prepared dependencies stay byte-identical across all 6,217 files.
The table reports medians; bytes include fresh input cloning, callback creation,
encoding and duplicate preparation.

| Cells / route | Original B/op | Candidate B/op | Change | Original allocations | Candidate allocations |
| --- | ---: | ---: | ---: | ---: | ---: |
| cells-0/owned-false | 2,395 | 2,395 | +0.00% | 19 | 19 |
| cells-0/owned-true | 1,818 | 1,057 | -41.86% | 16 | 10 |
| cells-40/owned-false | 66,600 | 66,598 | -0.00% | 274 | 274 |
| cells-40/owned-true | 50,773 | 33,878 | -33.28% | 269 | 177 |
| cells-400/owned-false | 690,924 | 692,837 | +0.28% | 2,447 | 2,447 |
| cells-400/owned-true | 518,269 | 343,406 | -33.74% | 2,440 | 1,624 |

For 400 cells on the ordinary owned route, median bytes fall from 518,269 to
343,406 (33.74%), and median allocations from 2,440 to 1,624 (33.44%). Borrowed
allocation counts are unchanged. The human was online; these sequential
microbenchmarks establish phase allocation evidence, not quiet-host frame
latency, native memory reclamation or statistical significance. Timing samples
are retained without using them to qualify the native performance budget.

## Checks and review

The complete adapter root-package offscreen race suite passes (Go output 90.867s;
wrapper 98.539s), as does adapter vet (2.392s). The native cycle fixture compiles
with `vellum_native_checks` (3.173s), but its binary is not executed. An initial
build without that required tag correctly refused the build constraints; its
record is retained. Final v3 checks bind tests that capture expected values
before input ownership transfer; the tagged build is v4. Earlier v1/v2 results
are retained as historical endpoints. The report is updated after the checks;
its earlier version is preserved with their source manifests. These durations are check durations, not UI measurements.

Regression coverage includes two interleaved scopes, detached pointer storage,
current callbacks, retirement, direct cache invalidation, metadata-only focus
changes, equal JSON with distinct scalar types, custom/mutable values, borrowed
updates and a custom-to-scalar transition. Existing lifecycle regressions cover
queued actions resolving the newest callback, pending revocation, mixed scopes,
failed encoding, removal and close. Disabling reuse makes the detached-storage
regression fail; suppressing invalidation makes its regression fail. Both faults
are overlays, and production files remain unchanged.

An internal design review found a preparation/publication invalidation gap before
the final checks. The identity now changes atomically with cache installation.
The exact source endpoint receives internal review and one requested Prism
checkpoint review using Anthropic Claude Sonnet 5.5 before commit. Before commit, raw results, coverage and measured inputs must be
retained in `records/evidence.tar.gz` with `archive-index.json`, and
`qualification.json` must state the final review outcome.

This is a scoped Go-side optimization. It does not resolve the failed 123/1,000
native endurance campaign or qualify the full Mac framework, native/GPU retention,
physical interactions, sustained frame budgets or public distribution audit.
Live desktop testing remains disabled while the human is online. No input
sources, VoiceOver, system preferences or general clipboard are changed.
