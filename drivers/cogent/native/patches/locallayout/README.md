# Fixed-allocation local layout

Stage 40 adds Core's explicitly owned `LocalLayoutRoot` boundary and
`Scene.NeedsLocalLayout`. At most 64 live initialized roots are queued per scene;
identity coalescing, bounded ancestor membership and fixed host-envelope checks
protect admission. Native mutation uses the existing originating render context.

Ordinary rebuild/restyle/full layout consumes admitted requests before callback
work while retaining an atomic `GlobalLayoutPending` fence. Later global/local
layout and render requests survive for another frame. Failed local layout falls
back to full layout. Abnormal full/render paint leaves a fenced full retry.
Successful local layout repaints the complete scene at original painter order.
Scene destruction clears queued references before and after the existing timer
and widget cleanup. This is not a foreign geometry mutation or GPU upload API.

Slider Render now calculates its thumb position without requesting another
paint from inside paint. Existing setters still request painting. The old slider
otherwise produces perpetual painting once legitimate render callback requests
are preserved.

Source-bound tests cover identity/capacity/envelopes, reentrant global/local/render
requests, full precedence/fallback, abnormal unwind/retry, detached/retired roots,
painter order, atomic fences, quiescence and existing timer lifetimes. Original
requeue and slider overlays fail actual assertions. Input/output pins cover all
six prepared files. Native driver integration, full-frame performance and broader
Mac acceptance remain separate qualifications.
