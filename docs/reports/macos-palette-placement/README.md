# Palette placement checkpoint

This checkpoint implements typed, reusable palette-to-canvas placement through
public Go APIs and the independent `examples/placement` module. See the
[contract](../../../specs/placement.md) and [demo](../../../examples/placement/README.md).
It does not complete the Mac framework or P9/P11 acceptance gates.

The portable controller owns one gesture, stable catalog/target identities,
transformed coordinate validation, optional snapping, rejection and an
application command callback. Managed Button sources and Canvas2D targets share
native window capture and ordinary application ownership. Source/target removal,
visibility or disabled transitions, capture/focus loss and window retirement
cancel without creating an object. Valid release retires before command delivery
and consumes the native event. Keyboard activation shares validation and command
routing. The demo records one history change per placement; Redo restores the
same object ID and captured geometry.

Automated evidence is deliberately scoped. Portable tests cover clipped/scrolled
origins, pan/zoom and scale arithmetic, command/history identity, rejection,
copy/reentrancy refusal and cancellation. Actual offscreen widget routes cover
owner batches 1/7/64, DPI 96/160/240, release consumption, lifecycle cleanup,
reserved cancellation under a full ordinary queue, admission exhaustion, stale
window domains, queued keyboard/semantic origin leases and viewport coalescing. A native ancestor scrollbar supplies real
scrolled geometry for pointer placement; software canvas pixel changes establish
visible preview output, and native sprite removal establishes terminal cleanup.
The separate public consumer routes all three tools, keyboard activation,
rejection, Undo/Redo, selection, pan and wheel zoom.

Code checks include race tests, vet, pinned preparation reproduction, private SDK
packaging/relocation and Mac native compilation. Checkpoint acceptance requires exact results and source hashes
in `records/qualification.json` and the evidence archive. Failed
fixture iterations remain in the archive and are excluded from acceptance.
Internal review and Prism using Anthropic Claude Sonnet 5.5 must complete before commit;
provider findings and their dispositions are retained without rewriting the raw
provider outcome.

No native executable is launched for this checkpoint while the human is online.
Physical Mac pointer capture, dragging outside/re-entering a window, cross-display
scale changes, GPU presentation and external accessibility remain unverified.
Only same-native-window placement is implemented; cross-window or cross-app
transport, OS drag sessions and drag autoscroll are unsupported. The managed
Canvas2D is a retained rectangle/navigation surface; application domain geometry
and undo policy remain application-owned. It does not change the older GPU
MeshCanvas viewport-unit contract or qualify a CAD application.
