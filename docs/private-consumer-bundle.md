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

Documents and Document Starter require explicit private data directories. Launch
with `open -a '/path/Vellum Documents.app' --args -data-dir /absolute/private/test-workspace`
or use the bundle executable with the same flags. The generated starter refuses
to default to another application's notebook. Documents separates A/B data and
routes process-wide menus through the native active-window callback.

These are private ad-hoc, unnotarized builds. Existing native distribution audit,
backend, rendering/pacing and qualification policy are unchanged. A matching
source identity and successful checks are not permission for a public release,
notarization or a new signing identity. Human IME/VoiceOver and other hosts remain
unqualified. Native UI observations are separate from automated SDK/routing tests.

Only the SDK/app archives named in SHA256SUMS are payloads. The build/ subtree
retains intermediate binaries, extracted sources and signature checks for local
diagnosis; it is not covered by the top-level archive checksum list. Copy only
the named archives, README, checksum list and BUNDLE_MANIFEST for delivery.
