# macOS native accessibility publication checkpoint

Managed grid frames continue to publish atomically with their semantics. Native
render callbacks now omit a second grid geometry payload that Cocoa already
ignored, while full manual inspection and editor text-range geometry remain
available. The native transaction reuses its validated node map and reads prior
scope ownership before mutation. Async accessibility producers and consumers
have explicit Cocoa pool lifetimes. Pooled p95 changes are mixed across these four matched cases; the distributions below are the observation, without a general speedup claim.
Every one of the 40 ordinary processes still fails both unchanged 4 ms p95 /
6 ms p99 UI-work targets. This checkpoint does not complete Mac or P6 acceptance.

## Implementation and scope

The preceding clean canonical checkpoint is `b995782`. The candidate changes
seven owned driver source/test files and two specs; pinned dependencies remain
exactly the same 6,140 prepared files. There is no new cache, partial transaction,
semantic delta protocol, notification policy or layout/rendering shortcut.
Native-only rectangle traversal skips a bound DataGrid's host, rows, cells,
headers and private editor. Workbench recursion carries the same policy through
active panel namespaces. Ordinary legacy rectangles retain their previous path;
the existing Cocoa guard still refuses all scoped legacy frame writes. Full
manual bounds and the separate exact-client/text geometry channel are retained.

Whole-snapshot validation returns its checked ID map for commit. Previous scope
host parents come from the existing old element dictionary during the complete
pre-mutation allocation loop. Epoch rejection, object identity, hierarchy order,
notification decisions and pending action refusal keep their existing contracts.

JSON producers copy C bytes and enqueue strongly capturing NSData blocks inside
local autorelease pools. Text geometry uses the same byte transport, avoiding a
string round trip. Action/text-query strings are copied inside producer pools;
consumer pools cover parsing, actions and queries. A delayed inert fixture
mutates/frees the original C buffer, consumes or discards the captured block, and
checks exact bytes, parseability, queued lifetime and weak-object release before
its outer pool drains. This proves those fixture lifetimes; ARC can eliminate
some factory autoreleases, so the change does not establish a prior leak size or
process/native/GPU memory reclamation.

## Verification

Eleven source-bound preparation/race/vet/build checks pass: preparation tests,
full offscreen/native driver race suites, three repetitions of publication,
geometry, payload and native envelope groups, both driver vet modes, loaded-grid,
grid and replay consumer race tests, and release/generic-smoke builds. The
accepted endpoint maps contain 653 selected source/specification files and the
same 6,140 prepared files. The baseline was freshly built from clean canonical
`b995782`, with unchanged 651/6,140 maps. A first baseline build attempt hit a
sandbox restriction on the default Go cache; the accepted second build uses the
controlled writable cache. It is retained separately as failed build evidence.

The new manual-scene test checks four logical DPIs (96/120/144/192), an ordinary
toolbar mixed with a Grid, a private inline editor, pending layout, scrolling,
disabled state and an active Workbench namespace. The semantic projection is the
independent omitted-ID oracle: native bounds must equal the full map minus scoped
IDs, and computing them must not mutate full inspection. The separate text
geometry retains the editor. Direct new tests do not exercise a Grid inside an
actual modal Scene, nested Workbench namespaces, a mixed DataTree or delayed
NSString action/query payloads; existing suites/native generic checks remain
broader regression evidence rather than claims of those exact new fixtures.

Three separate live Cocoa processes pass: fixed-10k and variable-million stored
400-cell accessibility checks, plus the generic accessibility smoke. They inspect
actual native values/indices, pixel-to-screen frames, hit testing, routed scroll
identity reuse, replacement, current native cell action and attachment refusal
after close. Candidate/generic binary hashes match the qualified build receipt
before and after every live child. No input source, VoiceOver or OS preference is
changed. These checks do not certify human screen-reader workflows.

## Ordinary frame comparison

Forty fresh ordinary processes are interleaved by repetition, five per build for
each fixed/variable 10k/million stored-table case. Each completes 128 guarded
warmups and retains 120 measured frames: all 4,800 observations, including
outliers. All paired geometry records are identical across builds and between
10k/million cases. Both artifacts have diagnostics disabled. The independently
reviewed byte allowlist matches the complete candidate source delta; the fresh
baseline and qualified candidate binary hashes are checked throughout. Pooled
nearest-rank UI-work times are in milliseconds:

| Case | Build | p50 | p95 | p99 | Max |
| --- | --- | ---: | ---: | ---: | ---: |
| fixed-10000 | baseline | 11.477 | 12.258 | 12.457 | 12.804 |
| fixed-10000 | candidate | 11.326 | 12.133 | 12.426 | 13.329 |
| fixed-1000000 | baseline | 11.174 | 12.121 | 12.663 | 13.634 |
| fixed-1000000 | candidate | 11.084 | 12.019 | 12.397 | 14.134 |
| variable-10000 | baseline | 11.574 | 12.356 | 12.888 | 24.540 |
| variable-10000 | candidate | 11.415 | 12.207 | 12.500 | 14.592 |
| variable-1000000 | baseline | 11.277 | 11.880 | 12.304 | 12.448 |
| variable-1000000 | candidate | 11.118 | 11.972 | 12.399 | 12.739 |

All 11 phase/queue distributions and per-process variation are retained in the
raw summary. Percentiles from separate phases must not be added. The comparison
is one host/workload observation and does not attribute a saving to one of the
three changes. A separate 512-sample diagnostic CPU profile motivated source
inspection; profiling is excluded from ordinary qualification, and Go CGO
samples do not reliably attribute asynchronous Cocoa/GPU execution. Physical
presentation, native/GPU allocation attribution, cold caches, sustained workloads,
P6 tree latency and complete Mac acceptance remain open.

## Evidence and reviews

`records.tar.gz` and its byte index retain the raw check/build/native/performance
records, all logs, source/prepared maps, producer entry copies, baseline failure,
separate diagnostic profile and internal reviews. `audit.json` binds the current
selected source, exact prepared files, qualified artifacts, reviews and archive.
An internal source review approves the nine-file implementation/spec batch with
no material findings. The final independent receipt review checks raw-record
statistics, geometry and source/binary links. One significant-batch Prism review
uses Anthropic Claude Sonnet 5.5; exact provider output/coverage is in `review/`.

Prism returned exit 1 for one medium conditional testing finding (confidence
0.3), `eb6ac100ffff8f63`: fatal assertions would run on the wrong goroutine if
`Application.Do` dispatched callbacks. Independent source review refutes that
premise for the current API and all three call sites: `Do` invokes the callback
inline on the owning subtest goroutine. The source, tests and specs are unchanged
after provider review. Its raw finding/exit code and the separate local
disposition are retained; the provider result is not represented as a
zero-finding pass. This paragraph is a separately reviewed documentation-only
addendum after Prism.
These selected endpoint bindings are not a whole compiler or SDK closure claim.
