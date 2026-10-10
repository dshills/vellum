# Retained macOS GPU drawing

Vellum now offers immutable indexed 2D meshes, world-space strokes, exact waveform
min/max envelopes and a paint-ordered rectangle spatial index. The new native
`MeshCanvas` uses the UI window's shared WebGPU device, a retained flat shader,
4× MSAA resolve and linear-premultiplied color interpolation/blending. Geometry
uploads follow immutable mesh identity; camera-only changes update one 16-byte
transform, and unchanged native visits reuse the completed texture.

The independent `examples/drawing` module exercises public APIs with ordinary UI
controls, two windows, a 100,000-object schematic and a million-sample waveform.
The race-built native fixture passes gradient/alpha pixels, transparent export,
independent resources, acknowledged unchanged visits, pan retention, visible
culling/reduction, budget refusal preserving the prior frame, resource rebuild,
three allocation-changing resizes, empty-frame clearing and closed-window refusal.
It also queries renderer stats from the owner-held error callback, checking that
native/resource locks have already been released. The expected injected budget
error appears once in the successful native log.

At the fixture's schematic viewport, 2,409 source rectangles were tested and
2,310 were submitted from 100,000. The waveform submitted 2,000 triangles for
1,000 visible envelope columns from 1,000,000 samples. These are distinct source,
query and submitted-work counts. Startup/OS layout can change renderer visit
counts; the raw report records this run's values.

Fresh `scripts/check.sh` passes all 87 Python regressions, portable/native race
tests and vet, selected owned-backend race regressions, native workbench build,
hello/drawing independent module checks and public SDK contracts. Five extra
focused race repetitions pass, including lifetime/retirement, clipping math,
logical-device admission, translated-mesh precision, stale source rejection and
serialized error callbacks. Exact inputs, commands, exit codes and raw log hashes
are in [source-and-check-proof.json](source-and-check-proof.json).

On this Apple M4 Pro, three portable benchmark repetitions measured indexed
100,000-object visible mesh construction at 46.2–58.0 µs/op, 410,280 B/op and
18 allocations/op; million-sample envelope queries into 1,024 columns measured
13.1–15.5 µs/op, 24,576 B/op and one allocation/op. Source construction occurs
outside the timed loops. These runs overlapped other checks and are not isolated
CPU measurements, GPU/presentation timings or sustained P4 certification.

Internal review resolved lifetime and late-input guards, float64 local rebasing,
logical-device/checked allocation admission, computed viewport validation,
acknowledged retention checks and consistent error callback ownership. Final
internal review found no remaining actionable defect in this phase. Early
exploratory fixture failures exposed a startup stage-read race, premature
retention assertion and resize expectations below the content minimum; those
were corrected before the recorded final checks.

The deterministic clipping tests and offscreen pixels do not certify scroll-clipped
native compositor output or MSAA edge quality. P4's 3D viewport, 4096² changed image
tiles, compute preparation, float targets, asynchronous screenshot readback,
sustained frame/input and global resource budgets remain open. Physical GPU loss,
external OS workflows and the other [Mac gates](../../../specs/macos-completion.md)
remain open. This checkpoint is not framework completion or public distribution
approval. No OS input-source or VoiceOver preferences changed; fixture windows closed.
