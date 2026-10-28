package main

import (
	"context"
	vdoc "github.com/dshills/vellum/document"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStudyRoundTrip(t *testing.T) {
	m := newWorkbench()
	if e := m.add(); e != nil {
		t.Fatal(e)
	}
	s := m.captureStudy()
	b, e := encodeStudy(s)
	if e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(t.TempDir(), "study.vforge")
	store := vdoc.FileStore{}
	if _, e = store.Write(context.Background(), p, b, studyLimit); e != nil {
		t.Fatal(e)
	}
	b, e = store.Read(context.Background(), p, studyLimit)
	if e != nil {
		t.Fatal(e)
	}
	got, e := decodeStudy(b)
	if e != nil {
		t.Fatal(e)
	}
	n := newWorkbench()
	if e = n.installStudy(got); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(n.captureStudy(), s) {
		t.Fatal("round trip changed study")
	}
	if e = n.add(); e != nil {
		t.Fatal(e)
	}
}
func TestInvalidStudyLeavesModel(t *testing.T) {
	m := newWorkbench()
	before := m.captureStudy()
	s := m.captureStudy()
	s.Document.Features[0].Object = "missing"
	if m.installStudy(s) == nil {
		t.Fatal("accepted missing body")
	}
	if !reflect.DeepEqual(before, m.captureStudy()) {
		t.Fatal("changed on rejection")
	}
	if _, e := decodeStudy([]byte(`{"Version":99}`)); e == nil {
		t.Fatal("accepted version")
	}
}

func TestRecoveryRoundTripAndCleanRemoval(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	r, e := openRecovery()
	if e != nil {
		t.Fatal(e)
	}
	m := newWorkbench()
	m.recovery = r
	if other, e := openRecovery(); e == nil {
		other.stop(newWorkbench())
		t.Fatal("second writer admitted")
	}
	if e = m.add(); e != nil {
		t.Fatal(e)
	}
	path := r.path
	r.stop(m)
	b, e := (vdoc.FileStore{}).Read(context.Background(), path, studyLimit)
	if e != nil {
		t.Fatal(e)
	}
	s, e := decodeStudy(b)
	if e != nil || len(s.Document.Objects) != 8 {
		t.Fatalf("recovery: %v", e)
	}
	r, e = openRecovery()
	if e != nil {
		t.Fatal(e)
	}
	m.recovery = r
	m.cleanClose = true
	r.stop(m)
	if _, e = (vdoc.FileStore{}).Read(context.Background(), path, studyLimit); e == nil {
		t.Fatal("clean close kept recovery")
	}
}
