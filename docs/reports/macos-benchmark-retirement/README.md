# Successful benchmark process retirement

The grid benchmark formerly retired its process group only on failure or while
the direct child remained alive. A successful producer could close its output,
exit zero and leave a same-group descendant alive. The case still reported PASS.
An injected retirement failure on that path was likewise never observed.

The Mac/POSIX runner now retires every launched case's owned group after its
measured work. Cleanup errors reject the case while preserving the direct-child
status, raw records and exception type/message. Windows retains its existing
failure/live-child policy. No application, rendering or timing API changes.

An actual Python producer emits valid synthetic benchmark records and exits
zero while its descendant sleeps with closed inherited pipes. The original
runner leaves that descendant live; the corrected runner terminates it. The
fixture always retires its own group, including when testing the original bug.
An absent process or a terminated orphan awaiting OS reaping is accepted by this
regression; it does not establish an empty native process group or allocation
inventory for Vellum applications.

The original two regressions fail. The final benchmark's 42 CPU-only checks and
the unchanged bounded process-helper suite pass. One intermediate failure exposed
loss of the cleanup exception message; that diagnostic is now preserved.
Acceptance requires retaining exact sources, original/final raw results, reviews
and producer identities in [records](records/qualification.json).

No Vellum native application or GPU workload runs for this checkpoint. The
independent 1,000-window native campaign uses its own frozen source and controller.
No new native timing, retention, overnight, human/physical, distribution or full
Mac qualification follows from this process-cleanup correction.
