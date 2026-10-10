# Vellum

Vellum is an experimental Go-native professional UI framework. Applications use
ordinary Go values, stable widget IDs, explicit commands and cancellable tasks.
A separate native driver reuses Cogent Core's windowing and WebGPU stack.

The framework implementation starts here; `prototypes/` preserves research and
foundation evidence. Production readiness, native accessibility, international
input, and Windows/Linux runtime validation remain open. System input sources
are not changed by the workbench.

## Build and run

The root module is dependency-free and locally named `example.com/vellum` until a
publication namespace is chosen. Run `go test -race ./...` here. The native driver
is a separate module so logic tests do not initialize an OS or GPU runtime.

Bootstrap the pinned, hash-checked patched dependencies, then build:

```sh
cd drivers/cogent
export P0_CACHE_DIR=/tmp/vellum-p0-cache
./prepare.sh
. ../../prototypes/foundationintegration/env.sh
go run ./cmd/workbench
```

The driver module uses repository-relative replacements for the prepared,
reviewed Cogent and go-text copies. It requires Go, a C compiler, the platform's
native graphics libraries, and a supported GPU. Current runtime evidence covers
macOS only. The initial adapter supports one native window per process and must
run from main. Additional windows and platform adapters remain future work.

`app.Application.Post` provides bounded worker delivery with explicit refusal.
Application state changes in `Do`, posted callbacks, or UI actions; views run
under the same serialized owner. Calls to `Do` and `Drain` are not reentrant.
Views only describe UI; they do not start workers or mutate domain state.

`drivers/cogent/prepare.sh` downloads exact pseudo-versions, verifies patch input
SHA-256 hashes, then applies the committed patches to owned source copies.
`scripts/check.sh` explicitly requires that bootstrap. The macOS CI workflow
runs bootstrap from a fresh checkout before tests and the native build. This is
a repository-local development distribution, not yet a published Go dependency.

Queue capacity bounds pending callbacks, plus the one callback currently executing.
Driver wake functions are nonblocking notifications and never enter the UI owner.
Native wakes use Cogent's lossless custom-event deque; they are coalesced while
one wake is outstanding. A panic closes the application and propagates.
