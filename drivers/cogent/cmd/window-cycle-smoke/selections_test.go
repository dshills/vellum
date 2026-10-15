//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"fmt"
	"math"
	"testing"

	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/acceptance"
)

func listPublication() cogent.AccessibilityInspection {
	s := cogent.AccessibilityInspection{Reachable: true, NativeWindow: true, Elements: []cogent.AccessibilityElement{{ID: "list-native", DomainID: acceptance.SelectionListID, ObjectIdentity: 1, Role: "AXList", Enabled: true, Value: "Last component", Frame: [4]float64{0, 0, 180, 90}}}}
	for i, id := range []string{"a", "b", "locked", "c", "d", "e", "f", "last"} {
		e := cogent.AccessibilityElement{ID: fmt.Sprint(i), DomainID: id, Parent: "list-native", ObjectIdentity: uint64(i + 2), Role: "AXRow", Enabled: id != "locked", Selected: id == "last", Name: []string{"Rejected first", "Accepted initial", "Locked", "Resistor", "Capacitor", "Connector", "Label", "Last component"}[i]}
		if i < 3 {
			e.Frame = [4]float64{0, float64(i * 30), 180, 30}
		}
		s.Elements = append(s.Elements, e)
	}
	return s
}

func TestSelectionListCompletionRefusesMalformedAndReplacedRows(t *testing.T) {
	var identity selectionListIdentity
	if !selectionListCompleted(listPublication(), &identity) || !selectionListCompleted(listPublication(), &identity) {
		t.Fatal("accepted/rejected/same-value list publication refused")
	}
	for name, mutate := range map[string]func(*cogent.AccessibilityInspection){
		"missing":   func(s *cogent.AccessibilityInspection) { s.Elements = s.Elements[:8] },
		"selection": func(s *cogent.AccessibilityInspection) { s.Elements[1].Selected = true },
		"disabled":  func(s *cogent.AccessibilityInspection) { s.Elements[3].Enabled = true },
		"zero-visible": func(s *cogent.AccessibilityInspection) {
			for i := 1; i < len(s.Elements); i++ {
				s.Elements[i].Frame = [4]float64{}
			}
		},
		"nonfinite-hidden": func(s *cogent.AccessibilityInspection) { s.Elements[8].Frame[0] = math.NaN() },
		"outside":          func(s *cogent.AccessibilityInspection) { s.Elements[1].Frame[0] = 190 },
		"duplicate":        func(s *cogent.AccessibilityInspection) { s.Elements = append(s.Elements, s.Elements[1]) },
		"replaced":         func(s *cogent.AccessibilityInspection) { s.Elements[1].ObjectIdentity = 90 },
	} {
		t.Run(name, func(t *testing.T) {
			s := listPublication()
			mutate(&s)
			if selectionListCompleted(s, &identity) {
				t.Fatal("bad list publication qualified", s)
			}
		})
	}
}

func canvasPublication(selected string, generation uint64) cogent.AccessibilityInspection {
	s := cogent.AccessibilityInspection{Reachable: true, NativeWindow: true}
	for i, id := range []string{acceptance.SelectionCanvasID, acceptance.SelectionLeftID, acceptance.SelectionRightID} {
		e := cogent.AccessibilityElement{ID: fmt.Sprint(i), DomainID: id, ObjectIdentity: uint64(i + 1), Role: "AXButton", Enabled: true, Selected: id == selected, Parent: "0", GridScope: "0", GridIncarnation: 7, GridGeneration: generation, GridTargetGeneration: generation, Frame: [4]float64{float64(i * 60), 20, 50, 50}, Name: []string{"", "Left object", "Right object"}[i]}
		if i == 0 {
			e.Role = "AXGroup"
			e.Parent = ""
			e.Frame = [4]float64{0, 0, 240, 160}
			e.Name = "Selectable components"
			e.Value = "Left object"
			if selected == acceptance.SelectionRightID {
				e.Value = "Right object"
			}
		}
		s.Elements = append(s.Elements, e)
	}
	return s
}

func TestSelectionCanvasCompletionBindsOneCurrentNativeEnvelope(t *testing.T) {
	var identity selectionCanvasIdentity
	if !selectionCanvasCompleted(canvasPublication(acceptance.SelectionLeftID, 3), acceptance.SelectionLeftID, &identity) || !selectionCanvasCompleted(canvasPublication(acceptance.SelectionRightID, 4), acceptance.SelectionRightID, &identity) {
		t.Fatal("newer completed selection did not preserve native incarnation")
	}
	for name, mutate := range map[string]func(*cogent.AccessibilityInspection){
		"unnamed-host":     func(s *cogent.AccessibilityInspection) { s.Elements[0].Name = "" },
		"wrong-host-name":  func(s *cogent.AccessibilityInspection) { s.Elements[0].Name = "Another drawing" },
		"empty-host-value": func(s *cogent.AccessibilityInspection) { s.Elements[0].Value = nil },
		"stale-host-value": func(s *cogent.AccessibilityInspection) { s.Elements[0].Value = "Left object" },
		"wrong-host-value": func(s *cogent.AccessibilityInspection) { s.Elements[0].Value = "Another object" },
		"pending":          func(s *cogent.AccessibilityInspection) { s.Elements[0].GridPending = true },
		"target":           func(s *cogent.AccessibilityInspection) { s.Elements[1].GridTargetGeneration++ },
		"scope":            func(s *cogent.AccessibilityInspection) { s.Elements[1].GridScope = "other" },
		"parent":           func(s *cogent.AccessibilityInspection) { s.Elements[1].Parent = "other" },
		"incarnation":      func(s *cogent.AccessibilityInspection) { s.Elements[2].GridIncarnation++ },
		"selection":        func(s *cogent.AccessibilityInspection) { s.Elements[2].Selected = false },
		"nonfinite":        func(s *cogent.AccessibilityInspection) { s.Elements[1].Frame[2] = math.Inf(1) },
		"duplicate":        func(s *cogent.AccessibilityInspection) { s.Elements = append(s.Elements, s.Elements[2]) },
		"replaced":         func(s *cogent.AccessibilityInspection) { s.Elements[2].ObjectIdentity = 90 },
	} {
		t.Run(name, func(t *testing.T) {
			s := canvasPublication(acceptance.SelectionRightID, 5)
			mutate(&s)
			if selectionCanvasCompleted(s, acceptance.SelectionRightID, &identity) {
				t.Fatal("bad canvas publication qualified", s)
			}
		})
	}
	if selectionCanvasCompleted(canvasPublication(acceptance.SelectionRightID, 4), acceptance.SelectionRightID, &identity) || selectionCanvasCompleted(canvasPublication("unknown", 5), "unknown", &identity) {
		t.Fatal("stale generation or unknown selection qualified")
	}
}

func TestSelectionCanvasActionTargetCannotCrossNativeScope(t *testing.T) {
	s := canvasPublication(acceptance.SelectionLeftID, 3)
	e := s.Elements[1]
	if !selectionCanvasTarget(s, e) {
		t.Fatal("current target refused")
	}
	for _, mutate := range []func(*cogent.AccessibilityElement){
		func(e *cogent.AccessibilityElement) { e.Parent = "another-canvas" },
		func(e *cogent.AccessibilityElement) { e.GridScope = "another-canvas" },
		func(e *cogent.AccessibilityElement) { e.GridIncarnation++ },
		func(e *cogent.AccessibilityElement) { e.GridPending = true },
		func(e *cogent.AccessibilityElement) { e.GridGeneration-- },
	} {
		wrong := e
		mutate(&wrong)
		if selectionCanvasTarget(s, wrong) {
			t.Fatal("foreign/stale target admitted", wrong)
		}
	}
}
