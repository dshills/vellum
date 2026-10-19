# Legacy theme and text compatibility

The app-surface additions changed the theme JSON hash even when unused, and
Core allocation alignment shifted inherited native paragraph alignment. Frozen
workspace replay detected both regressions. The original records and goldens
remain unchanged.

Only the five new optional style fields omit unset values from JSON. Explicit
zero values remain serialized. Core stage97 makes allocation layout opt-in;
managed labels enable it for resolved TextAlign/GrowX/GrowY and reset it on
removal. Existing control internals and scalar Grow retain legacy behavior.
Drawing, selection and hyperlink mapping still share the same text origin.

All six workspace runs (10k/1m rows, batches1/7/64) pass:1,296 operations and162
captures with exact original model and image comparisons. Root race/vet,
focused driver/Core race, driver vet, full preparation reproduction and29
preparation checks pass. The external artifact folders preserve initial failed
runs and first-capture isolation; no golden migration was used.

Independent source review found no material issues. Prism reviewed all11
implementation/spec/recipe files using Anthropic Claude Sonnet5.5. Its one
medium finding was independently dismissed: Label construction guarantees
*core.Text and kind changes remount the record. Re-resolving style is a bounded
refactor opportunity without a demonstrated material cost. Raw review and
adjudication are retained beside the certificate.

These are automated software results. Corrected native fixtures must be rebuilt
and renewed; physical pointer, real presentation performance and full Mac
completion remain separate gates.
