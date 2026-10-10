// Package eventqueue distinguishes lossless actions from coalescible progress.
package eventqueue

import "cogentcore.org/core/events"

func Send(source *events.Source, data any) {
	e := &events.CustomEvent{}
	e.Typ = events.Custom
	e.Data = data
	e.Init()
	e.SetUnique()
	source.Deque.Send(e)
}
