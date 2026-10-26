
## Record bounded Mac content interaction checks and reproducible fixtures

Prism: `mac-integration-interactions.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete. Addressed the PDF offset observation by assembling bytes before recording xref offsets; fixtures are deterministic ASCII and validated with an independent PDF offset/audio-duration reader. Native pointer media seek/volume/play, file Browse/Reveal and calendar range accepted; PDF navigation/search exercised and concrete visible-layout/selection defects assigned to checkpoint 2. Clipboard Copy approval is pending; native file dropping and corrected/focus checks remain explicitly open for subsequent checkpoints, not claimed passed.

Focused checks: `mac-integration-baseline-checks.json`, `mac-integration-fixtures-final-checks.json`.
