package cogent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
)

func prepareCommandMenus(groups []app.MenuGroup, registry *command.Registry) ([]app.MenuGroup, error) {
	if len(groups) == 0 {
		return nil, nil
	}
	if len(groups) > 16 || registry == nil {
		return nil, fmt.Errorf("cogent: command menus require a registry and at most16 groups")
	}
	result := make([]app.MenuGroup, len(groups))
	count := 0
	for i, g := range groups {
		count += len(g.Commands)
		if count > 256 || len(g.Label) > 1024 || strings.TrimSpace(g.Label) == "" || !utf8.ValidString(g.Label) || strings.ContainsRune(g.Label, 0) {
			return nil, fmt.Errorf("cogent: invalid command menu group")
		}
		result[i] = app.MenuGroup{Label: g.Label, Commands: append([]string(nil), g.Commands...)}
		for _, id := range g.Commands {
			c, ok := registry.Lookup(id)
			if !ok || len(id) > 128 || len(c.Label) > 1024 || !utf8.ValidString(c.Label) || strings.ContainsRune(c.Label, 0) {
				return nil, fmt.Errorf("cogent: invalid menu command %q", id)
			}
		}
	}
	return result, nil
}

type commandMenuSnapshot struct {
	groups []nativeMenuGroup
	key    string
}
type commandMenuFailure struct {
	err error
	key string
}

type commandMenuPump struct {
	service  *platformServices
	groups   []app.MenuGroup
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	wake     chan struct{}
	latest   atomic.Pointer[commandMenuSnapshot]
	ready    atomic.Bool
	last     string // UI owner only
	failure  atomic.Pointer[commandMenuFailure]
	install  func(context.Context, []nativeMenuGroup, *platformServices) error
	retryKey string // owner only
	retries  int    // at most one same-state retry
}

func newCommandMenuPump(service *platformServices, groups []app.MenuGroup) *commandMenuPump {
	if len(groups) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(service.owner.Context())
	p := &commandMenuPump{service: service, groups: groups, ctx: ctx, cancel: cancel, done: make(chan struct{}), wake: make(chan struct{}, 1), install: nativeSetMenu}
	go p.run()
	return p
}

// capture runs only on the UI owner; one latest snapshot and one wake are kept.
func (p *commandMenuPump) capture() {
	if p == nil || !p.ready.Load() || p.ctx.Err() != nil {
		return
	}
	p.reportFailure()
	groups := make([]nativeMenuGroup, len(p.groups))
	for i, g := range p.groups {
		groups[i].Label = g.Label
		for _, id := range g.Commands {
			c, ok := p.service.commands.Lookup(id)
			if !ok {
				continue
			}
			shortcut := c.Shortcut
			if c.ShortcutScope == command.AfterFocusShortcuts {
				shortcut = ""
			}
			groups[i].Items = append(groups[i].Items, nativeMenuItem{Label: c.Label, Command: id, Shortcut: shortcut, Disabled: p.service.commands.Enabled(id) != nil})
		}
	}
	data, _ := json.Marshal(groups)
	key := string(data)
	if key == p.last {
		return
	}
	if key != p.retryKey {
		p.retryKey = key
		p.retries = 0
	}
	p.last = key
	p.latest.Store(&commandMenuSnapshot{groups: groups, key: key})
	select {
	case p.wake <- struct{}{}:
	default:
	}
}
func (p *commandMenuPump) run() {
	defer close(p.done)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.wake:
			snapshot := p.latest.Load()
			if snapshot == nil {
				continue
			}
			if err := p.install(p.ctx, snapshot.groups, p.service); err != nil && p.ctx.Err() == nil {
				p.failure.Store(&commandMenuFailure{err: err, key: snapshot.key})
				// Report on owner without self-invalidating into an unbounded retry loop.
				// If its queue is full, capture reports the retained error next refresh.
				_ = p.service.owner.Post(p.reportFailure)
			}
		}
	}
}
func (p *commandMenuPump) close() {
	if p == nil {
		return
	}
	p.cancel()
	<-p.done
}

// reportFailure is owner-only. Permit one later same-state retry; repeated
// errors cannot create an unbounded native/owner wake loop.
func (p *commandMenuPump) reportFailure() {
	failure := p.failure.Swap(nil)
	if failure == nil {
		return
	}
	if failure.key == p.last && p.retries < 1 {
		p.last = ""
		p.retries++
	}
	if p.service.onError != nil {
		p.service.onError(failure.err)
	}
}
