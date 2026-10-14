# Interacting lifecycle fixture and failed endurance campaign

The composed lifecycle fixture now includes the reusable color picker, expanding
each fresh window from 20 to 24 native Cocoa selector steps. Red and whole-hex
proposals require exact owner callback counts and all five accepted native values,
positive field identities, visible geometry and the matching preview. The Colors
scroll page needs Grow to allocate its fields; the original native preflight and
an actual offscreen layout regression exposed the missing allocation. Corrected
model/layout race tests, native harness race tests, vet and build pass.

The corrected five-window preflight passes all 120 ordered selector steps in
49.47 seconds, with 18 tree AX rows and 1,000 contained labels per window. The
longer frozen campaign fails after 343 completed windows in 3,260.78 seconds.
The 344th attempted window, native-cycle-0343, completes its first 18 steps, then its tree-row press
reports native accessibility attachment closed after one attempt and admission.
Owner/native completion is false. The process ended before the human returned;
this result is a failure, not an interrupted or qualified 1,000-window campaign.
The triggering attachment-close path is not identified by these records.
The public error indicates Go accessibility bridge retirement during a query or
action request; lookup, malformed snapshot and stale-action refusal have different
outcomes. Existing offscreen retirement cancellation and bounded-request tests
pass under race, preserving failure instead of retrying a closed attachment.

Four successfully completed actions have two attempts and queue admissions each:
About tab selection in cycles 23, 208 and 303, and disclosure expansion in cycle
48. Queue admission does not imply model acceptance. The fixture joins each
admitted request through an owner FIFO acknowledgement, rejects duplicate accepted
callback counts and waits without reissuing accepted requests for native
publication. The original validator incorrectly required one admission for every
action. Its source and raw anomaly record are retained; a versioned correction
requires positive integer Admissions <= Attempts and reports every retry.
The disposition of each earlier unaccepted admission remains unobserved. This
correction does not change the independent terminal failure.
The failed window's first check/set/false step also completes with two attempts
and admissions. Its trailing retry is listed separately: four retries belong to
fully completed windows and one to the incomplete window, five in total.

Fourteen forced-GC samples cover 0 through 325 completed windows, retaining the
512 MiB Go heap and 1 GiB RSS guards. Samples from 100 onward show heap
44,114,344–50,446,240 bytes and RSS 382,763,008–392,069,120 bytes. These are
partial diagnostic observations, not a retention qualification. The final heap
profile and early owned vmmap snapshot are preserved; the late snapshot was not
reached. Native teardown reports main-thread cleanup, callbacks stopped, zero
windows, AX registrations, native operations and pending callbacks.

The source/prepared endpoint matches the launch baseline at terminal, including
all 6,204 prepared dependency files. Subsequent harness unit tests explicitly
cover retry after a drained unaccepted admission and refusal to qualify or retry
after inspection failure. These post-campaign test additions have a separate
source binding. No framework production close behavior is changed based on an
unproven cause. No preferences or clipboard were touched. Live testing is paused
while the human uses the machine.

Committing this scoped source, spec and report checkpoint requires internal review
and Prism Anthropic Claude Sonnet 5.5 review. The accompanying evidence archive
must retain original failed layout, strict validator assumptions, versioned
producers, tests, preflight, failed long run, heap profile, early snapshot and
review records. The final [review manifest](records/qualification.json) and
[evidence index](records/evidence-index.json) record issued reviews and preserved
members. Full P6/P11, native/GPU
allocation budgets, physical input/displays, human VoiceOver, overnight
background and distribution acceptance remain open. Vellum remains an
experimental Mac alpha.
