package notes

import (
	"context"
	"github.com/dshills/vellum/app"
)

// QuitGroup coordinates a bounded, fixed set of document windows on the UI
// owner. No window closes until every document approves and clears recovery.
// Repeated requests coalesce; Cancel/failure unlocks every surviving document.
type QuitGroup struct {
	owner     *app.Application
	files     []*Files
	running   bool
	epoch     uint64
	revisions []uint64
	approved  int
}

func NewQuitGroup(owner *app.Application, files ...*Files) (*QuitGroup, error) {
	if owner == nil || len(files) == 0 || len(files) > 32 {
		return nil, ErrInvalid
	}
	seen := map[*Files]bool{}
	for _, f := range files {
		if f == nil || f.owner != owner || seen[f] {
			return nil, ErrInvalid
		}
		seen[f] = true
	}
	return &QuitGroup{owner: owner, files: append([]*Files(nil), files...)}, nil
}
func (g *QuitGroup) Request() {
	if g.running {
		g.valid(g.epoch)
		return
	}
	var live []*Files
	for _, f := range g.files {
		if f.ref.Context().Err() == nil {
			if f.blocked() || f.model.pending {
				f.model.status = "Finish the current operation or dialog before quitting"
				_ = g.owner.Invalidate("")
				return
			}
			live = append(live, f)
		}
	}
	if len(live) == 0 {
		return
	}
	g.files = live
	g.running = true
	g.epoch++
	g.revisions = make([]uint64, len(live))
	for _, f := range live {
		f.quitHeld = true
	}
	g.prepare(0, g.epoch)
}
func (g *QuitGroup) abort() {
	g.running = false
	g.epoch++
	for _, f := range g.files {
		f.quitHeld = false
		f.quitDone = nil
		if f.prompt == "quit" {
			f.prompt = ""
		}
	}
	_ = g.owner.Invalidate("")
}
func (g *QuitGroup) valid(epoch uint64) bool {
	if !g.running || g.epoch != epoch {
		return false
	}
	for _, f := range g.files {
		if f.ref.Context().Err() != nil {
			g.abort()
			return false
		}
	}
	return true
}
func (g *QuitGroup) prepare(index int, epoch uint64) {
	if !g.valid(epoch) {
		return
	}
	if index == len(g.files) {
		g.clear(epoch)
		return
	}
	f := g.files[index]
	f.quitDone = func(ok bool) {
		if !g.valid(epoch) {
			return
		}
		if !ok {
			g.abort()
			return
		}
		g.revisions[index] = f.model.History().Revision
		g.prepare(index+1, epoch)
	}
	err := f.model.request(func() error {
		// Native settlement may outlive an aborted Quit or another window.
		if !g.valid(epoch) {
			return context.Canceled
		}
		if err := f.withQuitAccess(f.model.flush); err != nil {
			f.finishQuit(false)
			return err
		}
		if f.model.History().Dirty {
			f.prompt = "quit"
			_ = g.owner.Invalidate("")
		} else {
			f.finishQuit(true)
		}
		return nil
	}, true)
	if err != nil && g.valid(epoch) {
		g.abort()
	}
}
func (g *QuitGroup) clear(epoch uint64) {
	if !g.valid(epoch) {
		return
	}
	for i, f := range g.files {
		if f.model.History().Revision != g.revisions[i] || f.model.pending {
			g.abort()
			return
		}
	}
	g.approved = 0
	for _, f := range g.files {
		if !g.valid(epoch) {
			return
		}
		f.quitDone = func(ok bool) {
			if !g.valid(epoch) {
				return
			}
			if !ok {
				g.abort()
				return
			}
			g.approved++
			if g.approved != len(g.files) {
				return
			}
			for i, f := range g.files {
				if f.ref.Context().Err() != nil || f.model.History().Revision != g.revisions[i] || f.model.pending {
					g.abort()
					return
				}
			}
			g.running = false
			for _, f := range g.files {
				f.quitHeld = false
				_ = g.owner.CloseWindow(f.ref)
			}
		}
		f.intent = "quit-clear"
		f.clear = true
		f.startRecovery()
	}
}
func (f *Files) finishQuit(ok bool) {
	done := f.quitDone
	f.quitDone = nil
	if done != nil {
		done(ok)
	}
}

// Only synchronous preparation/capture runs with the hold bypassed, on the
// same owner. Native controlled Change callbacks still settle through the FIFO.
func (f *Files) withQuitAccess(action func() error) error {
	held := f.quitHeld
	f.quitHeld = false
	defer func() { f.quitHeld = held }()
	return action()
}
