# Direct exact glyph lookup

The stage-59 patch removes callback dispatch from the two hot normalized-path
lookup loops. Validation still runs first. Hash operation order and full exact
equality remain the same; cold copying, cache size and ownership are unchanged.

Five interleaved original/candidate captures of a procedural 400-cell actual
software-driver batch measured median 5.241 ms versus 3.908 ms (25.4% lower).
All pixels and encoded shapes matched. Allocation count remained 106,337 per
capture, approximately 9.28 MB. The fixture uses supplied NotoSans15, DPI 96, fixed
rows and two benchmark threads, with CPU profiling active in both arms. These
are profiled local capture observations. The fixture excludes native
window/input/presentation and
does not qualify whole-frame latency, cold font setup, or full P6/P11.

The existing glyph coverage race tests and a deterministic 2,000-path comparison
with the pinned original walker pass. Individual command/coordinate corruption
and path-length mismatches are checked. Internal source review found no issues.
Normal preparation reproduces two pinned outputs (6,209 to 6,210 prepared files),
with every other dependency leaf unchanged. All 48 actual-driver reference
records match across four DPI settings and fixed/variable rows, including batch
and full-composition pixels, encoded shape and accounting. Renderer race/vet,
adapter vet and preparation tests pass. Two faulty overlays fail the original
hash and per-coordinate collision checks. A first native build used the wrong
command directory; the corrected native loaded-grid build passes. Its binary
hash is retained; the subsequent native campaign launched it as recorded below.

The full adapter race suite needs local socket access: the sandboxed pass found
a denied Unix socket bind in the workbench test, while the main adapter passed.
The rerun with local socket access passes the full suite. The successful main
adapter tests were reused by Go from the preserved sandboxed run; the workbench
socket test reran successfully. The report was updated after the recorded
code checks; each recorded source endpoint was stable, and every production
source and prepared dependency remained unchanged.
Renderer tests generated 17 PNGs, preserved separately before restoring the
6,210-leaf prepared tree; finalization initially rejected those additions.
The code checkpoint passed internal review and one bounded Prism review using
Anthropic Claude Sonnet 5.5. Its subsequent ordinary native comparison completed
80 interleaved processes and 9,600 samples across four stored-table cases in CPU
and opt-in GPU modes, comparing `4cf9e36`/6,209 prepared leaves against
`9686e38`/6,210. Geometry remained exact; four gated native GPU roots passed without
skips on the candidate. Native differences were mixed, with observed tail regressions. Both arms missed
the unchanged warm targets. See the [source-bound native follow-up report](../macos-native-followups/README.md).
The reported original and candidate arms still miss the 4 ms p95 / 6 ms p99 targets.

The follow-up report names the observed p99 increases explicitly. Their cause
is unassigned; allocation or local capture gains do not dismiss those tails.
