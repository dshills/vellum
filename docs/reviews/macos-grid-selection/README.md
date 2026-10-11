# Captured grid row-selection review

Internal review covers the portable representation, admitted-row validation,
native modifier/lifetime handling, consumer controls, tests and contracts.
Draft-1 retains two original failing portable regressions and exact source bytes;
draft-2 replays their unchanged bodies against the corrected runtime. Source
archives use `.go.txt` to remain inert. Absolute private paths in original records
are execution provenance; historical-source-bindings.json maps retained copies.

The consumer's initial typed modifier helper failed compilation. Its corrected
suite exposed a real native border-sizing defect through the unchanged variable
geometry oracle. Both failures are preserved with their tested inputs. Native
geometry review verifies that cell/header exterior edges remain indexed and
rounded consistently while the cursor border stays inside. Initial native fixture
error-sink/layout assumptions are explicitly exploratory history.

Qualification wrapper review also requires owned process-group timeout cleanup,
failure records and refusal to overwrite older evidence. Current wrappers use
the existing process helper and exclusive output markers. Injected failure checks
and the helper's real descendant-cleanup tests are separate from GUI acceptance.

Prism reviews the completed implementation, tests, specifications and selected
README/wrapper files once using Anthropic Claude Sonnet 5.5. The process record,
input map, raw result and finding resolution are retained here. Final evidence
and source binding are audited internally before commit and promotion. None of
these records substitute for external AX, human input or full P6 performance.
