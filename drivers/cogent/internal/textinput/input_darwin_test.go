//go:build darwin && cgo

package textinput

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"strings"
	"testing"
)

var nativeFixtureResults map[string]error

// Pin the startup thread during package initialization so TestMain remains on
// Cocoa's process main thread even if the scheduler yields before it starts.
func init() { runtime.LockOSThread() }

// Native selector fixtures run on the process main thread, before testing starts
// worker goroutines. They create inert views, never windows or desktop loops.
func TestMain(m *testing.M) {
	pool := fixtureInputPool()
	nativeFixtureResults = runInputNativeFixtures()
	fixtureDrainInputPool(pool)
	os.Exit(m.Run())
}

type inputObservation struct {
	Text     string `json:"text"`
	Sequence uint64 `json:"sequence"`
	Client   uint64 `json:"client"`
	Selected [2]int `json:"selected"`
	Marked   int    `json:"marked"`
	Baseline bool   `json:"baseline"`
}

func inputObserve(i *Input) inputObservation {
	var s inputObservation
	if err := json.Unmarshal([]byte(fixtureInputSnapshot(i)), &s); err != nil {
		panic(err)
	}
	return s
}
func fixtureAssert(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func runInputNativeFixtures() map[string]error {
	results := map[string]error{}
	run := func(name string, body func()) {
		results[name] = nil
		func() {
			defer func() {
				if p := recover(); p != nil {
					results[name] = fmt.Errorf("%v", p)
				}
			}()
			body()
		}()
		if fixtureInputViews() != 0 {
			results[name] = fmt.Errorf("fixture retained %d views", fixtureInputViews())
		}
	}
	rect := [4]float64{0, 0, 0, 20}
	run("validation", func() {
		view := fixtureInputView()
		defer fixtureReleaseInputView(view)
		i := Attach(view, func(Update) bool { return true }, func() {})
		fixtureAssert(i != nil, "attachment failed")
		defer i.Close()
		fixtureAssert(i.Update("A😀B", 1, 3, -1, -1, true, 0, rect), "valid initial update refused")
		before := inputObserve(i)
		cases := []struct {
			text         string
			a, f, ms, me int
			rect         [4]float64
		}{
			{"A\x00B", 0, 0, -1, -1, rect}, {string([]byte{0xff}), 0, 0, -1, -1, rect},
			{"A😀B", 2, 2, -1, -1, rect}, {"A", math.MaxInt32 + 1, 0, -1, -1, rect},
			{"A", 0, 0, -2, -1, rect}, {"A", 0, 0, -1, -1, [4]float64{math.NaN(), 0, 0, 20}},
			{"A", 0, 0, -1, -1, [4]float64{0, 0, -1, 20}},
		}
		for _, c := range cases {
			fixtureAssert(!i.Update(c.text, c.a, c.f, c.ms, c.me, true, 0, c.rect), "invalid Update accepted")
			fixtureAssert(inputObserve(i) == before, "invalid Update changed native state")
		}
		i.Close()
		i.Close()
		fixtureAssert(!i.Update("A", 0, 0, -1, -1, true, 0, rect), "closed update accepted")
		fixtureAssert(!i.UpdateClient(1, "A", 0, 0, -1, -1, true, 0, rect), "closed client update accepted")
		accepted, err := i.UpdateGeometry("A", 0, 0, -1, -1, true, 0, rect, Geometry{Viewport: [4]float64{0, 0, 10, 10}})
		fixtureAssert(!accepted && err != nil, "closed geometry publication succeeded")
		i.EnableDeltas()
		i.RequestLimit(1, 64)
		i.Caret(rect)
		fixtureAssert(!i.Settle(0), "closed settle accepted")
		fixtureAssert(i.Probe(0) == "" && i.ProbeInactive() == "" && i.QueryProbe(0, 0, [2]float64{}) == "", "closed probe available")
		fixtureAssert(i.RenderedRectToScreen(rect, [2]int{100, 100}) == [4]float64{}, "closed coordinate oracle available")
	})
	run("geometry-result", func() {
		view := fixtureInputView()
		defer fixtureReleaseInputView(view)
		i := Attach(view, func(Update) bool { return true }, func() {})
		fixtureAssert(i != nil, "attachment failed")
		defer i.Close()
		i.EnableDeltas()
		g := Geometry{Viewport: [4]float64{0, 0, 100, 100}, FocusLine: -1}
		accepted, err := i.UpdateGeometry("A", 1, 1, -1, -1, true, 0, rect, g)
		fixtureAssert(accepted && err == nil, "valid geometry refused")
		fixtureInputEdit(i, "B", 1, 0, 0)
		before := inputObserve(i)
		accepted, err = i.UpdateGeometry("A", 1, 1, -1, -1, true, 0, rect, g)
		fixtureAssert(!accepted && err == nil, "stale geometry rejection was hidden")
		fixtureAssert(inputObserve(i) == before, "stale geometry changed state")
		accepted, err = i.UpdateGeometry("A\x00B", 0, 0, -1, -1, true, 1, rect, g)
		fixtureAssert(!accepted && err != nil, "NUL geometry admitted")
	})
	run("view-lifetime", func() {
		view := fixtureInputView()
		i := Attach(view, func(Update) bool { return true }, func() {})
		fixtureAssert(i != nil, "attachment failed")
		defer i.Close()
		fixtureReleaseInputView(view)
		fixtureAssert(fixtureInputViews() == 1, "external release destroyed borrowed view")
		fixtureAssert(i.Update("retained", 0, 0, -1, -1, true, 0, rect), "retained view unavailable")
		i.Close()
		fixtureAssert(fixtureInputViews() == 0, "explicit close did not release view")
	})
	run("attach-failure", func() {
		view := fixtureRejectedInputView()
		fixtureAssert(view != nil, "collision fixture allocation failed")
		defer fixtureReleaseInputView(view)
		fixtureAssert(Attach(view, func(Update) bool { return true }, func() {}) == nil, "incompatible subclass accepted")
		fixtureAssert(fixtureInputViews() == 1, "rejected attachment changed external ownership")
	})
	run("observer-layer", func() {
		view := fixtureInputView()
		defer fixtureReleaseInputView(view)
		i := Attach(view, func(Update) bool { return true }, func() {})
		fixtureAssert(i != nil, "attachment failed")
		defer i.Close()
		fixtureAssert(fixtureOverlayInputView(view), "observer overlay failed")
		i.Close()
		fixtureAssert(fixtureInputDelegates(view), "close removed observer layer or inherited input methods failed superclass delegation")
		fixtureAssert(Attach(view, func(Update) bool { return true }, func() {}) == nil, "attachment stacked above retired input methods")
	})
	run("deltas", func() {
		view := fixtureInputView()
		defer fixtureReleaseInputView(view)
		shadow := "A😀Z"
		updates := []Update{}
		commits := 0
		allow := true
		i := Attach(view, func(u Update) bool {
			if !allow {
				return false
			}
			expanded, err := ExpandDelta(u, shadow)
			fixtureAssert(err == nil, "invalid admitted delta")
			fixtureAssert(expanded.BeforeText == shadow, "wrong whole-shadow before bytes")
			_, err = Decode(expanded)
			fixtureAssert(err == nil, "surrogate-split native delta")
			shadow = expanded.Text
			updates = append(updates, u)
			return true
		}, func() { commits++ })
		fixtureAssert(i != nil, "attachment failed")
		defer i.Close()
		i.EnableDeltas()
		fixtureAssert(i.UpdateClient(7, shadow, 1, 3, -1, -1, true, 0, rect), "initial delta context refused")
		fixtureInputEdit(i, "😁", 1, 2, 1)
		fixtureAssert(shadow == "A😁Z" && len(updates) == 1 && updates[0].ReplaceStart == 1 && updates[0].ReplaceEnd == 3, "known replacement range lost")
		fixtureInputEdit(i, "", 0, 0, 2)
		fixtureAssert(shadow == "A😀Z" && commits == 2, "bulk cancellation did not restore exact surrogate baseline")
		fixtureInputEdit(i, "日本", 1, 2, 1)
		fixtureInputEdit(i, "日本", -1, 0, 0)
		last := updates[len(updates)-1]
		fixtureAssert(last.Kind == "insert" && last.BeforeText == "日本" && last.Text == "日本", "same-visible composition commit discarded actual replacement")
		before := inputObserve(i)
		allow = false
		fixtureInputEdit(i, "refused", 0, 0, 1)
		fixtureAssert(inputObserve(i) == before && commits == 4, "failed preflight mutated native state")
		fixtureInputEdit(i, "A\x00B", 0, 0, 0)
		fixtureAssert(inputObserve(i) == before, "bulk NUL check accepted native edit")
		allow = true
		shadow = strings.Repeat("a", 1<<20) + "😀Z"
		fixtureAssert(i.UpdateClient(7, shadow, 0, 0, -1, -1, true, uint64(commits), rect), "large exact publication refused")
		fixtureInputEdit(i, "a", 1<<20-1, 1, 0)
		last = updates[len(updates)-1]
		fixtureAssert(last.ReplaceStart == 1<<20-1 && last.ReplaceEnd == 1<<20 && last.BeforeText == "a" && last.Text == "a", "large same-byte edit did not use known replacement span")
		fixtureAssert(inputObserve(i).Text == shadow, "large same-byte edit changed document")
		i.RequestLimit(7, 64)
		shadow = "A"
		fixtureAssert(i.UpdateClient(7, shadow, 0, 0, -1, -1, true, uint64(commits), rect), "small per-client context refused")
		before = inputObserve(i)
		count := len(updates)
		for _, value := range []string{strings.Repeat("a", MaxDocumentBytes+1), strings.Repeat("界", 30), strings.Repeat("a", 64)} {
			fixtureInputEdit(i, value, 0, 0, 0)
			fixtureAssert(inputObserve(i) == before && len(updates) == count, "oversized incoming/prospective text reached admission or mutated state")
		}
	})
	run("secure-query", func() {
		view := fixtureInputView()
		defer fixtureReleaseInputView(view)
		edits := 0
		i := Attach(view, func(Update) bool { edits++; return true }, func() {})
		fixtureAssert(i != nil, "secure attachment failed")
		defer i.Close()
		i.SetSecureClient(7)
		fixtureAssert(i.UpdateClient(7, "secret😀", 8, 8, -1, -1, true, 0, rect), "secure publication failed")
		var query struct {
			Substring string `json:"substring"`
			Range     [2]int `json:"substring_range"`
		}
		fixtureAssert(json.Unmarshal([]byte(fixtureInputSubstring(i, 0, 8)), &query) == nil && query.Substring == "" && query.Range[0] == -1, "secure substring exposed")
		fixtureAssert(inputObserve(i).Text == "", "secure fixture snapshot exposed")
		i.SetSecureClient(0)
		fixtureAssert(!i.UpdateClient(9, "bad", 99, 99, -1, -1, true, 0, rect), "invalid publication accepted")
		fixtureAssert(inputObserve(i).Text == "", "refused switch cleared protection")
		// Native editing still reaches the application with correct UTF-16 semantics.
		fixtureInputEdit(i, "X", 0, 1, 0)
		fixtureAssert(edits == 1, "secure editing blocked")
		fixtureAssert(i.UpdateClient(9, "public", 0, 0, -1, -1, true, 1, rect), "public switch failed")
		fixtureAssert(inputObserve(i).Text == "public", "protection leaked into next client")
		i.SetSecureClient(7)
		fixtureAssert(i.UpdateClient(7, "next-secret", 0, 0, -1, -1, true, 1, rect), "public-to-secure switch failed")
		fixtureAssert(inputObserve(i).Text == "", "new secure client was exposed")
		fixtureAssert(i.UpdateClient(7, "next-secret", 0, 0, -1, -1, false, 1, rect), "secure deactivation failed")
		fixtureAssert(json.Unmarshal([]byte(fixtureInputSubstring(i, 0, 11)), &query) == nil && query.Substring == "" && query.Range[0] == -1, "new secure substring exposed")
	})
	return results
}
func TestNativeInputPublicationValidationAndClosedMethods(t *testing.T) {
	if err := nativeFixtureResults["validation"]; err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputGeometryRejectionResult(t *testing.T) {
	if err := nativeFixtureResults["geometry-result"]; err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputOwnsViewUntilExplicitClose(t *testing.T) {
	if err := nativeFixtureResults["view-lifetime"]; err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputAttachFailureReleasesBorrowedView(t *testing.T) {
	if err := nativeFixtureResults["attach-failure"]; err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputObserverOverlaySurvivesClose(t *testing.T) {
	if err := nativeFixtureResults["observer-layer"]; err != nil {
		t.Fatal(err)
	}
}
func TestNativeInputKnownRangeBulkCancelAndAtomicRefusal(t *testing.T) {
	if err := nativeFixtureResults["deltas"]; err != nil {
		t.Fatal(err)
	}
}

func TestNativeInputSecureSubstringAndClientSwitch(t *testing.T) {
	if err := nativeFixtureResults["secure-query"]; err != nil {
		t.Fatal(err)
	}
}
