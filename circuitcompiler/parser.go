package circuitcompiler

import (
	"errors"
	"io"
)

type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

type Constraint struct {
	// v1 op v2 = out
	Op      Token
	V1      Token
	V2      Token
	Out     Token
	Literal string
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) scan() (tok Token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.Scan()

	p.buf.tok, p.buf.lit = tok, lit

	return
}

func (p *Parser) unscan() {
	p.buf.n = 1
}

func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) ParseLine() (*Constraint, error) {
	/*
		in this version,
		line will be for example s3 = s1 * s4
		this is:
		val op val op val
		ident op ident op ident
	*/
	c := &Constraint{}
	var lit string
	c.Out, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	_, lit = p.scanIgnoreWhitespace() // skip =
	c.Literal += lit
	c.V1, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	c.Op, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	c.V2, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	if c.Out == EOF {
		return nil, errors.New("eof in parseline")
	}
	return c, nil
}

func (p *Parser) Parse() (*Circuit, error) {
	circuit := &Circuit{}
	for {
		constraint, err := p.ParseLine()
		if err != nil {
			// return circuit, err
			break
		}
		circuit.Constraints = append(circuit.Constraints, *constraint)
	}
	return circuit, nil
}
