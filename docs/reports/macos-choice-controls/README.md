# Radio groups and segmented controls

The public framework adds `ui.RadioGroup` and `ui.SegmentedControl` with 1..64
keyed options and one required accepted SelectedID. Labels may repeat; domain
IDs remain unique within the group. Disabled selected values remain readable.
The retained private native options survive label changes and reordering;
removal retires their binding. Native requests capture ID and binding identity,
recheck enabled ancestry and the latest callback on bounded owner delivery,
and cannot change a new same-ID binding after removal. Pending requests preserve
exclusive accepted selection, including application rejection and mailbox refusal.

Each group contributes one managed Tab stop. Arrows wrap through enabled choices,
Home/End choose the first/last, and Enter/Space request the focused option.
Radio groups default vertical and segmented controls horizontal; both support
explicit Orientation. Native semantic groups publish selected labels; option
radio-button records have generated IDs, boolean selection, measured frames and
focus/press actions. Set uses domain IDs. Focus admission validates a live binding;
pending semantic focus intentionally targets the current visible/enabled same-ID
binding at application time. Disappearance, disabled scope and hidden tabs/modals
refuse focus. Docked/modal namespaces retain this policy.

The gallery shares controlled state among its selector, radio group and segments.
See [the public catalog](../../controls.md) and [contract](../../../specs/controls.md).

## Verification and fixes

The final portable and complete offscreen adapter race suites pass. The focused
choice, capture and gallery regressions pass three repetitions under race;
adapter vet and the native gallery build also pass. The gallery was compiled,
not launched. These five checks bind the same nineteen changed source leaves
and unchanged 6,198 prepared dependency identities.

Actual Manual software input routes arrows, Home/End, Space, measured pointer
selection and Tab after each group, with completed pixel changes. Tests exercise
keyed retention/reorder, current callbacks, disabled/retired requests, owner
mailbox refusal, strict domain semantic actions and stale handlers. Separate
inert Cocoa fixtures verify AXRadioGroup and AXRadioButton roles. Additional
fixtures cover current-ID replacement focus, removed/hidden/disabled refusal,
and qualified option focus in docked modal content.

Internal review found quiet pointer focus was omitted by visual focus-state
flags, breaking subsequent managed Tab navigation. Native focus ownership now
identifies the correct managed group or ordinary control. Manual captures use
focusable native proxies, avoiding ancestor-container focus; active inline-grid
editor reporting retains its separate identity. Another finding showed a local
option ID could shadow an unrelated checkbox in Manual's DomainID index. Full
AX identity and qualified managed lookup now keep captures distinct. The same
collision fixture fails both kinds before the fix (unchecked checkbox falsely
captured as selected true), and the ordinary quiet-pointer fixture fails before
focus capture is corrected. Both pass on final sources. Raw pre-fix failures and
the exact inspected Manual/fixture payloads are retained.

25 raw command receipts and 50 logs are archived. The
10 exit-1 attempts include constructor-type/unused-import compilation
errors, the genuine initial group Tab regression and diagnostic run, the capture
collision/quiet-focus regressions, and an early focus-policy fixture that omitted
explicit invalidation and used the old index lookup, and the genuine retired
native mutation regression. None is a passing result.
Only the five final-source checks qualify this checkpoint. Historical receipts
preserve changed-source hashes and available payloads; a payload index explicitly
marks missing historical leaves. The initial and both final review snapshots remain
immutable. Neither the historical archive nor changed-leaf manifests claim full
compiler, native binary or SDK closure.

## Reviews and remaining gates

The internal source review approved the final nineteen leaves after its findings
were fixed. Prism using Anthropic Claude Sonnet 5.5 first covered nineteen files
in 4 chunks, 49473 input/5596 output
tokens and raw exit 1, with one lifecycle-guard finding. Normal retirement does
not clear the choice state, so its claimed normal-path nil panic was not established.
A narrower direct late-event regression did reproduce native mutation after
retirement. The guard fixes that boundary; the same fixture passes afterward.
The two changed source leaves were then reviewed in a single chunk, 10641 input/1487 output tokens, raw exit 0 and no findings.
The seventeen other files retain their reviewed identities, giving complete
nineteen-file composite coverage with the sole finding resolved. A separate internal evidence review binds
this report, final source map, reviews, raw receipts/producers and sealed archive
before commit. The [checkpoint](records/checkpoint.json) binds that review and the
[archive](records/evidence.tar.gz) with its [member index](records/evidence-index.json).
Binaries and caches are excluded. No desktop UI/input/settings or general
clipboard writes occurred. Physical Mac input, human VoiceOver, native/GPU memory,
sustained performance/endurance and full Mac qualification remain open.
