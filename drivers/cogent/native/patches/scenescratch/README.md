# Bounded scene scanner reuse

Patch 0054 reuses one private Rasterx scene-capture scanner under the existing
serialized renderer owner. Borrow removes it from the idle cache. The unchanged
fresh Dasher resets bounds and winding; capture restores the original winding.
Return detaches the current Spanner, clears clip and accumulation, and retains
only combined row/cell backing capacities of at most 64 KiB per renderer.
Resize on size change and Release retire the cache. No Packet, Builder, source
image or Font reference survives through the cached scanner.

The bound covers idle scanner array backing, not peak capture allocation or total
process/native memory. Oversized backing is discarded without changing successful
capture admission. Original whole preflight, work charges, immutable packet
emission and CPU fallback remain unchanged. GPU scene rendering remains opt-in.

An independent copied fresh-capture oracle checks Encode and software pixels,
previous-packet immutability, repeated winding/bounds, preflight and late builder
refusal, retirement and capacity. Native parity and interleaved frame measurements
are recorded in the Mac scene-scratch report. New helpers/tests are MIT; upstream
files preserve their original notices. Preparation verifies original/output hashes
and the patch SHA256.
