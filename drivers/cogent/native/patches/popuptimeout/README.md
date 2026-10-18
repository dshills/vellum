# Owned popup timeouts (stage 82)

Timed native popup destruction acquires the render context captured at arming
before reading live Stage/Scene fields. Each arming has a fresh ticket. Rearm,
disabling and Stage deletion revoke the ticket before nonblocking timer Stop;
a callback already waiting for the owner becomes a no-op. Manual execution keeps
its existing owned FIFO callback and does not recursively acquire the owner.
Successful timeout removal requests one on-demand paint. Native event-loop exit
also acquires the owner before destructive stage cleanup and defers its wait-group
decrement so normal and panicking cleanup release the wait after the owner.

Five new offscreen methods exercise native time.AfterFunc callback locking,
exactly-once repaint, popup retirement, the actual event-loop exit path, rearming,
disabling and manual FIFO expiry/cleanup. This is software scheduling/race evidence,
not physical input, macOS presentation or complete native lifecycle acceptance.
Input/output hashes and the patch digest bind independently prepared source.
