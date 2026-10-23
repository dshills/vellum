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
