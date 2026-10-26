//go:build darwin && cgo

package cogent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixturePDF() []byte {
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 400] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>", "<< /Length 42 >>\nstream\nBT /F1 18 Tf 30 300 Td (Vellum PDF) Tj ET\nendstream"}
	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, o := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(offsets))
	for _, o := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", o)
	}
	fmt.Fprintf(&b, "trailer\n<< /Root 1 0 R /Size %d >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref)
	return []byte(b.String())
}
func TestNativePDFReadAdmission(t *testing.T) {
	p := filepath.Join(t.TempDir(), "a.pdf")
	if e := os.WriteFile(p, fixturePDF(), 0600); e != nil {
		t.Fatal(e)
	}
	if !nativePDFDecodeFixture(p, 1) {
		t.Fatal("valid PDF rejected")
	}
	if e := os.WriteFile(p, []byte("broken"), 0600); e != nil {
		t.Fatal(e)
	}
	if !nativePDFDecodeFixture(p, 0) {
		t.Fatal("invalid PDF accepted")
	}
	if !nativePDFDecodeFixture(filepath.Dir(p), 0) {
		t.Fatal("directory admitted")
	}
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	e = f.Truncate(64<<20 + 1)
	f.Close()
	if e != nil {
		t.Fatal(e)
	}
	if !nativePDFDecodeFixture(p, 0) {
		t.Fatal("oversize admitted")
	}
	if !nativeFrameFixture() {
		t.Fatal("raw frame clipping")
	}
}
