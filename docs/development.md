# Development

## Modules and dependencies

The root `example.com/vellum` module contains portable public packages and uses
only the standard library. The `drivers/cogent` module depends on that local root
module and the pinned native stack. Keep driver-specific dependencies in that
module. Each prototype has its own module and records its experimental inputs.

The driver module's repository-relative replacements point to prepared Cogent and
go-text source copies under `prototypes/`. `drivers/cogent/prepare.sh` downloads
exact pseudo-versions, verifies patch-input SHA-256 hashes and applies committed
patches to owned copies. Preparation preserves the reproducible experimental
foundation; this is a repository-local development distribution.

## Checks

Run portable checks from the repository root:

```sh
go test -race ./...
go vet ./...
```

For the full portable and native check, bootstrap first:

```sh
export P0_CACHE_DIR=/tmp/vellum-p0-cache
./drivers/cogent/prepare.sh
./scripts/check.sh
```

The script runs root race tests and vet, native-module tests and vet, then builds
`drivers/cogent/build/vellum-workbench`. The macOS CI workflow follows the same
bootstrap and check sequence from a fresh checkout. Native checks need the C
compiler, graphics libraries and GPU dependencies described in the root README.

To run the native example after preparation:

```sh
cd drivers/cogent
. ../../prototypes/foundationintegration/env.sh
go run ./cmd/workbench
```

Generated `build/` directories and scratch reviews under `docs/reviews/tmp/`
are ignored. Committed
prototype evidence and review records remain part of the repository. Capture-time
paths and hashes in evidence are historical records; moving documentation does
not regenerate measurements or certify additional platform behavior.
