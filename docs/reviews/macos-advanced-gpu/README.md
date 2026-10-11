# Advanced GPU review

Internal agent review checked device/resource lifetime, lock ordering, source
generations, depth, checked arithmetic, callback borrowing and native assertions.
It identified three corrected issues: separate Prepare/Draw helper leases;
coordinate conversion after stale-input guards; one-shot Destroy preserving a
remounted program claim. Manifest order and statistics wording were corrected.
The settled resize assertion was strengthened after internal review.

[Initial Prism report](prism-initial.json): Anthropic Claude Sonnet5.5,37 nonempty
files,125624 bytes,7 chunks,11 calls,4 medium and0 high, complete coverage with
no exclusions/skips/truncation. The staged proof also includes one empty successful
vet log. [Adjudication](adjudication.json) independently confirms one input bug,
refutes owner-race and pipeline-leak claims, and clarifies logical-versus-physical
resource accounting. Calling Destroy on pending preparation handles was rejected.

[Targeted correction review](prism-delta.json):11 nonempty files,28321 bytes,
2 chunks,2 calls,0 findings, complete coverage with no exclusions/skips/truncation.
This reviewed only new changes; the unchanged broad batch was not repeated.
Actual key/scroll propagation tests pass; affected driver race/vet and independent
consumer race/vet/build/native checks pass after correction.

[Final reviewed input hashes](staged-inputs.json) combine the initial and delta
proofs. The commit must match every hash; these immutable review records are added
after the recorded input snapshot. [Native evidence](../../reports/macos-advanced-gpu/README.md)
and [contract](../../../specs/advanced-gpu.md) preserve outstanding Mac/P4 gates.
