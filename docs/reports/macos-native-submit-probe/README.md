# Native submission topology probe

Prepared follow-up to the completed matched calibration. The audited locked
wgpu-hal Metal queue commits presentation in a separate command buffer, whereas
the native reference combined render and presentation. This difference is a
hypothesis to test, not an established cause or framework repair.

The [protocol](../../../tools/macos-calibration/submit/PLAN.md) freezes three
combined/split native pairs, the existing match, timing boundaries, stop rules,
and a four-minute execution bound. GPU execution timestamps come from the Metal
render command buffer; scheduled/completed callbacks record CPU observation times.
The probe does not modify the SDK, framework, user application or release gate.

Preparation: warning-free native build; 18 unchanged parent classifier controls
and seven submission classifier test methods (including nine invalid timing
subcases). Automated synthetic controls do not verify physical input or pixels.
Results will be added only after the reviewed source checkpoint is committed.
