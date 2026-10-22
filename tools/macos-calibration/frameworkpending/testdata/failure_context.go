// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
package main

import "errors"

// Successful inputs neither collect extra native context nor mutate the journal.
// Preserve the first failed state even if a later teardown observation differs.
func cohortStateError(out map[string]any, sequence int, state map[string]any, valid bool, context func() map[string]any) error {
	if valid {
		return nil
	}
	if _, exists := out["foreground_failure"]; !exists {
		out["foreground_failure"] = map[string]any{
			"sequence": sequence,
			"state":    state,
			"context":  context(),
		}
	}
	return errors.New("foreground/configuration changed during cohort")
}
