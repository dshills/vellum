# Private completed texture-staging prototype: byte correctness first

This prototype lives only in a private clone of the CPU upload observer and in
these tools. No framework, SDK, shipping library or Placement app is modified.
Keep the original Metal shared/write-combined cache policy in both modes. All
package names/versions, lock entries, headers and native ABI remain unchanged
apart from diagnostic-only configuration/snapshot exports.

## Ownership and scope

Reuse is restricted to Metal, single-layer 2D RGBA8/BGRA8 uploads. Eligibility is
recorded per CPU texture-write call. Zero extents, other formats/dimensions/layers,
public mapped staging and write_buffer retain fresh allocation. The test host is
the pinned unified-memory M4 Pro; other hardware is not qualified by this prototype.

The existing staging object keeps the sole raw buffer and Arc<Device> until
successful submission tracking binds its scalar resource ID to the submission.
After the existing completion tracker drains that submission, retire its encoder
before explicitly considering staging for cache admission. Require matching
submission, completed fence/index and valid Device; never pool from generic Drop.
Queue shutdown's successful wait is also completion proof; device-loss cleanup
does not cache. Failed/unsent work keeps original destruction. The replay feature
is disabled in this native build.

Idle storage is per Device, exact size, at most eight entries and 32 MiB. Entries
hold raw Box<DynBuffer>, never Arc<Device>, so they do not create a Device cycle.
Checkout removes the sole idle owner before remapping. Oversized entries remain
fresh and are destroyed after use. Mapping failure destroys its fresh/completed
raw handle before returning the original error. Eviction and device drain invoke
the original backend destruction outside the cache mutex, before backend/fence
teardown. The raw-taken flag prevents a second destruction after explicit transfer.
All other validation, coherent storage, transitions, row/layer copying and pending
write ownership remain unchanged. The limits bound idle storage, not total GPU
memory or outstanding uploads.

The cache mutex is acquired under existing pending-write or completed-life locks.
While held, operations only move bounded entries and record scalar host clocks;
they never acquire those outer locks or call backend create/map/destroy. Device
teardown has exclusive ownership and drains through get_mut. Completed temporary
resources stay alive while their encoder is retired; other temporary destruction
is delayed until after that retirement, within the same completed-submission loop.
This ordering and all failure/teardown paths require review before execution.

A fixed 4096-row scalar journal records device scope, fresh allocation/checkout,
successful submission, completed retirement, cache/eviction, destruction and drain.
It retains no native handles or pointers. Record actual CPU-call ordinal at acquire;
bind every eligible upload to its CPU span. Refuse overflow, incomplete records,
unstable snapshots, unknown/duplicate leases, mismatched device/size/index, early reuse, failed caps or
unclosed resources. Destruction events mean backend calls returned, not physical
resident-memory reclamation or whole-GPU allocation balance.

## First native protocol

Build and review one headless Go/WebGPU executable using public NoDisplayGPU,
Device/Queue/Texture/Buffer APIs and BufferReadSync. No compiler overlay or window
is needed. The fixture has exactly 24 CPU texture-write calls per mode:

- Zero-area success, with no staging or transfer.
- Two distinct small RGBA patterns and a BGRA pattern, each 3×2.
- 7×5 unaligned rows, then 7×5 with 40-byte source rows and 16-byte offset.
- Two distinct 1519×1266 patterns, then 4096×2049 (above the 32 MiB idle cap).
- Three distinct 3×2 textures queued before one submission. Their pending leases
  must be distinct; verify every texture afterward. This forces queued ownership,
  not physical overlap of GPU execution.
- Nine exact-size changes, widths 2 through 10 at height 3, forcing entry eviction.
- A 3×2 upload on a second Device, with independent ownership and explicit release.
- A 3×2 R8 upload: full six-byte readback, eligibility false and no reuse acquisition.
- An unsent 3×2 upload discarded at queue/device release, never admitted to cache.

Every actual transfer (22 per mode) copies through GPU texture→map-read buffer,
waits for successful mapping and compares all unpadded bytes before unmapping.
Readback buffers, textures, devices, adapter and instance are explicitly released.
The zero-area and unsent calls do not claim a GPU transfer. Distinct queued and
large pattern hashes, CPU stage accounting, exact dimensions/offsets and all
eligible resource/submission/fence/cap/destruction joins are checked. Native
format exclusion, oversize rejection, eviction and device isolation must execute.
No shader performance, physical displayed pixel, native fault injection or total
allocation-balance claim follows from this protocol.

One off/on pair, same executable/library and both observers enabled. Thirty
seconds per native process, ninety seconds globally including configuration
probes. Use the existing exclusive lease, owned watchdog/retirement, pinned
sources/binary/library/logs, before/after configuration/persistent geometry and
immutable SDK inventory. No foreground requirement: this is headless correctness.
One attempt each; stop/seal at the first error, invalid record, drift or deadline.
No unchanged retry or broad campaign. Keep partial/failure data.

Snapshots require quiescent native callers; concurrent writers trigger a retryable
pending/unstable result. Configuration is strictly pre-device, not a concurrent
control API. The owned temporary recipe is per-host, not a portable SDK installer;
unexpected poison/unwind aborts the diagnostic rather than recovering corrupt state.

Focused policy and lifecycle/schema refusal controls, exact source audit and
Rust/native/Go builds precede internal review and Prism Anthropic Claude Sonnet
5.5, then commit. Only then execute this byte protocol, independently verify,
review and commit results. Native/Go build failures and preparation revisions are
retained; they are not GPU test attempts. Only after byte/lifetime proof, prepare
and review a separate bounded framework timing pair. No cache policy is adopted
automatically. Features frozen, experimental rendering/pacing disabled, 33.3 ms
receipt-to-actual-presentation gate unchanged and M3–M7 open.
