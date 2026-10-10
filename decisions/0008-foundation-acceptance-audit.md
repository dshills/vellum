# ADR 0008: Hold G1 after the remaining local acceptance audit

Date: 10 October 2026. Status: accepted for experimental evaluation; production foundation unaccepted.

The remaining local probes establish exact long-buffer/grapheme behavior and native UTF-16 context clipping. They reproduce a failure between native batches and legacy shortcuts: a shortcut sees only 16 of 32 previously admitted updates. They also show logical bidi navigation, insertion-only native geometry, and approximately 53 MiB retained font-map heap across 64 family queries resolving 16 locations. These results do not establish visual editing or global memory bounds.

Retain Go application ownership over Cogent Core, go-text and native adapters as the experimental direction. Hold broad workbench expansion. Resolve cross-path input ordering first, followed by visual caret/selection and geometry contracts and broader resource/performance admission. Do not infer live IME or full VoiceOver success from deterministic protocol probes or unverified synthetic input-source switching.

Core/go-text and focused race suites pass, but the broad text regression run fails. Parser panics reproduce on the pinned unmodified control; image mismatches do not reproduce in isolated repeated rendering runs and remain unexplained. Windows/Linux runtime and trusted clean-machine release evidence remain absent. Local instrumentation and review do not close those gates. See [the acceptance report](../FOUNDATION_ACCEPTANCE_AUDIT.md) for evidence, limitations and reproduction.
