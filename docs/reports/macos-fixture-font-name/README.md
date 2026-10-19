# Renamed invalid-variation fixture

The modified test font now uses Vellum Invalid Variation Fixture for its internal
primary names. Original copyright, trademark and OFL remain. The historical
testdata filename and source history remain documented. Only four of 26 name
records change; all other table payloads are preserved except the required head
checksum adjustment. Empty glyf and malformed variation data deliberately remain.
See [the contract](../../../specs/fixture-font-naming.md).

Nine independent generator controls, 29 preparation controls, five real-parser
race methods, default font vet, module verification and four notice-collector
controls pass. Fresh pinned preparation changes exactly four typesetting paths.
The original-font private copy fails the four intended primary-name assertions
and still passes InvalidGVAR. All 4438 selected embedded assets and 12 default
runtime fonts match the preceding current-build inventory byte for byte.

The first integration attempt exposed the existing 64 KiB overlay limit. The
explicit bounded 512 KiB stage allowance fixes that; default and aggregate limits
remain intact. Internal review also found unguarded standalone input reads;
nonblocking no-follow regular-file reads and controls address that. A subsequent
font test attempt failed at setup because declared standalone module dependencies
were absent from the offline cache. Only the three exact go.sum-verified public
dependencies were fetched; later offline race/vet/verification passed. Earlier
failed receipts and source review findings remain preserved externally.

Prism then identified a concrete descriptor leak when stream construction rejects
a directory after os.open succeeds. Both the generator and preparation readers
now close the raw descriptor from one outer finally, with closefd=False on the
stream. New controls cover actual directory rejection, injected wrapper failure,
bound rejection and successful reads. The existing short/growing-read proxy was
updated to forward closefd; its first failed fixture attempt is retained. Fresh
prepared dependency vectors match the already tested font race/vet input exactly.
Font bytes remain unchanged; the generator manifest is repinned. The self-hash
finding is covered by the existing deterministic manifest equality test. The
filtered name-storage finding has no unused source string data and no format-1
language tags. The first provider report and certificate are preserved.

Actual collector copying retains the original OFL and updated font NOTICE with
unchanged pinned module version/sum and own module license. FONTLOG travels in SDK
source; it is outside the unchanged conventional notice filename policy. Font
licensing remains OFL, separate from MIT generator code. This does not establish
historical naming permission or approve the broader font/asset distribution audit.
Mixed icon origins and final M6 decisions remain open.

Private SDK integrity, optimized native compilation, relocated consumer checks
and unsigned/unlaunched app packaging have separate checkpoint receipts. No
native app, physical input, OS settings, actual presentation, endurance or measured
performance was exercised here. The 27 unchanged demo programs were built in the
event-coalescing checkpoint and were not repeated. Measured 60 Hz responsiveness
remains the release gate, with 4 ms retained as an optimization target.
