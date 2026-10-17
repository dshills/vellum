# Grid intrinsic sizing and native comparison

Source `599442af1bf037abc75ed90591ce71aac74f5849` preserves completed intrinsic
sizing for eligible passive grid labels when only fractional-scale allocation
rounding changes. The source passed full driver root offscreen race tests,
driver vet, consumer race/vet and compile-only checks before commit. Native
results below have their own matched executable provenance and review.

## Source behavior and regression evidence

The predicate requires equal natural dimensions, the same scene and plain Text
leaf, existing shaped content, no queued dirty work, exact WhiteSpacePre,
zero Grow/Min/Max, no interactive abilities, children, Parts or shadows. A fixed
44-value scan refuses allocation/parent-dependent, custom, unknown and nonfinite
units. Existing content, state, parent, visibility and pending work retains
ordinary sizing. Pane positioning still supplies rounded final bounds and clips;
full-layout, editor, theme, scene, configuration and DPI fences remain effective.
No new cache or retained widget pool is added. See the
[contract](../../../specs/grid-native.md#reusing-intrinsic-grid-text-across-rounded-cell-edges).

Seventy mixed-text comparisons span DPI96/120/144/162/192, fixed/variable rows,
scroll/cursor steps, captured pixels, public geometry/semantics and private
intrinsic/line/run/rune geometry against full layout and the frozen prior rule.
Twenty real dependent-style reference cases preserve unsupported behavior. A
size-sensitive shaper proves that lazy allocation-shaped paint cannot replace
pending intrinsic work. Deliberate missing-font-unit and lost-dirty-work faults
fail actual glyph/offset geometry; the prior rule fails useful reuse. The first
pending-work fixture failed because it had not drained its owner queue; that
failed receipt is retained alongside the corrected passing fixture.

Four serial ABBA offscreen kernel processes measured owner dispatch, model
scroll, bounded projection sync, eligibility and local layout at DPI162. Median
fixed-row time was 818,550.5 to 274,801.5 ns; variable-row time was 706,803 to
279,828.5 ns. Allocation observations fell from 16,026 to 3,662 and 13,402.5 to
3,751 per operation, respectively. These kernels omit paint, capture and native
frames. They qualify neither retained/native memory nor whole-machine isolation.

One complete Prism invocation reviewed the exact four-file 29,772-byte payload
using Anthropic `claude-sonnet-5-5`. Its raw medium finding and child exit 1 are
retained: it suggested a possibly missing `math` import, already present at line
5 in both original and candidate. Internal review independently refuted it;
compiled production checks confirm the import. No source/test delta or second
provider call followed. Internal checkpoint and commit reviews are separately
bound. This evidence-only report introduces no runtime change.

## Native execution and scope

Fresh builds use clean source `599442af1bf037abc75ed90591ce71aac74f5849` with 4,683 tracked and 6,249 prepared files, unchanged and equal in both checkouts. The baseline is a compiler overlay restoring only the exact `99b632b` grid sizing file, with all other current source unchanged; it is not a clean historical build. Both metadata records identify the current commit, and the overlay has its own complete source vector and argv.

Baseline executable SHA256 is `d8fb48059ae70efc854c09b7c51ddc1053f43d7983d2bcc391788c0379a02628`; candidate is `9f7a951c4428fcc010402daf0f9f2b5b52f65bf47f96c5e2523b2957ffca31ee`. The earlier uncommitted compile is separately identified in the source checkpoint and is not used here. All 43 native processes exit zero, complete their logs and have their owned groups observed absent. The campaign is sealed. Native endpoint data is retained in structured results.

One million-row Cocoa fixture verifies a 400-visible-cell minimum, overlap
identity reuse, replacement retirement, native cell press delivery and
closed-attachment refusal. Controlled native selectors/actions remain separate
from physical pointer events, external notification delivery and human
accessibility. No desktop input events or input-source, VoiceOver, clipboard or
OS preference changes are generated; fixture geometry persistence is disabled.

The forty ordinary processes comprise five adjacent alternating AB/BA pairs for
each CPU/GPU-requested fixed/variable group. Each has 128 warmup and 120 measured
callbacks. All 4,800 measured tickets and 2,400 full geometry pairs must reconcile
exactly. Endpoints establish matching sampled geometry, not continuous global
host isolation or absence of transient display changes. GPU requests do not
establish native GPU execution or presentation. Timings report TotalUIWork
from measured owner phases; presentation latency remains outside this metric.

| Group | Prior median process p95 (ms) | Candidate median process p95 (ms) | Candidate p95 lower in pairs | Candidate meets 4 ms p95 | Candidate meets 6 ms p99 |
|---|---:|---:|---:|---:|---:|
| cpu-fixed | 7.065 | 6.068 | 5/5 | 0/5 | 0/5 |
| cpu-variable | 7.315 | 6.406 | 5/5 | 0/5 | 0/5 |
| gpu-fixed | 6.239 | 5.490 | 5/5 | 0/5 | 4/5 |
| gpu-variable | 6.404 | 5.709 | 5/5 | 0/5 | 4/5 |

Candidate TotalUIWork p95 is lower in all twenty matched ordinary pairs.
Median paired reductions are 14.2% for CPU fixed, 12.3% for CPU variable,
11.4% for GPU-requested fixed and 10.4% for GPU-requested variable. These
are scoped observations from this matched workload and campaign. All twenty
candidate processes still miss 4 ms p95. Eight meet 6 ms p99 (four of five
in each GPU-requested group); twelve miss it. All twenty baseline processes
miss both targets. Neither the kernel savings nor this improvement closes
the frame-budget, broad no-regression, native-memory or full Mac gates.

Two separate GPU-requested fixed-row allocation diagnostics complete 512 measured
callbacks each and match 512 paired full geometries/endpoints. Before/after
profiles are retained with their hashes; this report does not interpret them as
exact allocation savings, ordinary latency or retained/native/GPU memory.

Owned builds finish before native timings. The guarded serial controller checks
universal/local stops and a shared lease, bounds each child to 180 seconds and
8 MiB per stream, requires terminal EOF and absent owned process groups, and
seals the campaign. Its three finalization and twenty refusal controls precede
launch. These controls and source/endpoint checks do not prove exclusive use of
the machine.

P4/P6/P9/P11, broader workspace/control/canvas performance, memory/endurance,
physical input, human IME/VoiceOver, SDK closure and release/distribution gates
remain separate. No full macOS qualification, signing, notarization or public
release is claimed. See [structured results](result.json),
[manifest](evidence.manifest.json) and [raw evidence](evidence.tar.gz).
