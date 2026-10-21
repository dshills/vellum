package document

import (
	"os"
	"testing"
)

func TestWindowsDirectorySyncRetainsUnexpectedHandleErrors(t *testing.T) {
	directory, err := os.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := syncParentDirectory(directory); err != nil {
		t.Fatal("valid directory sync prevented a successful save", err)
	}
	if err := directory.Close(); err != nil {
		t.Fatal(err)
	}
	if err := syncParentDirectory(directory); err == nil {
		t.Fatal("closed directory handle error was suppressed")
	}
}
