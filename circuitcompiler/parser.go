package circuitcompiler

import (
	"errors"
	"io"
	"regexp"
	"strings"
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
	V1      string
	V2      string
	Out     string
	Literal string

	Inputs []string // in func delcaration case
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
		val eq val op val
	*/
	c := &Constraint{}
	tok, lit := p.scanIgnoreWhitespace()
	c.Out = lit
	c.Literal += lit

	if c.Literal == "func" {
		// format: `func name(in):`
		line, err := p.s.r.ReadString(':')
		if err != nil {
			return c, err
		}
		// read string inside ( )
		rgx := regexp.MustCompile(`\((.*?)\)`)
		insideParenthesis := rgx.FindStringSubmatch(line)
		varsString := strings.Replace(insideParenthesis[1], " ", "", -1)
		c.Inputs = strings.Split(varsString, ",")
		return c, nil
	}

	_, lit = p.scanIgnoreWhitespace() // skip =
	c.Literal += lit

	// v1
	_, lit = p.scanIgnoreWhitespace()
	c.V1 = lit
	c.Literal += lit
	// operator
	c.Op, lit = p.scanIgnoreWhitespace()
	c.Literal += lit
	// v2
	_, lit = p.scanIgnoreWhitespace()
	c.V2 = lit
	c.Literal += lit
	if tok == EOF {
		return nil, errors.New("eof in parseline")
	}
	return c, nil
}

func addToArrayIfNotExist(arr []string, elem string) []string {
	for _, v := range arr {
		if v == elem {
			return arr
		}
	}
	arr = append(arr, elem)
	return arr
}
func (p *Parser) Parse() (*Circuit, error) {
	circuit := &Circuit{}
	circuit.Signals = append(circuit.Signals, "one")
	for {
		constraint, err := p.ParseLine()
		if err != nil {
			break
		}
		if constraint.Literal == "func" {
			circuit.Inputs = constraint.Inputs
			continue
		}
		circuit.Constraints = append(circuit.Constraints, *constraint)
		circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.V1)
		circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.V2)
		circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.Out)
	}
	return circuit, nil
}
