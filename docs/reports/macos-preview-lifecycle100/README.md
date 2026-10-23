# Quiet native lifecycle and alpha.3 functional checks

The existing fixture passes 100 composed same-name cycles in 1,051.84 seconds,
with 99 preceding-window replacement checks. This is the outer monotonic
elapsed time through process exit in run.json; terminal.json's final in-process
sample is slightly earlier, at 1,051.671 seconds. Cycle indices 1–100 are present
exactly once. Every composed action reaches owner/native completion. Some modal
close actions require two admissions, retained in summary.json; this is not an
exactly-one-admission claim. These are automated Cocoa/semantic routing checks,
not physical pointer or human accessibility qualification.

The fixture was built from SDK source d8de4fc. Retained manifest comparison against
the alpha.3 candidate proves driver, prepared sources, lifecycle fixture and
Placement product source hashes unchanged. The only differing Go files are the
root version constant and SDK contract test expectation. This applies to unchanged
lifecycle code, not fresh alpha.3 application acceptance. Offscreen builds ran
during part of the owner-confirmed quiet desktop interval; this is not a system-idle
or performance benchmark.

Tracked terminal state: no Vellum window titles, accessibility windows, native
outstanding work or pending callbacks; callbacks stopped and normal process exit.
Warmed Go heap stays about 69.7–69.9 MB with 6–8 goroutines. RSS reaches about
460.7 MB; full samples and small late drift remain visible. Samples do not prove
native/GPU allocation ownership or an eventual global memory bound. Full P11,
native heap and GPU memory qualification remain false.

One invisible, untitled NSKVONotifying_TUINSWindow with NSView remains in the
terminal inventory. The earlier failed 83-cycle run and aborted soak each also
ended with one; no multiplicity accumulation is observed in these terminal
snapshots. Ownership remains unclassified; safe ownership or deallocation is not
claimed. A separate metadata-only process did not load that class and supplied no
ownership answer. This remains an R3 resource limitation to resolve or report.

The earlier breadcrumb failure stays FAILED in its original journal. The owner
confirmed other desktop activity then, making it inconclusive as a controlled
reproduction. The quiet run does not reproduce it, but cannot prove the cause or
exclude a rare bug. A separate sandboxed startup emitted no fixture events because
native services were unavailable and was terminated. The earlier 152-second failed
soak remains separate, with unknown-actor native close requests; no two-hour pass.

After lifecycle completion, the alpha.3 independent starter launches, verifies GPU
readback and exits normally in 2.28 seconds. The existing alpha.3 gallery native
text/editor/dialog script passes in 5.24 seconds. Its retained result describes
Unicode/edit/history/composition cases and session cleanup. Routed automation is
separate from real IME/VoiceOver/physical-input acceptance. No input-source or
accessibility preferences changed.

The compressed lifecycle journal reproduces the original raw bytes, bound by
SHA-256 in summary.json. The full text fixture stdout is likewise preserved as a
hash-bound gzip; text-native-summary.json retains result, step labels/count and
session cleanup, keeping repeated detailed snapshots outside the text review. JSON receipts normalize user-home paths. The real-repository public
export refusal confirms the repaired tooling still blocks pending audit approval
and creates no public output. Remaining packaged-app/activity/soak acceptance,
rights decisions and publication authorization are not closed here.
