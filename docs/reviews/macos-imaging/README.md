# Compute, HDR and asynchronous capture review

Internal agent reviews checked borrowed mapped memory, teardown lock order,
resource accounting, cancellation reservations, owner delivery and native fixture
precision. The independent consumer and reproducible backend patches were reviewed
separately. No operating-system preferences were changed.

[Initial Prism review](vellum-imaging-prism.json) used Anthropic Claude Sonnet 5.5:
45 nonempty files, 191546 bytes, 11 chunks, 18 calls, six medium and no high findings.
The proof includes one empty successful log. [Targeted corrections](vellum-imaging-prism-delta.json)
reviewed 31 nonempty files, 134207 bytes, seven chunks, 11 calls, four medium and
no high findings. [Three-file metadata follow-up](vellum-imaging-prism-metadata.json)
reviewed 22358 bytes, one chunk, two calls, one medium and no high findings.
Every review reports complete coverage with no exclusions, skips or truncation.
The unchanged broad batch was not repeated.

[Adjudication](adjudication.json) records each finding. A blocking diagnostic send
was corrected and tested under full-channel admission/completion. The closure
fixture now verifies domain cancellation, registration removal, program release,
native closure and a native teardown lock barrier. That stronger fixture exposed
a real backend window-registry race, corrected by reproducible patch25, plus an
incorrect fixture assumption about raw native closure and domain registration.
Historical failures remain in the acceptance evidence.

Evidence references and final/historical digests were corrected. The last finding
asked for proof of existing stock backend vet warnings. A private prepared copy
had patch25 reversed; all original hashes and absence of the added test were
verified. [Baseline proof](vet-baseline.json) and its full log establish identical
stock vet diagnostics and exit1 before/after patch25. This is a known upstream
failure, explicitly retained, not a passing gate. The separate scoped vet run
excludes composites/unreachable and does not certify those analyzers. No source
change or repeated broad review was needed for this attribution proof.

[Final input hashes](staged-inputs.json) merge the initial, correction and metadata
proofs in that order. All hashes must match the commit. These immutable review
records and baseline evidence were added afterwards. A separate private SDK from
the clean commit will reproduce all prepared native inputs and verify the archive,
standard consumers and independent native imaging after relocation.

The bounded mapped-row copy holds native/resource locks to protect borrowed
mapped memory and can delay paint or resize; no latency guarantee is claimed.
Private decoder/reservation destruction after native closure was not measured by
the public consumer. Physical device loss, presented-frame/external accessibility,
sustained performance and public distribution remain qualification work. See
[native acceptance](../../reports/macos-imaging/README.md) and the
[Mac completion audit](../../reports/MACOS_COMPLETION_AUDIT.md).
