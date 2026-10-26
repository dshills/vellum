package app

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharePayloadValidation(t *testing.T) {
	valid := ShareOptions{AnchorID: "share", Items: []ShareItem{{ShareText, "Synthetic text"}, {ShareURL, "https://example.com"}, {ShareFile, filepath.Join(t.TempDir(), "explicit.txt")}}}
	if e := valid.Validate(); e != nil {
		t.Fatal(e)
	}
	for _, o := range []ShareOptions{{Items: valid.Items}, {AnchorID: "share"}, {AnchorID: "share", Items: make([]ShareItem, 17)}, {AnchorID: "share", Items: []ShareItem{{"unknown", "x"}}}, {AnchorID: "share", Items: []ShareItem{{ShareFile, "relative.txt"}}}, {AnchorID: "share", Items: []ShareItem{{ShareURL, "file:///tmp/a"}}}, {AnchorID: "share", Items: []ShareItem{{ShareText, "x\x00y"}}}} {
		if o.Validate() == nil {
			t.Fatal("bad share admitted", o)
		}
	}
	if !errors.Is((ShareOptions{AnchorID: "share", Items: []ShareItem{{ShareText, strings.Repeat("x", 65537)}}}).Validate(), ErrServiceTooLarge) {
		t.Fatal("payload limit")
	}
}
