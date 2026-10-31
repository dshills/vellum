# Current project and macOS status

Documentation checkpoint: 31 October 2026, source release **v1.1.3** at
`0b11d3b092455b9967bad891fa42ae4ba5b6f8ae`. The version in
[release metadata](../release/version.json) and `vellum.Version` identifies source;
[stable-v1 compatibility](../specs/stable-api.md) retains its v1.0.0 API baseline.
Later documentation changes do not rebuild or relabel installed binaries.

## Source, SDKs and applications

| Item | Recorded identity and scope |
| --- | --- |
| Latest source tag | [v1.1.3](https://github.com/dshills/vellum/tree/v1.1.3), current framework source and build tools |
| Selected private SDK/five apps | build25-ed4f71f49bbf, SDK v1.1.1, source ed4f71f49bbf623dd8e205571669539212068c62; minimum macOS 27.0, inspected on 27.2 |
| Latest locally verified demo suite | v1.1.2 build5, source 04f1174f883469f6f62ce0975f1349127d17db9f; 19 signed apps plus replay; synced archive delivery and restoration verified |
| Exported-SDK consumer | Work Log 0.1.0, generated from the build25 SDK; native editing, history, Save/Open and restart checked |
| Published prepared native downloads | [macos-v1.0.0-build13](https://github.com/dshills/vellum/releases/tag/macos-v1.0.0-build13), source 64bf9cf, SDK v1.0.0 and six apps; older than the latest source |

The [candidate index](../release/private-macos-candidate.json) binds private
artifact hashes. Its `native_runtime: not executed` and `desktop_checks: deferred`
fields preserve the **producer-time** record, not the outcome of later sessions.
Subsequent build25 [Documents checks](reviews/next-five-stability/item-5.md) cover
three short native sessions; they do not establish every app's native acceptance.
The broader build24 [Quit/recovery records](reviews/document-priorities/) retain
their original identity. A later v1.1.2 physical body/title click check is in the
[application-readiness records](reviews/application-readiness/README.md).

## What works and what remains bounded

Applications can use stable lifecycle, commands/history, document/editor, controls,
virtual grids/trees and canvas APIs. The native driver provides macOS windows,
titlebar toolbars, menus and services. Palette placement is same-window.
The [demo suite](../examples/demo-suite/README.md),
[application guide](building-macos-apps.md) and
[Work Log recipe](../examples/work-log/README.md) provide runnable entry points.

Personal-use support is limited to recorded Apple Silicon Mac configurations.
Managed controls use Mac-style rendering; they are not all AppKit controls.
Human IME candidate usability, VoiceOver, other hardware/displays/OS versions and
downloaded-quarantine launch remain unqualified. Windows/Linux runtime is
unqualified; Wayland is unsupported. Replay, semantic automation and broader
GPU/workbench workloads retain experimental exclusions. See the
[Mac support guide](macos-v1.md) and [platform contract](../specs/platform-support.md).

## Finite macOS closeout

1. Verify all required CI workflows on the exact final source. Check
   [Framework](https://github.com/dshills/vellum/actions/workflows/framework.yml),
   [API](https://github.com/dshills/vellum/actions/workflows/api.yml) and
   [platform workflows](https://github.com/dshills/vellum/actions/workflows/platforms.yml)
   against the full commit SHA, not a previous tag's green result. The v1.1.2
   Framework failure was repaired in source; v1.1.3 has its own CI outcome.
2. Complete [multi-day operator evaluation](daily-use/SESSION-LOG.md) on one
   frozen candidate. The ledger currently records zero qualifying days; short
   automated runs and the physical focus check do not satisfy this gate.
3. Choose a final green source, produce a matching SDK/app packet and run scoped
   native acceptance on those exact bytes. Existing verified packages are useful
   evidence, not an unbuilt v1.1.3 native release.
4. Record the supported host and any deliberately retained human/hardware
   exclusions. IME/VoiceOver acceptance requires actual human results before
   claiming that scope.
5. Deliver the selected packet with hashes, supported configuration, known limits
   and [upgrade/recovery guidance](upgrading-macos-apps.md). Public native assets,
   Developer ID signing and notarization remain separate scope decisions.

No new widgets, backend rewrite, performance campaign or cross-window placement
is required to finish the currently bounded personal-use scope. Historical full
performance goals remain unqualified; they have not been silently passed or
replaced. The chronological plans remain evidence of earlier decisions rather
than a fresh work queue.

The low dependency alert is [triaged against the recorded artifact](reviews/application-readiness/item-5-advisory.json):
it points to an upstream development dependency declaration in notice material;
rand is absent from that compiled-source inventory. GitHub's alert remains open.
This is not a blanket statement about future native builds or upstream benchmarks.
