# Build an application with Vellum

This tutorial uses the private alpha SDK. Vellum's module namespace is
`github.com/dshills/vellum` and its own code uses the MIT license. Native
distribution audit and full target qualification remain gates.
The alpha version is available as `vellum.Version`; see the
[release contract](../../specs/release.md).

## Start from a checkout or relocated SDK

In a checkout, run `drivers/cogent/prepare.sh` once. A private SDK archive already
contains the pinned patched Cogent/typesetting source trees; it does not require
the research prototypes or the original checkout. You still need the declared Go
version, cgo C toolchain and external Go modules, obtained from your module cache
or the network. Native platform requirements are in
[platform support](../../specs/platform-support.md).

Create an application in a new directory. The generated starter retains its MIT
LICENSE notice alongside the copied template source:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-app-cache
. /path/to/vellum/drivers/cogent/env.sh
python3 /path/to/vellum/scripts/new-app.py '/path/to/My Application' \
  --module example.com/my-application
cd '/path/to/My Application'
go test -race ./...
go vet ./...
go run .
```

The generated module supplies all four local replacements explicitly: Vellum,
its Cogent driver, the pinned Cogent source and pinned typesetting source.
Dependency replacements do not propagate from another module's `go.mod`.
Keep the SDK directory in place; if you move it later, regenerate replacements
with `go mod edit`. The generated example contains a counter and GPU canvas.

## Use ordinary Go state and controlled views

A minimal application can use these imports and entry point:

```go
package main

import (
    "context"
    "log"

    "github.com/dshills/vellum/app"
    "github.com/dshills/vellum/drivers/cogent"
    "github.com/dshills/vellum/ui"
)

func main() {
    application, err := app.New(app.Config{})
    if err != nil { log.Fatal(err) }
    defer application.Close()
    checked := false
    err = application.Open(app.Window{
        ID: "main", Title: "My application", Width: 640, Height: 480,
        View: func() ui.Node {
            return ui.Node{
                ID: "root", Kind: ui.Column, Grow: true,
                Style: ui.Style{Padding: ui.Some(ui.UniformInsets(16))},
                Children: []ui.Node{{
                    ID: "preview", Kind: ui.Checkbox, Text: "Preview",
                    Checked: checked,
                    Accessibility: ui.Accessibility{Name: "Enable preview"},
                    Change: func(event ui.ChangeEvent) {
                        checked = event.Checked
                        application.Invalidate("main")
                    },
                }},
            }
        },
    })
    if err != nil { log.Fatal(err) }
    if err := application.Run(context.Background(), cogent.Driver{}); err != nil {
        log.Fatal(err)
    }
}
```

`View` and UI callbacks run under the application owner. Stable node IDs preserve
native widgets across redraws. Change callbacks request edits to your model;
the next view supplies the accepted value. A missing style property restores its
default, while `ui.Some(float32(0))` explicitly clears a gap or grow factor.
Driver themes are application-local snapshots.

## Commands, background work and multiple windows

Register actions in `command.Registry` and give it to `cogent.Driver.Commands`.
Menus and shortcuts invoke the same registry, rechecking availability under the
owner. Use stable command IDs rather than displayed labels.

Use `tasks.New` and `tasks.Start` for bounded workers. Start tasks under the UI
owner and use the relevant `application.WindowContext` as their parent. Workers
honor context cancellation and return owned results; commit callbacks run on the
owner. Direct `PostWindow` also delivers a worker result to a living window.
Call platform services from workers, outside UI callbacks, `View` and `Do`.

`application.Open` registers another window during a run. IDs remain unique for
the whole application lifetime. Closing a window cancels its scope; closing the
final window ends the application. Keep documents/undo history in application
model objects, and use `workbench.Session` for the separate window/panel layout.
The same document can move between docked/floating windows without becoming
layout-owned data.

## Test and package deliberately

The independent `examples/sdk-contract` module tests controlled values,
explicit-zero styles, commands, bounded worker commits and docking layout
snapshots through public APIs. Run it with `GOWORK=off` to check consumer usage.

Build and validate your own private SDK archive. The producer digest below is
local self-attestation: it checks integrity against the bytes you built, and does
not establish independent authenticity. Use this flow only for your own source.
For an archive from someone else, obtain its expected digest through a separate
trusted channel before executing its code. The default packages tracked files.
For your own precommit development snapshot, add `--include-untracked` explicitly
only after checking the whitelisted untracked files; their shipped names are recorded
in the manifest. A clean committed build needs no opt-in:

```sh
python3 scripts/release-sdk.py --output /tmp/vellum-private-alpha.tar.gz > /tmp/vellum-private-alpha.json
VELLUM_SDK_SHA256=$(python3 -c 'import json; print(json.load(open("/tmp/vellum-private-alpha.json"))["sha256"])')
python3 scripts/check-sdk.py --archive /tmp/vellum-private-alpha.tar.gz --sha256 "$VELLUM_SDK_SHA256"
```

The checker verifies every recorded file, extracts into a new path containing
spaces, generates an independent consumer, and checks that every local module
replacement points to its exact directory in that extracted SDK. It tests/vets/builds the consumer
and public API scenarios. Use the SHA-256 from the
successful producer JSON; the commands above capture that exact value. Execution requires that digest or explicit
`--allow-self-attested` for your own freshly built source. The checker runs archive
code with your filesystem privileges, preserved proxy/TLS transport settings (which
may include credentials), and desktop settings only during `--runtime`; its environment
allowlist is not a process
or filesystem sandbox. `--sha256` accepts a trusted producer digest; an
unsigned self-contained manifest proves internal integrity, not authenticity.
On filesystems without hardlinks, the output copy is complete only after a
successful producer return; interrupted output must not be used or assumed valid.
`--runtime` also opens, verifies and closes the relocated GPU example; serialize
it with other live native checks.

The alpha toolkit still has platform/accessibility/performance gates. Private SDK
checks are not production or redistribution approval. Finish dependency/native/font
audits before a public release. Signing,
notarization and store publication require separate owner configuration.

For a private macOS app bundle from a prebuilt executable:

```sh
go build -o /tmp/my-application .
python3 /path/to/vellum/scripts/package-app.py --binary /tmp/my-application \
  --output '/tmp/My Application.app' --name 'My Application' \
  --identifier org.example.myapp --version 0.1.0 --build-version 1 --platform darwin
```

This inspects native dependency closure and writes an unsigned private bundle.
The numeric app release and build versions enter `Info.plist`; the SDK semantic
release remains in `APP_MANIFEST.json`. Increment the build version when
distributing another build.
Pass `--sign-identity` only after choosing a signing configuration. Windows/Linux
archive layouts also exist; their dynamic runtime closure and actual target
execution require separate acceptance.
