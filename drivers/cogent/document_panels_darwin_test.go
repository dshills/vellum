//go:build darwin && cgo

package cogent

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestNativeColorPanelRecipeWithoutPresentation(t *testing.T) {
	valid := `{"Kind":"color","Options":{"Title":"Colors","Color":{"R":255,"G":2,"B":3,"A":128},"ShowsAlpha":true}}`
	if nativeDocumentPanelRecipeFixture(valid) == "" {
		t.Fatal("valid color recipe")
	}
	for _, wire := range []string{`[]`, `{}`, `{"Kind":"color","Options":{}}`, `{"Kind":"color","Options":{"Title":"Colors","Color":{"R":256,"G":0,"B":0,"A":255},"ShowsAlpha":true}}`} {
		if nativeDocumentPanelRecipeFixture(wire) != "" {
			t.Fatal("malformed color recipe")
		}
	}
}
func TestNativeDocumentPanelPreviewCoalescesAndRetires(t *testing.T) {
	withMenuCommitState(t)
	token := serviceToken.Add(1)
	result := make(chan nativeResult, 1)
	q := make(chan string, 1)
	nativeCallbacks.Lock()
	nativeCallbacks.operations[token] = nativeOperation{ctx: context.Background(), result: result}
	nativeCallbacks.Unlock()
	documentPanelPreviews.Lock()
	documentPanelPreviews.events[token] = q
	documentPanelPreviews.Unlock()
	defer func() {
		documentPanelPreviews.Lock()
		delete(documentPanelPreviews.events, token)
		documentPanelPreviews.Unlock()
	}()
	for i := range 1000 {
		deliverDocumentPanelPreview(token, fmt.Sprint(i))
	}
	if len(q) != 1 || <-q != "999" {
		t.Fatal("preview queue is not latest-only")
	}
	nativeServiceFinish(token, nativeResult{value: "done"})
	deliverDocumentPanelPreview(token, "retired")
	if len(q) != 0 {
		t.Fatal("retired preview published")
	}
}
func TestNativeDocumentPanelParentRevokesBeforeAsyncCancellation(t *testing.T) {
	withMenuCommitState(t)
	window, cancelWindow := context.WithCancel(context.Background())
	joined, end := context.WithCancel(context.Background())
	defer end()
	ctx := alertWindowContext{Context: joined, window: window}
	_, err := runDocumentPanel(ctx, nil, nil, func(token uint64, _ []byte) {
		cancelWindow()
		if nativeServiceActive(token) {
			t.Error("retired exact parent remains active")
		}
		nativeServiceFinish(token, nativeResult{value: "late"})
	}, func(token uint64) { nativeServiceFinish(token, nativeResult{err: context.Canceled}) })
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	documentPanelPreviews.Lock()
	remaining := len(documentPanelPreviews.events)
	documentPanelPreviews.Unlock()
	if remaining != 0 {
		t.Fatal("preview registry leaked")
	}
	nativeCallbacks.Lock()
	pending := len(nativeCallbacks.operations)
	nativeCallbacks.Unlock()
	if pending != 0 {
		t.Fatal("terminal operation leaked")
	}
}

func TestNativeDocumentPanelCanceledAdmissionHasNoPreviewEntry(t *testing.T) {
	withMenuCommitState(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := runDocumentPanel(ctx, nil, nil, func(uint64, []byte) { t.Error("canceled admission started native work") }, func(uint64) {})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	documentPanelPreviews.Lock()
	remaining := len(documentPanelPreviews.events)
	documentPanelPreviews.Unlock()
	if remaining != 0 {
		t.Fatal("refused admission leaked preview entry")
	}
}

func TestNativeFontPanelRecipeWithoutPresentation(t *testing.T) {
	valid := `{"Kind":"font","Options":{"Title":"Fonts","Name":"HelveticaNeue","Size":13}}`
	if !strings.Contains(nativeDocumentPanelRecipeFixture(valid), "Use Font") {
		t.Fatal("valid font recipe/action title")
	}
	for _, wire := range []string{`{"Kind":"font","Options":{"Title":"Fonts","Name":"HelveticaNeue","Size":0}}`, `{"Kind":"font","Options":{"Title":"Fonts","Name":"","Size":13}}`} {
		if nativeDocumentPanelRecipeFixture(wire) != "" {
			t.Fatal("bad font recipe")
		}
	}
}
