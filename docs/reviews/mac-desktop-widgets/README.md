# Mac desktop widget checkpoints

All Prism reviews use Anthropic claude-sonnet-5-5, baseline none, complete coverage.
Raw results are retained; findings below are dispositions, not claims of zero findings.

## Combo box

- Prism f9b989b6b545776a: fixed by validating the current option against its rendered
  snapshot. Removal, disabling or label changes refuse the older callback. New
  views admit the updated option. Controlled text acceptance remains owner-owned.
- Focused ui race tests and actual offscreen suggestion selection/rejected text
  commit/draft and identity tests pass. Existing menu/field tests pass. Native
  visual inspection is deferred to the final packaged demo; no physical pointer
  claim is made from offscreen evidence.

## Context menus

- ccbbb9e90d2b8a65: hardened AX show-menu with explicit menu-state guard and clear
  pointer position before opening at the host.
- 9eaf29f3b0b17839: not reproduced. syncMenuButton clones current options, including
  nil on removal; changed options rotate epoch and menuAvailable closes absent
  menus. Removed unused source pointer. Pointer identity cannot be compared:
  ordinary views recreate equivalent descriptors on every refresh. Existing
  removal/disable tests pass without disabling the host.
- fd7fb5cc72f1f190: added list-row/host shared-event test proving handled input
  opens one popup. No duplicate row/host request or popup replacement.
- ui race and popup/menu/combo focused driver race tests pass. Context opening
  uses actual native widget events in software tests because the public Manual
  harness deliberately refuses OS-context gestures. Popup item clicks use actual
  manual routing. Physical right/Control-click verification remains separate.

## Popover

- d6402637ae16f709: fixed. Once the originating native scene is ready, a missing
  anchor cancels even a not-yet-active Open panel. Dismissed state bounds callback
  delivery. Added initially-open valid/missing anchor cases.
- Focused geometry, Escape, focus return, keyed content retention/reopen and
  anchor-removal tests pass. Edge placement uses clamping; small-window fixture
  originally clamped to the top, so below-anchor assertion now uses a larger
  window. This is a test setup correction, not a claim of edge reversal.
