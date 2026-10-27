// Package quantity parses bounded arithmetic length expressions without evaluation
// of code, function calls or member access. Canonical lengths are millimeters.
package quantity

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

var units = map[string]float64{"mm": 1, "cm": 10, "m": 1000, "in": 25.4, "ft": 304.8}

func Unit(unit string) (float64, bool) { v, ok := units[unit]; return v, ok }

type value struct {
	n         float64
	dimension int
}
type parser struct {
	s         scanner.Scanner
	lookahead rune
	count     int
	vars      map[string]float64
}

func (p *parser) next() error {
	p.lookahead = p.s.Scan()
	p.count++
	if p.count > 128 {
		return fmt.Errorf("expression is too complex")
	}
	return nil
}
func (p *parser) expr(min, depth int) (value, error) {
	if depth > 16 {
		return value{}, fmt.Errorf("expression is too deep")
	}
	var a value
	var e error
	switch p.lookahead {
	case '+', '-':
		op := p.lookahead
		if e = p.next(); e != nil {
			return a, e
		}
		a, e = p.expr(3, depth+1)
		if e == nil && op == '-' {
			a.n = -a.n
		}
	case '(':
		if e = p.next(); e != nil {
			return a, e
		}
		a, e = p.expr(0, depth+1)
		if e == nil && p.lookahead != ')' {
			e = fmt.Errorf("missing closing parenthesis")
		}
		if e == nil {
			e = p.next()
		}
	case scanner.Int, scanner.Float:
		a.n, e = strconv.ParseFloat(p.s.TokenText(), 64)
		if e == nil {
			e = p.next()
		}
		if e == nil && p.lookahead == scanner.Ident {
			if f, ok := Unit(p.s.TokenText()); ok {
				a.n *= f
				a.dimension = 1
				e = p.next()
			} else {
				e = fmt.Errorf("unknown unit %q", p.s.TokenText())
			}
		}
	case scanner.Ident:
		var ok bool
		a.n, ok = p.vars[p.s.TokenText()]
		a.dimension = 1
		if !ok {
			e = fmt.Errorf("unknown length parameter %q", p.s.TokenText())
		}
		if e == nil {
			e = p.next()
		}
	default:
		e = fmt.Errorf("expected a number, length parameter or parenthesis")
	}
	if e != nil {
		return value{}, e
	}
	if p.lookahead == scanner.Ident {
		if f, ok := Unit(p.s.TokenText()); ok {
			if a.dimension != 0 {
				return value{}, fmt.Errorf("unit suffix requires scalar expression")
			}
			a.n *= f
			a.dimension = 1
			if e = p.next(); e != nil {
				return value{}, e
			}
		}
	}
	for {
		op := p.lookahead
		precedence := 0
		if op == '+' || op == '-' {
			precedence = 1
		} else if op == '*' || op == '/' {
			precedence = 2
		}
		if precedence == 0 || precedence < min {
			break
		}
		if e = p.next(); e != nil {
			return value{}, e
		}
		b, err := p.expr(precedence+1, depth+1)
		if err != nil {
			return value{}, err
		}
		switch op {
		case '+', '-':
			if a.dimension != b.dimension {
				return value{}, fmt.Errorf("add lengths with matching units")
			}
			if op == '+' {
				a.n += b.n
			} else {
				a.n -= b.n
			}
		case '*':
			a.n *= b.n
			a.dimension += b.dimension
		case '/':
			if b.n == 0 {
				return value{}, fmt.Errorf("division by zero")
			}
			a.n /= b.n
			a.dimension -= b.dimension
		}
		if math.IsNaN(a.n) || math.IsInf(a.n, 0) || math.Abs(a.n) > 1e12 || a.dimension < -2 || a.dimension > 2 {
			return value{}, fmt.Errorf("expression exceeds numeric bounds")
		}
	}
	return a, nil
}

// Length accepts units mm/cm/m/in/ft, arithmetic and caller-owned named length
// parameters. A scalar result is interpreted in defaultUnit. Intermediate and
// final values are finite and bounded. Only a length or scalar result is valid.
func Length(text, defaultUnit string, variables map[string]float64) (float64, error) {
	f, ok := Unit(defaultUnit)
	if !ok || len(text) == 0 || len(text) > 512 || len(variables) > 64 {
		return 0, fmt.Errorf("invalid length expression or unit")
	}
	for k, v := range variables {
		if len(k) == 0 || len(k) > 64 || math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e9 {
			return 0, fmt.Errorf("invalid parameter")
		}
		for i, c := range k {
			if !(unicode.IsLetter(c) || c == '_' || i > 0 && unicode.IsDigit(c)) {
				return 0, fmt.Errorf("invalid parameter name")
			}
		}
		if _, reserved := Unit(k); reserved {
			return 0, fmt.Errorf("parameter uses unit name")
		}
	}
	p := parser{vars: variables}
	p.s.Init(strings.NewReader(text))
	p.s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats
	p.s.Error = func(*scanner.Scanner, string) {}
	if e := p.next(); e != nil {
		return 0, e
	}
	v, e := p.expr(0, 0)
	if e != nil {
		return 0, e
	}
	if p.s.ErrorCount > 0 || p.lookahead != scanner.EOF {
		return 0, fmt.Errorf("invalid expression suffix")
	}
	if v.dimension == 0 {
		v.n *= f
	} else if v.dimension != 1 {
		return 0, fmt.Errorf("expression must produce a length")
	}
	if math.IsNaN(v.n) || math.IsInf(v.n, 0) || math.Abs(v.n) > 1e9 {
		return 0, fmt.Errorf("length exceeds bounds")
	}
	return v.n, nil
}
