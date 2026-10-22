# Private completed staging: native byte and lifetime result

The corrected **off/on pair passes** all declared byte/lifetime checks. Both modes
verify 22 full GPU readbacks (48,956,350 bytes each), across 24 CPU texture writes.
Enabled mode exercises six completed-buffer checkouts, five FIFO evictions and an
eight-entry idle peak. This supports the private prototype's next timing check;
no shipping policy or UI responsiveness gate closes.

| Mode | Fresh eligible staging buffers | Completed-buffer checkouts | Evictions | Peak idle entries | Peak idle bytes |
|---|---:|---:|---:|---:|---:|
| Off | 22 | 0 | 0 | 0 | 0 |
| On | 16 | 6 | 5 | 8 | 7,692,572 |

Every fresh eligible buffer has one matching returned backend destruction call,
and both Core device scopes end per mode, with no live/cached lease left. The
scalar source-bound journal has 92 events off and 111 on. The summary separates
raw-buffer resource IDs from per-checkout leases: earlier leases normally have no
destroy marker when the same resource is reused. Each resource-lifetime row binds
its unique fresh event, all acquisition calls and the one final destruction. This is not physical
resident-memory reclamation or accounting for every native GPU resource.

All requested transfers compare every unpadded byte after successful GPU mapping:
different small RGBA/BGRA and large patterns, unaligned rows, padded/offset source,
large above-cap input, size changes and a second device. Three same-size uploads
coexist before a single submission, use distinct leases and preserve all three
payloads. This proves pending separation, not forced physical GPU overlap. R8
readback succeeds outside reuse; zero-area success acquires no staging, and an
unsent upload is destroyed without cache admission. Original shared/write-combined
storage and cache policy stay fixed in both modes. Public mapped staging and
write_buffer remain on their original allocation path; those paths are excluded
by source eligibility, not claimed as a new full native regression campaign.

The 22 October 2026 campaign runs 00:05:15.179–00:05:19.767 UTC
(21 October, 19:05:15–19:05:19 Chicago), 4.587718 seconds within 90. All 18 owned
native/probe process groups retire. Source/library/binary/log pins, configuration,
persistent geometry and immutable SDK inventory (8,321 files) pass. Primary
AdapterInfo.Device identifies Apple M4 Pro, Metal backend; both optional vendor/
description strings are blank. No window, surface, input or settings changes.

The first scope remains rejected: it used Core's derived display name, which
ignores the primary Device field. Its successful off process and posthoc-only
synthetic-name replay are retained, never accepted as native pooling proof. The
fresh scope follows the concrete, reviewed diagnostic metadata fix 593bbb9, with
the same native library, matrix and bounds. Enabled mode never ran in that first
scope. Backlog the harmless Core display-name fallback; framework scope stays frozen.

The [protocol](../../../tools/macos-calibration/stagingreuse/PLAN.md),
[initial build identity](build-identity.json), [corrected identity](metadata-fix-build-identity.json),
[source audit](source-audit.json), [exact result](byte-result-summary.json) and
indexed archives retain original source snapshots, failures, raw journals and
all review dispositions. Actual private policy5, protocol/lifecycle9 and scalar
concurrency1 checks pass; earlier compiler rejection controls are reused. Complete
Prism Anthropic Claude Sonnet 5.5 reviews cover preparation (15 files), ownership/
snapshot corrections (four) and the metadata fix (five). The last empty-name
finding is handled before writes. Unexpected poison/unwind, concurrent configuration,
physical device faults, other hardware and general GPU allocation balance remain
outside this bounded protocol.

Next prepare/review one bounded framework off/on timing pair using the same native
prototype, original cache policy and existing source/actual-presentation checks.
Headless bytes do not prove displayed pixels, ordinary binary equivalence, observer
neutrality or the unchanged 33.3 ms receipt-to-actual-presentation requirement.
No SDK, app, shipping library, renderer/pacing option or feature changes; M3–M7 open.
