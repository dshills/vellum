# Covered uniform GPU encoding

Owned patch 0057 adds Packet.EncodeGPU and uses it in the opt-in GPU scene
adapter. It omits only an owned Uniform Src immediately followed by an equal
clipped-rectangle Uniform Over with 16-bit alpha 65535. The packet still retains
every original operation; all Builder validation, byte charges, original encoded
operation/byte/tile-reference admission, CPU reconstruction and independent-reset
proof remain unchanged. Compacted output owns exact operation IDs and tile lists.
Packet.Encode preserves the original ABI bytes and frozen digest tests.

Independent mixed-stream and tile-outer scalar execution compare serialized output
against original replay. Tests cover masks/pixels/spans, clipping, alpha 65534,
interrupted/unequal rectangles, detached results and original limit boundaries.
Three deliberately faulty overlays fail the alpha, index and refusal assertions.
The actual-driver software matrix compares all pixels in 48 full compositions
across four DPI settings, fixed/variable rows and mixed-axis scrolling. It reduces
193–477 encoded operations and 132,000–216,684 bytes per frame in this fixture.

Existing gated native integer/text regressions use the compact route, and a new
12-frame native test checks full and partial covered fills. These compile but
remain unexecuted while the human uses the machine. No native GPU/latency or
full Mac qualification is inferred. GPU scene rendering remains opt-in.

New tests are MIT; modified upstream files preserve their original notices.
Preparation validates original/output hashes and patch SHA256.
