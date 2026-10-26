
## Record bounded Mac content interaction checks and reproducible fixtures

Prism: `mac-integration-interactions.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete. Addressed the PDF offset observation by assembling bytes before recording xref offsets; fixtures are deterministic ASCII and validated with an independent PDF offset/audio-duration reader. Native pointer media seek/volume/play, file Browse/Reveal and calendar range accepted; PDF navigation/search exercised and concrete visible-layout/selection defects assigned to checkpoint 2. Clipboard Copy approval is pending; native file dropping and corrected/focus checks remain explicitly open for subsequent checkpoints, not claimed passed.

Focused checks: `mac-integration-baseline-checks.json`, `mac-integration-fixtures-final-checks.json`.

## Polish embedded Mac content layout, Find visibility and scoped appearance

Prism: `mac-integration-polish.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete. Replaced normalization-sensitive match-string filtering with per-search delegate identity tokens and weak/main-owned delivery; stale results cannot reveal an old selection. Opaque colors follow explicit scoped Dark changes, and forced layout occurs only on geometry/appearance changes. Wide/narrow actual-window regressions pass with process-local unsaved geometry isolation. Reproduced title and transparent/search-width defects have focused regressions; corrected physical selection/focus/Dark checks are carried forward to final package interaction checks. Backend, gate and stable API baseline unchanged.

Focused checks: `mac-integration-polish-final-checks.json`, `mac-integration-polish-reviewed-checks.json`.

## Use public checklist and date-range controls in Reminders

Prism: `mac-integration-reminders.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete. Addressed the duplicated pending-input concern with separate calendar/preset action guards and disabled nontext actions, clearing enabled calendar navigation metadata. Controlled date-field Change deliberately stays admitted during FIFO settlement; blindly disabling the entire picker would drop valid pending text. A regression proves both calendar refusal and accepted settled dates. Existing bounded stable-identity history, native viewport allocation and full-title preservation pass. The example remains fictional/in-memory with no user data access.

Focused checks: `mac-integration-reminders-checks.json`, `mac-integration-reminders-final-checks.json`.
