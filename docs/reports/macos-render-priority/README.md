# One render-worker priority candidate

This bounded session evaluates one actual implementation candidate under the
unchanged release gate. It adds a temporary USER_INTERACTIVE QoS override around
the common composition owner, pins the Go goroutine to its OS thread, and ends
the native override through deferred cleanup. Requested base QoS is unchanged;
no process-wide priority, OS preference, rendering or pacing option is changed.
The candidate is private and has not been adopted or run as a native application.

Two focused native Go race tests pass: 1000 successful scope retirements plus
panic unwinding, and eight concurrent workers with25scopes each. Native end calls
return success, each owner stays on its thread, repeated cleanup is safe and
requested base QoS remains identical. These are scoped native API/ownership
checks, not whole-framework resource/endurance qualification. The test thread
already requested class33/USER_INTERACTIVE before override. A priority deficit
in actual composition workers remains unproven; this is a low-confidence
candidate, not a demonstrated scheduling repair or a causal speedup claim.

The private Core differs in one common measured/ordinary composition owner plus
two platform helper files. Initial internal review found the first hook covered
only ordinary composition; the corrected common seam covers both trace branches.
The initial unrun build remains recorded. A subsequent build admission refused
before child creation because private pin keys mixed relative/absolute paths;
fixed, and the final optimized native build passes. The actual unauthorized CLI
refuses before native launch and retires its owned process. The native helper
source is unchanged from the race tests; no broad suite is repeated.

Reuse the frozen SDK baseline bd3cbc9, all8321files unchanged, original native
archive786aaad8 and the existing representative loadedgrid fixture/public strict
validator/owned executor/native lease. Candidate binaryf2097af2… is externally
bound to its exact private Core/module/source hashes; it is not the untouched SDK
binary. No profiler, GPU-scene or experimental pacing flag is enabled.

The [protocol](../../../tools/macos-render-priority/PLAN.md) permits one serial
CPU grid case with10000stored records,128warmups/120measured inputs. Invalid
accounting or p95>33.3ms stops the scope. Only a passing serial case admits paced
CPU grid128warmups/600states under unchanged cadence/latency checks. Limits are
60seconds/native process and150seconds/whole scope, additionally bounded by the
one-session deadline19:05:33UTC on22October. Preserve all samples and failures;
seal and review the actual result. The existing runner owns child retirement and
all geometry/configuration/SDK/hash checks. No new calibration harness is built.

Current human authority covers remaining release screen tests. Internal review
and Prism using Anthropic Claude Sonnet5.5 precede checkpoint commit and native
execution. No release requirement closes from preparation or this two-case check.
The full representative/five-run, resource/endurance, human/physical, final SDK/
rights and final audit requirements remain. Features and60Hz/33.3ms gates stay
unchanged. See the [checks](preparation-checks.json), [Core patch](candidate.patch)
and [indexed evidence](evidence-index.json). Absolute paths record local provenance;
member hashes/committed copies can be checked from the archive, while full source
revalidation requires the external pinned workspace. No public push is requested.
