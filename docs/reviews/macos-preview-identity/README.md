# Preview candidate identity checkpoint

Version/API metadata now identifies v0.1.0-alpha.3. The existing version contract
race test passes with Go1.25.6. No features, rendering/scheduling behavior, or
publication/audit guards changed. Private R2 SDK/app records remain alpha.2;
final alpha.3 packaging and acceptance are still required. The ongoing R3 runtime
fixtures use immutable SDK source d8de4fc, before this version-string-only change.

Prism using Anthropic Claude Sonnet5.5 reviewed all four changed files, complete
coverage, zero retained/discarded/suppressed findings, no exclusions/truncation.
Review receipts normalize the user-home prefix; originals stay in session records.
This closes only the identity sub-checkpoint, not all R4 distribution requirements.
