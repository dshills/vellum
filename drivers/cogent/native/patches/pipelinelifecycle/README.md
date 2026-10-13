# Pipeline replacement ownership

Patch 0052 publishes only fresh live shader modules and bind groups. Factory
refusal preserves the current reference; returned partial handles retire. Shader
reload retires the prior module explicitly after replacement. Existing native
pipelines retain their dependencies and require rebuild to use the new shader.

Binding replacement queues the old group only after successful admission. Refusal
preserves current ownership, generation and retirement queue, allowing retry
without duplicate aliases. Matching live generations reuse the binding. Pass
completion/shutdown retires prior groups; pipeline retirement detaches both cache
maps. All operations use one serialized owner and fresh-owned native factories.

Fault kernels cover partial/nil/invalid/error admission and retry; opt-in tiny
headless checks observe module replacement and duplicate-layout native refusal
followed by exact compute output. They do not qualify physical GPU faults, native
heap reclamation, presentation, performance or full Mac acceptance. New helpers
and tests are MIT; modified upstream files retain their BSD notices. Every owned
stage input, output and patch is pinned; module-cache sources are unchanged.
