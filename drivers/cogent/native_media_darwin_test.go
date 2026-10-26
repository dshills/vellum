//go:build darwin && cgo

package cogent

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func silentWAV() []byte {
	d := make([]byte, 44+16000)
	copy(d, "RIFF")
	binary.LittleEndian.PutUint32(d[4:], uint32(len(d)-8))
	copy(d[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(d[16:], 16)
	binary.LittleEndian.PutUint16(d[20:], 1)
	binary.LittleEndian.PutUint16(d[22:], 1)
	binary.LittleEndian.PutUint32(d[24:], 8000)
	binary.LittleEndian.PutUint32(d[28:], 16000)
	binary.LittleEndian.PutUint16(d[32:], 2)
	binary.LittleEndian.PutUint16(d[34:], 16)
	copy(d[36:], "data")
	binary.LittleEndian.PutUint32(d[40:], 16000)
	return d
}
func TestNativeMediaReadAdmission(t *testing.T) {
	p := filepath.Join(t.TempDir(), "silent.wav")
	if e := os.WriteFile(p, silentWAV(), 0600); e != nil {
		t.Fatal(e)
	}
	if !nativeMediaDecodeFixture(p, true) {
		t.Fatal("silent WAV rejected")
	}
	if e := os.WriteFile(p, []byte("bad media"), 0600); e != nil {
		t.Fatal(e)
	}
	if !nativeMediaDecodeFixture(p, false) {
		t.Fatal("broken media admitted")
	}
	if !nativeMediaDecodeFixture(filepath.Dir(p), false) {
		t.Fatal("directory admitted")
	}
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	e = f.Truncate(1<<30 + 1)
	f.Close()
	if e != nil {
		t.Fatal(e)
	}
	if !nativeMediaDecodeFixture(p, false) {
		t.Fatal("oversize media admitted")
	}
}
