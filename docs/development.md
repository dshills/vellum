# Development

## Modules and dependencies

The root `github.com/dshills/vellum` module contains portable public packages and uses
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
independent hello, drawing, 3D, imaging, editor, grid and placement modules, validates dependency
preparation inputs and race-tests upstream lifecycle regressions. Exact text and
shaped range fixtures run with the prepared typesetting replacement, installed
font cache and offscreen display initialization through
`scripts/check-text-backend.py`; native consumers separately exercise AppKit. The macOS CI workflow follows the same
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

## Editor native acceptance

After building `examples/editor`, run its `-smoke` mode for deterministic Cocoa
input, composition, native range geometry and document persistence checks. Run
native desktop fixtures sequentially. The portable and offscreen tests do not
qualify these AppKit behaviors.

A separate read-only accessibility inspector can qualify exact text, UTF-16
ranges, screen bounds, point lookup and RTL geometry using an already trusted
caller:

```sh
python3 scripts/check-editor-accessibility.py \
  --binary /tmp/vellum-editor \
  --output /tmp/vellum-editor-external-ax
```

The output directory must be new. The wrapper compiles its optional Swift
inspector, checks existing accessibility trust without prompting, and runs the
editor's dedicated 60-second `-ax-fixture`. It records binary/source hashes and
both process results. It refuses unsupported platforms or missing trust before
launching the application. It never changes input sources, VoiceOver or privacy
preferences. Swift is needed only for this acceptance tool. Human VoiceOver and
IME candidate-panel workflows remain separate checks. See the
[editor acceptance record](reports/macos-editor/README.md).

## Native menu and tray acceptance

Mac application-menu/tray acceptance has a separate, opt-in diagnostic build:

```sh
cd drivers/cogent
go build -tags=vellum_native_checks -o /tmp/vellum-native-services ./cmd/services-smoke
/tmp/vellum-native-services -cycles=3
```

Run it only on an available Mac desktop. It creates two temporary windows and
owned native tray/menu objects, invokes installed AppKit actions, and observes
cleanup before GLFW terminates. Its exported diagnostics exist only with that
build tag. It does not access the clipboard, request notification permission,
change preferences or synthesize physical input. Selector acceptance remains
separate from physical clicks and human screen-reader workflows.

## Native file-panel lifecycle acceptance

Mac file-panel lifetime has a separate opt-in check on an available desktop:

```sh
cd drivers/cogent
go build -tags=vellum_native_checks -o /tmp/vellum-file-panels ./cmd/file-panel-smoke
/tmp/vellum-file-panels -cycles=10 -shutdown-mode=open
/tmp/vellum-file-panels -cycles=10 -shutdown-mode=save
```

It observes real Open/Save panels before canceling their callers and checks
native acknowledgement retirement before each next panel. It then shuts down
with a visible panel and joins the application-owned operation. Partial progress
and terminal inventory are emitted on failure as well as success. It selects or
writes no files and changes no clipboard or OS preferences. See the
[lifecycle contract](../specs/file-panel-lifecycle.md); actual selections and
physical Finder drops have independent acceptance requirements.

## Native window cycle acceptance

The opt-in native window-cycle check exercises actual control/canvas mounting,
readback and closure with a surviving window:

```sh
cd drivers/cogent
go build -tags=vellum_native_checks -o /tmp/vellum-window-cycles ./cmd/window-cycle-smoke
/tmp/vellum-window-cycles -cycles=1000 -sample-every=25 -timeout=12m
```

Run it on an available Mac desktop. It records post-GC Go/RSS diagnostics and
stops on sampled safety limits; geometry persistence is disabled in this temporary
consumer. Explicit owner replies distinguish admission from completed checks.
It does not measure the full integrated/endurance or native/GPU memory gates.
See the [native lifecycle contract](../specs/native-window-lifecycle.md).

For the all-component lifecycle fixture, add `-composed` and allow up to 45 minutes:

```sh
/tmp/vellum-window-cycles -composed -cycles=1000 -sample-every=25 -timeout=45m
python3 native/patches/autorelease/check.py
```

The second command checks native temporary ownership without opening windows.
The composed mode mounts the full kind catalog, editor/history, table, outline,
docked GPU canvas, modal and 1000 dense labels. It checks native AX containment,
actual controlled edits and resource retirement; it does not measure frame latency
or certify physical display scale. The [AX retention report](reports/macos-ax-retention/README.md)
retains a failed campaign alongside the reviewed correction.


## Grid acceptance

`examples/grid` is a separate consumer module with independent 10,000-row and
million-row procedural windows. Its offscreen tests route primitive pointer,
wheel and keyboard events through the managed widgets and compare completed
software frames at different owner drain batches. `scripts/check.sh` includes
these tests; `scripts/check-example.sh` race-tests, vets and builds the module.

After preparing dependencies, source `drivers/cogent/env.sh` and run
`go run . -smoke` from `examples/grid` for sequential Cocoa measurement and
two-window lifecycle checks. This desktop smoke uses explicit owner commands;
offscreen routing, desktop rendering and physical input are separate evidence.
See the [grid example acceptance](reports/macos-grid-example/README.md).

`examples/loadedgrid` stores actual million-row ×20 numeric records and runs
generation, sort/filter indexing and immutable edit copies on bounded workers.
From the repository root, source `drivers/cogent/env.sh`, then run `go run .`
inside that module. `-rows 10000` selects a smaller input and `-variable-rows`
adds indexed heights. Its offscreen race tests and vet are included in
`scripts/check.sh`. Apply waits for the filter text field's native input barrier
before reading its controlled value; Save uses the inline-cell barrier.
Run `go run -race . -smoke` and then `go run -race . -smoke -variable-rows`
sequentially for the bounded Cocoa workflow. See the
[loaded-data report](reports/macos-grid-loaded/README.md) for qualification scope
and the remaining sustained performance and large-tree gates.

`examples/loadedtree` uses `ui.DataTree` and a worker-built `virtual.Outline` with
100,000 actual stored nodes. Load, expansion and ancestor filtering run on bounded
workers; owner navigation and admission use accepted indexes. Run its offscreen
tests for deterministic routing or `go run -race . -nodes 100000 -smoke` for the
bounded Cocoa owner workflow. This does not qualify tree frame latency or human
screen-reader workflows.

`examples/loadedgrid -perf` and `scripts/benchmark-grid.py` opt into one-window
native work tracing for a 400-cell stored table. See the
[measurement contract](../specs/grid-performance.md) for exact work boundaries,
raw records, paired 10k/million geometry and the unchanged 4 ms/6 ms warm targets.
Performance runs use optimized builds, separately from race-instrumented checks.

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

Collect Mac SDK distribution inputs without launching a window:

```sh
python3 scripts/distribution-inventory.py --output /tmp/vellum-distribution.json
```

Use a new output path outside the checkout. The report records SDK file hashes,
resolved embedded assets, listed cgo sources and native linker-search candidates
against the verified module notice graph. Review-required asset/native rights
and prebuilt archive component provenance remain separate from this inventory.
The collector does not approve or publish a distribution.

Set `VELLUM_CACHE_DIR` to reuse a chosen Go cache across preparation/check commands.
The legacy `P0_CACHE_DIR` variable is accepted for existing local environments.
Without either variable, sourcing the environment creates an isolated temporary cache.

## Local semantic automation

The workbench opens no external endpoint by default. Run it with `-automation`
when a trusted local controller needs domain inspection/actions. The single JSON
descriptor on stdout contains a secret capability. Pass it privately to clients
using `automation.Call`; inspect first, then invoke with that response's revision.
Do not automatically retry actions after a lost response. See the
[protocol and lifetime contract](../specs/automation.md). Headless socket tests
are included in root `go test -race ./...`; workbench domain tests run with the
native module's `offscreen` tag and create no windows.

## Clipboard data

Run preservation safety checks against a new private named pasteboard:

```sh
python3 scripts/check-clipboard.py --self-test --output /tmp/vellum-clipboard-safety
```

On an available Mac desktop, explicitly opt into the general-board workflow:

```sh
python3 scripts/check-clipboard.py --live --output /tmp/vellum-clipboard-normal
python3 scripts/check-clipboard.py --live --case exit-after-owned --output /tmp/vellum-clipboard-failure
python3 scripts/check-clipboard.py --live --case kill-after-owned --output /tmp/vellum-clipboard-termination
```

Each output directory must be new. Go pins must already be prepared; Swift/Clang
are needed for the separate guardian. The tool preserves bounded raw clipboard
items/flavors in a protected local backup and restores only known test ownership.
Original bytes/type names never enter ordinary output files or evidence. A retained
`private/job/original.plist` means uncertain ownership or restoration failed;
keep that protected directory for recovery. Never include it in a review/archive.
The compiled guardian's `--recover <private/job>` refuses a live recorded child,
foreign changes and ambiguous admitted state. These checks do not change input
sources, VoiceOver or notification preferences. See the [contract](../specs/clipboard.md)
for bounds, materialized-provider fidelity and non-atomic OS limitations.

Obtain `Application.Clipboard()` after the native driver publishes services and
inspect its capabilities. Read explicit formats such as `text/plain`, `image/png`
and `application/vnd.vellum+json`; write one slice of `app.ClipboardRepresentation`
from a worker. Keep input bytes immutable until the method returns. See the
[clipboard contract](../specs/clipboard.md) for limits and cancellation semantics.
`go test -race -run '^TestClipboard' .` in `drivers/cogent` checks the native
parser and injected queue without opening or reading a pasteboard. Live
interoperation tests remain separate and must wait for a free computer.

## Mac menu-bar item

From a worker after native services become available, obtain `Application.Tray()`
and inspect its capabilities. `SetTray` takes a complete title/icon/command-menu
snapshot; command IDs refer to the driver's registry. `ClearTray` removes the
application's item. Provide a descriptive tooltip/accessibility name for an
icon-only item. See [the contract](../specs/tray.md). The focused native-module
tests `TestTray` and `TestNativeService` exercise injected routing/admission and
offscreen Cocoa data/bitmap conversion without creating status items or panels.

## Mac notifications

Obtain `Application.Notifications()` from a worker after services are published.
Inspect capabilities and `NotificationSettings` before delivery. Only an explicit
`RequestNotificationPermission` may show a permission prompt; `Notify` does not.
Use a stable ID for replacement, and `WithdrawNotifications` for targeted removal.
Run through an app bundle with its own identifier; ordinary Go binaries return
Unavailable. See [the contract](../specs/notifications.md). Focused `TestNotification`
checks with `-tags offscreen` in `drivers/cogent` build data objects and injected
settings/results; they never access the center, send messages or request permission.

For separate nonprompting native acceptance, build the opt-in consumer and package
an isolated private app with the ordinary packaging tool:

```sh
cd drivers/cogent
go build -tags=vellum_native_checks -o /tmp/vellum-notifications ./cmd/notifications-smoke
cd ../..
python3 scripts/package-app.py --binary /tmp/vellum-notifications \
  --output /tmp/VellumNotificationChecks.app --name VellumNotificationChecks \
  --identifier com.vellum.native-checks.notifications.local --version 0.1.0 \
  --platform darwin --sign-identity -
/tmp/VellumNotificationChecks.app/Contents/MacOS/VellumNotificationChecks \
  -bundle-id=com.vellum.native-checks.notifications.local \
  -nonce="$(python3 -c 'import secrets; print(secrets.token_hex(32))')"
```

The output must be new. Run only on an available Mac desktop. The private bundle
uses ad hoc signing, independently of public release/signing/notarization. The
consumer never requests permission, changes a delegate, requests sound or changes
preferences. NotDetermined/Denied qualifies refusal and unchanged state. An
already Authorized/Provisional bundle additionally checks actual delivered
content, replacement and targeted withdrawal preserving a second item; that
branch can fail if OS presentation policy does not retain notifications. A pass
never proves human-visible presentation. `-previous-nonce` checks a completed
prior run's IDs in the same bundle. The bare executable accepts
`-expect-unbundled` with a fresh nonce and verifies unavailable operations.

The diagnostic inventory returns only requested test IDs, keeps pending/delivered
observations separate and enforces a 64 KiB aggregate reply limit. Cleanup waits
for native ACKs under a fresh deadline before withdrawal and observes empty owned
lists before closing the loop. This does not establish concurrent-writer
atomicity. See the [notification contract](../specs/notifications.md) for limits.

## Palette placement checks

`examples/placement` is an independent public API consumer: prepare dependencies,
source `drivers/cogent/env.sh`, then run `go run .` from that module. Its default
window demonstrates typed tools, target policy, optional snapping, selection,
pan/zoom and bounded command/history placement. The example and portable package
are included in the private SDK and normal demo checks. Run its `offscreen` race
tests for managed pointer/keyboard routing; no live window opens in that mode.

The driver tests include transformed/DPI coordinates, real ancestor scroll
geometry, preview pixels, consumed releases, exactly-once placement, queue
saturation and removal/window lifetime cleanup. The pinned stage-65 native
capture regressions run through the software window router. See the
[placement contract](../specs/placement.md) for units, bounds and same-window
limits. These checks do not certify physical Mac capture or cross-display input.


## Professional workspace replay

`examples/replay` includes `cmd/workspace`, a runnable public consumer composing
stored data, workbench panels, canvas placement and a shared inspector. The
software `cmd/workspace-replay` and frozen-record tests require `-tags offscreen`;
they explicitly step workers and clocks and do not open a desktop window.
See [commands and scope](../examples/replay/README.md#professional-workspace).
Native compilation and software replay remain distinct from physical/GPU checks.

## Canvas object semantics

The [placement demo](../examples/placement/README.md) also exercises opt-in named
Canvas2D objects. Follow the [contract](../specs/canvas-accessibility.md) and
[checkpoint](reports/macos-canvas-accessibility/README.md) for metadata budgets,
completed-frame coordinates, queued-action lifetime guards and the separation
between software/inert Cocoa evidence and live human accessibility acceptance.
