package circuitcompiler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Parser data structure holds the Scanner and the Parsing functions
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
	}
}

// NewParser creates a new parser from a io.Reader
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

func (p *Parser) scan() (tok Token, lit string) {
	// if there is a token in the buffer return it
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}
	tok, lit = p.s.scan()

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

// parseLine parses the current line
func (p *Parser) parseLine() (*Constraint, error) {
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
		allInputs := strings.Split(varsString, ",")

		// from allInputs, get the private and the public separated
		for _, in := range allInputs {
			if strings.Contains(in, "private") {
				input := strings.Replace(in, "private", "", -1)
				c.PrivateInputs = append(c.PrivateInputs, input)
			} else if strings.Contains(in, "public") {
				input := strings.Replace(in, "public", "", -1)
				c.PublicInputs = append(c.PublicInputs, input)
			} else {
				// TODO give more info about the circuit code error
				fmt.Println("error on declaration of public and private inputs")
				os.Exit(0)
			}
		}
		return c, nil
	}
	if c.Literal == "equals" {
		// format: `equals(a, b)`
		line, err := p.s.r.ReadString(')')
		if err != nil {
			return c, err
		}
		// read string inside ( )
		rgx := regexp.MustCompile(`\((.*?)\)`)
		insideParenthesis := rgx.FindStringSubmatch(line)
		varsString := strings.Replace(insideParenthesis[1], " ", "", -1)
		params := strings.Split(varsString, ",")
		fmt.Println("params", params)
		// TODO
		return c, nil
	}
	if c.Literal == "out" {
		// TODO
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

// Parse parses the lines and returns the compiled Circuit
func (p *Parser) Parse() (*Circuit, error) {
	circuit := &Circuit{}
	circuit.Signals = append(circuit.Signals, "one")
	nInputs := 0
	for {
		constraint, err := p.parseLine()
		if err != nil {
			break
		}
		fmt.Println(constraint)
		if constraint.Literal == "func" {
			// one constraint for each input
			for _, in := range constraint.PrivateInputs {
				newConstr := &Constraint{
					Op:  "in",
					Out: in,
				}
				circuit.Constraints = append(circuit.Constraints, *newConstr)
				nInputs++
			}
			for _, in := range constraint.PublicInputs {
				newConstr := &Constraint{
					Op:  "in",
					Out: in,
				}
				circuit.Constraints = append(circuit.Constraints, *newConstr)
				nInputs++
			}
			circuit.PublicInputs = constraint.PublicInputs
			circuit.PrivateInputs = constraint.PrivateInputs
			continue
		}
		if constraint.Literal == "equals" {
			// TODO
			fmt.Println("circuit.Signals", circuit.Signals)
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
			// if Out is "out", put it after first value (one) and before the inputs
			if !existInArray(circuit.Signals, constraint.Out) {
				signalsCopy := copyArray(circuit.Signals)
				var auxSignals []string
				auxSignals = append(auxSignals, signalsCopy[0])
				auxSignals = append(auxSignals, constraint.Out)
				auxSignals = append(auxSignals, signalsCopy[1:]...)
				circuit.Signals = auxSignals
				circuit.PublicSignals = append(circuit.PublicSignals, constraint.Out)
				circuit.NPublic++
			}
		} else {
			circuit.Signals = addToArrayIfNotExist(circuit.Signals, constraint.Out)
		}
	}
	circuit.NVars = len(circuit.Signals)
	circuit.NSignals = len(circuit.Signals)
	return circuit, nil
}
func copyArray(in []string) []string { // tmp
	var out []string
	for _, e := range in {
		out = append(out, e)
	}
	return out
}
