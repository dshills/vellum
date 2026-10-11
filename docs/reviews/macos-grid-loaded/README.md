# Stored-data grid reviews

This checkpoint received independent portable, consumer, native-field, contract,
cursor-source, cursor-regression and pinned-patch reviews. Each directory retains
the review, selected source identities, command proofs and inert snapshots. A
review certificate qualifies its stated source generation and scope; later
generation differences are recorded in the
[qualification ledger](../../reports/macos-grid-loaded/qualification.json).

One [Prism batch](prism.json) used Anthropic Claude Sonnet 5.5. It covered 44 files
in five complete chunks, excluding the consumer lockfile from review content;
all 45 selected files were identity-bound. Raw process exit 1 and all four
findings remain unchanged. [Final dispositions](findings.json) record three
source-backed refutations and one test-only correction. No production change
followed the batch. The report README and these final evidence records were
completed afterward and receive an independent metadata review.

[Portable adjudication](prism-portable-v1/review.json) checks the actual valid
token initializer and replaces exact callback counts with semantic phases. Its
same final test fails a reconstructed omission of the trailing input-health
check and passes the unchanged production implementation at race20. The previous
approved test and proof remain historical records.

[Consumer adjudication](prism-consumer-v1/adjudication.json) proves canceled and
refused worker saves can retry the same draft, and user read-only refusal happens
before task admission while explicit external immutable updates remain valid.
Three overlay tests pass 20 race repetitions each with frozen consumer sources.

The first Cocoa million-row process reached a workflow pass marker but exited 66
with a race. It remains a failure in the preserved consumer history. The final
fixed and variable processes passed only after fresh pinned cursor preparation,
using the required actual exit, stage and race-warning checks. No physical-input,
human IME/VoiceOver, sustained performance or public-release acceptance is claimed.
