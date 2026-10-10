//go:build darwin && cgo

// Package macbridge is a disposable AccessKit C bridge, not a general Cogent adapter.
package macbridge

/*
#cgo LDFLAGS: -L${SRCDIR}/../build/native -laccesskit -Wl,-rpath,@executable_path/native -framework AppKit -framework Foundation -framework CoreFoundation -liconv
#include "bridge.h"
*/
import "C"
import (
	"example.com/vellum/foundation/semantics"
	"runtime/cgo"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Bridge struct {
	mu          sync.Mutex
	snapshot    semantics.Snapshot
	handle      cgo.Handle
	adapter     *C.accesskit_macos_subclassing_adapter
	action      func(semantics.Action)
	active      atomic.Bool
	Actions     atomic.Uint64
	Activations atomic.Uint64
	Updates     atomic.Uint64
}

// Attach/Update/Close must be called on Cocoa's main thread with the NSView alive.
func Attach(view, window unsafe.Pointer, s semantics.Snapshot, action func(semantics.Action)) *Bridge {
	b := &Bridge{snapshot: s.Clone(), action: action}
	b.handle = cgo.NewHandle(b)
	b.adapter = C.vellum_attach(view, window, C.uintptr_t(b.handle))
	return b
}
func (b *Bridge) Update(s semantics.Snapshot) {
	b.mu.Lock()
	b.snapshot = s.Clone()
	b.mu.Unlock()
	C.vellum_update(b.adapter, C.uintptr_t(b.handle))
	b.Updates.Add(1)
}
func (b *Bridge) Close() {
	C.accesskit_macos_subclassing_adapter_free(b.adapter)
	b.handle.Delete()
	b.adapter = nil
}
func (b *Bridge) Active() bool { return b.active.Load() }

//export goVellumSnapshot
func goVellumSnapshot(h C.uintptr_t) *C.accesskit_tree_update {
	b := cgo.Handle(h).Value().(*Bridge)
	b.mu.Lock()
	s := b.snapshot.Clone()
	b.mu.Unlock()
	b.active.Store(true)
	b.Activations.Add(1)
	u := C.accesskit_tree_update_with_capacity_and_focus(C.size_t(len(s.Nodes)), C.accesskit_node_id(s.Focus))
	info := C.accesskit_tree_info_new(1)
	C.accesskit_tree_update_set_tree_info(u, info)
	for _, n := range s.Nodes {
		cn := C.accesskit_node_new(role(n.Role))
		if n.Invalid {
			C.accesskit_node_set_invalid(cn, C.ACCESSKIT_INVALID_TRUE)
		}
		setString(cn, n.Label, false)
		if n.Role == "text" || n.Role == "multiline" {
			setString(cn, n.Value, true)
		}
		C.accesskit_node_set_bounds(cn, C.accesskit_rect{x0: C.double(n.Bounds[0]), y0: C.double(n.Bounds[1]), x1: C.double(n.Bounds[2]), y1: C.double(n.Bounds[3])})
		if len(n.Children) > 0 {
			ptr := C.malloc(C.size_t(len(n.Children)) * C.size_t(unsafe.Sizeof(C.accesskit_node_id(0))))
			ids := unsafe.Slice((*C.accesskit_node_id)(ptr), len(n.Children))
			for i, id := range n.Children {
				ids[i] = C.accesskit_node_id(id)
			}
			C.accesskit_node_set_children(cn, C.size_t(len(ids)), (*C.accesskit_node_id)(ptr))
			C.free(ptr)
		}
		for _, a := range n.Actions {
			C.accesskit_node_add_action(cn, action(a))
		}
		if n.Checked != nil {
			v := C.accesskit_toggled(C.ACCESSKIT_TOGGLED_FALSE)
			if *n.Checked {
				v = C.ACCESSKIT_TOGGLED_TRUE
			}
			C.accesskit_node_set_toggled(cn, v)
		}
		if n.Expanded != nil {
			C.accesskit_node_set_expanded(cn, C.bool(*n.Expanded))
		}
		if n.Selected != nil {
			C.accesskit_node_set_selected(cn, C.bool(*n.Selected))
		}
		if n.RowIndex >= 0 && n.Role == "row" {
			C.accesskit_node_set_row_index(cn, C.size_t(n.RowIndex))
		}
		if n.RowCount > 0 {
			C.accesskit_node_set_row_count(cn, C.size_t(n.RowCount))
		}
		if n.Role == "status" {
			C.accesskit_node_set_live(cn, C.ACCESSKIT_LIVE_POLITE)
		}
		C.accesskit_tree_update_push_node(u, C.accesskit_node_id(n.ID), cn)
	}
	return u
}
func setString(n *C.accesskit_node, s string, value bool) {
	p := C.CString(s)
	defer C.free(unsafe.Pointer(p))
	if value {
		C.accesskit_node_set_value(n, p)
	} else {
		C.accesskit_node_set_label(n, p)
	}
}

//export goVellumAction
func goVellumAction(h C.uintptr_t, r *C.accesskit_action_request) {
	b := cgo.Handle(h).Value().(*Bridge)
	a := semantics.Action{Target: uint64(r.target_node), Kind: actionName(r.action)}
	if v := C.vellum_action_value(r); v != nil {
		a.Value = C.GoString(v)
	}
	b.Actions.Add(1)
	b.action(a)
}
func role(s string) C.accesskit_role {
	switch s {
	case "button":
		return C.ACCESSKIT_ROLE_BUTTON
	case "text":
		return C.ACCESSKIT_ROLE_TEXT_INPUT
	case "multiline":
		return C.ACCESSKIT_ROLE_MULTILINE_TEXT_INPUT
	case "checkbox":
		return C.ACCESSKIT_ROLE_CHECK_BOX
	case "table":
		return C.ACCESSKIT_ROLE_TABLE
	case "row":
		return C.ACCESSKIT_ROLE_ROW
	case "tree":
		return C.ACCESSKIT_ROLE_TREE
	case "treeitem":
		return C.ACCESSKIT_ROLE_TREE_ITEM
	case "cell":
		return C.ACCESSKIT_ROLE_CELL
	case "status":
		return C.ACCESSKIT_ROLE_LABEL
	default:
		return C.ACCESSKIT_ROLE_WINDOW
	}
}
func action(s string) C.accesskit_action {
	switch s {
	case "set_value":
		return C.ACCESSKIT_ACTION_SET_VALUE
	case "focus":
		return C.ACCESSKIT_ACTION_FOCUS
	case "expand":
		return C.ACCESSKIT_ACTION_EXPAND
	case "collapse":
		return C.ACCESSKIT_ACTION_COLLAPSE
	default:
		return C.ACCESSKIT_ACTION_CLICK
	}
}
func actionName(a C.accesskit_action) string {
	switch a {
	case C.ACCESSKIT_ACTION_CLICK:
		return "click"
	case C.ACCESSKIT_ACTION_SET_VALUE:
		return "set_value"
	case C.ACCESSKIT_ACTION_FOCUS:
		return "focus"
	case C.ACCESSKIT_ACTION_EXPAND:
		return "expand"
	case C.ACCESSKIT_ACTION_COLLAPSE:
		return "collapse"
	default:
		return "unsupported"
	}
}

func ResidentBytes() uint64 { return uint64(C.vellum_resident_bytes()) }

func (b *Bridge) Focus(focused bool) {
	events := C.accesskit_macos_subclassing_adapter_update_view_focus_state(b.adapter, C.bool(focused))
	if events != nil {
		C.accesskit_macos_queued_events_raise(events)
	}
}
