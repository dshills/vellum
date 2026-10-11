# Custom GPU drawing application

This independent module consumes Vellum's public APIs. It offers retained triangle
gradients and strokes, a spatially indexed 100,000-rectangle schematic and a
million-sample waveform reduced to the visible sample interval and pixel columns.
The 4K image mode uses a real 4096×4096 RGBA source retained in immutable
128×128 tiles, with a shared-device native texture below the mesh.
Drag to pan and scroll to zoom. Buttons and status are ordinary framework UI.

From the repository root, after native preparation:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
. ./drivers/cogent/env.sh
cd examples/drawing
go run .
```

`go run -race . -smoke` opens two fixture windows, checks native GPU pixels,
retention, culling/reduction, budgets, rebuild/resize and closure, then exits.
Image checks include one 64 KiB tile patch without reallocation, independent
image-only resources, camera/placement/filter retention, premultiplied linear
filter pixels, nearest selection, translated placement and image removal.
It changes no OS input or accessibility preferences. It has a bounded watchdog;
the printed JSON reports source data separately from submitted triangles.

The full check script runs this module's race tests, vet and build. See the
[drawing contract](../../specs/drawing.md) and [image contract](../../specs/images.md)
for ownership, precision, resource
accounting and acceptance limits. Snapshot tests inspect offscreen GPU pixels;
they do not measure presentation latency or qualify native clipped composition.
Mac runtime remains experimental. The example's code is MIT licensed.
