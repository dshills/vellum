# Direct managed-grid accessibility construction

This rejected experiment halves local accessibility construction cost, but its
ordinary native p95 is slower in all five matched pairs. The runtime candidate
and its experimental tests were reverted; no implementation gain is claimed.
The ordinary native target remains open; this is not full Mac qualification.

## Experimental implementation and correctness

Grid records append directly to the owner snapshot and handler map. Visible row
identity and union bounds are computed once per row. Detached frame values are
copied from the immutable completed measurement. Late stamping reuses them only
while that exact measurement is current and all original readiness/editor fences
pass. Measurement replacement, unknown descendants and pending layout use the
unchanged legacy stamping path; only producer-added frame fields are cleared.
This avoids a frame-only replacement becoming a pending scope with no retry.
The experiment introduces no public API, retained cache or native transport
change. Its full source/test/spec diff is retained in the archive, based on
ea20486; the committed runtime remains ea20486.

The frozen ea20486 constructor/stamper are retained as archived offscreen
tests. A recorded parity check confirms they differ only by names/whitespace.
Tests compare complete Go records and semantic JSON at DPI96/162/192, disabled
states, preceding records/handlers, selected rows and empty rows/headers seeded
before construction. Deterministic replacement/pending/unknown-child and real
inline-editor tests cover fallback. Original projection indexes remain captured
by actions; the full root suite covers routing, stale actions and lifecycle.

Full driver-root offscreen race passes in 284.295s; vet and optimized
native build pass. Existing AppKit deprecation and duplicate-library warnings are
retained. The initial preparation attempt did not install the new test file;
its existing-test output is retained but is not the new regression evidence.
Focused-v2 installed it and passed; the complete final race includes the later
same-pointer test strengthening.

## Measurement

Five same-process 200ms benchmark observations per path, GOMAXPROCS2, Go1.27.2,
without race/profiler: legacy median 185,579ns,
563,416B/4,804allocations;
direct median 89,124ns,
283,008B/2,776allocations.
This covers 400-cell construction, merging and stamping, excluding full refresh,
JSON and native publication. It is not a native frame result.

Native timing uses the existing grid runner and shared exclusive live lease,
normal runtime GOMAXPROCS/GOMEMLIMIT unset, GPU requested, five interleaved pairs,
128 warm events and 120 samples per run at 10k fixed rows. Both binaries exit 0;
matched geometry is checked per trajectory sample. Actual window/display metadata,
raw sample logs, binary hashes/linkage and module origins are retained. Baseline
reuses the previous verified stage0078 runtime binary; its prepared dependencies
and original root source leaves match clean ea20486. Candidate is an optimized
working-tree build, without overlay/VCS stamping. Neither is presented as a new
independent clean historical rebuild.

| Build | Per-run p95 UI work (ms) | Median run p95 (ms) | All runs meet 4 ms/6 ms |
| --- | --- | --- | --- |
| baseline | 5.480, 4.565, 5.404, 4.729, 4.601 | 4.729 | False |
| candidate | 5.604, 4.977, 5.473, 5.343, 5.746 | 5.473 | False |

All five candidate p95 values exceed the paired baseline. This result is a
rejection decision, not proof of a particular causal mechanism or a statistically
isolated estimate. Median run p95 rises from
4.729 ms to 5.473 ms. The source is reverted despite
lower allocation observations and passing semantic checks. Phase medians are
retained for diagnosis, without summing independent percentiles.

Process allocation observations include runtime and measurement overhead; they
are not framework-only allocation or retained-memory qualification. GPU requested
is not proof of actual route selection in these ordinary runs. Routed native
scrolling is separate from physical-pointer, screen-reader and presentation
verification. Complete table/tree/general-UI/pacing/memory/endurance/SDK gates stay
open; this first workload does not substitute for the full matrix.

## Review and evidence

Internal source/test/spec review approved. One complete Prism Anthropic
Claude Sonnet 5.5 review covered 7 files/26,999 bytes in 1 chunk/1 fresh provider call.
Raw retained medium 1c816691ecec034e and CLI exit 1 are preserved. It hypothesized a
nil handler map on a disabled private-helper call. Both production callers pass
initialized maps; the raw wrapper retains its original non-nil empty map, and
the enabled helper initializes nil itself. Independent internal review refuted a
reachable bug; no second unchanged provider call was made.

[records.json](records.json) identifies the bounded [raw archive](records.tar.gz),
its member hashes and summary. No SDK or executable copies are archived.


## Follow-up baseline diagnosis

After restoring the original runtime, two separate diagnostic cases each
completed 120 routed samples: source details and differential Go allocation
profiles. These instrumented cases are excluded from ordinary timing qualification.
The source case contains 121 accepted host records (120 captures plus one reused
composition), no CPU fallback; host median capture is 1.112 ms, encoding 0.355 ms.
Lease/resource/input timings are nested and cannot be added to one another.
The allocation estimate totals 876.72 MB over the measured profile interval.
Cumulative `syncAccessibility` is 368.31 MB (42.01%), and `ScenePacket` is
209.85 MB (23.94%); these are sampled Go allocation estimates, not retained memory
or exclusive components. The existing encoding scratch leaves only 7.98 MB flat
in `encodeWithScratch`, so more encoding byte scratch is not the next priority.
Whole accessibility publication and scene capture are the next investigation
boundaries. The raw profiles and analysis commands are in `baseline-diagnostic/`
in the same archive. All diagnostic native children returned exit 0.
