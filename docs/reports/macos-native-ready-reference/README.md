# Ready-window native Metal reference — 21 October 2026

**Both corrected native reference cohorts are valid. The framework latency gate
is still open; no shipping fix or policy change is adopted.** This closes the
specific readiness and target-timestamp diagnostic prerequisite from the
[preceding investigation](../macos-native-latency-investigation/README.md).
It does not qualify the framework or prove a universal macOS latency floor.

## Procedure and populations

A private Objective-C/Metal executable bypasses Vellum, Go, WebGPU, layout and
text. The raw path records its render-request timestamp before `nextDrawable`.
The direct CAMetalDisplayLink path records its receipt at the supplied-drawable
callback, after acquisition, and stores both native target timestamps. These
boundaries are different and must not be pooled or compared as a causal A/B pair.

Every startup record is retained as setup. A positive actual setup presentation
must precede the first cohort receipt. Then exactly 128 warmups and 120 measured
frames must have positive, strictly ordered presentation feedback; missing or
collapsed cohort feedback is refused. No failed framework prefix was repaired
or reclassified. The earlier reference and four failed candidates stay sealed.

One attempt per path ran during the authorized quiet window, 13:19:06–13:19:18
America/Chicago (12 seconds), on Apple M4 Pro, macOS 26.6.2/25G83, one LG
5120×1440/60 Hz display, AC power. The reference reports a 1519×1266 drawable,
scale 1, two drawables, vsync enabled and transaction presentation disabled.
The reference permits next-drawable timeout; the preceding SDK path disables it.
This configuration difference is another reason these are not a causal A/B pair.
Display/power/OS probe bytes match before/after each case; the framework geometry
store is unchanged. Source, executable, compiler and actual commands are retained.

The program requests ordinary app activation and a key window. **Actual foreground
and key-window state were not instrumented**, so this run must not establish a
foreground-input performance tier. No physical input or pixel-readback acceptance
is claimed. Presentation feedback establishes the native drawable timing; this
empty changing-clear reference is not a professional framework workflow.

## Results

| Path / measurement boundary | Measured p50 / p95 / p99 / max | Native interval p95 / max |
| --- | --- | --- |
| Raw Metal render request, before acquisition | 48.333 / 48.843 / 48.905 / 48.941 ms | 16.666792 / 16.666833 ms |
| Direct display-link supplied-drawable callback, after acquisition | 49.835 / 49.908 / 49.931 / 49.955 ms | 16.666792 / 16.666834 ms |

Each path retains all 128 warmups and 120 measured frames with valid feedback.
Each measured cohort has 119/119 intervals within 18.333 ms. Both have seven
setup records: raw retains six zero presentation times, direct link three.
Those zeroes remain explicit setup losses before cohort readiness, not missing
measured samples. Neither path is five-run qualification.

For direct link, requested `preferredFrameLatency=1` still produces these observed
targets:

- Callback to render deadline: p95 16.558 ms.
- Render deadline to target presentation: 33.333417 ms for all measured frames.
- Callback to target presentation: p95 49.891 ms.
- Target presentation to actual presentation: p95 0.017042 ms, maximum 0.017375 ms.

Actual presentation follows the supplied target closely. On this tested path,
removing Vellum work does not produce a ≤33.3 ms reference result. That is a
bounded observation, not proof that every native presenter, foreground state,
display mode or hardware configuration has the same limit. Apple's
[latency-request documentation](https://developer.apple.com/documentation/quartzcore/cametaldisplaylink/preferredframelatency)
explicitly allows longer final latency in windowed macOS; a one-frame preference
is not a one-frame presentation guarantee.

The initial 33.3 ms release threshold was not calibrated against a ready native
reference for this configuration. This gap in the release plan is now visible.
The result supports reassessing that assumption; it does not authorize silently
lowering the gate, enabling tearing, exempting framework workloads or claiming
that the framework contributes no excess latency. The preceding paced-grid
62.224 ms p95 remains failed under the existing policy.

## Verification and next decision

Nine focused classifier controls cover retained setup loss, receipt boundaries,
missing/collapsed feedback, readiness ordering, phase population, missing target,
closure and nonfinite values. They pass, as does the native build. Prism using
Anthropic Claude Sonnet 5.5 found a callback-data concurrency risk; the final code
uses separate C11 atomic presentation slots, captures immutable phase values,
drains the callback group before reading, and freezes the producer before
pausing. The controls were rerun and both actual native cohorts then passed.

Two native and sixteen probe process groups are reaped and observed absent;
both windows close and callbacks drain. This is not full native-resource
allocation/deallocation or endurance acceptance. The immutable SDK is unchanged.

Before choosing another backend repair or a different latency policy, verify
foreground/key-window state and the native scheduling configuration in a bounded
control. A proposed baseline-derived release policy needs owner approval,
predeclared criteria and fresh representative framework measurements; this
reference cannot itself qualify M3. Keeping the existing 33.3 ms gate requires
further native presentation investigation rather than adopting any rejected
candidate. M3 and M4–M7 remain open, with features frozen.

[summary.json](summary.json) records exact metrics and artifact-relative
source/executable pins, with explicit archive membership. Original execution
receipts retain local paths as provenance. External reused helpers are identified
by hash and the executable is excluded; this is not a standalone rebuild bundle.
[evidence-index.json](evidence-index.json) binds the
[evidence archive](evidence.tar.xz): source, validator controls, native raw records,
configuration, lifecycle receipts, source review and dispositions. It excludes the
executable, full SDK/prepared sources and historical archives. The final report
was reviewed with Prism using Anthropic Claude Sonnet 5.5. Its archive-mapping
finding was corrected and independently verified; no findings remain unresolved.
