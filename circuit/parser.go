package circuit

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
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
		// get func name
		fName := strings.Split(line, "(")[0]
		fName = strings.Replace(fName, " ", "", -1)
		fName = strings.Replace(fName, "	", "", -1)
		c.V1 = fName // so, the name of the func will be in c.V1

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
		c.V1 = params[0]
		c.V2 = params[1]
		return c, nil
	}
	if c.Literal == "return" {
		_, varToReturn := p.scanIgnoreWhitespace()
		c.Out = varToReturn
		return c, nil
	}
	if c.Literal == "import" {
		line, err := p.s.r.ReadString('\n')
		if err != nil {
			return c, err
		}
		// read string inside " "
		path := strings.TrimLeft(strings.TrimRight(line, `"`), `"`)
		path = strings.Replace(path, `"`, "", -1)
		path = strings.Replace(path, " ", "", -1)
		path = strings.Replace(path, "\n", "", -1)
		c.Out = path
		return c, nil
	}

	_, lit = p.scanIgnoreWhitespace() // skip =
	c.Literal += lit

	// v1
	_, lit = p.scanIgnoreWhitespace()

	// check if lit is a name of a func that we have declared
	if _, ok := circuits[lit]; ok {
		// if inside, is calling a declared function
		c.Literal = "call"
		c.Op = lit // c.Op handles the name of the function called
		// put the inputs of the call into the c.PrivateInputs
		// format: `funcname(a, b)`
		line, err := p.s.r.ReadString(')')
		if err != nil {
			fmt.Println("ERR", err)
			return c, err
		}
		// read string inside ( )
		rgx := regexp.MustCompile(`\((.*?)\)`)
		insideParenthesis := rgx.FindStringSubmatch(line)
		varsString := strings.Replace(insideParenthesis[1], " ", "", -1)
		params := strings.Split(varsString, ",")
		c.PrivateInputs = params
		return c, nil

	}

	c.V1 = lit
	c.Literal += lit
	// operator
	_, lit = p.scanIgnoreWhitespace()
	if lit == "(" {
		panic(errors.New("using not declared function"))
	}
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

func subsIfInMap(original string, m map[string]string) string {
	if v, ok := m[original]; ok {
		return v
	}
	return original
}

var circuits map[string]*Circuit

// Parse parses the lines and returns the compiled Circuit
func (p *Parser) Parse() (*Circuit, error) {
	// funcsMap is a map holding the functions names and it's content as Circuit
	circuits = make(map[string]*Circuit)
	mainExist := false
	circuits["main"] = &Circuit{}
	callsCount := 0

	circuits["main"].Signals = append(circuits["main"].Signals, "one")
	nInputs := 0
	currCircuit := ""
	for {
		constraint, err := p.parseLine()
		if err != nil {
			break
		}
		if constraint.Literal == "func" {
			// the name of the func is in constraint.V1
			// check if the name of func is main
			if constraint.V1 != "main" {
				currCircuit = constraint.V1
				circuits[currCircuit] = &Circuit{}
				circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *constraint)
				continue
			}
			currCircuit = "main"
			mainExist = true
			// l, _ := json.Marshal(constraint)
			// fmt.Println(string(l))

			// one constraint for each input
			for _, in := range constraint.PublicInputs {
				newConstr := &Constraint{
					Op:  "in",
					Out: in,
				}
				circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *newConstr)
				nInputs++
				circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, in)
				circuits[currCircuit].NPublic++
			}
			for _, in := range constraint.PrivateInputs {
				newConstr := &Constraint{
					Op:  "in",
					Out: in,
				}
				circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *newConstr)
				nInputs++
				circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, in)
			}
			circuits[currCircuit].PublicInputs = constraint.PublicInputs
			circuits[currCircuit].PrivateInputs = constraint.PrivateInputs
			continue
		}
		if constraint.Literal == "equals" {
			constr1 := &Constraint{
				Op:      "*",
				V1:      constraint.V2,
				V2:      "1",
				Out:     constraint.V1,
				Literal: "equals(" + constraint.V1 + ", " + constraint.V2 + "): " + constraint.V1 + "==" + constraint.V2 + " * 1",
			}
			circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *constr1)
			constr2 := &Constraint{
				Op:      "*",
				V1:      constraint.V1,
				V2:      "1",
				Out:     constraint.V2,
				Literal: "equals(" + constraint.V1 + ", " + constraint.V2 + "): " + constraint.V2 + "==" + constraint.V1 + " * 1",
			}
			circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *constr2)
			continue
		}
		if constraint.Literal == "return" {
			currCircuit = ""
			continue
		}
		if constraint.Literal == "call" {
			callsCountStr := strconv.Itoa(callsCount)
			// for each of the constraints of the called circuit
			// add it into the current circuit
			signalMap := make(map[string]string)
			for i, s := range constraint.PrivateInputs {
				// signalMap[s] = circuits[constraint.Op].Constraints[0].PrivateInputs[i]
				signalMap[circuits[constraint.Op].Constraints[0].PrivateInputs[i]+callsCountStr] = s
			}
			// add out to map
			signalMap[circuits[constraint.Op].Constraints[len(circuits[constraint.Op].Constraints)-1].Out+callsCountStr] = constraint.Out

			for i := 1; i < len(circuits[constraint.Op].Constraints); i++ {
				c := circuits[constraint.Op].Constraints[i]
				// add constraint, puting unique names to vars
				nc := &Constraint{
					Op:      c.Op,
					V1:      subsIfInMap(c.V1+callsCountStr, signalMap),
					V2:      subsIfInMap(c.V2+callsCountStr, signalMap),
					Out:     subsIfInMap(c.Out+callsCountStr, signalMap),
					Literal: "",
				}
				nc.Literal = nc.Out + "=" + nc.V1 + nc.Op + nc.V2
				circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *nc)
			}
			for _, s := range circuits[constraint.Op].Signals {
				circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, subsIfInMap(s+callsCountStr, signalMap))
			}
			callsCount++
			continue

		}
		if constraint.Literal == "import" {
			circuitFile, err := os.Open(constraint.Out)
			if err != nil {
				panic(errors.New("imported path error: " + constraint.Out))
			}
			parser := NewParser(bufio.NewReader(circuitFile))
			_, err = parser.Parse() // this will add the imported file funcs into the `circuits` map
			continue
		}

		circuits[currCircuit].Constraints = append(circuits[currCircuit].Constraints, *constraint)
		isVal, _ := isValue(constraint.V1)
		if !isVal {
			circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, constraint.V1)
		}
		isVal, _ = isValue(constraint.V2)
		if !isVal {
			circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, constraint.V2)
		}

		circuits[currCircuit].Signals = addToArrayIfNotExist(circuits[currCircuit].Signals, constraint.Out)
	}
	circuits["main"].NVars = len(circuits["main"].Signals)
	circuits["main"].NSignals = len(circuits["main"].Signals)
	if mainExist == false {
		return circuits["main"], errors.New("No 'main' func declared")
	}
	return circuits["main"], nil
}
func copyArray(in []string) []string { // tmp
	var out []string
	for _, e := range in {
		out = append(out, e)
	}
	return out
}
