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

The final image files and source/hash metadata are recorded beside this guide
when capture is complete. Application source and limitations are in
[the showcase guide](../../examples/showcase/README.md). The private app ZIPs
are separate from these Git-tracked images; no public release assets are created.
