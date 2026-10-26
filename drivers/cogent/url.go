package cogent

import (
	"context"
	"github.com/dshills/vellum/ui"
)

func (s *platformServices) OpenURL(ctx context.Context, value string) error {
	if err := ui.ValidateLinkURL(value); err != nil {
		return err
	}
	scope, end, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer end()
	return nativeOpenURL(scope, value)
}
