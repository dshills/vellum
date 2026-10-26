package app

import "context"

// URLServices opens an explicit HTTP(S) URL with the user's default handler.
// Call on a worker. Success means the OS accepted the open request, not that a
// page loaded. Implementations must validate the URL and respect cancellation.
// It is optional; existing Services implementations need no new methods.
type URLServices interface {
	OpenURL(context.Context, string) error
}

func (a *Application) URLs() (URLServices, error) {
	s, e := a.Services()
	if e != nil {
		return nil, e
	}
	v, ok := s.(URLServices)
	if !ok {
		return nil, ErrUnavailable
	}
	return v, nil
}
