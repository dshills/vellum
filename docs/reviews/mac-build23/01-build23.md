# Build23 private SDK and apps — internal review

Built from green main df07cd4ce70b221c443a2f90f8738f3b8a827a3c with Go 1.25.6
on Apple Silicon. SDK SHA256:
a07546224f00b389cf161cf4e85617c62ae4a00040f06c2315fe53b3608abd04.
The existing producer built Forge, Measure Pad, A/B Documents, generated Document
Starter and Notes against the same verified extracted SDK, then installed at
~/Applications/Vellum/build23-df07cd4ce70b. No app window was launched. Forge's
headless --write-icon export returns before UI/data initialization.

The initial Documents-based output failed strict codesign verification because
Finder metadata was attached to the bundle. Packaging correctly refused success;
that attempt has no completion/delivery receipt. Retried the same source in a
new nonsynced temporary directory, without stripping attributes or changing
security/signing policy. It succeeded. Phase four will add an early guard for
this demonstrated producer-location failure.

Internal review checked clean committed source before/after, SDK/source/hash
identity, generated consumer origins, exclusive output, five distinct bundle IDs,
build number 23, archived and installed signatures, source-bound manifest and
installation receipt, prior-version preservation and absence of app launch.
Private ad-hoc, unnotarized policy, audit records, backend, features and release
performance gates are unchanged. Existing AppKit deprecation/linker warnings are
not failures and do not authorize unrelated framework changes.

Validation: SDK relocation verification, both generated starter race/vet/build
checks and public API consumer checks passed with native runtime not executed.
All six archive digests matched the manifest after durable copy; all five
installed bundles independently passed codesign --verify --deep --strict and
Info.plist build/identifier checks. Prior build22 directories are retained.
The private mac-build23-20261030 delivery contains BUNDLE_MANIFEST, DELIVERY,
INSTALLATION, SHA256SUMS and VERIFICATION receipts plus the six archives.
No personal notebook, clipboard, global setting or desktop input was used.
