# Showcase screenshot reproduction

The README showcase images are unmodified captures of the running macOS app
windows. They are not HTML mockups, generated application screenshots, or
composited window frames. Artwork inside the apps is original generated sample
content; Pulse data is explicitly fictional. Captures do not qualify performance,
human accessibility, IME behavior or other hardware.

Prepare the checkout, source `drivers/cogent/env.sh`, then run:

```sh
cd examples/showcase
python3 package.py --output /private/tmp/vellum-showcase
```

Open each resulting app. Use a window around 1400×880. In Forma, choose Coral
accent and Fit artboard. In Luma, show All artwork, Canyon dawn and Original.
In Pulse, keep All services, the 24h chart and api-gateway selected at Sample001.
Use macOS Capture Selected Window (Shift-Cmd-4, then Space) to capture that app
only. Screen-sharing indicators or window controls belong to macOS, not Vellum.
Do not add hypothetical controls or replace the displayed UI after capture.

The final images [Forma](forma.jpg), [Luma](luma.jpg) and [Pulse](pulse.jpg) were
captured on macOS from private packages built at source commit
`b646a9c72ace856225e88abb373a297100c2303e`. The automated capture used the
computer-use service's per-app `getScreenshot` API and saved its JPEG bytes
unchanged. The manual procedure above is an equivalent reproduction route.
[Capture metadata](captures.json) records dimensions, state and SHA256 digests;
[package verification](../reviews/showcase-apps/final-package-verification.json)
records clean-source provenance, ZIP hashes and extracted signature checks.
macOS screen-sharing and pointer indicators are retained in the raw captures. Application source and limitations are in
[the showcase guide](../../examples/showcase/README.md). The private app ZIPs
are separate from these Git-tracked images; no public release assets are created.

## Forge workbench

[Forge](forge.jpg) is an unmodified JPEG from the running private build3 app, source
`b41f08fc5bcc97eb5df8a43005fbc23f1675da3f`. The final exact extracted bundle was rebound by its full
path in the computer-use service before capture, avoiding a superseded process
binding. See [capture metadata](forge-capture.json) and the
[private delivery record](../reviews/fusion-workbench/delivery/README.md).
Reproduce from a clean checkout with `examples/forge/package.py`, launch the app,
keep the default camera/seven-body study and capture its window after Add/Undo.
macOS recording/pointer indicators remain. No geometry, icons or controls were
added to the screenshot after capture. The private ZIP is not a public download.

### Forge completion capture

The root README now uses [forge-completion.jpg](forge-completion.jpg), an unmodified
native JPEG from the exact ZIP-extracted private build11. Its [metadata](forge-completion-capture.json)
records the source commit and hash. The round shaft, active Orbit treatment and
material-preserving selection outline are running app output. Reproduce with the
seven-body default study, Fit assembly, Orbit and selected Sliding carriage; keep
history empty. Native focus/pointer/recording indicators remain. Earlier Forge
captures and their historical metadata are retained unchanged.

### Forge detail gallery

The root README also shows [shaft inspection](forge-shaft.jpg) and the
[context action panel](forge-actions.jpg). Both are unmodified native captures
from the same exact ZIP-extracted private build11 and source as the overview.
[Gallery metadata](forge-gallery-captures.json) records each image hash and state.
Select Drive shaft for the inspection view; right-click inside the viewport for
the panel view. The shaft's display is cylindrical while its picking bounds
remain a box. Native focus, pointer and recording indicators are retained.
No new desktop capture, image editing or public app release was required.
