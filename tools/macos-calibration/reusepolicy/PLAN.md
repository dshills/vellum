# Completed texture-upload staging reuse: diagnostic design

This checkpoint is an ownership policy and source audit, not a native pool or
shipping implementation. Nothing imports policy.rs into Vellum, its SDK or the
private native library. The measured cache-mode change lowered CPU copy cost but
still missed the unchanged presentation gate. First/subsequent-copy evidence
motivates examining reuse; it does not establish that asynchronous reuse is safe.

## Source boundary

The locked wgpu-core 29.0.3 source already provides a useful completion boundary:

1. Queue::write_texture validates input before creating staging. Flushing removes
   CPU mapping; PendingWrites::consume retains FlushedStagingBuffer in TempResource.
2. PendingWrites::pre_submit transfers these resources into EncoderInFlight.
   Queue::submit calls the backend submit, advances last_successful_submission_index
   only on success, then registers active executions with LifetimeTracker.
3. Device::maintain reads the real backend fence value, refuses errors, and passes
   the observed completed index to Queue::maintain/LifetimeTracker::triage_submissions.
   Only active submissions at or below that index are drained. Their encoders and
   temporary resources are then destroyed. Queue shutdown also waits for successful
   submissions; device loss invalidates Device. Replay has a separate wait path and
   is not enabled in the retained native build.
4. FlushedStagingBuffer::drop destroys raw staging. It is also reachable through
   failure/unsent cleanup, so **generic Drop is not proof of GPU completion**.

Proposed reuse must therefore be an explicit action in completed-submission
retirement, after retiring the encoder, not a destructor replacement. Other drop,
failed-submit, unsent, canceled, device-loss and shutdown paths keep original
backend destruction. The audit binds exact source files, hashes and call sites.

## Proposed private prototype

- Base it on the private CPU upload observer with original write-combined cache
  policy fixed in both modes. Preserve all originals and the frozen SDK.
- Restrict eligibility to internal Queue::write_texture staging on the tested
  Metal/unified-memory path. Keep public create_staging_buffer and write_buffer
  allocations fresh. Do not expose cached contents to a caller-mapped buffer.
- Keep each checked-out raw Box<DynBuffer> under existing unique staging ownership.
  Track stable scalar resource IDs and submission indices. A successful submission
  binds its staging lease to that index; failed/unsent work cannot enter the cache.
- In triage_submissions, retain temporary resources while retiring/resetting their
  encoder, then explicitly admit eligible staging only with its matching submitted
  index, an observed fence at least that index and a still-valid matching Device.
  Never infer completion from time elapsed, CPU return, presentation or callbacks
  unrelated to that resource's queue submission.
- Cache exact-size, unmapped raw buffers per Device, with at most eight entries
  and 32 MiB of idle storage. No Arc<Device> belongs inside idle entries, avoiding
  a Device→cache→Device ownership cycle. In-flight resources remain exclusively in
  the existing submission tracker; these limits bound idle storage, not total GPU
  allocation or all outstanding uploads.
- Checkout removes the sole cache owner, remaps after proven completion and keeps
  original validation, coherent storage, transition and copy layout behavior.
  Oversized/different-size requests allocate fresh. Mapping/device errors and all
  evictions use original backend destruction outside the cache mutex. Drain cached
  handles through that backend before Device's raw backend/fence is destroyed.
- Audit cache-lock ordering against pending-writes, life-tracker, fence and device
  teardown. No operation while holding the cache mutex may acquire those outer
  locks or call arbitrary backend destruction. Preserve all existing initialization
  and error paths; padding/stale byte disclosure requires explicit review.
- Record a fixed, non-evicting scalar event journal: creation/checkout, submission,
  observed completion, cache admission/eviction and destruction. Refuse overflow,
  duplicate checkout, resource/device/index mismatches, admission before the real
  fence or exceeded caps. Observer state holds no native handles.

policy.rs models unique CPU-owned, submitted and completed values, exact-size
checkout, completion admission, FIFO eviction and caps. Submitted cannot be passed
directly to Pool::insert. It has no mutable-resource accessor. These types cannot
authenticate a supplied integer as a GPU fence, prevent a caller from dropping a
value, enforce thread synchronization, or replace the native lifetime tracker.
The backend-destruction escape hatch must never be used to manufacture a completed
lease. Real Box<DynBuffer> integration and source/event/readback proof remain required.

## Checkpoints and stop conditions

1. Review/commit this policy/source-audit checkpoint with focused Rust tests and
   compile-fail controls; no native resources or windows are created.
2. Implement only a private off/on prototype. Test mapping failures, disabled and
   oversized cases, failed/unsent work, device teardown/drain, exact-size changes,
   ownership/cycle/lock rules and event accounting. Review with Prism Anthropic
   Claude Sonnet 5.5 and commit before executing native code.
3. Use existing headless NoDisplayGPU/Texture/ReadData APIs for real WebGPU byte
   readback. Verify different patterns, non-aligned rows, multiple submissions,
   sizes and devices, then explicit release/drain. Separate simulated pending-state
   controls from actual observed GPU overlap; do not invent a physical fault test.
   Predeclare a small case matrix/watchdog and stop at the first invalid case.
4. Only after correctness/ownership proof, run one bounded off/on framework pair
   using identical binary, original cache policy and existing foreground/source/
   actual-presentation accounting. No unchanged retries or broad campaign. Report
   observer and ordinary-build limits; no automatic shipping/gate adoption.

If source integration cannot preserve explicit completion and safe teardown, stop
that implementation and retain the exact blocker. Do not loosen ownership, ignore
failed readback or modify the release gate to obtain a passing result. Features
remain frozen and experimental rendering/pacing remains disabled. M3–M7 stay open.
