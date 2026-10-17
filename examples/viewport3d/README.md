# Shared-device 3D viewport

This independent Go module consumes the public `cogent.GPUView` API. Its indexed
cube uses a camera uniform, retained geometry, a checked custom shader/pipeline
and four-sample depth testing. Drag or use arrow keys to orbit; scroll to dolly.
The buttons and status label are ordinary framework UI.

After preparing the native dependencies, run from the repository root:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
. ./drivers/cogent/env.sh
cd examples/viewport3d
go run .
```

`go run -race . -smoke` opens two fixture windows and runs bounded native checks.
A near red face is drawn before the overlapping far blue face, so the center
pixel assertion fails without depth testing. Checks also cover acknowledged
chrome retention, native UI camera actions, 64-byte camera-only uploads,
independent resources, callback-scope helpers, budget refusal, controlled callback
fault repair, resource rebuild, repeated resize/depth recreation, empty-frame
clearing and closure. A watchdog limits execution. No OS input/accessibility
settings are changed. JSON records actual submitted target sizes and counters.

`./scripts/check.sh` runs race tests, vet and builds this module. The
[advanced GPU contract](../../specs/advanced-gpu.md) describes ownership and
borrowing requirements. These snapshots verify offscreen GPU pixels; presented
clipping, latency, sustained performance, compute/floating-point workflows and
asynchronous readback need separate qualification. macOS remains experimental.
The example's code is MIT licensed.

The `--smoke` fixture disables native window-geometry loading and saving before
constructing the application and keeps that process-local setting through teardown.
It restores the prior process value on normal return; ordinary interactive runs
retain their usual geometry behavior. The JSON result reports the observed
`geometry_persistence_disabled_in_fixture` flag. This isolates fixture windows
from saved user geometry and avoids persisting the deliberate resize sequence.

The 3D smoke establishes a paired 100 ms quiet interval for both initial targets,
then captures both baselines together before its first chrome mutation. It refuses
later frame/resource drift with the original strict retention checks. This avoids
using a provisional first layout as the retention baseline; the quiet interval is
an operational fixture barrier, not a guarantee that all OS work has drained.
Bounded owner-thread observations retain independent-window state for diagnosis.
No frame, budget, depth, pixel, fault-repair, resize or closure assertion is waived.
