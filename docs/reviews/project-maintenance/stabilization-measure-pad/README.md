# Add Measure Pad as a real generated-starter application

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Document.Session guarantees Install and OnComplete run on the application owner; source contract explicitly prohibits worker model mutation. Decode receives bounded data from FileStore, with independent codec limit too. Added selection/history availability checks for Delete/Undo/Redo. Conversion, validation, history and actual file round trips pass.

Checks are offscreen/source checks; no new native qualification is claimed.
