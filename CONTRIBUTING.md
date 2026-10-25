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

Keep changes focused. The approved macOS-v1 work may extend framework APIs only within
[its five-area contract](specs/macos-v1-candidate.md); experimental rendering stays disabled. Include what changed, the checks
run, and any limitations in the commit or pull-request description.
