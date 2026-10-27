# Reproduce the packaging handoff

Use the unchanged public producers at packaging checkpoint 1fc594b with the
[approved packet's retained exact pre-sign inputs](../macos-v1-distribution/REPRODUCE.md).
The public commands and frozen/native input hashes in that recipe remain valid;
the current exact-payload audit is now approved in release/version.json.

The evidence archive retains the actual production controllers, stdout/stderr,
SDK independent-consumer receipt, fixed-profile Notes launcher and final
postlaunch.py. These are historical local controllers with explicit host paths,
not an unattended publisher. Their command timeouts and their earlier parsing
failures are visible. No controller contains an upload command or credential.

1. Verify retained input archive SHA 44ea387d… and reconstruct each app's exact
   inventory subset, not the union of all shared notices.
2. Use clean committed packaging source, fresh verified preparation, the approved
   native archive 786aaad8…, and the pinned Go1.25.6 toolchain. Run public SDK/app
   producers without private flags. Preserve default backend/rendering policy.
3. Verify all SDK manifest files and independent consumer with the supplied SDK
   digest; verify every ZIP extraction and strict signature. Compare the SDK file
   delta to the approved private candidate before calling it the same product.
4. Launch only separate verified extraction copies. Use a new disposable Notes
   profile and preserve personal notebooks/settings. Observe native content and
   close, then verify process retirement, signatures and signed hashes again.
5. Record actual new archive hashes; do not expect ad-hoc signatures/manifests or
   public metadata to reproduce older private-candidate hashes. Hold failures.
   Reuse this checkpoint's unchanged results instead of repeating these steps
   without a concrete change or failure.

All produced and recommended checksums differ intentionally: the Mac Shell
fixture is retained for traceability but excluded from the proposed upload set.
The sealed producer runtime labels are not edited after native observations.
