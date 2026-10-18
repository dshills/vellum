# Window-origin Custom routing (stage 81)

Adds an optional main-scene handler for Custom events before modal stage routing,
inside the existing render-context lock. Only the owning integration session's
private Custom requests are consumed. Ordinary Custom and every non-Custom event
retain normal popup/hit-test/capture routing. Install before native startup or
from its event loop; clear on teardown. A real software modal MenuStage regression
proves interception, fallback and revocation. It is routing evidence, not native
physical input or presentation qualification. No tracing lock spans the callback.
