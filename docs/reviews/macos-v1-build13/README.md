# macOS build13 packaging review

Prism used **Anthropic Claude Sonnet5.5** for the complete checkpoint and focused
follow-ups. Original JSON findings/receipts are retained unchanged. All calls
have coverage.complete=true, no skipped/truncated text and retired process groups.
Opaque archive/screenshots are represented by their named SHA256 evidence index;
no claim is made that the model inspected binary contents.

The complete first review covered21 text files/123,962 diff bytes and found one
release-copy clarity issue. Follow-ups covered only corrected proposal/validator
files. The latest review is not represented as a clean full-checkpoint result:
its two findings are addressed below and verified locally after correction.
No production source, feature, backend, performance policy or audit record changed.

| Finding | Disposition |
| --- | --- |
| 69138669f193bffa: ambiguous historical/tag versus build13 CI | Fixed: final proposed body contains only exact64bf9cf build-source CI, with original v1.0.0 tag/history clearly separate. |
| d3aae3c8943d18a7: binaries differ from proposed destination's tag | Fixed: propose a new macos-v1.0.0-build13 packaging tag/release exactly at64bf9cf; preserve the old v1.0.0 tag/release. New tag creation still requires the final owner decision and has not occurred. |
| 5e00f8786ad55e7f: unclosed stream / file_digest Python3.11 dependency | Fixed: explicit with blocks and streaming hashlib.sha256; no file_digest dependency remains. Full packet validation passes. |
| 685f0ef7ce16f28b: assertions disabled under -O | Fixed using the suggested guard: optimized execution refuses before importing verifier helpers or inspecting delivery data. A focused -O negative check proves nonzero refusal; normal validation passes. |
| d5d7f0aec33f94cf: archive names / size admission | Added explicit canonical relative-name checks, original app_notices.notice_name validation, ZIP per-member/aggregate and evidence-tar byte bounds. Exact seven archive hashes are already checked before archive inspection; this remains a read-only verifier of this named trusted set. Full packet validation passes. |

[Final packet validation](../../reports/macos-v1-build13/packet-validation.json)
checks all7 archives,8,861 SDK files, embedded app manifests/binaries/notices,
private retained-input hash,255 evidence members, exact source CI, proposed tag
source and unchanged historical audits. Final executable process absence and12
strict post-launch signature checks were recorded before report edits. The
[optimized-mode check](optimized-check.json) is a refusal test, not an app test.
Review corrections use focused local verification; unchanged app binaries/SDK,
consumer/native tests and existing green CI are reused. No additional broad
review, performance or qualification campaign was started.
