# macOS app-icon checkpoint

Requested: optional application icons in the existing Mac packaging flow. Source
base is 0c97300; verification.json pins the changed packaging source hashes used
for this precommit example. Original accepted artifacts remain unchanged.

Supports explicit ICNS or 1024×1024 PNG, standard/Retina sizes, original colors
and transparency. Artwork is installed under a fixed resource name before
signing and verified with its plist reference and manifest hashes before/after
signing and final publication. Existing public audit/notices and overwrite gates
remain in force. No stable Go API, renderer, global OS setting or user-data change.
The starter includes original reproducible artwork and a complete command.

Focused tests cover malformed/oversized/symlink/nonregular input, unsupported
targets, no-icon compatibility, native PNG/ICNS round trips and dimensions,
source replacement during native conversion, signing order and icon/plist
mutation at publication. CI invokes the icon suite; non-Mac hosts skip native
tool checks explicitly. Existing package and script-portability suites pass.

Built the existing Vellum personal starter, packaged with the sample icon and
ad-hoc signature, archived, extracted and verified the exact ZIP. All ten image
representations decode; AppKit confirms full color and transparency. Producers
are archived here. Build receipts are in the companion review directory.

The first Documents-stage signing attempt, and a Documents extraction attempt,
failed codesign due to Finder/resource-fork metadata injected by that storage.
Using local temporary staging/extraction passed. No user metadata was stripped.
This is documented in the packaging guide; it is not hidden as a passed attempt.

Private example: deliveries/mac-app-icons in the research workspace, with app
path and ZIP hash in verification.json. No desktop launch or Finder/Dock cache
pixel check was performed. Classic ICNS only; runtime replacement, Dock badges,
per-window icons and Icon Composer layered/appearance variants are outside scope.
