package ui

// MaxChoiceOptions bounds retained radio/segmented options. Larger collections
// belong in Selector or a virtual data view.
const MaxChoiceOptions = 64

// NextChoiceOption wraps through enabled IDs in a validated choice descriptor.
// step is +1 or -1. An absent current ID picks first/last enabled respectively.
// Disabled groups, invalid descriptors and all-disabled options have no target.
func NextChoiceOption(n Node, current string, step int) (string, bool) {
	if (n.Kind != RadioGroup && n.Kind != SegmentedControl) || n.Disabled || (step != 1 && step != -1) || ValidateControl(n) != nil {
		return "", false
	}
	index := -1
	for i, o := range n.Options {
		if o.ID == current {
			index = i
			break
		}
	}
	if index < 0 && step < 0 {
		index = 0
	}
	for range len(n.Options) {
		index = (index + step + len(n.Options)) % len(n.Options)
		if !n.Options[index].Disabled {
			return n.Options[index].ID, true
		}
	}
	return "", false
}
