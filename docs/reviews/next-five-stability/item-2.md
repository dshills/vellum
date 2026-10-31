# Item 2: disabled control geometry

The adapter now retains enabled maximum-decoration allowance for disabled managed
buttons/menu buttons, checkboxes and single-line text fields. Button elevation
space and checkbox/field focus-border space previously disappeared on disable,
shrinking the controls and moving unrelated content. The regression failed before
the correction at 96, 160, 162 and 192 DPI with default and Mac styling.

The final styler preserves only layout allowance; it does not restore disabled
focusability, callbacks, active shadows or painted focus borders. Managed explicit
borders and flat Mac buttons keep their existing presentation. The native backend,
dependency patches, stable API signatures and rendering/pacing policy are unchanged.

Removed the placement example's last-idle availability snapshot workaround. Its
original routed bounds regression now checks the shared adapter correction while
controls really become disabled during palette capture. Live command refusal,
release/Escape, valid placement, selection and Undo/Redo remain covered.

Validation: the four-DPI default/Mac availability regression checks button, menu,
checkbox, field and sibling bounds across repeated disable/enable transitions.
It also verifies disabled state/opacity and refuses disabled button activation.
Focused button, Mac focus and app-surface race checks passed (3.752 s); adapter
vet passed. The complete placement race suite passed after workaround removal.
The final menu-inclusive availability regression has its own passing run. Compiler
AppKit deprecation and duplicate-libobjc warnings remain unrelated.

This is software-routing/layout evidence. Item 1 qualifies the earlier example
fix with an operator; it is not relabeled as native qualification of this changed
adapter. A matching SDK/app build and separate runtime checks follow in item 4.
Default Prism review and dispositions accompany this checkpoint.

Default Prism (Wafer / GLM-5.3) completed with full coverage and zero findings.
The raw receipt retains its declared exclusion of the deleted workaround file;
its removal is visible in the checkpoint diff and remaining integration checks.
