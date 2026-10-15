# Scoped window pointer capture

Stage 65 adds one originating-render-owned primary mouse capture per native
window. Captured move/drag/up and keys route before popup/main-stage dispatch.
The token retires before terminal delivery; cancellation calls its loss callback
once and keeps only a boolean pending-release tombstone. The mouse-down default
tail cannot re-arm press, slide or repeat-click work after capture acquisition.
Escape, explicit cancellation, hidden/deleted owners, modal/popup exclusion,
window focus loss and stage retirement revoke the capability.

This patch provides routing, not application command ownership. Vellum's managed
placement adapter must apply owner/lifetime/geometry checks and clean its own
previews. Calls require originating render ownership. Offscreen event tests are
distinct from physical Mac pointer delivery and cross-display verification.
Cross-window placement is not supported.
