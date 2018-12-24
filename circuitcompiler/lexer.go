package circuitcompiler

import (
	"bufio"
	"bytes"
	"io"
)

type OperatorSymbol int
type Token int

const (
	ILLEGAL Token = iota
	WS
	EOF

	IDENT // val

	VAR   // var
	CONST // const value

	EQ       // =
	PLUS     // +
	MINUS    // -
	MULTIPLY // *
	DIVIDE   // /
	EXP      // ^

	OUT
)

var eof = rune(0)

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
func isDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9')
}

type Scanner struct {
	r *bufio.Reader
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if isWhitespace(ch) {
		// space
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		// letter
		s.unread()
		return s.scanIndent()
	} else if isDigit(ch) {
		s.unread()
		return s.scanIndent()
	}

	switch ch {
	case eof:
		return EOF, ""
	case '=':
		return EQ, "="
	case '+':
		return PLUS, "+"
	case '-':
		return MINUS, "-"
	case '*':
		return MULTIPLY, "*"
	case '/':
		return DIVIDE, "/"
	case '^':
		return EXP, "^"
	}

	return ILLEGAL, string(ch)
}

func (s *Scanner) scanWhitespace() (token Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	return WS, buf.String()
}

func (s *Scanner) scanIndent() (tok Token, lit string) {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}
	switch buf.String() {
	case "var":
		return VAR, buf.String()
	}

	if len(buf.String()) == 1 {
		return Token(rune(buf.String()[0])), buf.String()
	}
	if buf.String() == "out" {
		return OUT, buf.String()
	}
	return IDENT, buf.String()
}
