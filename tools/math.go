package tools

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokNumber tokenType = iota
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokPercent
	tokLParen
	tokRParen
	tokMathFloor
	tokEOF
)

type token struct {
	typ tokenType
	val float64
}

type lexer struct {
	input []rune
	pos   int
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) skipSpaces() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *lexer) hasPrefix(s string) bool {
	runes := []rune(s)
	if l.pos+len(runes) > len(l.input) {
		return false
	}
	for i, r := range runes {
		if unicode.ToLower(l.input[l.pos+i]) != unicode.ToLower(r) {
			return false
		}
	}
	return true
}

func (l *lexer) next() token {
	l.skipSpaces()
	if l.pos >= len(l.input) {
		return token{typ: tokEOF}
	}

	ch := l.input[l.pos]

	switch ch {
	case '+':
		l.pos++
		return token{typ: tokPlus}
	case '-':
		l.pos++
		return token{typ: tokMinus}
	case '*', '×':
		l.pos++
		return token{typ: tokStar}
	case '/':
		l.pos++
		return token{typ: tokSlash}
	case '%':
		l.pos++
		return token{typ: tokPercent}
	case '(', '[':
		l.pos++
		return token{typ: tokLParen}
	case ')', ']':
		l.pos++
		return token{typ: tokRParen}
	}

	if l.hasPrefix("Math.floor") {
		l.pos += 10
		return token{typ: tokMathFloor}
	}

	if unicode.IsDigit(ch) || ch == '.' {
		start := l.pos
		for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
			l.pos++
		}
		val, err := strconv.ParseFloat(string(l.input[start:l.pos]), 64)
		if err != nil {
			return token{typ: tokEOF}
		}
		return token{typ: tokNumber, val: val}
	}

	l.pos++
	return l.next()
}

type parser struct {
	lex     *lexer
	current token
}

func (p *parser) advance() {
	p.current = p.lex.next()
}

func (p *parser) expect(typ tokenType) error {
	if p.current.typ != typ {
		return fmt.Errorf("expected token %d, got %d at pos %d", typ, p.current.typ, p.lex.pos)
	}
	p.advance()
	return nil
}

func (p *parser) expression() (float64, error) {
	left, err := p.term()
	if err != nil {
		return 0, err
	}
	for p.current.typ == tokPlus || p.current.typ == tokMinus {
		op := p.current.typ
		p.advance()
		right, err := p.term()
		if err != nil {
			return 0, err
		}
		if op == tokPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *parser) term() (float64, error) {
	left, err := p.unary()
	if err != nil {
		return 0, err
	}
	for p.current.typ == tokStar || p.current.typ == tokSlash || p.current.typ == tokPercent {
		op := p.current.typ
		p.advance()
		right, err := p.unary()
		if err != nil {
			return 0, err
		}
		switch op {
		case tokStar:
			left *= right
		case tokSlash:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		case tokPercent:
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		}
	}
	return left, nil
}

func (p *parser) unary() (float64, error) {
	if p.current.typ == tokMinus {
		p.advance()
		val, err := p.unary()
		return -val, err
	}
	return p.primary()
}

func (p *parser) primary() (float64, error) {
	switch p.current.typ {
	case tokNumber:
		val := p.current.val
		p.advance()
		return val, nil
	case tokMathFloor:
		p.advance()
		if err := p.expect(tokLParen); err != nil {
			return 0, err
		}
		val, err := p.expression()
		if err != nil {
			return 0, err
		}
		if err := p.expect(tokRParen); err != nil {
			return 0, err
		}
		return math.Floor(val), nil
	case tokLParen:
		p.advance()
		val, err := p.expression()
		if err != nil {
			return 0, err
		}
		if err := p.expect(tokRParen); err != nil {
			return 0, err
		}
		return val, nil
	default:
		return 0, fmt.Errorf("unexpected token %d at pos %d", p.current.typ, p.lex.pos)
	}
}

func EvalMath(expr string) (string, error) {
	expr = strings.ReplaceAll(expr, "×", "*")
	lex := &lexer{input: []rune(expr)}
	p := &parser{lex: lex, current: lex.next()}

	result, err := p.expression()
	if err != nil {
		return "", fmt.Errorf("eval %q: %w", expr, err)
	}

	rounded := math.Round(result)
	if math.Abs(result-rounded) < 1e-6 {
		return fmt.Sprintf("%d", int64(rounded)), nil
	}
	return fmt.Sprintf("%d", int64(result)), nil
}
