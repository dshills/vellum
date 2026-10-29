# Contributing

Vellum is a Go library with a separate native driver and independently runnable
example modules. Start with the [repository layout](docs/repository-layout.md)
and [development guide](docs/development.md).

## Portable framework

Use the Go version declared in the root `go.mod`. From the repository root:

```sh
go test -race ./...
go vet ./...
```

Keep public framework packages in their existing directories and put tests beside
the code they exercise. Use `internal/` for genuinely private supporting packages
when needed; keep existing public import paths stable. Test fixtures belong in
`testdata/` beside their package. Format changed Go files with `gofmt`.

## Native driver and examples

Native dependencies belong to `drivers/cogent`, not the root module. Prepare the
pinned sources before working in that module or a native example:

```sh
./drivers/cogent/prepare.sh
. ./drivers/cogent/env.sh
cd drivers/cogent
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
```

Each [example](examples/README.md) has its own module. Run its checks from that
directory; root `go test ./...` does not cross module boundaries. For example,
after preparation, from the repository root:

```sh
. ./drivers/cogent/env.sh
cd examples/placement
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
go run .
```

Use focused checks for the changed module. The full prepared-repository workflow
is `scripts/check.sh`; live desktop checks are separate and documented in the
development guide. Keep generated outputs in ignored `build/` directories.

## Documentation and scope

Update application instructions or `specs/` when behavior changes. Guides belong
in `docs/`; historical reports and reviews retain their original paths and hashes.
The reports' `go.mod` is an archive boundary, not a dependency to import or test.
Optional personal `go.work` files are local development settings and are ignored;
the module-local replacements remain the reproducible dependency setup.

Keep changes focused. The current [macOS v1 closeout](specs/macos-v1-release.md)
freezes features and experimental rendering/pacing. Preserve the
[stable v1 API contract](specs/stable-api.md) when maintaining supported APIs.
Include changes, checks and limitations in commit or pull-request descriptions.
Use internal code review at significant checkpoints and record its scope,
findings, fixes and verification before committing. External review services
(including Prism) require a new explicit user request; historical receipts do not
authorize further source transmission.

Before changing a stable root API, run the separate developer guard:

```sh
(cd tools/apicompat && go test -race ./... && go vet ./...)
(cd tools/apicompat && go run . check -root ../.. -baseline ../../release/api/v1)
```

The baseline is reviewed release data; checks never regenerate it. Follow the
[guard's scope and conservative findings](tools/apicompat/README.md). Behavioral
and native-facade changes retain their existing testing/review requirements.

## Native dependency maintenance

Vellum owns a substantial pinned dependency patch set. Before changing a patch or
pin, follow [ADR 0012](docs/decisions/0012-native-dependency-maintenance.md).
`python3 -B scripts/native-patch-inventory.py --check` refuses unrecorded drift.
After reviewing patch ownership and upgrade impact, regenerate the inventory with
`python3 -B scripts/native-patch-inventory.py > release/native-patch-inventory.json`
and include it in the same reviewed checkpoint. The inventory is descriptive,
not a performance target or an approval to publish dependencies.
