# Vellum

Vellum is an experimental Go-native professional UI framework. Applications use
ordinary Go values, stable widget IDs, explicit commands and cancellable tasks.
A separate native driver reuses Cogent Core's windowing and WebGPU stack.

Production readiness, native accessibility, international input, and Windows/Linux
runtime validation remain open. The workbench does not change system input sources.

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
| `scripts/` | Repository checks and macOS fixture packaging |
| `prototypes/` | Independent experimental modules and historical foundation evidence |

Public packages stay at the module root so framework consumers can import them
directly. Executable code lives under `cmd/` in the module that owns its native
dependencies. Prototype modules remain separate from the framework API.

## Build and run

The dependency-free root module is locally named `example.com/vellum` until a
publication namespace is chosen. Test the portable packages from the repository root:

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
[implementation status](docs/implementation.md) for the five completed phases.

The [independent hello application](examples/hello/README.md) has its own module.
Use `python3 scripts/new-app.py /path/to/my-app --module example.com/my-app`
after preparation to create an application outside the checkout. Native patches
and preparation inputs belong to the driver; prototype directories are not needed.
See the [local SDK contract](specs/sdk.md) and run `scripts/benchmark.sh` for
portable baselines or `scripts/benchmark.sh --native` for visible GPU measurements.
