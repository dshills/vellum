# Private macOS v1 final closeout

**Complete for the approved private personal-use scope.** Runtime/SDK source is
`382e0e47e08044f18106b34b35f503a169bbfec0`; version `v1.0.0`, app build9.
The [release contract](../../../specs/macos-v1-release.md),
[stable API policy](../../../specs/stable-api.md) and
[Mac support guide](../../macos-v1.md) define the exact compatibility and runtime
claim. No public tags, release assets, signing identity or notarization were used.

## Checkpoints

1. `0391fb7` repairs the observed cold-build CI timeout. Required text binaries
   compile separately with a300s bound; their original90s execution deadline,
   selected tests, race/offscreen flags, count and package cwd remain. Missing
   binaries fail explicitly. Focused runner/failure/Windows-suffix and actual
   text-backend checks pass; Prism/Anthropic Claude Sonnet5.5 reviewed the repair.
2. `382e0e4` adopts the stable v1 source contract for13 named root packages and
   a whitelisted Mac driver facade. Replay/automation/diagnostics and optional
   native/GPU interfaces remain experimental. Current version/docs/consumer
   metadata agree. Source comparison against the accepted RC changes no public
   signature or runtime feature; only Version/comments differ in non-test Go
   source. Portable and consumer race/vet, SDK tooling and290 local links pass.
   Prism reviewed the scoped contract and its preserved release guards.
3. Final packaging uses that clean committed source, independent consumer and
   unchanged native capsule. Three apps and SDK/source archives are verified;
   Notes was launched and checked. This report-only checkpoint records final
   remote CI and actual package/relaunch evidence; it does not relabel built bytes.

## Required remote CI

Both workflows pass on the exact packaged source:
[Framework40232563478](https://github.com/dshills/vellum/actions/runs/40232563478)
and [Platform adapters40232563469](https://github.com/dshills/vellum/actions/runs/40232563469).
The Framework workflow includes scripts/check.sh and portable benchmarks.
X11 software-GPU smoke and Windows build/unit pass; interactive platform
qualification remains intentionally skipped. These are build/test outcomes, not
new hardware, display, human-input or performance qualification.

The original [Framework failure40231586819](https://github.com/dshills/vellum/actions/runs/40231586819)
remains a failure: its aggregate120s process limit expired during cold native
compilation. No test assertion or release responsiveness gate was relaxed.
[ci.json](ci.json) records source-bound outcomes and the specific repair.

## Independent delivery

[package-verification.json](package-verification.json) pins SDK, native and
app ZIP/signed-binary hashes. All29 build/verification commands return0. The
consumer's nine replacement origins are inside the extracted SDK; module
integrity, consumer race/vet and generated independent starter test/build pass.
The same SDK builds Notes, Mac Widgets and the small two-window shell fixture.
Info.plist versions are1.0.0/build9. Exact ZIP extractions and installed native
copies pass strict codesign checks and signed-byte comparison. Original notices
remain executable-bound private-package inputs; no public audit is implied.

The native capsule remains
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
The final SDK SHA256 is
`680728b78331698e253cb9f59d53557bac1259d7b71071ff903e208d4e254cdc`.
Canonical artifacts are clean ZIP/tar archives, with a README, manifest,
acceptance record and SHA256SUMS in the local macos-v1 delivery. Synced-folder
metadata can invalidate a loose app signature; verified launch copies use
temporary native staging and are reproducible from the canonical ZIPs.

## Final app launch

Packaging and the21:01:22Z launch check ran while remote CI was in progress.
The Framework workflow subsequently succeeded (run updated21:30:42Z) on that
same source. CI gates the final handoff; it did not precede the launch test.
No timestamps or earlier pending states are relabeled.

[native-relaunch.json](native-relaunch.json) records the actual app-local check.
The previous Notes process was already closed, so no Save/Close keystroke was
needed. The verified v1 package launched, mounted its editor and displayed Saved
with the existing seven-note notebook. Its notebook bytes are unchanged across
startup. The app was left running. Only the authorized app/path-restricted helper
was used; clipboard and global input/accessibility settings were unchanged. Raw
notebook values and the local AX journal are not included in this report.

The completed [RC evidence](../macos-v1/README.md) supplies unchanged physical
selection, Unicode/history, native panel focus, toolbar isolation and disabled
action checks. This final pass adds fresh packaging/launch evidence; it does not
claim another physical interaction campaign, fresh Save/recovery round trip or
human IME/VoiceOver acceptance.

## Remaining boundaries

Support is personal use on the recorded Apple Silicon/macOS26.6.2 configuration.
Human IME/VoiceOver, other hardware/macOS/displays and broader platform use remain
unqualified under the existing owner choice. The backend and performance policy
are unchanged; experimental rendering/pacing remains disabled. Existing
correctness/resource/format limits still apply. Optional polish remains backlog.
Public distribution, coordinated module tags, changed-payload audit and any
signing/notarization policy require a separate owner-authorized release. Current
audit metadata stays pending. Historical preview/RC records are preserved.
