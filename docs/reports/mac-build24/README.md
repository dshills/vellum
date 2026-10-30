# Private build24 delivery — 30 October 2026

Matched SDK and five apps are built and installed from green main
`b9212489afc921171fc5abbd5275d8919ae930dd`. This includes the e589f36e fix for
late Notes-family actions after exact-window retirement. Features, backend,
rendering/pacing and release/performance policy remain frozen.

- [Framework](https://github.com/dshills/vellum/actions/runs/41091357397),
  [Stable root API](https://github.com/dshills/vellum/actions/runs/41091357457) and
  [Platform adapters](https://github.com/dshills/vellum/actions/runs/41091357554)
  passed their first attempts on that exact source. Manual qualification remains
  skipped under existing policy.
- [Packaging summary](packaging.json): SDK relocation/module origins, generated
  minimal/document race/vet/build, public API scenarios, five builds, ZIP-extracted
  signatures and installed version/byte/signature verification passed.
- [Independent verification](verification.json): the durable copied delivery and
  selected installation passed source/checksum/receipt/metadata/byte checks and
  ten strict signatures without rebuilding, reinstalling or launching.
- [Static linkage](linkage.json): all five executables are thin ARM64 with 23 system
  dependencies each, system dyld, no rpaths, embedded environment overrides or
  extra Mach-O binaries. This is metadata inspection, not runtime acceptance.

## Actual deployment limit

All five binaries declare minimum macOS **27.0** and SDK 27.0; current inspection
host is 27.2. They cannot run on the historically qualified 26.6.2 host. The
existing framework support matrix is unchanged; this private current-Mac candidate
is not a new qualified release. No app window was opened and launch on 27.2 is
still unverified. Older-host compatibility must not be inferred from historical
qualification, successful compilation, signatures or the inspected version alone.

No public tag/asset, audit decision, signing identity, notarization, native pin or
notebook migration changed. The Forge headless icon-only command exits before
UI/data setup. Other hosts/hardware, human IME/VoiceOver, quarantine launch and
dynamic/plugin/transitive system dependency closure remain unqualified.

## Exact private identities

| Record | Identity |
| --- | --- |
| Source | `b9212489afc921171fc5abbd5275d8919ae930dd` |
| SDK SHA256 | `5a825eb07c480f2df248435b7c0a55e0b0cb5104ccd2de3b8f146c55c2798382` |
| Bundle manifest SHA256 | `516456cc39bc0b3c2b8faf14871cc820afda813a0a5f269e23da15fd97c77727` |
| Installation receipt SHA256 | `1c8a8b8425759434214631ff73b8ec9293af9c89dd82847f111cdea544a39fcd` |
| Installed apps | `~/Applications/Vellum/build24-b9212489afc9/apps` |
| Durable private packet | `~/Documents/ChatGPT/Go Native UI/deliveries/mac-build24-20261030/final-build24-b9212489` |

The packet contains the SDK, Forge, Measure Pad, A/B Documents, Document Starter,
Notes ZIPs, SHA256SUMS, original producer/delivery/installation/preflight receipts,
CI, PACKAGING, VERIFICATION and LINKAGE reports plus a handoff. Original archive
and manifest bytes retain their identities. Build23 and earlier installations
remain intact; they were not silently replaced or requalified.

Subsequent report/checker commits on codex/mac-build24-closeout are newer than
the packaged source. Their focused local checks are recorded in
[internal reviews](../../reviews/mac-build24/); no new hosted-CI result or SDK
content is claimed for them. Use the current checkout for the newly added static
checker. Report-only commits never relabel built artifacts.

The [current handoff](../../mac-document-workflows-handoff.md) and
[safe resumption checklist](../../mac-desktop-resumption.md) identify the remaining
bounded desktop checks. Preserve any old dirty B fixture before switching versions;
use only disposable data directories when desktop checks are eventually authorized.
No app launch is part of this delivery checkpoint.
