# Ordered shared-device scene paint

Stage 41 follows stage 40 and binds 32 changed/added upstream files. The scene
compiler owns immutable bounded paint packets; compute blends exact RGBA bytes
into the window drawer's existing device and appends a sampled source at the
original scene ordinal. A detached CPU journal reconstructs partial work for
snapshots and unsupported paint. Mutable image exposure preserves old image
semantics at the cost of CPU replay and later seed uploads.

Checked leases own source budgets, borrowed view lifetimes and safe deferred
retirement. Shared-to-owned slot recycling creates independent storage and
invalidates cached bind groups when handles change. The ordinary borrowed-texture
Redraw refusal remains. Native checks are explicit opt-in through
VELLUM_GPU_SCENE_TEST=1; default tests open no GPU device. Go module replacement
points direct upstream checks at the adjacent prepared bounded-font dependency.

The adapter is opt-in while representative performance and native memory
acceptance remain open. See [rendering contract](../../../../../specs/rendering.md).
