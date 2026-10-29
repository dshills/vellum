# Matching private SDK and application bundle

From a clean committed Apple Silicon checkout, prepare native dependencies and
source drivers/cogent/env.sh. Run:

```sh
python3 -B scripts/private-macos-bundle.py --output /absolute/new/private-bundle
```

The exclusive output contains an SDK, Forge, Measure Pad, the A/B Documents
consumer and the generated Document Starter. Every app compiles against the
extracted SDK, with replacement origins checked to remain inside it. Both starter
profiles receive race/vet/build checks. ZIP-extracted signatures, named checksums
and the source-bound BUNDLE_MANIFEST.json make the result reviewable. The script
checks source cleanliness before and after; partial failures retain diagnostics
and must not be presented as a finished bundle.

The A/B Documents diagnostic requires an explicit private data directory. Launch
with `open -a '/path/Vellum Documents.app' --args -data-dir /absolute/private/test-workspace`
or use the bundle executable with the same flags. Documents separates A/B data
and routes process-wide menus through the native active-window callback.
Document Starter starts normally from Finder with app-specific private Application
Support storage. Supply `--args -data-dir /absolute/private/document-data` for
isolated tests or an existing notebook. The generated starter never defaults to
Vellum Notes. See the [starter guide](building-macos-apps.md#complete-document-starter)
for identity persistence, override precedence and refused storage configurations.

These are private ad-hoc, unnotarized builds. Existing native distribution audit,
backend, rendering/pacing and qualification policy are unchanged. A matching
source identity and successful checks are not permission for a public release,
notarization or a new signing identity. Human IME/VoiceOver and other hosts remain
unqualified. Native UI observations are separate from automated SDK/routing tests.

Only the SDK/app archives named in SHA256SUMS are payloads. The build/ subtree
retains intermediate binaries, extracted sources and signature checks for local
diagnosis; it is not covered by the top-level archive checksum list. Copy only
the named archives, README, checksum list and BUNDLE_MANIFEST for delivery.

## Reproducible verification and handoff

Before packaging, the offscreen CI command can independently export and verify a
clean checkout (use a new output directory outside the checkout):

```sh
python3 -B scripts/check-sdk-ci.py --output /absolute/new/sdk-check
```

It verifies both generated starter profiles from a relocated SDK path containing
spaces, including race tests, static checks and compilation. The Framework
workflow runs it in a separate macOS job; it never opens native windows. The
result's replacement maps establish module resolution within the SDK. They are
not an OS filesystem sandbox or a public distribution certification.

Each private delivery should carry its exact source commit, SDK digest, named
archive checksums, signing route and separate automated/native observations.
Current working instructions supersede old handoffs that describe commits as
local or lifecycle checks as unfinished; keep historical source/hash records
unchanged. The [consumer closeout](reviews/consumer-closeout/README.md) describes
the completed lifecycle checks, exact-main CI and review policy. New checkpoints
use internal code review; historical Prism receipts retain their original scope.
