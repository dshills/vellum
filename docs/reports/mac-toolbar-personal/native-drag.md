# Packaged toolbar native drag check

Checked the private version-12 light Mac Widgets app at source checkpoint
`1b1f7a9`, using the user-authorized mouse-event helper restricted to its exact
bundle and window. No framework code changed for this check.

- Dragged Save from the third row before New. The resulting order was
  Save, New, Open, Separator, Undo, Redo, Delete.
- Hid Save, then dragged it after Open. Its row retained the full-color Save
  artwork, its visibility checkbox stayed off, and its order changed correctly.
- Dragged hidden Save toward the start and pressed Escape before release. The
  order and visibility remained unchanged and the customization panel stayed open.
- Released another drag on Done, outside the list. The panel stayed open and the
  accepted order/visibility did not change; the release did not click Done.

Each outcome was observed through a fresh accessibility snapshot; screenshots
separately confirmed full-color artwork and aligned controls. This is native
injected-pointer evidence, separate from deterministic Manual routing tests and
not a claim of a human pointer or VoiceOver session. The helper's first admission
attempt refused the older bundle-ID spelling before posting input; it was
corrected to match the actual `.native` package identity.

Earlier focused race/routing tests and Prism review are recorded in
[the widget review log](../../reviews/mac-personal-widgets/README.md). No broad
campaign was repeated for this documentation-only checkpoint.
