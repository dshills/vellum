# Exact small glyph coverage reuse

Patch 0056 retains ordered 16-bit scanner spans after the existing Alpha glyph
cache declines a small static TrueType fill-only glyph. Original outline,
segment, coordinate and extent work charges run before lookup. Exact fixed-path
geometry, including fractional rounding, authorizes a hit; the fingerprint only
indexes candidates. Original owned spanners perform blending and packet emission.

Admission requires an identity transform, value uniform fill, clean owned scanner,
and a guarded hull fully inside the current target and nonempty clip. Unsupported
fonts, boundary cases, scroll recording and custom spanners retain ordinary routes.
Per renderer FIFO limits are 512 entries and 1 MiB of exact path/span array backing,
with 512 path values and 256 spans per entry. Metadata and transient recording are
outside this bound. Resize and Release drop storage. Entries own no fonts, images,
packets or builders. Existing owner serialization remains required.

Tests compare original pixels and encoded packets, budget refusal/fallback state,
held packet immutability, independent entry/byte eviction, detached backing, hash
collision refusal and scanner ownership. Four deliberately broken overlays prove
specific assertions. A 48-frame actual-driver software grid matrix matches full
composition at four DPI settings and fixed/variable row sizes.

New helper/tests are MIT; modified upstream files preserve their original notices.
Preparation checks original/output bytes and patch SHA256. This stage does not
qualify native frame targets or full Mac acceptance. GPU scene rendering remains
opt-in; live checks are paused while the human uses the computer.
