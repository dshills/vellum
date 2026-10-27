# Coherent macOS v1 build13 — final local candidates

The SDK and all six apps were freshly built from clean green commit
`64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a`. Every app was compiled from the newly exported and independently
verified SDK, including ordinary stage0108 preparation. Features, stable APIs,
backend, experimental options and release/performance policy are unchanged.
This report checkpoint changes no production code.

[Publication packet](PUBLICATION_PACKET.md) · [Exact upload set](PUBLICATION_SET.json)
· [Checksums](SHA256SUMS) · [Package record](final-package.json)
· [Native results](native-result.json) · [Reproduce](REPRODUCE.md).

## Verified

- All three workflows passed on this exact source: [Framework](https://github.com/dshills/vellum/actions/runs/40504082685), [Platform adapters](https://github.com/dshills/vellum/actions/runs/40504082565), [Stable root API](https://github.com/dshills/vellum/actions/runs/40504082914). [Reused results](reused-ci.json) do not claim CI for this later report-only commit.
- Fresh SDK: 8,861 checksum-verified files, public producer guards, independent relocated consumer checks, and race-enabled extracted-SDK stage0108 regressions pass. Every SDK file still verifies after building all apps.
- Six builds select Vellum/Core/typesetting/native replacements inside that SDK. Actual external-link maps select native archive 786aaad8…. Effective dependency versions/sums and all 4,441 selected embeds per app match prior candidates. Original showcase icon PNGs regenerate identically.
- 72 successful bounded production/consumer commands and 12 post-launch strict signature checks. ZIP extraction, version 1.0.0/build13, source metadata, pre-sign/signed executables, manifests, notices and archive hashes verify.
- All six exact ZIP-derived app copies render and close on Apple Silicon/macOS 26.6.2. Notes uses a fresh disposable profile; the personal notebook is untouched. Widgets' managed Light/Dark controls retain original-color artwork. Forma's poster, Luma's image/filmstrip/histogram and Pulse's chart/table render.
- Shell: native close of edited B2 reveals untouched A0 **before any action in A**, then A closes via its application button. [B2](shell-b2.jpg), [A0](shell-a0-after-native-close.jpg). All six exact executable process paths are absent afterward. Exit codes were not collected and are not invented.

## Distribution delta

[SDK delta](sdk-delta.json) records all changed/added source files. Core's
stage0108 exposure repair and Shell's light chrome/background correction are
now present across the rebuilt delivery, without a manual SDK overlay.
**Prepared Go sources and all executable/archive bytes change.** The old mixed
source proposal and its archives retain their original identities.

All 559 bundled dependency declaration/native-notice files, third-party inventory
and native archive are identical. Original notice bodies/component references
are preserved: 418 files for Notes/Widgets/Shell, 419 for Forma/Luma/Pulse. Only
artifact_sha256 is rebound to each new pre-sign binary; the old binding rejects
it. Both compiled Go Regular/Bold TTF arrays are found directly in each showcase
executable, and the original Bigelow & Holmes font notice remains included.

The collection labels pending/review-required remain unchanged collection
semantics. They are separate from the unchanged approved release/native audit
guard. The [historical audit record](../../../release/macos-v1-distribution-audit.json)
is byte-identical; its exact old binary approval is not silently extended to
these new archives. This final delta is disclosed for the separate owner decision.
No new upstream rights grant is inferred; existing pinned objc2 caveats remain.

## Retained observations and limitations

The deliberately background-launched Notes capture was initially blank; native
Raise displayed correct prose without a model edit. Both captures are retained.
The first Luma close action left its window/process running; a fresh observed
native close retired it. One separate fresh launch followed by a **single**
native close also retired normally. The first process-check failure, intermediate
AX/capture and passing control remain in evidence. No sole cause is assigned.

Controller errors are retained: an unused historical packet lookup had the
wrong filename after four successful preparation commands; it was removed and
verified outputs reused. The tag check compared an annotated tag object instead
of its peeled commit; the assertion was corrected. A report-side trailing
newline was also corrected before JSON decoding. None changed product source
or weakened a release guard. Successful unittest summaries and requested
codesign replacement messages can appear on stderr with exit0.

Sealed manifests retain producer-time target_runtime=unverified. Final
runtime_result fields in final-package.json refer to native-result.json;
raw pre-launch records retain capture-time not-yet-run values. The
[indexed evidence](evidence-index.json) retains controllers, original logs,
captures and link maps. Local pre-sign inputs are retained separately and are
excluded from the upload set. This is native AX/render/close evidence, not new
physical-pointer, performance, soak, exhaustive-widget or human IME/VoiceOver
qualification. Reuse unchanged prior verification within its original limits.

Other OS/hardware/displays, cross-window canvas drag and downloaded-quarantine
launch remain unqualified. Apps are **ad-hoc signed and unnotarized**. No signing
identity, notarization, security bypass, global setting change, upload, new tag
or tag movement occurred. The v1.0.0 source tag remains 4b42fb8. The final proposal uses a separate
macos-v1.0.0-build13 packaging tag at64bf9cf, matching every artifact. No such
tag has been created. Publication is
pending; the bytes are ready locally in research workspace
`deliveries/macos-v1-build13-20261027/`. [Prism review](../../reviews/macos-v1-build13/README.md).
