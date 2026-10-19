# Routed input measurement checkpoint

`NewMeasuredDesktopSession`, `MeasureInput` and `Measurements` add bounded,
opt-in input evidence to the existing public native session. Original event
identities, owner drains and actual reconciled view generations determine the
response. An owner callback independently checks accepted app state; its model
revision stays separate from backend generation and native presentation scope.
The callback runs at most once and is skipped after known earlier failure.

Core stage 98 keeps intermediate frames unarmed and their acquisition identities
immutable. It records intermediate owner, reconciliation and rendering work,
including Manual's separate between-event owner drain. Only the complete input/
owner/reconciliation terminal can arm the final checked generation. Cancellation,
retirement, stale capability, no-op and capacity behavior retain bounded records
and clear callbacks/references. Ordinary Input/Frame behavior is preserved.

Independent source review and focused race tests preceded Prism. Prism's
premature owner-terminal claim was refuted; review found a related real omission
of intermediate CPU stages. That omission was corrected with exact stage-total,
failed-prefix, actual two-event FIFO and unarmed-observer regressions. Root
reviewed the correction; its focused driver/Core race, vet, native compile-only,
full preparation reproduction and 29 preparation checks pass. Prism reviewed the
full 13-file baseline and 8-file correction with Anthropic Claude Sonnet 5.5; final
correction review has no findings. The certificate verifies that both reviewed
patches reconstruct the final source bytes exactly.

This checkpoint does not qualify actual display performance. The native test
binary was compiled only; software routing cannot yield native composition
success. Real presentation must join tag/generation/scope through the separate
native recorder. Representative workspace/pacing consumers, native text/AX,
resize, affected secondary windows and five-run 60 Hz qualification remain open.
4 ms remains an optimization target. No M3 or full Mac release gate closes here.
