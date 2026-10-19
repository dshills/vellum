package calculator

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/command"
)

// Commands routes keyboard actions through the same Press validation as buttons.
// A calculator owns its window, so its basic keys precede focused buttons.
func (m *Model) Commands() *command.Registry {
	r := command.New()
	keys := []struct{ chord, key string }{{"Backspace", "delete"}, {"Escape", "AC"}, {"ReturnEnter", "="}, {"Enter", "="}, {"=", "="}, {"Shift++", "+"}, {"+", "+"}, {"-", "−"}, {"*", "×"}, {"Shift+*", "×"}, {"/", "÷"}, {"%", "%"}, {"Shift+%", "%"}, {".", "."}, {"Meta+Y", "history"}}
	for i := 0; i < 10; i++ {
		s := fmt.Sprint(i)
		keys = append(keys, struct{ chord, key string }{s, s})
	}
	for i, k := range keys {
		key := k.key
		if err := r.Register(command.Command{ID: fmt.Sprintf("calculator.key.%d", i), Label: key, Shortcut: k.chord, Execute: func(context.Context) error { m.Press(key); return nil }}); err != nil {
			panic(err)
		}
	}
	return r
}
