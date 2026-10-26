# Personal macOS starter

A minimal consumer of the frozen public API: one native window, full-color AppKit
toolbar, controlled text/range values and notification-driven system appearance.
It adds no framework controls and has no dependency on gallery internals.

From a prepared checkout or extracted personal SDK:

```sh
. ./drivers/cogent/env.sh
(cd examples/personal-mac && go run .)
```

The root range/text state is owned by the application. Change handlers accept
proposals; the toolbar callback uses the same owner and invalidates the window.
The toolbar catalogue stays immutable for this opening. `-exit-after=3s` is an
optional native launch/normal-close smoke, not an interaction or accessibility
qualification. Artwork is original two-color PNG generated from standard Go code.

For a completely separate application module, use the SDK's `scripts/new-app.py`
then add `FollowSystemAppearance: true` to its driver. See
[personal delivery instructions](../../docs/personal-macos-v01.md).

## Application icon

`assets/app-icon.png` is original full-color 1024×1024 RGBA artwork; regenerate
it with `python3 assets/create-icon.py`. Supply it when building an application
bundle (the icon is bundle metadata, so `go run .` does not install it):

```sh
# From this example directory after sourcing the driver environment:
go build -o build/personal-mac .
python3 ../../scripts/package-app.py --binary build/personal-mac \
  --output 'build/Vellum Personal.app' --name 'Vellum Personal' \
  --identifier com.example.vellum.personal --version 1.0.0 --build-version 1 \
  --icon assets/app-icon.png --sign-identity -
```

Open the resulting `.app` to use its Finder/Dock icon. See the
[app-icon packaging contract](../../docs/building-macos-apps.md#application-icons)
for formats, Retina sizes and limitations. This private starter command does not
supply distribution notices or approve public distribution.
