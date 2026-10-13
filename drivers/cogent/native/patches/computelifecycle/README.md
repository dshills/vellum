# Compute device and recording ownership

Patch 0050 closes ComputeSystem's private device and queue at shutdown while
preserving the borrowed GPU adapter. Serialized Release detaches owner fields,
abandons pending recording, completes prior work, retires pipelines/values and
then retires the device. Empty/repeated shutdown is safe. GraphicsSystem retains
its separate borrowed-device contract.

NewComputeSystemChecked refuses missing/invalid adapters and incomplete device
admission. The legacy constructor logs a refusal and returns nil. Pass admission
publishes only a complete live pair; a partial or refused factory result retires
acquired references. The pinned native BeginComputePass has no Try API and can
panic. Acquired command cleanup unwinds then, but native panics are not converted
to returned errors. No broad panic/resource-recovery guarantee is added.

EndComputePass still requires caller End. It detaches recording, releases the pass
before Finish, and uses patch48's submission helper to retire command/buffer on
success or refusal. Old bind groups retire on both paths. Abandoned recordings
are never submitted. Device.Release detaches and releases Queue independently of
Device, polling only a live device; a partial queue cannot survive an early return.

Original-compatible empty-End/Release and partial-device tests fail on the prior
implementation. Nine counted admission scenarios and panic unwind verify cleanup.
Opt-in tiny headless checks exercise real shader output, submitted/abandoned
recordings, private device retirement, child resources and borrowed-owner liveness.
Those checks do not qualify native/GPU heap reclamation, physical device loss,
window presentation, sustained workloads or performance. Modified upstream notices
remain BSD-style; new tests/helper are MIT. Inputs/outputs and patch are hash-pinned.
