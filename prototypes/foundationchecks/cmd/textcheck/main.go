// textcheck diagnoses the exact buffer primitives used by the multiline editor.
// This is not native keyboard, IME, or screen-reader verification.
package main

import (
	"cogentcore.org/core/text/lines"
	"encoding/json"
	"flag"
	"os"
)

type result struct {
	Name     string `json:"name"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Observed string `json:"observed"`
	Pass     bool   `json:"pass"`
}

func main() {
	flag.Bool("nogui", true, "use toolkit offscreen initialization")
	grapheme := flag.Bool("grapheme", false, "use patched editor grapheme movement")
	flag.Parse()
	var results []result
	for _, s := range []string{"AB", "e\u0301", "👩‍💻", "👨‍👩‍👧‍👦", "🇺🇳", "नमस्ते"} {
		ls := lines.NewLines().SetText([]byte(s + "\n"))
		end := ls.EndPos()
		start := ls.MoveBackward(end, 1)
		if *grapheme {
			start = ls.MoveBackwardGrapheme(end, 1)
		}
		ls.DeleteText(start, end)
		got := string(ls.Line(0))
		// The ASCII control leaves A; each other probe tests only whether a single
		// cursor move can end inside its final grapheme, using a separate segmenter
		// to calculate the required deletion result.
		want := previousGrapheme(s)
		results = append(results, result{"multiline_backspace_primitive", s, want, got, got == want})
		ls.Close()
	}
	for _, s := range []string{"one\ntwo", "one\ntwo\n", "one\ntwo\n\n", "日本語\n👩‍💻"} {
		ls := lines.NewLines().SetText([]byte(s))
		got := string(ls.Text())
		results = append(results, result{"multiline_exact_roundtrip", s, s, got, got == s})
		ls.Close()
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(map[string]any{"scope": "buffer primitives used by editor, not OS input", "grapheme_movement": *grapheme, "results": results}); err != nil {
		panic(err)
	}
}
