package cogent

import "cogentcore.org/core/events"

func isWindowWake(e events.Event) bool {
	ce, ok := e.(*events.CustomEvent)
	if !ok {
		return false
	}
	name, ok := ce.Data.(string)
	return ok && name == "vellum.wake"
}

// Owner work belongs to its originating window even while a popup is modal.
// Only our exact wake and this window's session requests bypass popup routing;
// unrelated Custom events retain the usual target and modality rules.
func bindWindowOriginCustom(w *nativeWindow, session *DesktopSession, custom func(*nativeWindow, events.Event)) {
	w.body.Scene.SetOriginCustomHandler(func(e events.Event) bool {
		if handleWindowActivation(w, e) || handleWindowVisibility(w, e) {
			return true
		}
		if !isWindowWake(e) && !session.owns(w, e) {
			return false
		}
		custom(w, e)
		return true
	})
}
