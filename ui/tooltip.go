package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// WithTooltip attaches plain-text, noninteractive help without another focus stop.
// Long hover shows it; F1 on the focused host also shows it. Escape, focus loss,
// pointer exit, disablement, hiding or removal dismisses it. No application timer
// or goroutine is created. Existing Accessibility.Help is preserved; when empty
// it inherits this text. Empty text removes explicit help-tag presentation.
func WithTooltip(host Node, text string) (Node, error) {
	if err := validateTooltip(text); err != nil {
		return Node{}, err
	}
	host.Tooltip = text
	if host.Accessibility.Help == "" {
		host.Accessibility.Help = text
	}
	if err := Validate(host); err != nil {
		return Node{}, err
	}
	return host, nil
}
func validateTooltip(text string) error {
	if len(text) > 4096 || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
		return fmt.Errorf("ui: tooltip requires at most 4096 UTF-8 bytes without NUL")
	}
	return nil
}
