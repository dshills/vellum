# Renamed regression font fixture

`VellumInvalidVariationFixture.ttf` is an OFL-1.1 derivative of the pinned
go-text invalid-GVAR test fixture. See [OFL.txt](OFL.txt), [FONTLOG.txt](FONTLOG.txt)
and [manifest.json](manifest.json). The font is not MIT-licensed; the generator
and its tests are Vellum code under MIT. It is not a default Darwin embedded font.

Reproduce it with the exact original input identified in the font log:

```sh
python3 generate.py --source /path/to/original/Selawik-VF-Subset.ttf --output /tmp/new-fixture
```

The output directory must be new. The generator accepts only the pinned input,
preserves font programs and the deliberately inconsistent variation data, and
renames four primary name records. A fresh dependency preparation overlays the
result onto the historical upstream test path and supplies the updated notice
and font log. It changes no runtime font or Go/C input behavior.

The independent tests parse raw SFNT/name records and verify all table payloads,
checksums, declarations and deterministic output. Run them with the same original
input; they never install fonts or open native applications.
