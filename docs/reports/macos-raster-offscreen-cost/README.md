# Offscreen raster cost: warm no-paint path replay is negligible

An uninstrumented frozen renderer replaying the generated offscreen patch fixture
spends about **0.17 ms per warm replay**. Its four unpainted paths together take
about **130 ns**. For this warm workload, these observations do not justify a path-elision
optimization or native performance campaign. The final solid fill alone has a similar cost to full replay.

| Separate benchmark case | Range across three observations |
| --- | ---: |
| Full 20-command replay | 167.693–172.643 µs/op |
| Eight image commands | 172.054–174.005 µs/op |
| Four no-paint paths | 0.1295–0.1306 µs/op |
| Last solid fill alone | 163.943–165.436 µs/op |
| Empty replay | 0.003612–0.003663 µs/op |

All cases report0bytes/0allocations per operation. Explicit-cpu=2 agrees with the
runtime assertion and benchmark suffix; one bounded process completed in9.099s.
The cases are independently timed subsets, **not additive phases or native cost
shares**. Image-only replay is slightly slower in these observations; this is not
evidence of a real regression or speedup. Full legacy/image-only pixel parity
is checked outside timing. No raster algorithm has been changed.

## Scope and focused verification

The [protocol](../../../tools/macos-calibration/rastercost/PLAN.md) and private
[test fixture](../../../tools/macos-calibration/rastercost/testdata/replay_test.go)
retain the original routed state/full-image checks. Two Go methods additionally
verify red/green/red full legacy replay parity, image-only parity, unchanged pixels
under unpainted paths, restored command state and detached mutable source/path
capture. One Python method refuses nine malformed/incomplete/allocation/CPU cases.
Preparation reproduces the fixture/module/Core bytes without another timing run.

Target1519×1266 and patch(8,8)–(1511,1258) agree with native scalar geometry. The
original manual96DPI setup produced5px insets and was refused; its diagnostic
rerun is preserved. DPI160 gives8px. A subsequent test bookkeeping error restored
the already-mutated uniform pointer rather than its original color; corrected
before timing, with failed sources/logs retained. Final tests and bounded capture
forwarding of renderer release pass. No native window/input/settings test ran.

This is a real generated offscreen command stream, not byte-identical native
replay: four overwritten opaque backgrounds differ black/white and the last
patch green/red. All20command indices/types, image rectangles/positions/operations/
color types and path lengths/transforms/no-fill/no-stroke flags agree with the
last native capture. That capture omitted full path coordinates. The complete
comparison and raw observations are in [summary.json](summary.json).

The complete private Core copy and all8,321frozenSDKfiles remain unchanged, with
no metadata hooks, uniform-fill candidate or pool/cache experiment. Logs, source,
module, compiler-overlay pins and owned process retirement are verified. Existing
compiler accessibility deprecation/duplicate-library warnings are retained.
[Indexed evidence](evidence-index.json) includes raw logs, initial failures, input
vectors, preparation reproduction, independent validation and checkpoint review.

## Remaining uncertainty

Hot repeated replay is much smaller than the earlier native roughly0.74ms raster
p95. The workloads have different timing/cache/scheduling and surrounding upload
conditions; these results neither explain that gap nor prove a native speedup.
Next investigate warm versus separated replay/cache conditions offscreen before
changing an algorithm. Do not retry native timing or skip no-paint paths on this
evidence. Features/defaults/SDK/app and the33.3ms receipt-to-presentation p95 gate
stay unchanged; M3–M7 remain open. This checkpoint is not release qualification.
