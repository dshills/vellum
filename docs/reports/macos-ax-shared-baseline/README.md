# Shared immutable accessibility wire baseline

Owned ready grid publications now compare wire data through ordered references
into their existing detached per-grid cache records. Ordinary controls retain
separately detached records. This removes the second full flat grid-node copy;
each grid still owns independent storage, so a pending survivor cannot retain a
removed sibling. Wire null/empty actions, callback authority, pending proof,
generation and metadata replacement rules remain unchanged. Borrowed/custom
publications, epoch exhaustion and closure clear both baseline forms.

## Results

The local changed-publication ownership benchmark uses five observations per
variant, Go 1.27.2, CPU2, 200 ms minimum, no race or profiler. Median flat versus
reference results are 462.620 / 450.481 us, 864,352 / 669,253 B and 847 / 842
allocations: about 2.6% less time and 22.6% fewer allocated bytes. It measures
fresh transferred containers, semantic encoding, preparation, baseline retention
and envelope construction. It excludes actual traversal, full update bookkeeping,
C/native queue admission/commit and full owner refresh.

Five interleaved native pairs request GPU scene mode with 10k fixed rows,
128 warm events and 120 measured events per process. GOMAXPROCS/GOMEMLIMIT are
unset; optimized ARM64 consumers have no race detector, profiler or compiler
overlay. Each paired geometry check passes: 400 nonempty clipped cells, 50 rows,
eight columns, 477 retained cells and nine retained headers, scale 1.6875,
1519x1266 client on the LG HDR DQHD / M4 Pro host.

| Run | Baseline p95 ms | Candidate p95 ms |
| --- | ---: | ---: |
| 1 | 5.544834 | 5.518667 |
| 2 | 5.438708 | 5.471291 |
| 3 | 5.487500 | 5.448165 |
| 4 | 5.382293 | 5.401417 |
| 5 | 6.953583 | 6.977999 |

Median run-p95 is **5.488 / 5.471 ms**, with mixed paired results and no
meaningful latency improvement established. The slower fifth pair remains in
both sets. Every p95 exceeds the unchanged 4 ms requirement; the candidate fails
the 4/6 ms p95/p99 pair. Median per-run median process-wide Go allocation
observations are **7,822,800 / 7,388,632 B**, about 5.5% lower. These observations
include helpers/runtime activity and do not qualify exact framework allocation
or retained Go/native/GPU memory.

Adopt this checkpoint for fewer ownership copies and allocated bytes only.
**No Mac acceptance gate closes.** The first workload still fails, so the broader
table/tree/application matrix is not rerun. GPU mode requests do not establish
accepted routes or presentation timing. No physical pointer evidence is claimed.

## Correctness and review

Focused race tests and the complete driver offscreen race suite pass (280.692 s
command duration); vet and optimized native build pass. Tests compare frozen
original preparation and pending-proof implementations, interleaved multiple
grids and ordinary records, null actions, repeated pending transitions, refusal,
borrowed/custom publication, callback/metadata replacement, sibling removal,
direct preparation and closure. Both baseline fields must clear when revoked.

The first focused run exposed two tests coupled to the previous private storage
representation; they now assert logical baseline or both-fields cleanup as
appropriate. A later new repeated-pending fixture incorrectly used input order
rather than the original preparation's emitted grouped order. The corrected
fixture follows the actual emitted baseline and passes. Raw failures are retained;
no native comparison samples or reference output were rewritten.

Internal review approved the implementation, requested direct frozen pending
proof coverage and clarified the spec lifetime wording; both were addressed.
One Prism Anthropic Claude Sonnet 5.5 call reviews the complete eight-file,
31,761-byte source/test/spec diff in one chunk, with no truncation or cache hit:
**zero findings**, including discarded/suppressed findings. No unchanged provider
re-review was requested.

## Identity and remaining work

Source base is clean f641776. The candidate is that base plus the recorded
working-tree diff; its binary SHA is
478cedd825bbcc48709fe45210c448588b72fa82ce3df41b4485ed1293d745ba.
The baseline reuses the verified stage0078 runtime binary
29aa5cc0e9018916bfffa4b92900a7103d9ae239c596e6e5f411cd0886963ad3;
Go driver bytes are unchanged from ea20486 through f641776, and all 6,253
prepared files match its receipt before/after. This is not a fresh clean
historical build. No prepared source changes are introduced.

Commands, raw outputs, source diff, identities, native samples and review records
are indexed in records.json and retained in records.tar.gz. No SDK or executable
is copied into this report. Native execution used a fresh stop marker and shared
lease. The runner observes direct-child exit/EOF and attempts group cleanup;
a later endpoint process query found no matching benchmark executable. Strict
per-PID/PGID whole-group absence was not recorded. No system preferences, input
sources, clipboard or VoiceOver settings were changed.

M3 needs a larger reduction in full refresh/scene rendering work, not further
small byte-encoding or accessibility-copy tuning. The next implementation is a grid-scoped incremental accessibility publication
path: separate reusable semantic records from changing frames/generations, admit
atomic scope updates, and retain a full-publication fallback. The existing ordinary
geometry cache excludes grid scopes, so this needs explicit incarnation, membership
and action-authority rules. Unchanged clipping, geometry, frozen-pane, selection,
editor and stale-input contracts must remain testable. Establish correctness and
one matched comparison before expanding qualification; scene reconstruction remains
the other measured substantial cost. M2 native integration, M3, M4 memory/endurance,
M5 human/display, M6 rights/final SDK and M7 final audit remain open.
