# Multiline adapter Prism review

Baseline: `60c4f9513ea9bd7af9a3ae886a61f2bcf8715968`. Final reports: `source-approved.json` (nine decoded toolkit files), `bridge-final.json` (twelve integration paths), `docs.json` (five documentation files) and `evidence-final.json` (three verifier/manifest/scope files). Every final report has complete coverage, no exclusions/truncation/skips and zero findings. `scopes.json` records exact staged diff and report hashes. The decoded source snapshot is reviewed separately because reviewing a patch embedded inside a repository diff would add a second layer of hunk syntax; the evidence manifest proves its files match canonical prepared patch 0012.

Initial findings are retained:

- `cafceb94caf7656c`: cache whole-document metadata by identity/revision; covered by metadata reuse regression and zero-allocation cached-state benchmark.
- `4f9fce9ab9cac8b9`, `88d3206408041a3f`: validate NUL/UTF-8 and signed C integer limits before raw client publication; native atomic-publication tests cover refusal.
- `eccac369ac536c5d`: restore dirty flags for explicit/implicit no-net composition commits; both have regression coverage.
- `56934c9a7786f278`: discard pending composition at buffer close; close/reuse regression prevents resurrecting its baseline.
- `2f83a34695e90edb`: map selection through cached line-start binary search and reuse anchor/focus positions.
- `e3fffcfe91d9c6c9`: clarify single-strong-caret geometry and downstream rendered-fragment affinity. Keeping multiple carets for one offset/physical line would violate the existing native geometry contract; secondary bidi affinity remains outside this adapter scope. Native point lookup uses character cells. A uniqueness regression, the final source review and the implementation report establish this explicit policy without claiming full visual editor navigation.
- Evidence findings `beaa1b3ade0cdcdd`, `6a6e45c7318ff3c3`, `d5fb76f9b0969ace`: enforce four-coordinate rectangle shape, five refusal snapshots, and shared Editor/TextField result comparisons. The corrected verifier passes recorded evidence and rejects malformed/missing geometry.

Reviews are not foundation acceptance. The full upstream lines/textcore run retains parser failures and a timeout; G1 remains open. Native tests are deterministic protocol calls and hidden-window lifetime checks, not live IME/VoiceOver workflows. Raw runtime logs preserve their original whitespace; standard unified patch context is also retained. The evidence manifest deliberately does not hash itself or its later review receipts.
