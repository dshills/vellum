# Covered uniform GPU encoding

Owned stage 57 preserves the original packet and Builder while omitting covered
uniform Src operations only from the opt-in GPU encoding. An immediately following
opaque uniform Over with the same owned clipped rectangle replaces every affected
pixel. All original operation, byte and tile-reference charges run before output
allocation. CPU reconstruction, packet byte accounting, reset proof and original
Encode bytes stay unchanged. The GPU stream owns compact IDs and tile lists.

## Evidence and limits

A 48-frame current-driver software matrix captures actual procedural 10,000-row
grid paint batches with fixed/variable rows, DPI 96/120/144/192 and six mixed-axis
scroll steps. Scalar execution of the serialized compact GPU ABI matches original
CPU raster and every full-composition pixel. Prior full-composition records also
match. This is an offscreen procedural fixture, not stored-data/native acceptance.
The first 400-cell frame reduces 4,397 → 3,980 encoded operations and 1,688,088 →
1,538,904 bytes. Across all frames it removes 193–477 operations and
132,000–216,684 bytes. Host/native frame times and GPU allocation were not measured.

Focused tests compare original Encode to an independently marked mixed stream,
including masks, pixels and ordered 16-bit spans. They cover clipped/default/full
rectangles, partial alpha, interruptions, unequal rectangles, detached encoding,
held packet state and original retained/encoded byte, operation and tile-reference
limits. Existing original frozen digests remain unchanged. Three faulty overlays
fail the alpha, compact index and original-reference refusal assertions.

Existing gated native integer/text tests switch to EncodeGPU; a new 12-frame
regression checks full and partial covered fills on native compute. They compile
but are deliberately unexecuted while the human uses the computer. No GPU device,
window, input event, preference or clipboard is touched by this checkpoint.
The first matrix failed compilation due to a missing image/draw import; the fixed
V2 and original failure are preserved with their bound source receipts.

Normal preparation reproduces four reviewed outputs through all 57 patch
numbers, adding one test leaf (6,207 → 6,208) with every other dependency leaf
unchanged. The first adoption script incorrectly equated stage-group count with
patch count and stopped before manifest mutation. Its partial files and an
unchanged-baseline preparation are retained; the corrected adoption checks exact
patch numbers and its new preparation proves the expected outputs. Final renderer race/vet, adapter offscreen race/vet, portable race, preparation
tests, native harness contract tests/vet and native build pass. Internal
source/evidence review and one bounded Prism review using Anthropic Claude Sonnet
5.5 precede commit.
Producers, overlays, failures, comparisons and review records will be retained in
the indexed records archive before commit; native binaries are excluded with their hash retained.

GPU scene rendering stays opt-in. Native 4 ms p95 / 6 ms p99 targets still fail;
full P6/P11, a fresh complete lifecycle campaign (prior run failed after 343
windows), physical input/displays, human screen-reader work, native/GPU memory and
Mac distribution remain open. Live tests are paused while the human uses the host.
