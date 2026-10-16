# Native API-attempt observations

`vellum_native_checks` builds expose a detached `NativeActivitySnapshot` through
the driver. `Available` means that instrumentation was compiled in, independently
of native backend initialization. The fixed process-wide counters cover admitted
desktop empty-event post attempts and held window-surface backend Present
attempts across windows. Refused pre-initialization/retired posts, idle surfaces
and RenderTexture no-ops do not count. An admitted call that panics still counts
as an attempt. Direct third-party/backend calls are outside this scope.

Counters saturate at uint64 maximum and become unusable; they never reset or
wrap. Saturation is derived from the same loaded values returned in the snapshot.
Reads are independently atomic per counter, not one atomic combined snapshot.
Returned values retain no callback, window, image, model or historical registry.
Ordinary builds use disabled counters. Their original surface Present and
PresentMeasured method bodies are retained separately from diagnostic routing.

The native soak reads start counters after its starting Cocoa state query and
before CPU0/quiet start. It reads end counters after CPU1/quiet end and before
the ending Cocoa query. Each read retains its own start/end UTC timestamps in
version-1 `framework_native_activity` data. The bracket is a superset of the CPU
quiet interval and may count activity at its margins; it is not exact per-event
attribution. Missing/unavailable/saturated/decreasing counters, wrong owned PID
or unbracketed reads refuse the sample without publishing usable deltas.

PID must match the actual controller-owned process receipt as well as both
snapshots. PID equality is not a process lifetime capability or artifact
authenticity proof. The strict `scripts/native_activity.py` component validates
the new schema, exact integer deltas and nanosecond brackets against a supplied
owned PID. It does not replace complete source/process/trace/window/CPU/cleanup
validation. Old archived validators and source vectors remain historical;
new native campaigns must require this observation component.

Event-post coalescing and actual OS wakeups are unavailable. Present attempts do
not establish GPU completion or compositor presentation. Existing
`app_originated_wakeups_available`, `present_count_available` and
`full_p11_qualified` remain false. CPU/paint/API-attempt data alone cannot qualify
the unchanged full idle target, ordinary latency, successful overnight, native
memory or Mac completion. Inert injected callbacks validate counter/ownership/
delta behavior; actual native calls and quiet measurements remain separate gates.
