# Historical foundation distribution checks

This page documents the earlier foundation fixture and its optional credentialed
route; it is not the current private SDK/app delivery procedure. Use
[private consumer packaging](private-consumer-bundle.md),
[installation verification](private-macos-installation.md) and
[current distribution boundaries](status.md). No new signing/notarization action
is authorized by the historical example below.

`package_macos.py` packages a previously built native integration fixture and AccessKit dylib, enforces application-relative dynamic-library resolution, signs each native artifact and seals/verifies the complete bundle. It rejects unknown runtime dependencies and existing output paths. Default identity `-` produces a development ad-hoc bundle. It does not establish Gatekeeper trust or a clean-machine result.

```sh
python3 scripts/package_macos.py \
 --binary prototypes/foundationintegration/build/integration-fixture \
 --library prototypes/foundationintegration/build/native/libaccesskit.dylib \
 --output /private/tmp/VellumFoundation.app \
 --manifest /private/tmp/vellum-package.json
```

For a release, supply `--identity 'Developer ID Application: …'` and `--notarize PROFILE` using an already provisioned notarytool keychain profile. The workflow uses hardened-runtime/timestamp signatures, waits for Accepted, staples/validates the ticket, performs Gatekeeper assessment, and verifies the final bundle again. Submission happens only when explicitly requested via this flag. Credentials are managed by Apple tools, never embedded in the repository. Failure preserves no partial final bundle. A credentialed release has not been executed here.

A runtime matrix must exercise Windows (D3D12 + NVDA), macOS (Metal + VoiceOver), Linux Wayland and X11 (Vulkan + Orca), and separately launch each packaged artifact on a clean user machine without development dependencies. Portable Go cross-compilation is compile evidence only. The current macOS bridge does not implement Windows/Linux native IME/accessibility adapters. Do not mark those gates passed from a hosted compile job. Foundation expansion remains blocked until the missing runtime, text, memory and assistive-technology contracts have passed.
