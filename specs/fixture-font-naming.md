# Modified regression font naming

The distributed typesetting invalid-variation fixture has distinct primary
internal names: Vellum Invalid Variation Fixture, with PostScript name
VellumInvalidVariationFixture-Regular. The historical testdata filename remains
Selawik-VF-Subset.ttf for test compatibility and provenance. This fixture is not
one of the default runtime fonts and is not installed on the host.

Start only from the pinned 454405-byte original, SHA-256
720afadd389c79e39a02502e952fecb6f5d78eeca5f095c14b9963ead45a8f1c.
Change Windows English name IDs 1, 3, 4 and 6. Preserve the other 22 name records,
including copyright, trademark, designer, license, version, style and variation
axis names. Preserve all non-name table payloads except the required head
checkSumAdjustment. Recompute directory offsets, checksums and alignment. Keep
the deliberately malformed empty glyf, loca and gvar data intact: this is a
regression fixture, not a font repair or a new rendering font.

The deterministic generator, independent raw SFNT checks, derivative, exact OFL,
FONTLOG and manifest live in drivers/cogent/native/assets/fixturefont. Font bytes
remain under the original OFL; generator code is MIT. The OFL modification and
Reserved Font Name guidance is at
https://openfontlicense.org/how-to-modify-ofl-fonts/ and
https://openfontlicense.org/ofl-faq/ (3.8, 5.3 and 5.9). Historical attribution and
reserved-name declarations remain; this change does not assert historical
permission, authorship of the original font or overall distribution approval.

Typesetting stage 0006 pins the input, derivative and changed text. Its explicit
asset_byte_limit is 512 KiB. The default stays 64 KiB; the hard per-asset cap is
512 KiB and aggregate cap stays 4 MiB. Exact integers only, with limits checked
before payload reads and captured for post-write verification. The standalone
generator rejects symlinks and nonregular inputs without blocking on FIFOs,
uses bounded reads and checks identity before and after reading. Both this reader
and the preparation reader keep a single raw-descriptor owner in an outer
finally and pass closefd=False to the stream, including constructor failures.

Verify fresh preparation, independent generator and asset-limit controls, real
parser name/declaration tests, InvalidGVAR and deferred-glyf regressions under
race, default font vet, module verification and actual notice transport. A private
copy with only the original font restored must fail all four new primary-name
checks while still passing InvalidGVAR. Notice copying retains the original OFL
and updated font NOTICE; FONTLOG is included in SDK source, not selected by the
conventional notice collector. Keep failed attempts and raw receipts.

The current 4438 selected embed bytes and 12 default runtime fonts remain
unchanged. Declaration availability is not applicability approval; mixed icon
origins and the rest of M6 remain open. SDK integrity, native compilation,
consumer relocation and unsigned app packaging are separate checks. No physical
input, native app launch, presentation, endurance or performance qualification
is implied. Measured 60 Hz responsiveness gates release; 4 ms is optimization.
