# Matching private SDK and application bundle

From a clean committed Apple Silicon checkout, put Go 1.25.6 on PATH, prepare
native dependencies and source drivers/cogent/env.sh. Build in a new nonsynced
location such as /private/tmp; copy only completed archives into Documents. Run:

```sh
python3 -B scripts/private-macos-bundle.py --output /absolute/new/private-bundle
```

The exclusive output contains an SDK, Forge, Measure Pad, the A/B Documents
consumer, the generated Document Starter and Notes. Every app compiles against the
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
the named archives, README, checksum list, BUNDLE_MANIFEST, PREFLIGHT and (when
installation was requested) DELIVERY for delivery.

## Build, verify and install without desktop checks

For a new private candidate, one command also verifies installation. Use a new
output directory outside the checkout, a deliberate numeric build version and a
nonsynced installation parent:

```sh
python3 -B scripts/private-macos-bundle.py \
  --output /absolute/new/private-bundle \
  --build-version 23 \
  --install-to "$HOME/Applications/Vellum"
```

Preflight rejects known Documents/Desktop/iCloud/CloudStorage output locations
and aliases, existing outputs, outputs inside the checkout, a Go version other
than go.mod, non-Darwin/ARM64/cgo configuration, an active prepare lock, or missing
prepared-module sentinel files. It reserves no output on admission failure.
PREFLIGHT.json records Go environment and SHA256 identities of the resolved Go,
compile and link tools plus sampled native files. Those samples are rechecked
after packaging and the receipt hash is bound into BUNDLE_MANIFEST and DELIVERY.
This is sampled evidence, not a complete or authenticated toolchain closure.
Full pinned preparation, module integrity and SDK relocation checks still apply.

Use Go 1.25.6 on PATH (the child environment uses GOTOOLCHAIN=local), prepare the
pinned native sources, and source drivers/cogent/env.sh first. The command builds
all five apps from the extracted SDK, verifies that SDK's clean source identity,
all six archive checksums and the checksum list, then calls the existing strict
versioned installer. It never launches or focuses an app. Existing installations
and notebook contents are preserved. Build versions accept one to six decimal
digits; omitting the option retains the older producer default of 22.

BUNDLE_MANIFEST.json means packaging completed. With --install-to, DELIVERY.json
is the final success record: it binds the SDK, source, bundle manifest and exact
INSTALLATION.json digest/path, with native_runtime recorded as not executed.
If installation fails, the completed bundle remains available for diagnosis,
but no delivery receipt is written. Use the standalone installer to retry an
unchanged completed bundle; it verifies an existing installation before reuse.
Do not reuse an output directory or overwrite prior delivery receipts.

These are reproducible steps and source checks, not a promise of byte-identical
compiler output across hosts. Physical app acceptance remains a separate task.

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
Use the [current document-workflow handoff](mac-document-workflows-handoff.md)
for exact source/package identities and the remaining deferred native checks.
Keep historical source/hash records unchanged. The older
[consumer closeout](reviews/consumer-closeout/README.md) records its own earlier
lifecycle/CI evidence; it does not qualify later workflow changes. New checkpoints
use internal code review; historical Prism receipts retain their original scope.

Install the completed bundle with the [versioned private installer](private-macos-installation.md).
The existing Notes app is also included for ordinary use; native tests use an
explicit disposable -data-dir and never migrate or edit an existing notebook.
