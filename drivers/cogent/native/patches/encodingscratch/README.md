# Bounded GPU encoding scratch

Stage 0078 lets the serialized GPU scene renderer borrow four scalar byte arrays
for its checked synchronous queue-input writes. Aggregate retained capacities
are at most 1 MiB. Counts/position scratch and mask lookups remain per call;
no packet, font, image, native handle or application closure is retained here.
Cross-shaped growth resets backing before replacement; accepted encodings larger
than the cache limit retain the ordinary detached allocation path. The limit
never changes the original packet byte, operation or tile-reference admission.

Encode and EncodeGPU still return detached output. EncodeGPUWithScratch explicitly
borrows only until the next scratch use/reset. Every nonempty logical word is
written, and empty output padding is zeroed. Renderer resize and Release drop
backing. The immutable replay journal owns Packet objects only; neither it nor
GPU retirement keeps an Encoded value. Checked queue writes copy input bytes
before returning. Refusal keeps the existing CPU reconstruction/fallback path.

Coverage compares exact bytes over mixed masks, pixels, spans and covered clears;
large/small and cross-shaped changes; exact/first-over-cap limits; original tile
refusals and recovery; detached ownership; and resize/repeated release. The native
checked-write test overwrites scratch with different data before waiting/readback,
then exercises actual write refusal and recovery. Native latency, complete memory
accounting and full Mac qualification remain separate requirements.
