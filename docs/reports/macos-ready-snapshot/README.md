# Exact ready accessibility snapshot reuse

The owner now checks an unchanged ready accessibility snapshot against its
existing detached, validated wire baseline before encoding it again. Exact
membership/order, every wire field, builtin scalar types/values, pointer-field
contents and the nil/empty action distinction must match. Builtin eligibility
stays false once any custom/mutable input is encountered, with checks
before/after both raw and restored-pending encodings. The completed-cache
semantic identity must remain current, and Go-only cell focus metadata is
checked separately. A match replaces the current handlers and activation guards
without retiring semantic generation or completed storage.

Changed, pending, borrowed, custom/mutable and uncertified inputs keep the
existing validation/encoding/preparation path. Empty snapshots use that path to
preserve `null` versus `[]`. Direct preparation invalidates certification;
closure clears the existing baseline and cache. No new retained cache, native
transport format or public API is introduced.

This addresses remaining duplicate encoding visible in current source. The
[older native profile](../macos-current-grid-profiles/README.md) identified
accessibility snapshot work, but subsequent storage/sizing fixes already
changed that baseline. The local experiment below measures this new path
directly; it does not attribute older native samples to the new source or
qualify a native frame improvement.

## Local measurement

The existing `BenchmarkAccessibilityDuplicatePublication` executes offscreen
without a native window. Its fixture now seeds the detached last-published wire
baseline that ordinary successful publication retains. Both variants use that
same fixture and pinned dependencies. Both variants are working-tree
experiments based on `3044c77`, rather than clean historical builds. Baseline
production source matches that commit; the candidate changes only the
production leaf `drivers/cogent/accessibility.go`. New regression declarations
do not run during the benchmark. Each sequential variant process takes five
observations per case, 200 ms minimum per observation, with Go 1.27.2,
`-cpu=2`, `GOMAXPROCS=2`, no race/profile flags and lower process priority. Raw
argv/source hashes and complete stdout/stderr are in `records/`.

Medians (time in microseconds; B/op are Go benchmark allocation observations):

| Fixture | Prior time | Candidate time | Prior B/op | Candidate B/op | Prior/candidate allocations |
| --- | ---: | ---: | ---: | ---: | ---: |
| Host only, owned | 1.146 | 0.285 | 1,136 | 704 | 11 / 7 |
| 40 cells, owned | 38.224 | 6.317 | 33,708 | 19,144 | 98 / 54 |
| 400 cells, owned | 365.782 | 58.320 | 340,125 | 185,896 | 825 / 420 |
| 400 cells, borrowed control | 420.807 | 425.996 | 675,631 | 677,030 | 846 / 846 |

The 400-cell owned duplicate median is 84.1% lower and allocates about 154 KB
less per operation in this local fixture. Borrowed timing is 1.2% higher in
this sequential comparison; a general no-regression claim is not established.
All other raw observations remain available, including allocator variation. The
comparison uses one baseline and one final candidate process, with the prior
process first; the earlier candidate process is separately retained. It is not
an interleaved native comparison, process-wide memory result, whole refresh
benchmark or full Mac performance qualification. The fixture still allocates
its fresh input nodes/handlers; those costs remain included.

## Correctness and review

The focused accessibility offscreen race run passes before the final guard test
addition. Existing coverage includes pending/completed publication, every wire
field, typed/custom/mutable values, direct preparation, duplicate callbacks,
detachment, stale action refusal and closure. New regressions check atomic
refusal for nonfinite scalar/frame/bound values, unsupported values and
malformed JSON, plus empty/nil semantic generations.

The reviewer-requested guard regression uses real admission and owner drain: an
action admitted under the old false guard stays refused after a duplicate
refresh; new true-guard admission succeeds; a false replacement refuses again;
removing the guard permits a new action. Semantic generation and
completed-cache storage remain unchanged throughout. The final complete driver
offscreen race suite passes in 296.564 seconds and vet in 3.021 seconds on the
final source. The fixed-runtime focused run passes 62 top-level accessibility
tests before a final test-only strengthening; that strengthening is included in
the full run.

A draft shortcut fails the retained custom-encoding counterexample: the
borrowed control passes while the owned path incorrectly preserves stale wire
bytes. The corrected source keeps builtin eligibility sticky across both
encoding points. The final regression also proves ordinary custom output was
published, left no reusable certification/baseline, and a fresh builtin
description advances the semantic generation. Exact failing source/test bytes
are retained as gzip. The initial broad success is not substituted for the
final-source suite.

Internal review approves the final design, regressions, contract and scoped
local adoption. One Anthropic Claude Sonnet 5.5 Prism call covers four files /
12,928 diff bytes with complete coverage. Its raw medium finding and CLI exit 1
remain retained. Independent review refutes its hypothetical
aliasing/concurrency concerns: every audited wire field and nested element is
compared by value, Go-only focus metadata is separate, and `b.mu` protects the
closed check through handler/guard assignment. Closure uses the same lock and
clears certification. Owner updates remain serialized. No source change or
second provider call follows that review. See `records/review-assessment.json`
for the exact disposition.

## Remaining Mac work

M3 remains incomplete. The existing ordinary frame-budget failures are
unchanged acceptance outcomes. A quiet desktop comparison on the current
candidate must measure the actual table/tree and composed UI workloads against
their original thresholds. Native input/accessibility, presentation, display
transitions, native/GPU memory, integrated endurance and final SDK/release
gates remain open. No native application, physical input, system preference
change or public release is part of this checkpoint.

The [subsequent native comparison](../macos-encoding-scratch/README.md) fails the
unchanged target under both the controlled two-processor and normal runtime.
The local duplicate result above does not translate into a demonstrated native
frame improvement.
