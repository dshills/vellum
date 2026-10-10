# Vellum

Vellum is an experimental Go-native professional UI framework. Applications use
ordinary Go values, stable widget IDs, explicit commands and cancellable tasks.
A separate native driver reuses Cogent Core's windowing and WebGPU stack.

The framework includes multiwindow docking, native macOS accessibility, standard
controls, virtual data views and a shared-device GPU canvas. Production readiness,
complete assistive/input acceptance and Windows/Linux runtime qualification remain
open. The workbench does not change system input sources.

## Repository layout

| Path | Purpose |
| --- | --- |
| `app/`, `ui/` | Application lifecycle, UI ownership and keyed widget descriptions |
| `command/`, `tasks/` | Commands and bounded asynchronous work |
| `workbench/`, `virtual/`, `canvas/` | Dock layouts, large-data views and canvas models |
| `drivers/cogent/` | Separate native driver module |
| `drivers/cogent/cmd/workbench/` | Executable workbench example in the native module |
| `examples/hello/` | Standalone application module consuming the local SDK |
| [docs/](docs/README.md) | Development guide, research, decisions, reports and review evidence |
| [specs/](specs/README.md) | Framework contract and foundation validation plan |
| `scripts/` | Repository checks, SDK release and private application packaging |
| `prototypes/` | Independent experimental modules and historical foundation evidence |

Public packages stay at the module root so framework consumers can import them
directly. Executable code lives under `cmd/` in the module that owns its native
dependencies. Prototype modules remain separate from the framework API.

## Build and run

The dependency-free root module is `github.com/dshills/vellum`; the native driver is
`github.com/dshills/vellum/drivers/cogent`. Test the portable packages from the repository root:

```sh
go test -race ./...
```

Prepare the pinned native dependencies and run the example:

```sh
cd drivers/cogent
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./prepare.sh
. ./env.sh
go run ./cmd/workbench
```

The native driver requires Go, a C compiler, platform graphics libraries and a
supported GPU. Runtime evidence currently covers macOS. It supports multiple native
windows in one process main loop and must run from main.

See the [development guide](docs/development.md) for the full check workflow,
[framework contract](specs/framework.md) for implemented behavior and limits, and
[implementation status](docs/implementation.md) for completed phases and remaining acceptance gates.

The [independent hello application](examples/hello/README.md) has its own module.
Use `python3 scripts/new-app.py /path/to/my-app --module example.com/my-app`
after preparation to create an application outside the checkout. Native patches
and preparation inputs belong to the driver; prototype directories are not needed.
See the [local SDK contract](specs/sdk.md) and run `scripts/benchmark.sh` for
portable baselines or `scripts/benchmark.sh --native` for visible GPU measurements.

The SDK identifies `v0.1.0-alpha.1`. Follow the
[application tutorial](docs/tutorials/building-an-app.md) and
[release contract](specs/release.md) to create and verify a relocatable private SDK.
Vellum's own code uses the [MIT license](LICENSE). The selected module namespace
and license are recorded in `release/version.json`; public packaging still requires
the third-party/native distribution audit and target qualification.
