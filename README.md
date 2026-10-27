# Vellum

Vellum is a Go desktop UI framework focused on macOS look and feel. Applications
compose UI descriptions with stable widget IDs and manage state through explicit
commands, cancellable tasks, document persistence and Undo/Redo history.

The native driver uses Cogent Core and WebGPU for windowing and rendering.
Managed widgets use shared Mac typography, spacing and themes; AppKit provides
native titlebar toolbars, file dialogs, alerts, Quick Look, and Colors/Fonts panels.
Full-color toolbar and picker artwork remains untinted.

## See it in action

Three original, runnable Mac apps built with Vellum's public Go APIs. These
windows use managed Mac-themed UI, native AppKit toolbars and full-color icons.
The screenshots are actual app captures; artwork and data are original fixtures.

### Forma — a space for ideas

A poster studio combining retained GPU geometry, an image layer, layer selection,
a color/size inspector and command-based Undo/Redo. Pan, zoom and fit the artboard.

![Forma design studio with an editable poster, layers and color inspector](docs/screenshots/forma.jpg)

Run: `(cd examples/showcase && go run ./cmd/forma)`

### Luma — a library of light

An image browser with a selectable filmstrip, collections, favorites, search,
nondestructive color variants and a histogram sampled from the displayed image.
Every landscape is generated locally in Go.

![Luma image library showing original landscape artwork, filmstrip and histogram](docs/screenshots/luma.jpg)

Run: `(cd examples/showcase && go run ./cmd/luma)`

### Pulse — a clearer signal

A dark telemetry explorer pairing retained GPU charts with a virtualized,
indexed table of 1,024 fictional service records. Filter, sort, inspect a service
or advance the repeatable sample data. No monitoring service is contacted.

![Pulse dark telemetry explorer with a chart, selected service and details](docs/screenshots/pulse.jpg)

Run: `(cd examples/showcase && go run ./cmd/pulse)`

Prepare and source the driver environment using the Quick start below first.
See the [showcase guide](examples/showcase/README.md) for controls, private app
packaging and limits, and [screenshot reproduction](docs/screenshots/README.md)
for source identities and capture details. These examples reset on launch and
use the existing experimental GPU surface APIs; they do not expand the stable
API or release qualification.

## Current status

**Latest tagged source release: [v1.0.0](https://github.com/dshills/vellum/releases/tag/v1.0.0).**
This release provides the Go framework source, macOS examples and build tools
for personal use. Read the [release notes](release/notes/v1.0.0.md) for highlights
and the [stable API contract](specs/stable-api.md) for compatibility guarantees
and experimental exclusions. See the showcase apps above for current UI examples.

Build native apps from the tagged checkout using Quick start below. GitHub's
source ZIP/tar archives are checkout source, not prepared SDKs or installed apps;
run the pinned dependency preparation step before native builds. The separately
versioned native driver is consumed through the checkout's local replacements.
This release creates the root `v1.0.0` tag; it does not add native-module tags.

The latest prebuilt native download remains the older
[v0.1.0-alpha.3 developer preview](https://github.com/dshills/vellum/releases/tag/v0.1.0-alpha.3).
Its SDK and Placement app do not contain the current Notes, widget or showcase
work. Current v1 SDK/app downloads remain unpublished. The owner approved the
[current payload audit](release/macos-v1-distribution-audit.json) on 27 October;
[final local packaging checks](docs/reports/macos-v1-public-packaging/README.md)
produced an SDK and five recommended app candidates. The optional Mac Shell fixture
is held for dark-appearance contrast. No v1 native download has been uploaded.
The candidates are ad-hoc signed and unnotarized. The
[Mac guide](docs/macos-v1.md) and [private release report](docs/reports/macos-v1-release/README.md)
retain their exact build identities, supported behavior and verification limits.

Native acceptance covers the recorded Apple Silicon macOS host. Human
IME candidate behavior and VoiceOver usability remain unqualified, as do other
macOS/hardware/display configurations and Windows/Linux. Wayland is unsupported.
Replay, automation, diagnostics and broader GPU/workbench workloads remain
experimental. See the [support matrix](docs/macos-v1.md#supported-api-contract).

## What is available

- **Mac controls and themes:** text/search fields, prose editors, selectors,
  steppers, sliders, calendars, lists/outlines, inspectors, sheets and popovers.
- **Toolbars:** full-color icons, multiple sizes and display styles, overflow,
  spacers, customization and persistence; native titlebar toolbars can have
  independent commands and availability for each window.
- **Content widgets:** controlled checklists, file/folder fields, date ranges,
  embedded PDFKit documents and AVKit media with original-color posters on macOS.
  [Contracts and native limits](docs/controls.md).
- **Application services:** multiple windows, command menus and shortcuts,
  clipboard, file dialogs, native panels and accessibility integration.
- **Application state:** commands, bounded Undo/Redo, document persistence,
  interrupted-session recovery and cancellable background work.
- **Data and graphics:** virtual grids/trees, dockable workbench panels,
  same-window palette placement, retained 2D scenes, images, meshes and custom 3D views.
- **Testing tools:** offscreen routing tests, deterministic replay and opt-in
  local semantic automation, with separate physical/native acceptance.

Browse the [widget demo](docs/mac-widget-demo.md),
[control contracts](docs/controls.md), [content integration examples](docs/mac-content-integration.md), [app-shell APIs](docs/macos-app-shell.md),
and [framework contract](specs/framework.md) for behavior and subsystem limits.

## Quick start on macOS

Install the Go version declared in [go.mod](go.mod), Python 3, Git, `patch`, and
Xcode Command Line Tools. Native applications require a supported GPU; current
personal-use testing targets Apple Silicon macOS.

```sh
git clone --branch v1.0.0 --depth 1 https://github.com/dshills/vellum.git
cd vellum

# Prepare pinned native sources and the local build environment.
export VELLUM_CACHE_DIR=/tmp/vellum-go-cache
./drivers/cogent/prepare.sh
. ./drivers/cogent/env.sh

# Explore Mac widgets with an AppKit titlebar toolbar and system appearance.
(cd drivers/cogent && go run ./cmd/gallery -mac-widgets -native-titlebar -mac-system-appearance)
```

Preparation verifies dependency sources and applies driver-owned patches. The
native driver and example modules use explicit local replacements. Build from
the checkout with this workflow, or use the independently extracted SDK route in
the [Mac guide](docs/macos-v1.md#consume-the-private-sdk).

## Example applications

After preparation, run these commands from the repository root with
`drivers/cogent/env.sh` sourced. Each command launches a separate application.

| Example | Run | Demonstrates |
| --- | --- | --- |
| [Forma](examples/showcase/README.md#forma--design-studio) | `(cd examples/showcase && go run ./cmd/forma)` | Original poster, retained GPU canvas, layer inspector and Undo/Redo |
| [Luma](examples/showcase/README.md#luma--image-library) | `(cd examples/showcase && go run ./cmd/luma)` | Original image assets, filmstrip selection, collections, color variants and sampled histogram |
| [Pulse](examples/showcase/README.md#pulse--telemetry-explorer) | `(cd examples/showcase && go run ./cmd/pulse)` | Dark theme, retained charts and virtualized service fixtures |
| [Mac Widgets](docs/mac-widget-demo.md) | `(cd drivers/cogent && go run ./cmd/gallery -mac-widgets -native-titlebar -mac-system-appearance)` | Frozen widget catalogue, scoped themes, full-color icons, toolbar customization and native panels |
| [Notes](examples/macapps/notes/README.md) | `(cd examples/macapps && go run ./cmd/notes)` | Proportional prose with emoji, folders/search, native toolbar and menus, local Save/Open, recovery and Undo/Redo |
| [Native shell](examples/macapps/cmd/macshell/main.go) | `(cd examples/macapps && go run ./cmd/macshell)` | Two windows with independent native toolbar commands and disabled states |
| [Drawing app](examples/placement/README.md) | `(cd examples/placement && go run .)` | Palette placement, pan/zoom, selection, editing, layers, Save/Open and Undo/Redo |
| [Mac starter](examples/personal-mac/README.md) | `(cd examples/personal-mac && go run .)` | Small public-API consumer with a full-color native toolbar and system appearance |
| [Hello](examples/hello/README.md) | `(cd examples/hello && go run .)` | Minimal public-API application |
| [Calculator](examples/macapps/calculator/README.md) | `(cd examples/macapps && go run ./cmd/calculator)` | Fixed-size calculator with arithmetic and keyboard input |
| [Weather](examples/macapps/weather/README.md) | `(cd examples/macapps && go run ./cmd/weather)` | Sample cities, forecasts and unit selection |
| [Reminders](examples/macapps/reminders/README.md) | `(cd examples/macapps && go run ./cmd/reminders)` | Public checklists, scheduled date ranges, details and Undo/Redo |
| [Workbench](drivers/cogent/cmd/workbench/) | `(cd drivers/cogent && go run ./cmd/workbench)` | Docking and composed UI |

The macOS-style examples approximate familiar application layouts and use sample
data; they do not connect to Apple's services. Notes stores its local notebook in
`~/Library/Application Support/Vellum Notes`; use `-data-dir /absolute/test/path`
for a separate test notebook. It remains a plain-text app without iCloud or rich
attachments. See its guide for storage, locking and recovery behavior.

The drawing app supports moving/resizing, grouping, alignment/distribution,
layer reordering and SVG export. Placement stays in one window, groups are flat,
and fills use opaque RGB. Rotation, nested groups and SVG import are not implemented.
The [example index](examples/README.md) covers additional data, graphics and
experimental automation consumers.

## Build your own application

Create a separate Go application module after preparing this checkout:

```sh
python3 scripts/new-app.py ../my-vellum-app --module example.com/my-vellum-app
(cd ../my-vellum-app && go test ./... && go run .)
```

The generated module uses replacements pointing at this checkout. The private
SDK supports the same generator with replacements contained inside its extracted
archive; no repository checkout is needed. Do not run `prepare.sh` over the SDK's
pinned native capsule.

Start with [building a macOS app](docs/building-macos-apps.md), the
[application tutorial](docs/tutorials/building-an-app.md),
[shared Mac presentation](docs/mac-presentation.md), and
[native shell APIs](docs/macos-app-shell.md). The [SDK contract](specs/sdk.md)
describes packaging and relocation.

## Verification

The release's automated checks include portable race/vet, focused native
routing/lifetime regressions, 20 repeated grid/tree cancellation checks, an
independent SDK consumer and a generated starter build. App ZIP extractions and
installed launch copies have strict signature and signed-byte verification.

Separate physical checks cover Notes pointer selection, text/history, native
file-panel focus restoration, per-window toolbar isolation, disabled-action
refusal, gallery field targeting and normal saved-notebook relaunch. These checks
do not establish human IME/VoiceOver usability or every workload/display. The
[acceptance report](docs/reports/macos-v1/README.md) distinguishes final results,
reused verification and the initial defects that were corrected. Performance
policy and the rendering backend remain unchanged.

## Repository layout

| Path | Purpose |
| --- | --- |
| `app/`, `ui/`, `canvas/`, `placement/` | Public application, UI and canvas APIs |
| `command/`, `history/`, `document/`, `editor/` | Commands, persistence and editing models |
| `tasks/`, `execution/`, `virtual/`, `workbench/` | Background work, scheduling, data views and docking |
| `replay/`, `automation/` | Experimental replay and local semantic automation |
| `drivers/cogent/` | Native driver module; executable tools under `cmd/` |
| [examples/](examples/README.md) | Independent application and contract modules |
| [docs/](docs/README.md), [specs/](specs/README.md) | Guides, contracts, research and historical evidence |
| `scripts/`, `release/`, `prototypes/` | Build/packaging tools, notices and isolated historical experiments |

The root module, `github.com/dshills/vellum`, uses only the standard library.
Native dependencies belong to `github.com/dshills/vellum/drivers/cogent`.
Historical reports have a separate module boundary and stay out of framework
module downloads. See the [layout guide](docs/repository-layout.md).

## Development and checks

From the repository root:

```sh
go test -race ./...
go vet ./...
```

Each driver/example module is checked separately. For example, after preparation:

```sh
(cd examples/placement && go test -race -tags offscreen ./...)
(cd examples/macapps && go test -race ./...)
```

Offscreen tests do not open desktop windows. Live input, accessibility, GPU and
display checks have separate verification requirements. The full prepared-repo
workflow is `scripts/check.sh`; see [CONTRIBUTING](CONTRIBUTING.md) and the
[development guide](docs/development.md) before running native campaigns.

## License and distribution

Vellum's own code is [MIT licensed](LICENSE). Third-party components retain their
original licenses and notices. The approved alpha.3 distribution decision covers
that exact payload and notice bundle. The owner has also approved the exact
[current v1 payload](release/macos-v1-distribution-audit.json). The v1.0.0 source
tag is public; final public-mode SDK/app candidates remain local and unpublished.
Their [publication packet](docs/reports/macos-v1-public-packaging/PUBLICATION_PACKET.md)
records matched notices, final checksums, a held fixture and the separate upload decision. See the [release contract](specs/release.md) and
[release metadata](release/version.json).
