# Compute and HDR imaging

This independent Go module consumes Vellum's shared-device `GPUView`, checked
`GPUResources`, and asynchronous capture APIs. The controls are ordinary framework
buttons: halve/double exposure, reset to 1, and capture. Interactive Capture reads
the completed GPU frame asynchronously, then opens a native PNG save chooser on a
worker. PNG saving uses a same-directory temporary, fsync and atomic replacement,
so encoding/cancellation errors preserve an existing document. The status label
reports progress; only one capture/save is retained by
this UI at a time. No GPU readback happens during normal rendering.

The real three-stage program runs entirely on the window's shared device:

1. Compute writes a 256 × 256 storage buffer with premultiplied HDR pixels. Left
   straight RGB is `(2, .5, 0)`, right is `(0, .25, 3)`, alpha is `.5`, and the top
   quarter is transparent. Exposure multiplies RGB before premultiplication.
2. A single-sample render pass copies storage pixels into an `RGBA16Float`
   attachment, preserving values above 1.
3. The main four-sample sRGB pass loads the float attachment, recovers straight
   RGB, tone maps `x / (1 + x)`, and premultiplies again for composition.

Tracked resources are 1,048,576 bytes of storage, 524,288 bytes of float texture,
and a 32-byte uniform. Only that uniform uploads on each changed frame. Normal
resize replaces the framework target while retaining storage, float attachment,
groups and pipelines. Program-owned groups/pipelines/views retire before the
framework releases helper resources. Creation uses checked APIs; the pinned
binding's sole panic-based compute-pass creation is converted to a program error.

After preparing native dependencies, run from the repository root:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
. ./drivers/cogent/env.sh
cd examples/imaging
go run .
```

`go test -race ./...`, `go vet ./...`, and `go build -race .` check this module.
`go run -race . -smoke` runs bounded native acceptance in two temporary windows.
It checks actual HDR/partial-alpha/transparent pixels at exposures 1 and 2,
ordinary native controls, acknowledged chrome retention, 32-byte-only frame
uploads, retained float/storage across resize, capture-time pixels/metadata independent of subsequent model/resize changes, callback
recursive busy admission, quiet cancellation/rebuild/closure, independent windows,
native nil/canceled/byte-budget refusals, explicit sync/async staging coexistence,
and capture completion while minimized without another frame. The fixture pumps
the application owner mailbox explicitly while minimized; GPU polling/decoding
progresses on the adapter worker without a paint ticker. It also initiates native
force-close directly and checks dropped delivery and surviving-window pixels.
Closed Stats cannot expose private staging/reservation counters. Adapter regressions
establish the fallback control flow and reservation release; ownership review checks
the private staging reference. This example does not measure private buffer release
after native closure.

The replacement directives point to this checkout's public root/driver modules
and prepared dependencies; module origins were checked when creating the module.
These checks verify offscreen GPU images, not native presentation timing, physical
device-loss recovery, Windows/Linux qualification, or a sustained performance
budget. No OS input sources, accessibility settings, or preferences are changed.
The code is MIT licensed with the repository.

The smoke reports whether old-frame delivery actually overlapped target replacement;
it does not assume that scheduling overlap. Adapter regressions establish delivery
of capture-time metadata across a normal target generation change independently.

Current macOS acceptance and its scope are recorded in
[acceptance/README.md](acceptance/README.md), including two passing native runs,
exact compiler-input hashes, and the preserved failures that led to the final fixes.
