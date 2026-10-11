# Source cost diagnostics

Stage 36 follows empty-sprite omission. Exact input, patch and output hashes
bind the optional native path. Ordinary ComposeMeasured and zero-value FrameTrace
keep their existing draw path. The explicit detailed composer invokes Start once,
each source once in the original order, and measured End once. Overflow refuses
diagnostic completion after drawing and submitting every source.

Known scene/sprite paint sources time the actual lazy raster Render and Copy
calls. Scrim/insets time their actual Copy calls. Unknown sources retain Draw
time with unavailable raster/copy breakdowns. Immutable numeric source kinds
are captured at gathering, not inferred from blending operations or live scenes.

GPUDrawer keeps a nil diagnostic pointer on the ordinary path. An explicit
composition begins cumulative scalar image counters after Start. Begin/Read/End
take separate mutex acquisitions; no drawer lock crosses source Draw. Deferred
End clears the diagnostic pointer on normal return, panic and Goexit. Existing
renderContext and native-window ownership serialize the full batch.

The actual existing-and-unchanged branch counts a cache hit. The existing Value
SetFromGoImage call records attempted input dimensions at checked four bytes per
pixel and elapsed API time, including conversion/allocation/queue/sampler work.
This is not confirmed transfer success, queue-write count, physical GPU bytes,
GPU execution, or display latency. The existing API logs errors without returning
them, so upload success/failure remains unavailable. Uniform Fill and resident
UseTexture need no Go-image attempts; unknown drawers expose unavailable counters.

Each ticket retains at most 32 scalar source records across intermediate and
final compositions, with composition and source ordinals. Aggregate count/byte
overflow refuses completion. Ticket-owned records detach from composer storage;
public Snapshot clones the nested details. Source records retain no images,
widgets, sources, arbitrary names, errors or user text. All source intervals are
nested inside RasterUploadNS and do not add to TotalUIWorkNS.

The separate public constructor accepts capacity 1..4096, while ordinary
NewFrameMetrics retains 1..65536. The loaded-grid source-details mode is explicit,
and ordinary warm-budget validation refuses it. Extra clocks and scalar checks
affect detailed elapsed time. Software pixel/copy tests and actual GPU diagnostic
observations have different scopes; neither qualifies the full P6 workload.
