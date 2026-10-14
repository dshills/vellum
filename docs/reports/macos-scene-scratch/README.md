# Bounded scene scanner reuse

Owned patch 0054 reuses one renderer-private scene-capture scanner. Borrow removes
it from the idle cache; return detaches the capture Spanner, clears clip and
accumulation, and retains at most 64 KiB of combined row/cell backing capacity.
Size changes and Release drop the cache. The existing serialized renderer owner
remains required. No Packet, Builder, source image or Font reference is retained.
This bounds idle array backing, not peak capture or total native/process memory.

Original whole preflight, logical work charges, packet emission and CPU fallback
remain unchanged. Oversized scratch is discarded without changing admission.
A copied original fresh-capture oracle compares complete Encode output and pixels
across repeated paths/text/winding and bounds. Previously sealed packets remain
unchanged after later accepted capture. Preflight refusal and a real late builder
refusal check reset/detachment, original CPU state and subsequent parity.

The frozen four-file overlay passes focused race tests repeated three times,
selected Vellum renderer race tests and headless native GPU compute/text/path/
shared-drawer fallback parity. Formal preparation produces exactly the four
reviewed outputs; every other prepared leaf matches the original baseline.
Preparation now expects all 54 owned patches and 6,202 prepared files.

## Measurements

A sequential five-repeat local scene-capture benchmark reports median short
capture 3,119 → 2,668 ns, 16,924 → 14,204 B and 89 → 81 allocations per operation.
Tall capture reports 4,798 → 3,094 ns, 25,884 → 14,204 B and 89 → 81 allocations.
These small owned path fixtures demonstrate local allocation removal; they do
not establish a whole-frame benefit.

The ordinary native comparison runs 40 processes: five alternating pairs per
fixed/variable 10k/1m-row case, each with 128 guarded warmup callbacks and 120
samples (4,800 total). Both binaries use the same observer and opt-in GPU scene
mode. All paired geometry, callback guards and process exits pass. Every run and
outlier is retained.

| Case | Original p95 / p99 (ms) | Candidate p95 / p99 (ms) |
| --- | --- | --- |
| fixed-10000 | 7.060 / 9.235 | 6.898 / 7.584 |
| fixed-1000000 | 6.672 / 6.901 | 6.637 / 6.870 |
| variable-10000 | 6.875 / 7.851 | 6.758 / 7.662 |
| variable-1000000 | 6.684 / 6.928 | 6.713 / 6.970 |

Whole-frame timing is mixed: mean UI work changes −2.11% to +0.52%; p95 changes
−2.30% to +0.43%. Median process allocation changes stay within 0.1% and do not
demonstrate a process-wide allocation reduction. Both binaries still fail the
unchanged 4 ms p95 / 6 ms p99 targets. Host UI-work timing excludes queue wait and
drawable acquisition and does not measure GPU completion/presentation. Default
CPU rendering was not part of this comparison. This checkpoint adopts a bounded
local allocation improvement without claiming a native frame-rate gain.

## Review and evidence

Final owned-patch preparation, selected renderer race, adapter race/vet and
preparation regression checks qualify the adopted bytes. The checkpoint requires an internal review and
one bounded Prism review using Anthropic Claude Sonnet 5.5 covering the owned
patch, manifest, specs and report. Raw receipts, overlay inputs, all measurement
records and reviews will be retained in an archive with a verified SHA256 index.
The original unused-import compilation failure is retained as superseded.

Physical GPU recovery, presentation/native allocation attribution, full P6,
integrated P11, human input/VoiceOver and Mac distribution acceptance remain open.
