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

## Multiple-selection list

- 52f538d1681befb6: clarified contract. Focus/range anchor are input navigation
  state, independent of accepted selection; rejecting a value does not reject
  native focus movement. This is intentional and specified. Removed/disabled
  anchor rows now clear the anchor on reconciliation; future ranges start at
  the focused row. No acceptance acknowledgment can be inferred from a void
  controlled callback.
- 015d8dfd381a5af1: fixed the fragile name slicing by capturing the stable option
  ID in each row styler. Styling follows existing originating render/owner locks.
- 5a1609a957e76a25: driver ranges already contain unique validated option IDs;
  hardened public Apply to deduplicate manually constructed edit IDs as well,
  with a regression for a repeated toggle ID.

Final modifier follow-up: native modifiers use bit indices with HasFlag or
1<<index masks. Actual Command/Shift routing now passes; corrected Shift-F10
and existing Shift-Tab comparisons and fixtures. Follow-up findings
48b1761424eb2694 / c2d58101b20e0063 do not reproduce: every admitted range includes
the enabled target enforced by the row guard, so a disabled-only empty range
cannot be generated. Public Apply with empty replacement IDs intentionally clears
selection; an app rejects by ignoring the edit, not by calling Apply. Both
representations of modifiers are the backend's documented bit-flag API.

## Date/time picker

- 656d06494c764e10: opening already resets the displayed month, and normal managed
  activation invalidates the app. Added reset on Dismiss and day selection too,
  bounding navigation state to the open calendar lifetime.
- 0ab5e2adc496f842: fixed. Parse civil fields in UTC then construct/compare in the
  accepted location, catching DST gaps without rejecting parser-supported month
  case/padding variants. Layouts with zone tokens are refused: this picker edits
  local fields and does not select time zones. Added variant/stale-layout tests.
- The discarded location-pointer suggestion would admit old callbacks after a
  real zone-rule change. Keeping identity deliberately revokes that older view;
  newly rendered callbacks accept the current location.
- Focused ui and actual offscreen date/popup/input checks pass. Calendar selection
  commits once, invalid text restores accepted value, Escape changes nothing and
  disabling/re-enabling cancels rather than reopens a calendar.

Date follow-ups:

- 8e4e93a58ba666b1: fixed with civil sentinel round-trip validation for each mode,
  rejecting layouts missing required date/hour/minute fields or mixing date and
  time in single-mode inputs. Seconds remain optional precision.
- 656d06494c764e10 follow-up: actual anchor down/up test exposed missing outside
  dismissal in the backend main-stage Dialog policy. Popovers now chain the
  originating scene EventBoundary, consume outside MouseUp and close through
  the existing dialog lifecycle. The routed test proves no reopen/change.
- Final concrete-fix review returned no retained findings (complete coverage).
  Earlier raw findings and the discarded location-identity suggestion remain
  recorded. No claim that all intermediate reviews were zero-finding.

## Internal review and packaging checks

Reviewed lifetime guards, current-handler delivery, bounded option/cell counts,
accepted-state restoration, modifier mask representation, coordinate units and
originating native context for teardown. No rendering/pacing changes or public
release publication. Focused ui race/vet and affected native driver/gallery
race/vet pass; native executable builds. Private light/dark packages retain
standard OS window chrome. Live appearance observations are recorded separately
after packaging.

## Packaged native follow-up

Live light-window inspection showed a clipped duplicate combo chevron. Icon-only
MenuButtons now omit the redundant indicator; labeled buttons retain/restore it.
Live keyboard/AX activation selected Inches and changed the field correctly.

The same native check exposed an invalid full AX snapshot: multi-list host Value
used []string, but the existing Objective-C validator admits scalar strings or
numbers only. Host Value now summarizes the selection count as a string; row
Selected flags retain the authoritative membership. A focused Darwin/offscreen
fixture runs actual produced nodes through the production native validator for
empty, single and multiple selections, rather than only inspecting Go snapshots.

Native follow-up Prism f1466a0a3b586dec: changed the scalar count to neutral
“1 selected”/“2 selected” phrasing. Existing native payload and menu/selection
routing tests pass. This is fixed English wording, not OS localization.

Live inspector follow-up: generic native Dialog defaults added a move handle,
24dp scene padding, a 50px sizing allowance and a narrow wrapping body. The
popover-only surface now has 240dp default content width (caller style wins),
zero outer scene/body padding, no move parts or title side effect, and scene
size clamped to the measured retained host. Existing Dialog behavior is unchanged.
Menu surfaces follow the application-local form surface in dark appearance.

Popup appearance review dispositions:

- a671d0650f6a5833: added explicit dialog/host guard before measuring the retained
  host. Host is normally created at mount, but defensive early return is harmless.
- 5fac631b1494c058: removed temporary geometry diagnostic logging from the test.
- User screenshot additionally showed outlined unequal rows offset at the arrow.
  Popup rows no longer inherit push-button fills/borders; they fill one shared
  surface. Combo MenuAnchorID names its outer field; native menu origin includes
  scene offset and matches the full field width. Tests assert exact anchor, width,
  equal rows and zero border.
- Calendar geometry regression exposed zero child hit boxes when scene/body
  shrink-wrapped after resize. Scene/body now fill the explicitly measured host
  allocation, complete originating layout, and retain compact host growth. The
  full popup/date/menu/context/combo race suite and portable ui race tests pass.

Painted-edge correction: native shadows reserve space outside the popup surface.
Menu origin/width now include those margins so the visible surface aligns with
the full combo field, rather than aligning only the shadow canvas. Popover size
includes scene/body outer space and a regression refuses spurious scrollbars.
The precise space/coordinate follow-up has a complete zero-finding Prism result.
Source list row growth is applied after text style inheritance so labels of
different widths still fill one consistent list width.

Final popup/list geometry checkpoint: complete Prism review with no findings;
affected multi-list, popover, date, combo, menu and context race tests pass.
