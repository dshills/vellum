# Focused offscreen raster replay costs

Question: are the four no-paint paths from the admitted native root structure a
worthwhile optimization target? Keep features, shipping algorithms, native archive,
SDK, render/pacing defaults and the 33.3 ms presentation p95 gate unchanged.

Use an exact fresh copy of the frozen SDK Core without metadata hooks. The test
compiler overlay reuses the existing routed patch fixture with dimensions changed
from 1500×1250 to 1519×1266 and explicit manual DPI160 (8px inset). This creates
real manual fixture commands, not a reconstruction claiming byte-identical native
replay. All20command types/indices, image geometry/operations/color types and path
length/transform/no-paint flags match the last native capture; four overridden
opaque backgrounds are black instead of white and final color green instead of
red. Native path coordinates were not recorded, so equivalence is bounded.

The test-only source-renderer wrapper delegates unchanged drawing and release.
At most8calls/32items each are captured into owned copies; uniform values and
path slices detach. Only builtinRGBA/NRGBA uniforms, unpainted/unmasked paths and
plain contexts are accepted. No wrapper is present in the timed replay loop.
Restore the global factory after the fixture and retire its manual owner before
benchmarking. Keep the capture helper local to testdata, outside Go discovery.

Two focused Go methods retain original routed full-image/state/unrelated-key
checks; compare full and image-only replay against the original pre-elision
pixel oracle for red/green/red, prove paths alone do not alter pixels, preserve
command state, and check detached source/path mutation. No production fast path
is implemented. The last-fill and empty cases are subsets, not complete scenes.

One benchmark process, five cases, three200ms observations each, explicit-cpu=2,
parallel1,90s test/120s owner deadline. Capture/build/pixel checking stay outside
timing. Cases: full replay, eight image commands, four no-paint paths, last solid
fill alone, empty replay. Require exact15rows, suffix2, positive iterations/cost,
zero allocation and successful terminal; one parser method checks nine refusals.
Report raw values/ranges/medians, never native latency/p95/phase sums or speedup
from differences inside run noise. Hot repeated CPU replay omits native120ms gaps,
CPU/GPU upload sharing, scheduling and presentation.

Retain initial96DPI geometry refusal, its exact diagnostic rerun and a test-only
color-restoration bookkeeping failure; fix before timing. Current private Core
and8321SDKfile vectors remain exact, logs/source/module/overlay pins hold, and all
owned groups retire. Known existing compiler deprecation/duplicate-library warnings
are retained, not new native runtime warnings. No desktop app is launched.

To prepare elsewhere with the same pinned Go1.25.6/cache and prepared SDK:

```sh
python3 tools/macos-calibration/rastercost/prepare.py --sdk "$SDK" --core "$NEW_CORE" --artifacts "$NEW_OUTPUT"
cd tools/macos-calibration/stagingreuse/framework
# Use the pinned offline Go environment; these are headless test binaries.
go test -tags=offscreen -modfile="$NEW_OUTPUT/framework.mod" -overlay="$NEW_OUTPUT/overlay.json" -run='^TestRasterReplay' -count=1 -cpu=2 -v .
go test -tags=offscreen -modfile="$NEW_OUTPUT/framework.mod" -overlay="$NEW_OUTPUT/overlay.json" -run='^$' -bench='^BenchmarkRasterReplayBreakdown$' -benchtime=200ms -count=3 -cpu=2 -parallel=1 -timeout=90s .
```

New output/Core paths must not exist. The retained owned executor adds network
deny,120s timeout, source/SDK/log vectors and process-group retirement. Preparation
reproduction checks exact fixture/test/module/sum and complete Core bytes without
rerunning a benchmark. Internal review, Prism Anthropic Claude Sonnet5.5, specs,
indexed evidence and a local checkpoint complete this investigation. M3–M7 open.


## Separated-condition follow-up

After warm checkpoint7a2614d, one new focused method primes four persistent
renderers outside timing, toggles the patch color, and times one full replay.
Eight rounds rotate the starting condition: immediate,120ms pause, immediate
32MiB scratch touch,120ms pause plus32MiB scratch touch. Touch each64-byte region
outside replay timing; this is a footprint, not proof of cache eviction. Retain
all32values, actual gaps and pressure times. FixedCPU2,10s sample-phase/15s Go/
45s owner deadlines; check allfourfinalfullimages against legacy replay after
timing. Original routed/pixel fixture and helpers unchanged. No new hot benchmark.
One parser method refuses nine wrong-scope/order/partial/duration/color cases.
No cache/CPU/scheduler/GC causality or native-p95/adoption claim follows. See the
[separated result](../../../docs/reports/macos-raster-separated-cost/README.md).
