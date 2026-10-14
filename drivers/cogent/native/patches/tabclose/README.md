# Controlled tab close admission

Patch 0058 adds optional Tab.CloseRequest(*Button). When present, a private close
button consumes the gesture and delegates before native parent/index lookup,
selection or deletion. Nil preserves the original unmanaged close path.
The Vellum adapter retains a guarded callback through retirement and validates
the actual originating private button plus an immutable owner capability.
The new Core test checks origin identity, no selection/deletion, detached-parent
safety and original nil-seam deletion. It creates no native window.

Modified upstream tabs retain their original notices; the new test is MIT.
Preparation checks both original/output files and the patch SHA256.
