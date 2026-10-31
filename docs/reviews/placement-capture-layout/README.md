# Palette capture layout correction

Internal review of a user-reported v1.1.0 placement demo defect. The full managed
consumer regression reproduced mouse-down reflow: starting a palette gesture
changed command availability, and native disabled styling removed reserved border
or shadow space. Toolbar heights shrank and the canvas moved by 18–24 logical
pixels in the recorded 160-DPI fixture. Simple isolated button presses did not
reproduce the application-wide change.

The placement consumer now retains only node/option disabled presentation from
its last idle view during a palette capture. It copies no callbacks, payloads or
entire UI descriptors. Live callback checks and command-registry admission remain
unchanged. Release or Escape renders current idle availability again. The change
is confined to the placement example; framework APIs, native dependency patches,
rendering policy and historical releases are unchanged.

Review checked capture lifetime, missing/new controls, option identity and live
execution refusal. Snapshot maps are replaced at each idle render; new controls
without prior presentation retain their live disabled value. The ordinary
consumer's routed regression now measures all existing elements on palette
mouse-down, short release and Escape; it also verifies that an attempted command
is refused during capture and enabled again afterward. Existing valid placements,
Undo/Redo, selection and the remaining consumer checks still pass at owner batches
1, 7 and 64.

Validation: the complete examples/placement race suite with offscreen routing
passed (22.089 s), as did offscreen vet. The final expanded routed regression
passed with race detection (17.032 s). Compiler warnings about deprecated native
accessibility methods and duplicate libobjc remain unrelated. Native demo refresh
is a separate follow-up observation; these tests do not certify all physical
pointer, accessibility, hardware or display behavior.
