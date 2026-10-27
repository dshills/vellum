# Vellum macOS v1 — build13 local publication candidates

All seven archives were built from 64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a.
Each app is version 1.0.0/build13. No v1 native assets have been uploaded.

Verify in this directory before extraction: `shasum -a 256 -c SHA256SUMS`.
Extract an app ZIP with Archive Utility, then launch the extracted app. The
apps are ad-hoc signed and unnotarized; downloaded-quarantine launch is
unqualified. No security bypass is provided. private/ contains retained local
inputs, excluded from proposed downloads. Native checks used exact ZIP-derived
copies on Apple Silicon/macOS 26.6.2.

- Notes: local plain-text notebook, folders/search, native toolbar, saving,
  recovery and Undo/Redo. Validation used a disposable profile; normal launches
  use the existing personal notebook. Preserve backups of important notebooks.
- Mac Widgets: existing Mac component catalogue and full-color toolbar examples;
  appearance switches are app-scoped.
- Mac Shell: independent A/B document toolbars. Edit B, then close it; A keeps
  its own count. The exposure repair is in this SDK and every app.
- Forma: poster/layers, colors, pan/zoom/fit and Undo/Redo. Luma: generated
  landscapes, filmstrip/looks/histogram. Pulse: fictional metrics, chart/table
  and filters. These demos reset on launch and use existing experimental GPU APIs.

SDK: extract vellum-v1.0.0-build13-darwin-arm64-sdk.tar.gz; keep its internal
vellum-v1.0.0 directory. Use Go 1.25.6, Python 3 and Xcode Command Line Tools.
From the SDK run `python3 scripts/new-app.py /absolute/new/app --module example.com/my-app`,
then build/run from that generated app. Replacements must resolve into the SDK.
It is already prepared: do not run prepare.sh or overlay a different native
library. scripts/check-sdk.py accepts --archive and the trusted --sha256 for
independent verification.

Known limits: other OS/hardware/displays, cross-window canvas drag, human
IME/VoiceOver, downloaded-quarantine launch and exhaustive physical-pointer
coverage are unqualified. Existing backend/performance policy is unchanged.
An initial background Notes capture needed native Raise; no model edit was
required. One Luma close attempt needed a fresh native target; a separate
single-close control passed. No product cause was established for the first
action. The final report retains the full observation. This build's source 64bf9cf
is newer than the unchanged v1.0.0 source tag 4b42fb8.
