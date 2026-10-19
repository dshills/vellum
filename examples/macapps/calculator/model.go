// Package calculator is a small, independently reusable Vellum example.
package calculator

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

const MaxHistory = 32
const maxEntry = 18
const maxParts = 127
const maxNumberBytes = 64
const maxExponent = 300
const maxRationalBits = 4096

// Budget before parsing scientific notation: a short exponent token must not
// allocate an arbitrarily large integer on the UI owner.
func checkNumberToken(s string) error {
	if len(s) == 0 || len(s) > maxNumberBytes {
		return fmt.Errorf("number exceeds budget")
	}
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exponent, err := strconv.Atoi(s[i+1:])
		if err != nil || exponent < -maxExponent || exponent > maxExponent {
			return fmt.Errorf("number exceeds exponent budget")
		}
	}
	return nil
}
func checkRational(r *big.Rat) error {
	if r.Num().BitLen() > maxRationalBits || r.Denom().BitLen() > maxRationalBits {
		return fmt.Errorf("arithmetic exceeds budget")
	}
	return nil
}
func (m *Model) setNumber(r *big.Rat) bool {
	if err := checkRational(r); err != nil {
		m.failure = err.Error()
		return false
	}
	s := decimal(r)
	if err := checkNumberToken(s); err != nil {
		m.failure = err.Error()
		return false
	}
	m.entry = s
	return true
}

type Calculation struct{ Expression, Result string }

// Model is owned by the application's UI owner, as are its view callbacks.
// The example is a bounded basic calculator, not a scientific calculator.
type Model struct {
	entry, expression, failure string
	parts                      []string
	fresh, result              bool
	lastOp, lastOperand        string
	history                    []Calculation
	HistoryOpen                bool
}

func New() *Model { return &Model{entry: "0"} }
func (m *Model) Display() string {
	if m.failure != "" {
		return "Error"
	}
	return m.entry
}
func (m *Model) Expression() string {
	if m.result || m.failure != "" {
		return pretty(m.expression)
	}
	s := strings.Join(m.parts, "")
	if !m.fresh {
		s += m.entry
	}
	if s == m.entry {
		return ""
	}
	return pretty(s)
}
func (m *Model) History() []Calculation { return append([]Calculation(nil), m.history...) }
func pretty(s string) string            { return strings.NewReplacer("*", "×", "/", "÷", "-", "−").Replace(s) }
func decimal(r *big.Rat) string {
	if r.Sign() == 0 {
		return "0"
	}
	return new(big.Float).SetPrec(128).SetRat(r).Text('g', 11)
}
func number(s string) (*big.Rat, error) {
	if err := checkNumberToken(s); err != nil {
		return nil, err
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, fmt.Errorf("invalid number")
	}
	if err := checkRational(r); err != nil {
		return nil, err
	}
	return r, nil
}
func evaluate(parts []string) (*big.Rat, error) {
	if len(parts) == 0 || len(parts)%2 != 1 || len(parts) > maxParts+1 {
		return nil, fmt.Errorf("invalid expression")
	}
	term, err := number(parts[0])
	if err != nil {
		return nil, err
	}
	sum := new(big.Rat)
	sign := 1
	for i := 1; i < len(parts); i += 2 {
		r, err := number(parts[i+1])
		if err != nil {
			return nil, err
		}
		switch parts[i] {
		case "*":
			term.Mul(term, r)
		case "/":
			if r.Sign() == 0 {
				return nil, fmt.Errorf("cannot divide by zero")
			}
			term.Quo(term, r)
		case "+", "-":
			if sign < 0 {
				sum.Sub(sum, term)
			} else {
				sum.Add(sum, term)
			}
			term = r
			sign = 1
			if parts[i] == "-" {
				sign = -1
			}
		default:
			return nil, fmt.Errorf("invalid operator")
		}
		if err := checkRational(term); err != nil {
			return nil, err
		}
		if err := checkRational(sum); err != nil {
			return nil, err
		}
	}
	if sign < 0 {
		sum.Sub(sum, term)
	} else {
		sum.Add(sum, term)
	}
	if err := checkRational(sum); err != nil {
		return nil, err
	}
	return sum, nil
}
func (m *Model) reset() {
	m.entry = "0"
	m.expression = ""
	m.failure = ""
	m.parts = nil
	m.fresh = false
	m.result = false
	m.lastOp = ""
	m.lastOperand = ""
}

// Press accepts only the visible basic keys. Both native buttons and keyboard
// commands use this path; arbitrary expressions cannot be evaluated as Go code.
func (m *Model) Press(key string) bool {
	key = strings.NewReplacer("×", "*", "÷", "/", "−", "-").Replace(key)
	if len(key) == 1 && key[0] >= '0' && key[0] <= '9' || key == "." {
		if m.failure != "" || m.result {
			m.reset()
		}
		if m.fresh || strings.ContainsAny(m.entry, "eE") {
			m.entry = "0"
			m.fresh = false
		}
		if len(m.entry) >= maxEntry {
			return true
		}
		if key == "." {
			if !strings.ContainsAny(m.entry, ".e") {
				m.entry += "."
			}
			return true
		}
		if m.entry == "0" {
			m.entry = key
		} else if m.entry == "-0" {
			m.entry = "-" + key
		} else {
			m.entry += key
		}
		return true
	}
	if key == "AC" || key == "Escape" {
		m.reset()
		return true
	}
	if key == "history" {
		m.HistoryOpen = !m.HistoryOpen
		return true
	}
	if m.failure != "" {
		return key == "delete" || strings.Contains("+-*/=%±", key) && len(key) == 1
	}
	switch key {
	case "+", "-", "*", "/":
		if m.result {
			m.parts = nil
			m.result = false
			m.expression = ""
		}
		if m.fresh {
			m.parts[len(m.parts)-1] = key
			return true
		}
		if len(m.parts)+2 > maxParts {
			m.failure = "Expression is too long"
			return true
		}
		m.parts = append(m.parts, m.entry, key)
		m.fresh = true
		m.lastOp = ""
		return true
	case "±":
		if strings.HasPrefix(m.entry, "-") {
			m.entry = strings.TrimPrefix(m.entry, "-")
		} else {
			m.entry = "-" + m.entry
		}
		m.result = false
		m.fresh = false
		return true
	case "%":
		r, err := number(m.entry)
		if err != nil {
			return true
		}
		r.Quo(r, big.NewRat(100, 1))
		if len(m.parts) > 0 && !m.fresh && (m.parts[len(m.parts)-1] == "+" || m.parts[len(m.parts)-1] == "-") {
			base, err := evaluate(m.parts[:len(m.parts)-1])
			if err != nil {
				m.failure = err.Error()
				return true
			}
			r.Mul(base, r)
		}
		if !m.setNumber(r) {
			return true
		}
		m.result = false
		m.fresh = false
		return true
	case "delete":
		if m.result {
			m.parts = nil
			m.result = false
			m.expression = ""
		}
		if m.fresh {
			m.entry = m.parts[len(m.parts)-2]
			m.parts = m.parts[:len(m.parts)-2]
			m.fresh = false
		} else if strings.ContainsAny(m.entry, "eE") {
			m.entry = "0"
		} else if len(m.entry) > 1 {
			m.entry = m.entry[:len(m.entry)-1]
		} else {
			m.entry = "0"
		}
		if m.entry == "-" || m.entry == "" {
			m.entry = "0"
		}
		return true
	case "=":
		parts := append(append([]string(nil), m.parts...), m.entry)
		if m.result && m.lastOp != "" {
			parts = []string{m.entry, m.lastOp, m.lastOperand}
		}
		expression := strings.Join(parts, "")
		r, err := evaluate(parts)
		m.expression = expression
		if err != nil {
			m.failure = err.Error()
			return true
		}
		if len(parts) > 1 {
			m.lastOp = parts[len(parts)-2]
			m.lastOperand = parts[len(parts)-1]
		}
		if !m.setNumber(r) {
			return true
		}
		m.parts = nil
		m.fresh = false
		m.result = true
		if len(m.history) == MaxHistory {
			copy(m.history, m.history[1:])
			m.history = m.history[:MaxHistory-1]
		}
		m.history = append(m.history, Calculation{pretty(expression), m.entry})
		return true
	}
	return false
}

func (m *Model) UseHistory(index int) {
	if index < 0 || index >= len(m.history) {
		return
	}
	r := m.history[index]
	m.reset()
	m.entry = r.Result
	m.expression = r.Expression
	m.result = true
}
