# Bounded originating input measurements

Stage0098 adds an opt-in bounded FrameTrace event-group transaction. It preserves
actual event identities and waits for dispatch, owner drain and reconciliation;
only a separate checked commit arms a composition. In-flight/intermediate
compositions keep their original unarmed identity. Cancellation preserves input
and owner terminals, while window retirement releases the final reservation.
Ordinary single-event tracing is unchanged. Controlled Core tests do not provide
real native presentation, physical input or Mac release evidence.

Grouped owner/reconciliation accounting starts after the first original input
has actually dispatched and accumulates intermediate drains/refreshes. Only the
final input terminal plus a following drain permits commit; CPU accounting
cannot arm an earlier composition. Pre-input admission/setup is excluded.
Tests cover exact intermediate/final totals, canceled groups and ordinary-trace
compatibility. CPU stage intervals are not exclusive per-callback attribution.

Unarmed frames after first dispatch accumulate their render/compose work without
inheriting final generation. GroupReady and CommitGroup share checked terminal
prerequisites so a known earlier failure skips app-state observation. Final
retained group totals include later input/owner work after prefix failure.
