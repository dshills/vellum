package main

import "github.com/go-text/typesetting/segmenter"

func previousGrapheme(s string) string {
	r := []rune(s)
	var seg segmenter.Segmenter
	seg.Init(r)
	it := seg.GraphemeIterator()
	last := 0
	for it.Next() {
		last = it.Grapheme().Offset
	}
	return string(r[:last])
}
