# Sparse staging bands for scanner spans

Stage 42 follows the ordered GPU scene stage and changes only span construction,
its CPU reference/allocation checks and the native multi-row parity fixture.

BeginSpans allocates no row headers. Accepted spans grow a charged geometric row
band within the existing target-height cap, moving unused headroom when necessary
and preserving every touched row in either arrival direction. Combined row/span
reservation precedes allocation. Flush refunds unused headers and makes the same
exact relative-row packet as before; pixel arithmetic, operation order, shader
layout, tile encoding, detached ownership and fallback policy remain unchanged.

The meaningful regression compares prior full-target construction bytes and
reference scan pixels for arbitrary row order, gaps, overlap and both operators.
A scoped benchmark compares the same tall-target path corpus via inert overlays.
Native parity exercises row-band growth and relocation in the existing 96-packet
integer compositor corpus. End-to-end allocation/latency remains separately
measured; the scene renderer stays disabled by default.
