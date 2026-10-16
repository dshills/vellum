# Mac event-driven queue checkpoint

The Mac desktop adapter now waits for events without the recurring 100 ms native
wait deadline. A single bounded ingress publishes each accepted queue packet
before posting its wake. The native thread retains callback execution and output
ownership. Stop retires pending packets and joins ingress before native cleanup;
Windows and Linux retain their finite wait. See the [dispatch contract](../../../specs/native-dispatch.md).

This removes a source-observed recurring deadline. Its contribution to measured
CPU has not been attributed, and this checkpoint claims no performance gain or
idle-budget acceptance. The separately running overnight fixture uses its frozen
pre-change executable; its observations cannot qualify this implementation.

Owned patch 0069 pins six input/output leaves, including two new files. Four
existing prepared leaves change and two are added; the regenerated dependency
vector has 6,229 files. A pre-existing unkeyed Mac MIME-data literal now names its
existing fields, resolving the actual vet failure without changing values.
The preparation test expects the complete ordered patch sequence 1 through 69.

Final checks passed on the regenerated six-leaf endpoint:

- Fifty repetitions of the selected base/desktop queue, failure, event-barrier and
  new ingress race regressions. The new cases cover lost wakes, publication before
  wait, FIFO, bounded pending ownership, active completion, concurrent stop,
  callback panic/nil panic/Goexit, wait/post failure and stop-before-start.
- Static vet for the base and desktop packages and all 11 preparation unit tests.
- An optimized `mainqueue-smoke` build, with its complete prepared vector and
  binary digest recorded. The executable has not been run at this checkpoint.

A separate overlay control deliberately omitted the ingress join. It failed the
intended shutdown-order assertion, demonstrating that this regression catches
premature loop return. The first vet failure and earlier source endpoints are
retained alongside the final passing results. Inert checks were bounded and
scheduled between recorded native quiet intervals; their receipts do not certify
global machine inactivity. The first two check-producer versions were recovered
by reversing known edits and matching the original recorded SHA-256 exactly;
the evidence labels them as verified reconstructions.

The native probe has independent `normal`, `panic`, `idle` and `stop` cases. The
last two omit the deliberate event flood and test delayed dispatch or worker-only
stop. They still perform ignored late-post probes after native termination.
Run them sequentially after the ongoing native campaign has retired. Native
startup, event delivery, shutdown, windowed interactions, idle CPU, app-wide
wakeups/presents and full Mac acceptance remain open for this new implementation.
No input source, VoiceOver, clipboard or OS preference changes are required.

Internal source review precedes the single requested Prism checkpoint.
Checkpoint acceptance requires retaining raw code checks, producers, source
associations, reviews and their qualification record in `records/`.
Executable outputs remain private.
