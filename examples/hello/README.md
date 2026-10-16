# Independent Go application

This module consumes Vellum's portable and Cogent modules through its own `go.mod`.
The counter and GPU rectangle use public APIs. Native replacements appear here
explicitly because Go does not inherit replacements from dependency modules.

From the repository root:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./drivers/cogent/prepare.sh
./scripts/check-example.sh
. ./drivers/cogent/env.sh
cd examples/hello
go run .
```

For an application outside this checkout, run `python3 scripts/new-app.py /path/to/my-app
--module example.com/my-app` from the repository root. The generated module points
at absolute local source paths; keeping or relocating this checkout changes those
paths. The framework namespace is `github.com/dshills/vellum`; public distribution
qualification remains open.

`-benchmark` launches the same application, verifies the rendered rectangle through
GPU readback, emits a JSON measurement, waits quietly, and closes. The normal app
performs no measurement readbacks. Benchmark mode disables native window-geometry
loading and saving before app construction, keeps the process-local switch set
through teardown, and restores its prior value on normal return. Its JSON report
includes the actual `geometry_persistence_disabled_in_fixture` flag. Generated
applications inherit this benchmark behavior; ordinary interactive mode retains
normal geometry persistence. See [SDK contract](../../specs/sdk.md).
