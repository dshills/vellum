# Bounded span header scratch

Owned patch 0066 reuses exact power-of-two construction header bands, one each
for sizes 1..256, within one sceneplan Builder. The nine slots bound backing to
511 slice headers (12,264 bytes on64-bit); larger/clamped non-power bands bypass
reuse. Active stream operation storage is inline. These are construction costs,
separate from unchanged logical packet admission and exact retained payload.

Reservation remains before borrowing; growth copies row references before
clearing only the old header container. Flush detaches exact packet rows/span
storage before clearing and retaining construction headers. It then clears its
inline operation. Success/refusal Seal drops all scratch and stream references.
No global pool, font/model retention, packet alias or cross-renderer reuse exists.

The unchanged original builder oracles cover state/admission and original/GPU
encodings/pixels. Added tests check bounds, zeroed references, packet isolation,
clamped bands and budget-before-borrow. Software allocation comparisons and
native latency qualification are separate; this patch does not close P6.
