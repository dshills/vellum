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
