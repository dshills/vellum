# macOS application icons

Authorized extension: add application icon support to the existing bundle
packager. No stable Go API, backend, performance gate or release baseline changes.

- Optional `scripts/package-app.py --icon` accepts existing ICNS or a 1024×1024
  PNG, only for macOS. Omission retains prior behavior; other targets reject it.
- Snapshot one explicit regular nonsymlink file, at most 16 MiB. Native tools
  consume the snapshot, never reread a changing source path. Reject empty,
  malformed, non-square/wrong-sized PNG and undecodable ICNS before publication.
- PNG uses standard/Retina 16, 32, 128, 256, 512 point representations. Preserve
  original colors/alpha; ICNS input retains individually authored representations.
- Fixed Resources/AppIcon.icns name and CFBundleIconFile metadata avoid input
  path/name collisions. Include detached source and installed artwork hashes.
- Finish embedding before existing optional signing. Verify installed icon bytes
  and metadata at each existing bundle verification boundary. Failure preserves
  preexisting output; remove only the publisher's owned partial output.
- Public audit/notices, signing identity and notarization policy are unchanged.
  Do not mutate accepted historical SDK/apps or install global Finder overrides.
- Cover absent icons, malformed/oversized/nonregular inputs, unsupported targets,
  PNG/ICNS native round trips, all Retina dimensions, signing order, tampering and
  extraction/signature integrity in focused checks. Distinguish bundle/OS image
  resolution evidence from actual Finder/Dock pixel verification.

Classic ICNS only. No runtime icon API, badges, per-window icons, Icon Composer
layers or Light/Dark asset catalog variants. The package icon is visible through
macOS app-bundle resolution; unbundled go run is outside this contract.
