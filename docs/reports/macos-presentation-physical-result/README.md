# Corrected phase-control attempt — native phase mismatch

The owner authorized one corrected live check with up to three minutes of screen
use. The existing fail-first native-reference/framework protocol stops after
**40.713 seconds**, within its 100-second overall bound. The native reference
returns normally in 34.051 seconds, but its receipt-phase admission fails.
**The corrected framework is not launched. Its startup remains unverified and
no matched native/framework latency comparison exists.** No unchanged retry occurs.

The exact failure is measured request **215**, phase **2** (fraction 5/8):
receipt is **2.143599 ms** after target, above the predeclared 0..2 ms diagnostic
phase interval. The excess is 0.143599 ms. This phase criterion determines
comparability; it is separate from the unchanged 33.3 ms release latency gate.
No sample is dropped, replaced or rounded into acceptance and neither cutoff changes.

An [offscreen raw audit](phase-rejection.json) checks all 248 requests, including
128 warmups and 120 measured requests, after capture. This is the sole phase
violation. Every target/anchor, receipt clock/tag and acquisition/drawable/callback
join verifies; the inherited native validator, display-clock checks and complete
pending replay pass. There are 251 retained attempts (including startup), 1001
events and 1986 display callbacks. Pending/acquiring finish at zero. These facts
do not admit the rejected pair or measure Metal pool availability, reference
lifetime, compositor depth, optical latency or release performance.

[Admission](admission.json) verifies all 6403 compiled pins, both full 6031-file
original/private Core vectors, unchanged binaries and the previously reviewed
linked-hook audit. HEAD 340a1bd adds only reviewed offscreen result documentation
after preparation 1124f8c; the local preparation metadata is explicitly rebound
without rebuilding. Native reference remains 44c85638…, corrected framework
6825d90e…, and the native archive remains 10e39953…. Original preparation/source
reports and the first failed capture remain sealed and unchanged.

The [result verification](execution-sources/verify_live_results.py.txt) checks
all nine owned groups (one native command and eight configuration probes),
complete raw logs, source/binary pins, geometry/configuration and all 8321 SDK
files. Every group retires and is observed absent. The [raw audit source](execution-sources/audit_rejection.py.txt)
independently reproduces the precise rejection across the full receipt population.
Existing classifier/controller/build/factory/pixel regression evidence is reused
through unchanged pins; no broad test campaign is repeated.

The [indexed evidence](evidence-index.json) retains the admission/rebind record,
exact command lines, source audit/verifier snapshots, raw journals, observations,
results and stop marker in [the archive](result-evidence.tar.gz). Index keys refer
to archive members; explicit committed-copy mappings prove snapshot byte identity.
Absolute paths describe local provenance; external source trees and binaries are
not included. Internal review and Prism using Anthropic Claude Sonnet 5.5 precede
commit; see the review completion receipt.

The desktop is free and the attempt is sealed. No additional UI run is authorized
by this result. The direct unresolved question is corrected framework startup;
a separately scoped startup-only check would answer that without first requiring
a new native phase-matching pair. No such additional check runs here. Features,
the existing Cogent/WebGPU/Metal backend, SDK/app, render/pacing defaults and release
gate remain unchanged; M3–M7 stay open. No performance repair, startup success,
Mac release qualification or public push is claimed.
