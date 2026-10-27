# Native macOS application examples

Four independently runnable public-API consumers of Vellum:

| Example | Run from this module | Working core |
| --- | --- | --- |
| [Calculator](calculator/) | `go run ./cmd/calculator` | Arithmetic, keyboard, bounded history |
| [Notes](notes/) | `go run ./cmd/notes` | Folders, search, plain-text editing, local save/open/recovery, Undo/Redo |
| [Weather](weather/) | `go run ./cmd/weather` | Sample cities, search, units, forecasts |
| [Reminders](reminders/) | `go run ./cmd/reminders` | Lists, details, completion, Undo/Redo |

Prepare the native driver using the repository development guide first. From the
repository root: `. ./drivers/cogent/env.sh`, then `cd examples/macapps` and the
command above. The existing `scripts/check-example.sh` checks and builds all four.
`go test -race ./calculator ./notes ./weather ./reminders` tests models without
opening desktop windows; full `go test -race ./...` also compiles executable
wiring. Each app owns its model on the UI owner. Callbacks use the same public
commands, editors and history paths available to other applications.

These are independent demonstrations seeded with fictional data. They do not
connect to Apple services or access personal Notes/Reminders/weather locations.
Notes now persists an explicit local notebook with recovery; the other examples
remain in memory. Closing those other demonstrations discards their state. They follow Tahoe layouts but do
not yet reproduce Apple's glass, native toolbar items or rich Notes formatting.
The measured 52Dp header reserve is specific to the tested configuration; no
adaptive safe-area support is implied. See each package README and the
[acceptance specification](../../specs/macos-app-examples.md) for exact limits.

Apple reference screenshots are not part of the executable or distributed
assets. Bundled weather symbols retain their pinned Apache-2.0 notices in the
driver assets directory. Native pointer, visual and accessibility evidence is
separate from automated model/routing results.

## Multiwindow shell fixture

`go run ./cmd/macshell` opens two independent titlebar-toolbar windows. The fixture
uses a fixed light content theme, paints its root with that theme’s window color,
and explicitly selects Light native appearance
for each window, so text remains legible when the desktop uses Dark appearance.
It does not change system appearance. Each toolbar edits only its own document;
closing one window leaves the other alive. This fixture is not a new feature or
an adaptive appearance demonstration.
