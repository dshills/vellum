# macOS grid accessibility publication checkpoint

Managed grids now keep their last completed accessibility subtree readable while
new layout is pending. Pending objects refuse actions; a completed frame publishes
matching cell semantics and bounds together. Overlapping cells retain native
identity within the same live binding. Replacement, reparenting and actual removal
retire old objects. Race, native and consumer checks pass. The pre-correction performance
campaign was refused for unequal rendered workloads, so this checkpoint makes no
latency improvement claim. Mac completion and the unchanged 4ms p95 / 6ms p99
targets remain open.

## Behavior and implementation

At `f68941b`, a scroll's pre-frame refresh could publish an empty grid subtree,
retiring visible Cocoa objects before completed layout republished them. This is
a source-backed control-flow diagnosis; historical CPU samples do not isolate
asynchronous Cocoa parsing or establish a promised saving.

The owner now stamps each visible grid scope with binding incarnation, target and
completed generation, pending state and copied completed pixel frames. A detached
cache holds one completed subtree per currently visible scope, retaining no model
or action callback. Pending publication removes focus, text-client authority and
all actions. Unrelated controls and other ready grids publish independently.
Namespace, parent, incarnation, visibility and close boundaries clear old cache
eligibility. Startup without a completed frame publishes only a disabled host.

Queued actions recheck owner lifetime, current model generation, measurement
incarnation/epoch, viewport/stamp and private editor session/text. Eligible stale
cell actions preserve the existing `ErrGridStale` error callback; pending, retired
and geometry-only refusals remain quiet. Geometry withdrawal requests publication
once. Same-generation geometry or editor changes invalidate completion, while an
unchanged frame ordinal does not. Typed scalar, cell/header and editor comparisons
cover completed publication data without a reflective traversal in capture.

Cocoa validates the entire scoped envelope before commit. Native objects reuse
only matching live identifiers, scope/incarnation and parent for grid scopes.
Ordinary unscoped controls preserve their existing identity and cached geometry
when reparented. Pending state also
refuses native action and focus selectors independently of Go. Scoped pixel frames
commit with semantics; legacy unversioned geometry updates cannot overwrite them.
Transport retains copied UTF-8 data directly. Detailed native text-range geometry
keeps its separate client/text contract. Opaque object serials support inspection
without exposing pointer addresses.

## Verification

Final source-bound receipts have equal before/after maps: 642 selected source and
specification files, and 6,139 prepared dependency files. Ten checks pass: portable
race, full offscreen/native driver race, 17 new test groups each run three times,
offscreen/native vet, loaded-grid/grid/replay consumer race, and loaded-grid vet.
The new groups cover detached caches, mixed scopes, startup and retirement,
namespace/lifecycle changes, queued actions, editor previews, geometry withdrawal,
actual completed cell/header/row frames, malformed native transactions and the
production Go JSON encoding. Disabled workbench encoding is tested through the
real semantic wrapper; it does not claim full Manual Extension routing.

Three faulty source overlays fail behavioral assertions: missing completed-subtree
retention, obsolete private-editor actions, and silent geometry withdrawal.
An isolated source copy restores the original Objective-C parent predicate and
fails the new ordinary-control reparent identity/frame assertion. The earlier
Go `.m` overlay exited zero and is preserved as a refused negative proof; it did
not exercise that altered native source. Production source remains unchanged
during all checks.

The optimized native artifact is bound to the same source/prepared map. Separate
stored-data acceptance processes query actual attached Cocoa objects with at least
400 clipped cells. They check values, indices, hit testing, pixel-to-screen cell
frames, six routed scrolls, replacement, exact owner selection after a native press,
and attachment refusal after close. Fixed-10k runs retain 2,268 overlapping cell
identities and replace 459 scoped identities each. The variable-million run retains
2,307 and replaces 468. Generic native accessibility smoke also passes. These are
in-process Cocoa assertions, separate from human VoiceOver acceptance.

The first full gate caught suppression of three existing stale-cell error tests.
That failed receipt is preserved; the guard and narrow new test were corrected
before the final full gate passed. Earlier fixture/setup failures, a source-drifted
native interval and superseded review maps remain historical evidence.

## Refused performance campaign

The pre-Prism-correction optimized artifact ran ten fresh ordinary processes: five fixed-10k and
five fixed-million, each with 120 samples after 128 warmups. Individual native
record validation and child exits pass. The parent campaign exits 1 and records
`FAILED: 10k/million rendered workload differs` before launching either variable
case. Fixed pairs 1–3 match all clips; pairs 4–5 differ at all 120 samples each.
Bounds, scale and column count match, but starting vertical offsets differ:
43 versus 12 in pair 4, and 12 versus 33 in pair 5. These yield 50 versus 51
visible rows and 400 versus 408 cells. The cause of the offset discrepancy is
unproven. The current corrected acceptance artifact has a separate recorded
binary hash and has not received a new performance campaign. No runs are removed or selected, and no retry is represented as a pass.

All 1,200 observations, outliers, refused summary and raw child records are
retained. They do not establish matched four-case performance or a speedup.
The latest qualified preceding checkpoint still misses the unchanged targets.
Full P6 table/tree, memory, cold-cache and endurance qualification remains open.

## Evidence and review

[records.tar.gz](records.tar.gz) preserves final and historical receipts, exact
logs/maps, faulty overlays, inert source/producer copies, diagnostics and internal
reviews. [records-index.json](records-index.json) binds regular-file membership;
[audit.json](audit.json) verifies current source, prepared inputs and evidence
linkage. Binary identity is recorded without claiming full compiler/cache/SDK
closure or a second canonical native measurement.

Independent reviews approve the exact 17 source/specification/document inputs and
audit the final checks separately. The significant 18-file source/test/spec/report
batch receives one Prism review using Anthropic Claude Sonnet 5.5. The
[review records](review/process.json) preserve its payload, exit 1 and three
medium findings. Two were corrected and independently reviewed/tested; the
cache-lock finding is refuted by the helper's existing mutex and closed check.
Local triage and supplemental review bind the resulting source; no second
provider review or empty-findings claim is made. The public namespace remains `github.com/dshills/vellum` under MIT. No
public release or operating-system preference change is part of this checkpoint.
