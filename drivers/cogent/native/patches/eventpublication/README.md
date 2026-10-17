# Key-event publication ownership

Stage 0077 ends Source.Key access to the event after Deque.Send transfers it to
the consumer. Chord decisions use the original code/modifier arguments. Calling
a Base value-receiver method on the published event copies mutable fields and
races with the render consumer's legitimate SetLocalOff/SetHandled mutations.
The original key/chord order, mapped keypad runes, NumLock, modifier behavior,
producer state and injected event clock remain unchanged. No lock, delayed
delivery, event clone, polling timer or broader source-serialization claim is added.

The producer/consumer regression acknowledges each of 128 ordinary keys and joins
its consumer. It exercises real deque ownership with bounded in-flight work.
A separate twelve-case matrix checks key/chord identity, order, modifiers/runes,
producer state and injected timestamps. Run the regression with the Go race
detector. An old-source counterfactual must expose the prior race; ordinary
success alone does not show that a test distinguishes the defect.

The native imaging exit-66 failure that exposed this race remains failed. New
source, race regressions and a fresh native result need independent review;
physical input, full IME/accessibility behavior and complete Mac qualification
are separate. Preparation binds exact input, patch and output hashes. Added
test code and owned edits are MIT; original upstream notices remain in place.
