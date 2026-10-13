package ui

// SplitRange resolves an omitted Numeric configuration to a 5..95% range and
// a 1% keyboard step. Explicit bounds must remain strictly inside 0..1 so neither
// retained pane is implicitly collapsed. Number is the first pane's accepted fraction.
func SplitRange(config Numeric) Numeric {
	if config == (Numeric{}) {
		return Numeric{Min: 0.05, Max: 0.95, Step: 0.01}
	}
	return config
}
