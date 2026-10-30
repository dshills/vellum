# Checkpoint 4: static native linkage and deployment metadata

Internal review, 30 October 2026. Added a developer-only checker for the existing
private five-app system-library-only packaging route. No framework/app binary,
backend, native pin, signing policy, performance gate or support-matrix change.
The tool is newer than packaged source b9212489 and runs from this checkout.

Reviewed explicit installation receipt pin and exact app/name/identifier/build
admission, symlink/special-file and file-count bounds, one Mach-O per app,
ARM64 executable header, all supported dylib load-command forms, canonical
system paths, loader, rpaths, dyld overrides and unique macOS deployment metadata.
Tools have 20-second deadlines; no app, dyld, LaunchServices or GUI API is invoked.
The checker is not a general dependency resolver or publisher authenticity check;
prior byte/signature verification and no concurrent installation writer are
preconditions. Static paths do not prove dlopen/plugin or transitive system closure.

Six focused tests cover current/older host version comparisons, all five dylib
load kinds with absolute/relative/rpath/traversal failures, loader/environment/rpath
refusal, malformed/duplicate/foreign/missing deployment data, legacy minimum
metadata, pinned five-app orchestration, and unexpected embedded Mach-O rejection.
Tools are mocked in the synthetic orchestration test; only lipo/otool are admitted.
Added this suite to the existing tooling stage; dispatcher equivalence still passes.

Actual inspection passed for all five installed build24 apps: thin ARM64
executables, 23 system dependencies each, /usr/lib/dyld, no rpaths or dyld overrides,
no extra Mach-O closure. Exact signed executable hashes are in the linkage report.

Material limitation: every binary declares minimum macOS 27.0 and SDK 27.0.
The current host reports 27.2, so its version satisfies that declared minimum;
no application was launched to establish behavior. Historical qualification is
26.6.2 and does not transfer; these binaries cannot run there. This is a private
current-machine candidate with explicit limits, not a newly qualified release.
No arbitrary deployment clamp or new release gate was introduced to hide this.

Checks: focused linkage suite (6 tests plus path/load subcases), stage-dispatch
suite (2), actual all-five static inspection, diff whitespace. All pass within the
stated route and current-host metadata boundary. No personal data, desktop,
clipboard, global settings, binary rewrite, repackage or release action.
