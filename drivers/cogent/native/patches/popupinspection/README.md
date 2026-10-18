# Bounded actual popup inspection (stage 83)

`Scene.InspectPopups` requires the caller's existing render ownership. It takes
at most 64 current popup stage pointers under the stack lock, releases that lock,
and copies completed clipped scene/button bounds. Only the top event-accepting
popup's visible, enabled direct standard buttons are offered as targets.
Background-main, pending/unshown/layout/render, invalid budget and exhausted ID
states refuse safely. Custom/nested menus keep their existing inspectors.

Each real scene/button receives an opaque monotonic incarnation ID in its own
metadata. Owner checks reject copied metadata; saturation refuses rather than
wrapping. There is no historical widget registry. A detached label clones only
its admitted bytes, so a small substring cannot retain a large backing string.
Every visited record is charged before inspection and each label is capped.
The public DesktopSession separately charges generated output identifiers and
fixed records against its remaining shared semantic budget.

Core regressions use real Chooser popup scenes and the software FIFO: transformed
window bounds, visibility, actual click, exactly-once change, fresh reincarnation,
modal/tooltip precedence, budgets, pending geometry, metadata copies, saturation
and independent bounded label backing. Driver tests add same-frame stale-popup
refusal and shared public-output budget controls. Software evidence is distinct
from native execution and physical input. See the native-controls report.
