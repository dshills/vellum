# Vellum v0.1.0-alpha.3 — macOS developer preview

A reusable Go UI SDK and the Placement drawing example for Apple Silicon macOS,
using the default Cogent/WebGPU/Metal backend. The app includes palette placement,
selection, pan/zoom, snapping, multi-object editing, groups/layers, Undo/Redo and
Save/Open. Features are frozen; APIs remain experimental.

Download the app ZIP to try Placement, or the SDK archive to build a Go app.
SHA256SUMS covers both archives. Verify with shasum -a 256 -c SHA256SUMS, then
extract the chosen archive. Open Vellum Placement Preview.app; its adjacent README
describes the gestures and limitations. Use temporary drawings while evaluating.

**The app is ad-hoc signed and not notarized.** macOS may require explicit user
approval for a downloaded app; use the normal macOS approval flow if you trust
this download. Do not disable system security protections. Developer ID signing,
notarization and quarantine/first launch on an independent Mac are unqualified.

The SDK includes pinned prepared dependency sources and the ARM64 native archive.
Use Go 1.25.6 and Xcode command-line tools. Inside the extracted SDK, run
python3 scripts/new-app.py ../my-vellum-app --module example.com/my-vellum-app,
source drivers/cogent/env.sh, and follow the generated README. The verified
external consumer needs no private checkout or Rust build.

Verification: extracted external consumer race/vet/build and native startup;
packaged app editing/persistence/recovery; 100 lifecycle cycles; ten-minute
composed activity; and a four-phase two-hour native soak with 24 eligible quiet
samples, tracked cleanup and normal exit. Candidate macOS framework checks and
benchmarks, Linux smoke and Windows build jobs pass in GitHub CI. Optional
self-hosted qualification was skipped. Each check retains its recorded source
and automation scope. Linux/Windows CI does not extend preview platform support.

Tested on the recorded Apple Silicon/macOS 26.6.2, 60 Hz display configuration.
Same-window placement only. Human physical-pointer, VoiceOver/IME, other machines,
Intel, cross-display/high-refresh, full native/GPU heap and production qualification
remain open. One stable hidden Apple TUINS window was observed before process
exit; its retaining reference and full native/GPU deallocation remain unqualified.
Bounded resource checks do not establish a leak-free guarantee.

Historical results from the older bd3cbc9 SDK remain disclosed: serial grid/tree
p95 about 48 ms; paced grid p95 62.224 ms, with 599/599 observed intervals within
18.333 ms in one run. These are not measurements of this packaged alpha.3 app,
physical input-to-photon latency or general sustained cadence. Full responsiveness
targets remain unmet/unqualified; the owner approved this narrower preview policy.

The owner approved distribution for the exact recorded native archive, app source
binary and 418-file notice bundle. Original applicable notices accompany the
artifacts. That decision does not cover future changed payloads. Four repository
security alerts were checked against this SDK: three concern excluded historical
Go prototype versions; one concerns a copied original Rust dev-dependency absent
from selected compiled native components. No alerts were dismissed and no general
security-audit claim is made.

See [PR #1](https://github.com/dshills/vellum/pull/1) and its release reports for
source CI, detailed evidence and limits. Archive manifests retain actual build
sources: SDK ed359484a5be720948e92430f7496b36ec5cd352; app source binary
0d7539594ad742844f629ea1ab6080edb0a86ddb8ffbffe8192d3b53d636a16a, with unchanged
product inputs. Subsequent release-checkpoint commits change documentation only.

| Asset | SHA-256 |
| --- | --- |
| vellum-v0.1.0-alpha.3-darwin-arm64-sdk.tar.gz | 28a95b5c2354673d97161df30e828008a5aa18465ce18545a42e8caf0a82f269 |
| vellum-v0.1.0-alpha.3-macos-arm64.zip | bd58d2eb3d94486e07683a289a77299dde03fd450bd59bd65240f83d6d2f0e50 |
