# Final integrated SDK checks

Captured from the canonical repository after the phase-four implementation and
final backend ownership fixes. Fresh driver-owned preparation, full repository
checks, native driver race tests, and an application generated outside the repository
at a path containing spaces all pass. That application uses GOWORK=off and explicit
local replacements; it still depends on this checkout. The seventeen historical
patches are byte-identical to their recorded sources; the provenance JSON records
that comparison. Source and backend patch hashes identify the checked implementation.

Three native processes each verified a GPU rectangle and measured ten synchronous
render/readback durations, resident memory and quiet CPU. Application-entry-to-GPU
startup ranged 290.3–310.8 ms.
Readback min/median/max was 1.732/2.047/2.470 ms.
RSS ranged 253.8–255.7 MiB;
quiet CPU samples ranged 2.21–3.30% of one core.
Host: macOS-26.6.2-arm64-arm-64bit-Mach-O. Image sizes: 3054x722.

These measurements use the boundaries in [the SDK contract](../../../../specs/sdk.md);
readback is not compositor presentation latency. Short ps CPU samples have limited
precision. No production performance guarantee or sub-1% idle CPU acceptance is
claimed. Public SDK distribution, Windows/Linux runtime and full native accessibility
remain release gates. No OS input sources or preferences were changed.
