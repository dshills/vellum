# Framework implementation phases

This implementation proceeds alongside the unresolved foundation checks. Those
checks still prevent a production-ready claim. They do not prevent building and
assessing the workbench API. Each phase receives Prism review, relevant tests,
and a separate commit.

1. Application lifecycle and keyed UI API — implemented; native driver and executable example.
2. Commands and cancellable tasks — pending.
3. Docking, document tabs, inspector and layout persistence — pending.
4. Virtualized tables and trees — pending.
5. Shared-device GPU canvas with navigation and selection — pending.

The first driver deliberately exposes one native window. Root UI ownership and
logic tests are portable; native runtime support is not claimed across platforms.
