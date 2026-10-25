# Examples

These applications are independent Go modules consuming Vellum's public APIs.
Run a command from the example's own directory. Prepare native dependencies using
the [development guide](../docs/development.md) before running native examples.

| Module | Purpose |
| --- | --- |
| [hello](hello/README.md) | Minimal native application |
| [personal-mac](personal-mac/README.md) | Frozen personal Mac starter: system appearance and full-color native toolbar |
| [placement](placement/README.md) | Internal drawing test app: palette/canvas, selection, editing, files and Undo/Redo |
| [macapps](macapps/README.md) | Calculator, Notes, Weather and Reminders demonstrations under `cmd/` |
| [editor](editor/README.md) | Managed multiline editing and document persistence |
| [grid](grid/README.md) | Procedural large-grid demonstrations |
| [loadedgrid](loadedgrid/README.md) | Stored-data grid and background sort/filter |
| [loadedtree](loadedtree/README.md) | Stored large-tree navigation and filtering |
| [drawing](drawing/README.md) | Retained GPU drawing |
| [imaging](imaging/README.md) | Image processing and frame capture |
| [viewport3d](viewport3d/README.md) | Custom GPU viewport |
| [replay](replay/README.md) | Deterministic input replay and workspace consumers |
| [automation](automation/README.md) | Local semantic automation client |
| [sdk-contract](sdk-contract/) | Independent public-API contract checks |

From the repository root, after preparation:

```sh
. ./drivers/cogent/env.sh
cd examples/placement
go test -race -tags offscreen ./...
go run .
```

Root `go test ./...` does not include these nested modules. Follow each README
for its commands and applicable offscreen/live checks. Build outputs go under
ignored `build/` directories. See [CONTRIBUTING](../CONTRIBUTING.md) for the module
and verification workflow.
