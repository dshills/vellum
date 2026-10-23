# alpha.3 private developer-preview: final delivered-byte checkpoint

22 October 2026. **Private candidate verified; public prerelease not complete.**
The app and SDK are usable local artifacts. Features remain frozen and default
Cogent/WebGPU/Metal is unchanged. No source, tag or release asset was published.

## Exact delivered artifacts

| Artifact | SHA-256 |
| --- | --- |
| SDK `final-private-sdk.tar.gz` | `bb5bbba8161efb88a102b3189e189da8038a5c18063beaedbae230d5e5b23887` |
| App `Vellum-alpha.3-build28-private-macos-arm64.zip` | `acd90220d2a27f386d48e7277c620fd555357c54a87cfe96b8057e72f95c6e74` |
| Extracted signed app executable | `144e76be8a84654a45baab46f721d38b1f740a16b00950bf46f79823ab3d6f5e` |

SDK source is clean `10beef424c532395802065b2be45119bec88efe2`. The app's input
binary was built from SDK source `412e5b8266fb8655935bcdad203b73887923c9db` and
has SHA `0d7539594ad742844f629ea1ab6080edb0a86ddb8ffbffe8192d3b53d636a16a`.
Manifest comparison establishes that all Go/native/dependency inputs match;
the six changed SDK files are documentation and packaging scripts/tests, listed
in `final-sdk-product-equivalence.json`. This is code equivalence, not a claim
that the app was recompiled at the latest documentation commit.

The SDK verifier checks 8,468 file hashes, relocation/replacements, external Go
module cache verification, standalone consumer race/vet/build and public API
race/vet using Go 1.25.6. A new starter was generated from a fresh extraction
outside the checkout, built, launched, checked its GPU rectangle readback and
exited normally. This single readback is not latency qualification.

The app is numeric version 0.1.0/build 28, SDK identity `v0.1.0-alpha.3`, Apple
Silicon, ad-hoc signed and **not notarized**. Its 418 original notice files and
175 inventory entries bind the selected pre-sign executable. Audit status is
still pending; notice presence does not grant redistribution permission.

## Actual ZIP acceptance

The app was staged under `/private/tmp`, zipped, freshly extracted outside the
checkout, strictly signature-verified, then launched from that extraction.
App-only native CGEvent input verified one Block placement, empty-space pan and
scroll zoom. Native accessibility actions verified Undo removed the object and
Redo restored the same `block-1`. Save As wrote a disposable drawing; New emptied
the canvas; native Open restored the object and 94% zoom with a clean history.
The saved JSON records exact geometry, camera and snap state. Normal CmdQ retired
the exact owned process. Strict signature verification still passes afterward.

The earlier [full app acceptance](../macos-preview-app-acceptance/README.md)
remains applicable through unchanged product inputs. It covers group/multiselection
editing, history, rejection/cancellation, corrupt-file protection and controlled
recovery. These checks use native automation, not human physical pointer input,
VoiceOver or real IME acceptance. A stale last-action status can show 100% while
the dedicated zoom control and saved camera correctly show 94%; cosmetic backlog.

## Try it / limits

Extract the ZIP and open `Vellum Placement Preview.app`. The adjacent README
explains palette placement, selection, pan/zoom, snapping, history and Save/Open.
Keep experiments in new temporary drawings. For SDK use, generate a new external
module with the extracted `scripts/new-app.py`, source `drivers/cogent/env.sh`
and follow the generated instructions with Go 1.25.6. Do not disable macOS security
or alter global input/accessibility preferences to run this private candidate.

Support is limited to the tested Apple Silicon macOS configuration. Placement is
same-window. Cross-display, high-refresh, Intel/other-machine, human VoiceOver/IME
and full production/endurance qualification remain open. Historical latency gaps
are disclosed in the release ledger; no new latency claim is made here.

The [resource checkpoint](../macos-preview-resource-check/README.md) passes active
editing and visible idle samples, but the required two-hour soak remains blocked
at the native occlusion transition. Public distribution also awaits the concrete
component-rights decision, applicable candidate CI and final owner publication
approval. This report does not change those gates or label a private build a
completed public release. Public source push/draft PR is separately awaiting the
owner's answer; historical remote CI is not CI for this local branch.

Local checksums bind the candidate bytes, not independent authenticity. Records
normalize home paths and retain test-only timestamps/process/window observations.
