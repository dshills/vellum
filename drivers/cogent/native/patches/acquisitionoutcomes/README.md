# Measured acquisition outcomes (stage 85)

An opt-in measured render pass carries a detached value containing source,
acquisition stage, actual attempt count and first/last backend statuses. The
existing acquisition protocol still yields on unavailability and retries a
lost/outdated surface at most once, preserving failed-handle release order.
Ordinary acquisition is unchanged. No polling, extra acquire/present, readback,
queue wait, retained history or native handles are added by the observation.

The existing measured APIs preserve their return and ownership conventions;
new optional outcome methods extend them. Legacy measured renderers report
unrecorded detail, and offscreen textures do not invent surface attempts or
statuses. Surface ownership/configuration/view work remains in the same timing
boundaries. Metadata work is charged as measured UI work.

Measured composition includes the outcome in an unavailable-pass failure or
wraps an existing failure with outcome context and errors.Is identity. A precise
acquisition result does not identify a later encoder/pass error as its cause.
Outcome descriptions are bounded by scalar enums/counts/statuses; arbitrary
existing backend error text is not newly bounded. Strict supported/submitted/
error-free frame completion is unchanged. The old failed structural scope is
immutable and cannot be diagnosed retrospectively from these new fields.

Injected regressions compare ordinary and observed acquisition/release/retry
protocols, unavailable/recovered/null/error/unknown outcomes, legacy compatibility,
actual GraphicsSystem no-frame routing, composition error propagation and strict
FrameTrace failure. These are separate from native device and physical acceptance.
