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
	_, lit = p.scanIgnoreWhitespace()
	c.Op = lit
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

func existInArray(arr []string, elem string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
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
	nInputs := 0
	for {
		constraint, err := p.ParseLine()
		if err != nil {
			break
		}
		if constraint.Literal == "func" {
			// one constraint for each input
			for _, in := range constraint.Inputs {
				newConstr := &Constraint{
					Op:  "in",
					Out: in,
				}
				circuit.Constraints = append(circuit.Constraints, *newConstr)
				nInputs++
			}
			circuit.Inputs = constraint.Inputs
			continue
		}
		circuit.Constraints = append(circuit.Constraints, *constraint)
		isVal, _ := isValue(constraint.V1)
		if !isVal {
			circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.V1)
		}
		isVal, _ = isValue(constraint.V2)
		if !isVal {
			circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.V2)
		}
		if constraint.Out == "out" {
			// if Out is "out", put it after the inputs
			if !existInArray(circuit.Signals, constraint.Out) {
				signalsCopy := copyArray(circuit.Signals)
				var auxSignals []string
				auxSignals = append(auxSignals, signalsCopy[0:nInputs+1]...)
				auxSignals = append(auxSignals, constraint.Out)
				auxSignals = append(auxSignals, signalsCopy[nInputs+1:]...)
				circuit.Signals = auxSignals
			}
		} else {
			circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.Out)
		}
	}
	circuit.NVars = len(circuit.Signals)
	circuit.NSignals = len(circuit.Signals)
	circuit.NPublic = 0
	return circuit, nil
}
func copyArray(in []string) []string { // tmp
	var out []string
	for _, e := range in {
		out = append(out, e)
	}
	return out
}
