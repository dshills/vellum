# Opt-in text allocation compatibility

Stage97 corrects stage95's effect on native text whose paragraph alignment was
already inherited from a control. Allocation layout is explicit per Text and is
false by default. The driver enables it for resolved Label TextAlign/GrowX/GrowY
and resets it when those properties disappear; internal control labels keep the
legacy origin and sizing. The old scalar Grow alone remains compatible.

Explicit layout retains shared draw/selection/link origin and final growth while
measurement wrapping stays enabled. Core tests cover default center/end origin,
selection and links alongside explicit alignment/multiline/growth. Driver tests
cover actual glyph pixels and retained updates/removal. Unmodified frozen
workspace records and screenshot goldens are the integration compatibility gate.

New changes/tests are MIT. Existing upstream notices remain intact. Preparation
checks input/output hashes and the patch digest. No OS preference changes or
physical pointer, performance or full Mac acceptance claim.
