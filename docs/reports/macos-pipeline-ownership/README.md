# Mac pipeline replacement ownership

Shader reload previously replaced its module pointer without releasing the old
owned reference. Bind-group replacement queued the current group for retirement
before creating its replacement; refusal left that group both cached and queued.
Repeated refusal could accumulate aliases and later invalidate a cached binding.

Patch 52 admits a fresh live candidate before publication. Shader refusal preserves
the current module; success explicitly retires the prior one. Existing native
pipelines retain their dependencies and require rebuild to use new shader code.
Binding refusal keeps the current reference, generation and retirement queue
unchanged. Successful replacement queues the old group once; matching live
bindings reuse it. Pipeline retirement detaches both cache maps. Factories return
fresh owned references and one owner serializes use and configuration.

## Verification and scope

Final GPU/gpudraw tests pass three race repetitions. The full offscreen adapter
race suite, vet, eleven preparation tests and native window-cycle consumer
compilation pass. The consumer was compiled only. Controlled kernels cover
partial/error/nil/invalid candidates, repeated refusal/retry, invalid-cache
replacement, matching-generation reuse and 1,000 shader reference replacements.

Two tiny actual headless cases retire sixteen prior shader modules and the final
module, preserve the current module after invalid WGSL, and refuse three duplicate
layout bindings before successful replacement/reuse and four exact shader result
words. Old and current bind groups retire at pass completion and pipeline shutdown.
The refusal is native layout creation within VarGroup.bindGroup, not an
incompatible-buffer TryCreateBindGroup failure. No windows, surfaces, input,
settings changes, presentation or performance measurement were used.

The unchanged parent prepared dependency with the final test-only overlay fails
both cases: sixteen retained-module assertions and one premature-retirement-queue
assertion. Explicit fixture cleanup retires unique detached references and avoids
releasing current aliases twice. This is a correctness comparison, not a native
heap or physical-device-fault qualification.

Fifteen command receipts preserve twelve passes and three raw exit-1 attempts.
The initial invalid-buffer fixture was automatically repaired and failed its
refusal oracle; its source and log are retained. The corrected duplicate-layout
fixture passes. A nonexistent native consumer path failed compilation; the actual
window-cycle-smoke path compiles. The original regression is the third raw failure.
The first corrected native build passed while Git staging crossed its two-part
source inventory capture, omitting four newly staged paths in that capture.
Its unstable inventory is preserved and excluded from final qualification; the
final repeat compiles with a stable complete source inventory.
An initial registration guard incorrectly compared forty stages with fifty-one
patches and stopped before manifest mutation; its producer is preserved. The
separate preparation before registration prepared the unchanged 6,195 files.
Final fresh preparation verifies 6,198 files and the complete 1–52 patch sequence.

## Reviews and evidence

Internal review identified the faulty initial native oracle and incomplete
fixture cleanup, then approved the corrected candidate and independently checked
owned patch replay, both static pin maps, prepared outputs and specs. Prism ran
once using Anthropic Claude Sonnet 5.5: complete eight-file coverage,
1 chunks, 12701 input/1731 output tokens,
raw exit 0 and no findings. A separate internal evidence review binds this report,
raw receipts, source variants and sealed archive before commit.

The [checkpoint](records/checkpoint.json) binds reviews and the
[archive](records/evidence.tar.gz) with its [member index](records/evidence-index.json).
All 6,198 prepared endpoint identities are recorded; this is not a full compiler,
SDK or static-native-component closure. Native/GPU heap reclamation, actual live
window endurance, sustained performance and full Mac qualification remain open.
Desktop testing remains deferred while the user uses the computer.
