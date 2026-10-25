# Private macOS v1 candidate acceptance

All five bounded candidate areas are complete. This is a private `v1.0.0-rc.1` for personal evaluation, not published stable-v1 qualification. [Support/API matrix](../../macos-v1.md); [approved finish contract](../../../specs/macos-v1-candidate.md).

| Area | Result |
| --- | --- |
| Shared presentation | `6c6e95a`: shared emoji-aware Mac typography, compact fields and prose surface; actual light/dark geometry, retained input/insets and portable race/vet pass. |
| Native shell | `6ecc3a9`: per-window native toolbars, original-color icons, owner-checked availability and shared command menus. ABA, modal/customization, close/reopen isolation, bounded retry and shutdown checks pass. Physical A/B isolation and disabled refusal pass in final package. |
| Native interactions | `48c57e8`: fixed observed file-panel focus handoff, retained exact-parent/no-steal guards. Corrected package physically verifies Open/Escape then keyboard delivery without refocus. Notes title/prose Unicode, actual pointer selection and focused Undo/Redo were observed before this nonvisual fix. |
| Reliability | `99de0b4`: canceled diagnostic FIFO delivery checked synchronously, still preserves admitted input and earlier terminals. Actual grid/tree metrics pass20 race repetitions. Ordinary foreground Notes screenshot was visibly painted; background-only capture and other hardware are not qualified. |
| Candidate delivery | `18854e5` defines private candidate identity/support. Final SDK and all three apps are built from `48c57e81cf5bc2226df87a38fce0a1327fba32df`. External consumer module origins all9 stay inside the SDK; module integrity, race/vet and generated independent starter build pass. Clean ZIP extractions and installed native copies pass strict signatures and signed-byte comparisons. |

Runtime and archive identities are in [delivery-verification.json](delivery-verification.json). [Native acceptance](native-acceptance.json) records final physical cases and limits; [initial observations](native-build7.json) retain the real focus failure and service interruption instead of relabeling them passes. The final acceptance asserts only cases actually exercised. Scalar native restoration tests cover identity/state policy; actual closing/inactive/other-window responder scenarios are not simulated by that fixture.

The final Notes build8 was closed/relaunched normally and left foreground, with original count7 and prior selected n10. Only the newly created n13 was deleted; Undo restored the same ID/body and Redo removed it before Save. Native Shell and Gallery test windows were closed. Global input/accessibility settings stayed unchanged; clipboard-restoring paste was used only in the initial disposable note.

The initial CUA app-selection call took2840.7475sec despite its requested timeout. Its expired desktop authorization was renewed before final actions; this is recorded as a service-call interruption, not a proven application defect. A first immediate relaunch refusal during normal process draining was followed by a successful new-process launch.

## Remaining limits and backlog

Human IME candidate usability and VoiceOver remain unqualified by owner choice. Additional macOS/hardware/display configurations, stable API acceptance, final signing/notarization and public distribution require separate decisions/results. No new numeric performance gate or rendering/pacing policy was introduced. Replay/automation/diagnostics and broad professional/GPU workloads remain experimental. Existing soak/catalogue evidence is reused for unchanged paths; no new broad campaign is claimed.

Notes intentionally permits Undo/Redo with a native settlement adapter attached because edits may still need committing before history availability is known. With empty settled history these commands can report ErrEmpty. Improving the idle enabled appearance is a cosmetic follow-up, not incorrect history or lost input. Notes remains local plain text without iCloud/rich attachments.

Canonical delivery consists of clean SDK/app/source archives and SHA256SUMS. A loose app copied into the synced Documents folder acquired resource/Finder metadata and failed strict signature verification; that copy was removed. Verified local launch copies are in temporary staging and can be recreated from the clean ZIPs. Later report-only commits do not relabel SDK/app source bytes.
