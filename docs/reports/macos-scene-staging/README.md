# Scene staging: offscreen checkpoint

Stage 0044 reduces scene-construction copying with private 64-operation blocks
and flattens once at seal. It also reuses tile-count scratch after writing the
output row table. Packet ownership, painter order, logical limits and wire bytes
stay fixed. This is an offscreen checkpoint; Mac acceptance remains open.

The operation limit is checked before allocating another block. Per builder,
at most 1024 blocks exist and the last has at most 63 unused slots. Successful
and refused seals retire all construction references. Refused span streams skip
compaction, with byte admission after spare-header refund. Retained operations
have exact capacity. Block metadata, spare slots, flattening overlap and encoder
scratch are excluded from the logical packet payload limit.

Checks passed: full sceneplan/rasterx/gpuscene offscreen race; root and driver
offscreen race; root, driver and prepared-renderer vet; preparation tests; and
fresh pin-verified preparation. Eleven encoding digests captured from unchanged
fdb67bc match, including block boundaries through 65,536 operations and mixed
image/mask/span payloads. A separate tile-outer oracle checks row counts and
painter-ordered references. Existing randomized image/draw source-mutation and
exact scanner arithmetic tests remain. New checks cover exact byte/operation
boundaries, reset detection and both retirement paths.

Local offscreen allocation fixtures ran five times per revision:

| Fixture | Baseline B/op | Candidate B/op | Allocations/op |
| --- | ---: | ---: | ---: |
| OperationStaging/512 | 198948–198962 | 133601 | 529 → 527 |
| OperationStaging/8192 | 5726753–5727139 | 2064867–2064883 | 8220–8222 → 8327 |
| TileScratch | 393414–393427 | 262341–262343 | 7 → 6 |
| SpanBandsTallTarget | 874680–874711 | 801106–801111 | 12413 → 12410 |

Operation fixtures include synthetic uniform source creation and sealing.
The 8192-operation fixture allocates fewer bytes but more objects because blocks
replace geometric slice growth. TileScratch uses a 2048×2048 target and one
small operation; SpanBandsTallTarget uses the existing scanner-band fixture.
These numbers describe Go allocated bytes/objects in these fixtures, not whole
frames, peak/retained/native/GPU memory or UI responsiveness. Wall timing is
excluded because the computer is shared and live comparison is deferred.

Internal source review approved the implementation. One Prism review used
Anthropic Claude Sonnet 5.5 for nine source/pin/test/contract files, with complete
coverage and no reported findings. Runtime evidence was excluded from provider
inputs. Pin reproduction changes exactly three prepared files; the complete
prepared set contains 6172 files.

Two development errors were corrected: a test initially compared a struct
containing a slice, and the staging producer initially confused grouped stage
count with patch number. The failed compilation and the preparation sequence
are retained in evidence; the final tests and flattened pin-order guard pass.
Before promotion, internal review strengthened the proposed Git parent guard
to require exactly one parent. The original unexecuted proposal and both
wrapper reviews are preserved.

Live windows/input, actual native GPU parity and full-frame allocation/latency
checks are deferred until the user explicitly says the computer is free. No OS
preferences were changed and no public release or publication ran. Prior warm
performance failures and other Mac completion gates remain unresolved.

[Checkpoint](records/checkpoint.json), [archive index](records/archive-index.json)
and [raw evidence](records/evidence.tar.gz) bind the source, tests and local
allocation observations. This report does not qualify broad native acceptance.
