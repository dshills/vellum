# Mac desktop widgets — personal-use expansion

Owner approved the next five widgets: editable combo box, context menu, anchored
popover, multi-selection source list and date/time picker. This supersedes the
older internal-app API freeze in CONTRIBUTING.md for these five capabilities only.
Rendering, pacing and released alpha.3 artifacts remain unchanged. Each checkpoint
gets focused tests, Prism Anthropic claude-sonnet-5-5 review and a commit.

## 1. Editable combo box

ui.ComboBox composes the retained TextField and managed MenuButton; text uses the
existing native editing bridge. Typed values and suggestions propose ComboEvent
through one owner callback, never auto-add options. Stable option IDs are separate
from labels. Current disabled/removed options and changed component IDs refuse
old proposals. Changed labels also revoke callbacks from the older rendered option. At most 64 suggestions, with no background work. Empty options
leave a text-only field. Mac appearance uses one compact outline around the row.
Tab reaches the suggestions affordance; Down/Return opens the ordinary managed
menu, arrows choose and Escape dismisses. Down in the text entry itself does not
open suggestions in this initial composite. No live filtering or completion.

All controls are custom-rendered, not AppKit wrappers. Fixed MacTheme colors and
Helvetica Neue are retained; automatic system appearance/accent tracking and
pixel-identical AppKit claims are outside scope.
