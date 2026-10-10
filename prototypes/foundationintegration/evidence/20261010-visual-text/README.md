# Visual navigation and native range geometry evidence

10 October 2026, macOS 26.6.2 (25G83), Apple M4 Pro arm64, 64 GiB, Go 1.27.2. Experimental TextField scope; G1 remains open. The manifest records the staged implementation, decoded toolkit and evidence SHA-256 values.

The native bridge and lifetime checks pass; final rendered queries pass in both the research and canonical Vellum builds. Each rendered probe checks 15 visible graphemes against an independent conversion from physical rendered cells through the actual native view/framebuffer dimensions. It verifies rectangles within 0.05 point, UTF-16 actualRange and hit indices. Synthetic queries cover wrap/bidi fragmentation, clipping, insertion affinity, scale, stale/invalid state and accepted/refused edits. The lifetime fixture creates 1,000 hidden windows and drains 11,000 native updates; it exercises the legacy publication API repeatedly, with geometry contracts separately checked once.

The protocol fixture acknowledges 13 synthetic input updates, rejects no application edit, and closes with zero outstanding credits. During saturation there are 256 outstanding updates; admission refusal preserves native logical state. These probes invoke Cocoa protocol methods and do not establish real IME candidate-panel or VoiceOver behavior.

The full offscreen core package, focused editing/shaping/raster tests, bridge/transport/semantics race tests and vet pass. Actual raster images `textfield-scroll.png` and `textfield-wrap.png` were manually inspected for visible Hebrew, selected cells, wrapping and clipping. Benchmarks are short local samples, not whole-app performance acceptance. Historical parser/TeX/font-golden discrepancies were not reclassified or retested here.

Commands run from `prototypes/foundationintegration`, after `export P0_CACHE_DIR=/tmp/vellum-p0-cache` and `. ./env.sh`:

```sh
python3 prepare.py
go vet -modfile=integration.mod ./...
go test -race -count=1 -timeout=60s -modfile=integration.mod ./transport ./inputbridge ./semantics
VELLUM_FONT_CACHE=../fontresources/build/font-cache VELLUM_VISUAL_EVIDENCE="$PWD/evidence/20261010-visual-text" go test -tags offscreen -modfile=integration.mod cogentcore.org/core/core cogentcore.org/core/text/shaped/shapers/shapedgt cogentcore.org/core/paint/renderers/rasterx -run 'TestVellum|TestFoundation' -count=1 -timeout=60s -v
go test -tags offscreen -modfile=integration.mod cogentcore.org/core/core -count=1 -timeout=120s
go test -tags offscreen -modfile=integration.mod cogentcore.org/core/text/shaped/shapers/shapedgt -run '^$' -bench BenchmarkVellumInteractionManyBidiRuns -benchtime=3x
go test -tags offscreen -modfile=integration.mod cogentcore.org/core/paint/renderers/rasterx -run '^$' -bench BenchmarkVellumSelectedTextReplay -benchtime=3x
go build -tags offscreen -modfile=integration.mod -o build/foundationaudit ./cmd/foundationaudit
./build/foundationaudit -nogui -font-cache=../fontresources/build/font-cache
go build -modfile=integration.mod -o build/integration-fixture .
go build -modfile=integration.mod -o build/lifecyclecheck ./cmd/lifecyclecheck
./build/lifecyclecheck -cycles=1000
./build/integration-fixture -geometry-probe -exit-after=5s -report=evidence/20261010-visual-text/rendered-native.json
./build/integration-fixture -input-probe -saturation -exit-after=14s -report=evidence/20261010-visual-text/native-fixture.json
python3 verify_acceptance.py evidence/20261010-visual-text
```

Commands redirect their output into the adjacent evidence files. Schema 2 requires bidi and native geometry evidence independently. Temporary negative artifacts confirm missing geometry and failed bidi each produce verifier exit 2. Schema 1 preserves historical evidence validation.

The canonical checkout independently prepares and builds the same sources; its rendered probe writes `canonical-rendered-native.json`. All 14 files decoded from patch 0011 match the tested research layout, canonical preparation and source snapshot reviewed by Prism. `patch-application.txt` checks every unified hunk count and dry-run application against the pre-0011 sources. Local heap profiles and superseded `render-tests.txt` are excluded from the committed evidence. No input sources or VoiceOver settings were changed.
