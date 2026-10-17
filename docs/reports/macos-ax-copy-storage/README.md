# Detached accessibility storage and native comparison

Committed source `3a14e53777072caf19f5bcf4ce3e27f12c1e8a9a` passes scoped Cocoa
grid correctness and local allocation regressions. Fresh native measurements
establish no latency improvement: all 20 candidate processes miss 4 ms p95 and
19 miss 6 ms p99. Candidate fixed-row CPU p95 is higher in all five pairs.
Performance, native memory and full Mac qualification remain open.

## Source and review

Batch snapshot copies use fresh, exact-counted typed backing slices for mutable
fields. Input/output, cross-node and later-snapshot mutations stay isolated;
Actions cannot append into neighboring records. Optional pointers, exact float
bits, arbitrary-value admission, pending null-versus-empty actions, callback
retirement and removal/close retain their existing contracts. There is no
cross-snapshot pool. See [the contract](../../../specs/accessibility.md#detached-snapshot-field-storage).

The complete affected driver root package's offscreen race suite and vet pass
before a later test-only clarification. The final copy regressions and vet pass
after that clarification. A compiler-overlay control builds the exact prior
copy and fails the allocation-bound assertion at 193 allocations for 32 fully
populated nodes; its actual exit 1 is retained. Five-observation local 400-node
benchmarks show 2,401 to 7 copy allocations, with 223 additional bytes per
operation. The 400-cell ready publication fixture shows 815 to 15 owned-path
allocations and 820 to 20 borrowed-path allocations, each with 224 additional
bytes. Those benchmarks report 14 CPUs despite the child GOMAXPROCS2 request;
no exclusive-core, native latency or retained-memory claim follows.

One complete Prism call reviewed the three source/spec files using Anthropic
`claude-sonnet-5-5`. Its two raw medium findings and child exit 1 are preserved.
Internal review added explicit multi-node all-nil and mixed nil/empty Actions oracle cases,
while confirming existing behavior was correct. The CI omission finding was
refuted by the macOS workflow's explicit offscreen race/vet command. Only tests
changed afterward; final test-only checks and commit review remain separately
bound. There is no zero-finding claim or second provider call.

## Native execution

Fresh release builds use the same clean `3a14e53` checkout and prepared source.
The baseline is a compiler overlay restoring the exact `5ba6e86` accessibility
copy file, with all other current source unchanged. It is not a clean historical
executable. Both build metadata records identify the current commit; the overlay
has its own complete source bytes, map, hash and build argv.

Baseline binary SHA256 is
`156f2e96df129896d80695d093897dc12715b2fa7a8a0d91bb4051b893e5bb23`;
candidate is `526a9a4852e5a6817c959a874a8e2933b24756045c5cc86dcce6b505044e71f9`.
The 4,669 tracked and 6,246 prepared files in both checkouts remain unchanged.
The source checkpoint's earlier uncommitted compile is separately retained and
is not used for these comparisons.

All 43 native processes exit zero, reach EOF on both streams and have their
owned groups observed absent before the next dispatch. The shared lease covers
execution and cleanup. The controller uses universal/local stop guards,
180-second watchdogs, 8 MiB per stream, fresh profile/output paths and independent
cleanup/finalization. Three finalization cases and 20 refusal controls pass
before launch. The campaign is sealed afterward. No desktop pointer/key events
or input-source, VoiceOver, clipboard or OS preference changes are made;
geometry persistence is fixture-disabled. Owned builds and offline analyses do
not overlap ordinary timings. Continuous global host isolation is not proven.

The candidate's actual million-row Cocoa fixture passes the 400-visible-cell
minimum, 2,268 overlap identity-reuse observations, 459 replacement retirements,
native cell press delivery and closed-attachment refusal. Reuse observations
are not distinct-object counts. Controlled selectors/actions are separate from
physical input, external notification delivery and human accessibility.

## Ordinary frames

Five adjacent AB/BA pairs cover each CPU/GPU-requested fixed/variable million-row
group, with 128 warmup and 120 measured events per process. All 4,800 tickets
complete exactly once, with no missing, dropped or pending records. All 2,400
paired full geometries and native endpoints match. Client/render size is
1519×1266 pixels, recorded DPI162/DPR1 on the same screen; the logical grid is
767×624 and meets the 400-cell/50-row/eight-column minimum. Endpoints do not
prove continuous absence of a transient display change.

| Group | Prior-copy median p95 (ms) | Current median p95 (ms) | Over 4 ms p95, prior / current | Over 6 ms p99, prior / current | Current p95 higher in pairs |
|---|---:|---:|---:|---:|---:|
| cpu-fixed | 6.661 | 6.767 | 5/5 / 5/5 | 5/5 / 5/5 | 5/5 |
| cpu-variable | 6.892 | 6.830 | 5/5 / 5/5 | 5/5 / 5/5 | 1/5 |
| gpu-fixed | 5.884 | 5.882 | 5/5 / 5/5 | 5/5 / 4/5 | 3/5 |
| gpu-variable | 5.996 | 5.986 | 5/5 / 5/5 | 5/5 / 5/5 | 3/5 |

Fixed-row CPU median p95 is about 1.6% higher, with all five pair ratios above
one. Other groups have mixed paired directions; GPU median changes are close to
zero. These results establish neither a latency gain nor broad no-regression.
The 4/6 ms targets remain failed. GPU requests do not prove actual GPU execution
or presentation for ordinary frames.

## Allocation diagnostics and remaining work

Two separate GPU-requested fixed-million-row processes record 512 measured
callbacks after 128 warmup events each. All 512 paired full geometries/endpoints
match. Each writes exclusive before/after cumulative allocation snapshots,
forces two GC cycles per snapshot outside recorded events and disables CPU
profiling. The runtime writer caps each file at 64 MiB. These runs do not qualify
ordinary timing.

Eight bounded offline commands read the four profiles using the retained,
source-bound Go cmd/pprof tool with `-symbolize=none`. Analysis does not execute
an application binary or request remote symbols. Reported process-wide totals
are 22,196,348 versus 20,469,999 object estimates and 4274.56 versus 4226.61
reported MB: about 7.8% lower estimated allocated-object count and 1.1% lower estimated
allocation space in this one pair. Observers and other goroutines are included. MB totals
are rounded; object totals are also sampled estimates. Cumulative rows overlap;
missing top/threshold rows are not zero. These observations prove neither exact
per-frame savings nor retained UI/native/GPU/private memory budgets. The local
microbenchmark's small byte increase remains part of the record.

The source change is retained as bounded detached field copying with scoped
correctness and allocation evidence. Frame budgets still require further work;
allocation counts cannot substitute for that acceptance. P4/P6/P9/P11, broader
workspace/memory/endurance, physical input, human IME/VoiceOver and release/
distribution gates remain open. No public release, signing or notarization occurs.

See [structured results](result.json), [manifest](evidence.manifest.json) and
[raw evidence](evidence.tar.gz). [Earlier profile attribution](../macos-current-grid-profiles/README.md)
retains its separate SDK source and diagnostic scope.
