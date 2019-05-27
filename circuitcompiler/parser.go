package circuitcompiler

import (
	"errors"
	"io"
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
	switch lit {
	case "def":
		c.Op = FUNC
		// format: `func name(in):`
		//todo this is all a bit hacky and unsafe
		line, err := p.s.r.ReadString(':')
		line = strings.Replace(line, " ", "", -1)
		line = strings.Replace(line, ":", "", -1)
		//set function name
		//c.Literal = strings.Split(line, "(")[0]
		c.Out = line

		if err != nil {
			return c, err
		}
		// read string inside ( )
		rgx := regexp.MustCompile(`\((.*?)\)`)
		insideParenthesis := rgx.FindStringSubmatch(line)
		varsString := strings.Replace(insideParenthesis[1], " ", "", -1)
		c.Inputs = strings.Split(varsString, ",")
		return c, nil
	case "var":
		//var a = 234
		//c.Literal += lit
		_, lit = p.scanIgnoreWhitespace()
		c.Out = lit
		//c.Literal += lit
		_, lit = p.scanIgnoreWhitespace() // skip =
		//c.Literal += lit
		// v1
		_, lit = p.scanIgnoreWhitespace()
		c.V1 = lit
		//c.Literal += lit
		break
	case "#":
		return nil, errors.New("comment parseline")
	default:
		c.Out = lit
		//c.Literal += lit
		_, lit = p.scanIgnoreWhitespace() // skip =
		//c.Literal += lit

		// v1
		tok, lit = p.scanIgnoreWhitespace()
		c.V1 = lit
		//c.Literal += lit

		// operator
		tok, lit = p.scanIgnoreWhitespace()

		c.Op = tok
		//c.Literal += lit
		// v2
		_, lit = p.scanIgnoreWhitespace()
		c.V2 = lit
		//c.Literal += lit
	}

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
func (p *Parser) Parse() (programm *Program, err error) {
	programm = NewProgram()

	var circuit *Circuit

	for {
		constraint, err := p.parseLine()
		if err != nil {
			break
		}
		if constraint.Op == FUNC {
			circuit = programm.addFunction(constraint)
		} else {
			circuit.addConstraint(constraint)
		}
	}
	//TODO
	//circuit.NVars = len(circuit.Signals)
	//circuit.NSignals = len(circuit.Signals)
	return programm, nil
}
func copyArray(in []string) []string { // tmp
	var out []string
	for _, e := range in {
		out = append(out, e)
	}
	return out
}
