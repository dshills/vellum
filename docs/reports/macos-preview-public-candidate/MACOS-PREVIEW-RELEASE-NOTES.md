# Vellum v0.1.0-alpha.3 — macOS developer preview

Candidate for final owner publication approval; no tag or downloadable assets published.

A reusable Go UI SDK and a Placement example app for Apple Silicon macOS, using the existing Cogent/WebGPU/Metal backend. Palette placement, selection, pan/zoom, snapping, multi-object editing, groups/layers, Undo/Redo and drawing Save/Open are available. The SDK includes prepared pinned dependencies and the selected native archive; a new external Go module builds and launches through public APIs.

The app is ad-hoc signed and not notarized. This is an early developer preview on the recorded Apple Silicon/macOS 26.6.2 host. APIs and same-window placement remain experimental. Human VoiceOver/IME and physical input, cross-display/high-refresh, other machines/architectures and full native/GPU heap/production qualification remain open. Historical latency disclosures remain workload-scoped; this release does not claim that full responsiveness targets are met.

Verification: external consumer race/vet/build and native startup; packaged app editing/persistence/recovery and final delivered-byte checks; 100 native lifecycle cycles; ten-minute composed activity; four-phase two-hour native soak with 24 eligible quiet samples and normal tracked cleanup. See the repository release ledger for evidence and limits. Candidate CI and final owner publication approval are still required.

Assets prepared locally: `vellum-v0.1.0-alpha.3-darwin-arm64-sdk.tar.gz`, `vellum-v0.1.0-alpha.3-macos-arm64.zip`, and `SHA256SUMS`. The tag target will be fixed to the approved CI-qualified source commit before publication.
