# Retained drawing review

One significant Prism batch used Anthropic Claude Sonnet 5.5 on 32 staged
source/specification/test/evidence files: 128,355 bytes, seven chunks and eleven
calls, with complete coverage, no exclusions, skipped chunks or truncation.
It reported four medium findings and no high findings. Independent internal
review refuted the model/view and status race findings through the actual
Application.Do call paths. The nil/closed-channel finding is unreachable under
this private fixture's confined, nonnil producer and never-closed channel.

The resize finding was partly valid: SetSize queried native geometry before
marshalling its mutation. The private fixture now uses SetWinSize with native
window coordinates, avoiding the worker-side GLFW queries. Application.Do alone
would not fix native thread affinity. Affected own-module race/vet/build and the
live race-built drawing fixture pass on the corrected source. A second Prism
review covered only that three-line correction (991 bytes, one call), with zero
findings and complete coverage. The unchanged significant batch was not repeated.
[adjudication.json](adjudication.json) retains each disposition, source identity,
raw report identities, final targeted commands and log hashes.

Internal review also resolved native lifetime guards, translated-mesh precision,
logical-device limits/checked target and submit errors, computed viewport bounds,
acknowledged retention, late input and consistent owner-held error reporting.
Final review found no remaining substantive defect in this phase. Error callbacks
run after native/resource lock release while holding the application owner;
the final live fixture successfully queries renderer stats from that callback.

`initial-staged-inputs.json` records all initial 32 index hashes, parent commit,
diff and original report hash. `staged-inputs.json` records the final 32 inputs
before these immutable review records were added. Only examples/drawing/smoke.go
changed between those records, exactly as covered by the corrective review.
The original [check evidence](../../reports/macos-drawing/README.md) retains its
original snapshot and logs; the corrected fixture's logs are stored here, so the
recorded original snapshot is not silently rewritten. The commit verifies all
final input hashes. Postcommit private SDK relocation evidence is generated
outside the repository because its future commit identity cannot appear here.

Native scroll-clipped compositor/MSAA edge quality, broader P4 workloads,
sustained performance, physical device loss and the other Mac acceptance gates
remain open. No OS input-source/VoiceOver changes or public distribution occurred.
