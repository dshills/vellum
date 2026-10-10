# Development

## Modules and dependencies

The root `example.com/vellum` module contains portable public packages and uses
only the standard library. The `drivers/cogent` module depends on that local root
module and the pinned native stack. Keep driver-specific dependencies in that
module. Each prototype has its own module and records its experimental inputs.

The driver module's repository-relative replacements point to prepared Cogent and
go-text source copies under `drivers/cogent/build/`. `drivers/cogent/prepare.sh` downloads
exact pseudo-versions, verifies patch-input SHA-256 hashes and applies committed
patches from driver-owned `native/` inputs to owned copies. Preparation has no
prototype pipeline dependency; this is a repository-local development distribution.

## Checks

Run portable checks from the repository root:

```sh
go test -race ./...
go vet ./...
```

For the full portable and native check, bootstrap first:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./drivers/cogent/prepare.sh
./scripts/check.sh
```

The script runs root race tests and vet, native-module tests and vet, then builds
`drivers/cogent/build/vellum-workbench`. It also race-tests, vets and builds the
independent `examples/hello` module, validates dependency preparation inputs and race-tests the upstream multiwindow
ownership regressions. The macOS CI workflow follows the same
bootstrap and check sequence from a fresh checkout. Native checks need the C
compiler, graphics libraries and GPU dependencies described in the root README.

To run the native example after preparation:

```sh
cd drivers/cogent
. ./env.sh
go run ./cmd/workbench
```

Generated `build/` directories and scratch reviews under `docs/reviews/tmp/`
are ignored. Committed
prototype evidence and review records remain part of the repository. Capture-time
paths and hashes in evidence are historical records; moving documentation does
not regenerate measurements or certify additional platform behavior.

## Standalone applications and benchmarks

`examples/hello` is a separately versioned consumer module. To create a consumer
outside the repository, source `drivers/cogent/env.sh` and run
`python3 scripts/new-app.py /path/to/my-app --module example.com/my-app`.
Its four explicit replacements point to the local framework and prepared native
modules. Run `go test -race ./...`, `go vet ./...` and `go build .` in the result.

Run `scripts/benchmark.sh` for portable allocation/time baselines, or add `--native`
to launch three desktop processes and measure verified GPU startup, synchronous
render/readback, RSS and sampled idle CPU. The [SDK contract](../specs/sdk.md)
defines measurement boundaries and residual release/platform gates.

Set `VELLUM_CACHE_DIR` to reuse a chosen Go cache across preparation/check commands.
The legacy `P0_CACHE_DIR` variable is accepted for existing local environments.
Without either variable, sourcing the environment creates an isolated temporary cache.
