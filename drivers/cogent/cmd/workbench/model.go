package main

import (
	"bytes"
	"github.com/dshills/vellum/virtual"
	"fmt"
	"strconv"
)

type part struct {
	Value             uint32
	Rotation, Package uint16
}
type partSource struct {
	parts    []part
	revision uint64
}

func newParts(count int) *partSource {
	s := &partSource{parts: make([]part, count), revision: 1}
	for i := range s.parts {
		s.parts[i] = part{Value: uint32(100 + i%997), Rotation: uint16(i % 4 * 90), Package: uint16(i % 2)}
	}
	return s
}
func (s *partSource) Len() int { return len(s.parts) }
func (s *partSource) Row(i int) virtual.Row {
	p := s.parts[i]
	pkg := "0603"
	if p.Package != 0 {
		pkg = "0805"
	}
	return virtual.Row{ID: "component:" + strconv.Itoa(i+1), Cells: []string{fmt.Sprintf("R%07d", i+1), strconv.Itoa(int(p.Value)) + " Ω", pkg, strconv.Itoa(int(p.Rotation)) + "°"}}
}
func (s *partSource) Less(i, j, column int) bool {
	a, b := s.parts[i], s.parts[j]
	switch column {
	case 1:
		return a.Value < b.Value
	case 2:
		return a.Package < b.Package
	case 3:
		return a.Rotation < b.Rotation
	default:
		return i < j
	}
}
func (s *partSource) Matches(i int, q string) bool {
	var storage [128]byte
	dst := storage[:0]
	dst = append(dst, 'r')
	n := i + 1
	digits := 1
	for x := n; x >= 10; x /= 10 {
		digits++
	}
	for j := digits; j < 7; j++ {
		dst = append(dst, '0')
	}
	dst = strconv.AppendInt(dst, int64(n), 10)
	dst = append(dst, ' ')
	p := s.parts[i]
	dst = strconv.AppendInt(dst, int64(p.Value), 10)
	dst = append(dst, []byte(" ω ")...)
	if p.Package == 0 {
		dst = append(dst, []byte("0603 ")...)
	} else {
		dst = append(dst, []byte("0805 ")...)
	}
	dst = strconv.AppendInt(dst, int64(p.Rotation), 10)
	dst = append(dst, []byte("°")...)
	return bytes.Contains(dst, []byte(q))
}
func (s *partSource) Snapshot() *partSource {
	return &partSource{parts: append([]part(nil), s.parts...), revision: s.revision}
}

func (s *partSource) Revision() uint64 { return s.revision }
