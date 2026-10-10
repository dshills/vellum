# Local SDK consumption and performance measurements

The root framework module stays standard-library-only. The native adapter is a
separate module, `github.com/dshills/vellum/drivers/cogent`, matching its repository
directory. The framework namespace is `github.com/dshills/vellum` and its own code
is MIT-licensed. Public distribution remains gated on the dependency/native audit
and target qualification.

## Native dependency preparation

`drivers/cogent/prepare.sh` and `env.sh` own preparation and Go cache setup.
Preparation reads only driver-owned inputs. Its manifest pins Cogent and go-text
module versions/checksums, patch digests, original input hashes and patch order.
It downloads immutable module source, retains upstream license files, copies into
writable owned staging directories, verifies hashes and applies patches there. The driver multiwindow safety stage
also verifies expected final source hashes and runs its ownership regression tests
through the repository check script.
Both dependencies must pass before replacing existing prepared directories. Failure
rolls back installed directories. If filesystem errors prevent full rollback,
preparation retains original backups in the reported recovery staging directory
and reports their exact paths. Inspect and restore those backups before removing
the recovery directory. A filesystem lock rejects concurrent preparation.
A killed process can leave the lock; inspect for an active process before removing it.

Driver replacements resolve to `drivers/cogent/build/upstream` and
`drivers/cogent/build/typesetting`. Prototype modules and evidence remain historical
and are not required to prepare, test, build or consume the framework modules.

## Independent application module

`examples/hello` has an independent `go.mod`, imports public framework packages and
the native driver, and renders a working counter and GPU rectangle. Its explicit
replacements cover both framework modules and both patched native dependencies.
Dependency-module replacements are not transitive in Go.

`scripts/new-app.py` creates the same application in a new directory, with a caller
selected application module path and absolute local framework/dependency replacements.
It retains the starter template's MIT LICENSE notice in the generated application.
It refuses existing destinations and requires preparation first. The generated
application can be tested and built outside the repository with `GOWORK=off`; it
still requires this checkout for its local SDK dependencies. Publishing versions,
installing a remotely fetched SDK, and relocating those dependencies are separate
release tasks.

## Repeatable measurements

`scripts/benchmark.sh` runs existing portable data benchmarks plus the standalone
application's UI construction/validation benchmark, with allocation reporting and
three repetitions. `--native` additionally builds and launches three visible native
processes via `benchmark-native.py` on a desktop with a supported GPU.

Every native run reports:

- Startup from application entry to a pixel-verified GPU rectangle; external process
  launch-to-report timing is also recorded. Both include initialization; first frame
  readiness is polled at 25 ms, so this is not compositor presentation latency.
- Ten synchronous GPU render/readback durations. Readback forces rendering; these
  observations include transfer/synchronization and are not ordinary frame FPS.
- Go heap allocation, native image dimensions, runtime/platform/machine metadata,
  and resident memory samples before/after the quiet interval.
- Process CPU-time difference during 60% of the quiet interval after a brief render-settling delay, expressed as seconds
  and percentage of one core. Linux uses `/proc` CPU ticks; other supported
  hosts use `ps` accounting. Reports record the sampling source and time quantum.
  Accounting precision limits small deltas; a zero observed delta does not prove
  zero idle CPU. Linux accounting parsing has fixture tests; native Linux runtime
  measurements remain unverified.

The harness fails if no verified GPU report arrives, the application exits with an
error, or sampling fails. Measurements provide host-specific baselines without
invented pass thresholds. Cold OS/font caches, other desktop processes, GPU/power
state and hardware affect results; comparisons should record matching conditions.
Windows/Linux runtime performance and accessibility/input correctness remain gates.

The public `document` package ships in private SDK archives alongside app/ui/history.
A relocated archive consumer exercises bounded file persistence after deleting
the fixture source checkout. Session/close policies follow [document.md](document.md).
