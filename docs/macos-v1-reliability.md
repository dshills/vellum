# macOS v1 candidate reliability

The intermittent tree cancellation assertion was a diagnostic race: context.AfterFunc runs asynchronously, allowing the offscreen unsupported-composition terminal to win after cancellation but before its callback. The reserved grid/tree FIFO event now observes an already-canceled context synchronously at actual dispatch, before acquiring render ownership. It still delivers the identical admitted scroll. AfterFunc remains for cancellation during waiting/in-flight work. Existing earlier terminal results remain immutable. No renderer, presentation policy or responsiveness gate changes.

Focused evidence covers deterministic delayed-notification precedence, immutable prior failures, the actual tree canceled-delivery owner transition, grid admission/FIFO, and bounded race repetitions. Ordinary foreground first paint is recorded separately in the native candidate check; offscreen capture cannot establish it.
