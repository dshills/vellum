# Vellum

Vellum is a desktop UI framework written in Go, with a current focus on macOS.
Applications compose UI descriptions from Go values and stable widget IDs, then
handle changes through explicit commands, cancellable tasks, and Undo/Redo history.
A separate native driver uses Cogent Core for windowing and WebGPU rendering.

**Status:** [v0.1.0-alpha.3](https://github.com/dshills/vellum/releases/tag/v0.1.0-alpha.3)
is a published macOS developer preview for Apple Silicon. Its SDK and Placement
app are experimental; the app is ad-hoc signed and unnotarized. Windows and Linux
remain unqualified. Full production and numeric performance qualification remain
separate from this preview under the [approved release plan](specs/macos-first-release-plan.md).

The current source also includes newer macOS widget work for personal evaluation.
Those widgets are not included in the published alpha.3 assets. See the
[Mac widget demo](docs/mac-widget-demo.md) and
[public-API integration guide](docs/mac-widget-integration.md) to try them from source.

## What is available

- UI controls, keyboard focus, native macOS text input and accessibility integration.
- Multiple windows, dockable workbench panels, menus, clipboard and file dialogs.
- Commands, bounded Undo/Redo, document persistence and cancellable background work.
- Virtual grids and trees with worker-built data views.
- Palette-to-canvas placement, retained 2D scenes, GPU meshes, images and custom 3D views.
- Offscreen routing tests, deterministic replay and opt-in local automation.

See the [framework contract](specs/framework.md) and
[implementation status](docs/implementation.md) for subsystem behavior and limits.

## Quick start on macOS

Install the Go version declared in [go.mod](go.mod), Python 3, Git, `patch`, and
Xcode Command Line Tools. Native applications require a supported GPU. Current
internal app testing targets Apple Silicon macOS.

```sh
git clone https://github.com/dshills/vellum.git
cd vellum

# Prepare the pinned native dependencies and local build environment.
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./drivers/cogent/prepare.sh
. ./drivers/cogent/env.sh

# Launch the drawing test application.
(cd examples/placement && go run .)
```

Preparation verifies pinned dependency sources and applies the driver-owned
patches. The native driver and example modules use explicit local replacements;
follow this preparation workflow when building from the checkout.

### Drawing test application

The [placement application](examples/placement/README.md) demonstrates:

- Drag components from a palette onto a pannable, zoomable canvas with optional snapping.
- Select, move, resize, group, reorder and edit component fills.
- Align, distribute, match dimensions and create repeated copies of selected units.
- Use a layers panel, command search, keyboard shortcuts and built-in help.
- Save/Open drawings, recover interrupted work, export SVG and Undo/Redo edits.

Placement stays within one window, groups are flat, and fills use opaque RGB.
Rotation, nested groups and SVG import are not implemented. See the application's
README for bounds, file compatibility and verification details.

## More examples

The following commands assume dependencies are prepared and the shell is at the
repository root with `drivers/cogent/env.sh` sourced.

| Example | Run | Purpose |
| --- | --- | --- |
| [Hello](examples/hello/README.md) | `(cd examples/hello && go run .)` | Minimal public-API application |
| [Mac widgets](docs/mac-widget-demo.md) | `(cd drivers/cogent && go run ./cmd/gallery -mac-widgets)` | Light/dark Mac input controls, color well, tags, font/path pickers and protected entry |
| [Workbench](drivers/cogent/cmd/workbench/) | `(cd drivers/cogent && go run ./cmd/workbench)` | Docking and composed UI |
| [Calculator](examples/macapps/calculator/README.md) | `(cd examples/macapps && go run ./cmd/calculator)` | Fixed-size calculator with arithmetic and keyboard input |
| [Notes](examples/macapps/notes/README.md) | `(cd examples/macapps && go run ./cmd/notes)` | Folders, search, plain-text editing and Undo/Redo |
| [Weather](examples/macapps/weather/README.md) | `(cd examples/macapps && go run ./cmd/weather)` | Sample cities, forecasts and unit selection |
| [Reminders](examples/macapps/reminders/README.md) | `(cd examples/macapps && go run ./cmd/reminders)` | Lists, details and completion |

The macOS-style examples use sample data and approximate familiar application
layouts; they do not connect to Apple's services. The [example index](examples/README.md)
also covers large-data grids/trees, GPU drawing, imaging, 3D, replay and automation.

## Build your own application

After preparation, create an independent application module outside this checkout:

```sh
python3 scripts/new-app.py ../my-vellum-app --module example.com/my-vellum-app
(cd ../my-vellum-app && go run .)
```

The generated module uses explicit replacements pointing at this local checkout.
See the [application tutorial](docs/tutorials/building-an-app.md) and
[SDK contract](specs/sdk.md) for consumption and relocation details.

## Repository layout

| Path | Purpose |
| --- | --- |
| `app/`, `ui/`, `canvas/`, `placement/` | Public application, UI and canvas APIs |
| `command/`, `history/`, `document/`, `editor/` | Commands, persistence and editing models |
| `tasks/`, `execution/`, `virtual/`, `workbench/` | Background work, scheduling, data views and docking |
| `replay/`, `automation/` | Deterministic input and local semantic automation |
| `drivers/cogent/` | Native driver module; executable tools under `cmd/` |
| [examples/](examples/README.md) | Independent application and contract modules |
| [docs/](docs/README.md), [specs/](specs/README.md) | Guides, contracts, research and historical evidence |
| `scripts/`, `release/`, `prototypes/` | Build/packaging tools, notices and isolated historical experiments |

The root module, `github.com/dshills/vellum`, uses only the standard library.
Native dependencies belong to `github.com/dshills/vellum/drivers/cogent`.
Historical reports have a separate module boundary and stay out of framework
module downloads. See the [layout guide](docs/repository-layout.md).

## Development and checks

From the repository root, portable checks are:

```sh
go test -race ./...
go vet ./...
```

Each driver/example module is checked separately. For example, after preparation:

```sh
(cd examples/placement && go test -race -tags offscreen ./...)
```

Offscreen tests do not open desktop windows. Live input, accessibility, GPU and
display checks have separate verification requirements. The full prepared-repo
workflow is `scripts/check.sh`; see [CONTRIBUTING](CONTRIBUTING.md) and the
[development guide](docs/development.md) before running native campaigns.

## License and release status

Vellum's own code is [MIT licensed](LICENSE). Third-party components retain their
original licenses and notices. The owner approved distribution of the exact
alpha.3 payload and its notice bundle. That decision does not cover future changed
payloads; new public artifacts still require the applicable audit, notice checks
and publication approval. See the [published release and limitations](https://github.com/dshills/vellum/releases/tag/v0.1.0-alpha.3),
[release contract](specs/release.md) and [release metadata](release/version.json).
