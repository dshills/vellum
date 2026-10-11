# Specifications

- [Framework contract](framework.md): implemented application ownership, commands,
  tasks, docking, virtualization and GPU canvas behavior, including current limits.
- [Keyed reconciliation](keyed-reconciliation.md): retained native identity.
- [Layout and themes](layout-theme.md): portable style values and local themes.
- [Standard controls](controls.md): interaction, focus and semantic metadata.
- [Windows and services](windows-services.md): scoped lifetimes and OS integration.
- [Workbench interaction](workbench-interaction.md): pointer docking, floating,
  redocking and transactional multiwindow session persistence.
- [Accessibility and keyboard workflows](accessibility.md): native macOS semantics,
  actions, focus scopes and bounded virtual data publication.
- [External file drops](file-drop.md): bounded native file references and owner callbacks.
- [Document persistence](document.md): asynchronous save/open, revision checks and unsaved-close policies.
- [Document history](history.md): bounded undo/redo, save revisions and native keyboard precedence.
- [Native dispatch](native-dispatch.md): synchronous callback failure and desktop event termination.
- [macOS completion](macos-completion.md): remaining framework and target acceptance work.
- [Retained GPU drawing](drawing.md): immutable meshes, strokes, waveform reduction, indexed culling and resource accounting.
- [Retained image drawing](images.md): immutable tile patches, native texture retention and alpha-safe filtering.
- [Rendering](rendering.md): demand painting, resource lifetimes and device loss.
- [Platform support](platform-support.md): Windows/Linux services and target gates.
- [Release artifacts](release.md): alpha API, private SDK integrity and packaging.
- [Local SDK](sdk.md): independent applications, pinned bootstrap and benchmarks.
- [Prototype validation plan](prototype-plan.md): research acceptance gates and
  recorded investigation progress; this is a historical foundation plan.

See [implementation status](../docs/implementation.md) for completed phases and
[foundation reports](../docs/reports/README.md) for the evidence behind the gates.
