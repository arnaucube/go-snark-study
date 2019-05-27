package circuitcompiler

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var variableIndicationSign = "@"

// Circuit is the data structure of the compiled circuit
type Circuit struct {
	Inputs []string
	Name   string
	root   *gate
	//after reducing
	//constraintMap map[string]*Constraint
	gateMap map[string]*gate
}

type gate struct {
	index      int
	left       *gate
	right      *gate
	funcInputs []*gate
	value      *Constraint //is a pointer a good thing here??
	leftIns    []factor    //leftIns and RightIns after addition gates have been reduced. only multiplication gates remain
	rightIns   []factor
}

func (g gate) String() string {
	return fmt.Sprintf("Gate %v : %v  with left %v right %v", g.index, g.value, g.leftIns, g.rightIns)
}

//type variable struct {
//	val string
//}

// Constraint is the data structure of a flat code operation
type Constraint struct {
	// v1 op v2 = out
	Op  Token
	V1  string
	V2  string
	Out string
	//fV1  *variable
	//fV2  *variable
	//fOut *variable
	//Literal string
	//TODO once i've implemented a new parser/lexer we do this differently
	Inputs []string // in func declaration case
	//fInputs []*variable
	negate bool
	invert bool
}

func (c Constraint) String() string {
	if c.negate || c.invert {
		return fmt.Sprintf("|%v = %v %v %v|  negated: %v, inverted %v", c.Out, c.V1, c.Op, c.V2, c.negate, c.invert)
	}
	return fmt.Sprintf("|%v = %v %v %v|", c.Out, c.V1, c.Op, c.V2)
}

func newCircuit(name string) *Circuit {
	return &Circuit{Name: name, gateMap: make(map[string]*gate)}
}

func (p *Program) addFunction(constraint *Constraint) (c *Circuit) {
	name := constraint.Out
	fmt.Println("try to add function ", name)

	b, name2, _ := isFunction(name)
	if !b {
		panic(fmt.Sprintf("not a function: %v", constraint))
	}
	name = name2

	if _, ex := p.functions[name]; ex {
		panic("function already declared")
	}

	c = newCircuit(name)

	p.functions[name] = c

	renamedInputs := make([]string, len(constraint.Inputs))
	//I need the inputs to be defined as input constraints for each function for later renaming conventions
	//if constraint.Literal == "main" {
	for i, in := range constraint.Inputs {
		newConstr := &Constraint{
			Op:  IN,
			Out: in,
		}
		//if name == "main" {
		//	p.addGlobalInput(*newConstr)
		//}
		c.addConstraint(newConstr)
		renamedInputs[i] = newConstr.Out
	}
	//}

	c.Inputs = renamedInputs
	return

}

func (circ *Circuit) addConstraint(constraint *Constraint) {
	if _, ex := circ.gateMap[constraint.Out]; ex {
		panic("already used FlatConstraint")
	}
	gateToAdd := &gate{value: constraint}

	if constraint.Op == DIVIDE {
		constraint.Op = MULTIPLY
		constraint.invert = true
	}
	if constraint.Op == MINUS {
		constraint.Op = PLUS
		constraint.negate = true
	}

	//todo this is dangerous.. if someone would use out as variable name, things would be fucked
	if constraint.Out == "out" {
		constraint.Out = circ.Name //composeNewFunction(circ.Name, circ.Inputs)
		circ.root = gateToAdd
	} else {
		constraint.Out = circ.renamer(constraint.Out)
	}

	constraint.V1 = circ.renamer(constraint.V1)
	constraint.V2 = circ.renamer(constraint.V2)

	circ.gateMap[constraint.Out] = gateToAdd
}

func (circ *Circuit) currentOutputName() string {

	return composeNewFunction(circ.Name, circ.currentOutputs())
}

func (circ *Circuit) currentOutputs() []string {

	renamedInputs := make([]string, len(circ.Inputs))
	for i, in := range circ.Inputs {
		if _, ex := circ.gateMap[in]; !ex {
			panic("not existing input")
		}
		renamedInputs[i] = circ.gateMap[in].value.Out
	}

	return renamedInputs

}

func (circ *Circuit) renamer(constraint string) string {

	if constraint == "" {
		return ""
	}

	if b, _ := isValue(constraint); b {
		circ.gateMap[constraint] = &gate{value: &Constraint{Op: CONST, Out: constraint}}
		return constraint
	}

	if b, name, inputs := isFunction(constraint); b {
		renamedInputs := make([]string, len(inputs))
		for i, in := range inputs {
			renamedInputs[i] = circ.renamer(in)
		}
		nn := composeNewFunction(name, renamedInputs)
		circ.gateMap[nn] = &gate{value: &Constraint{Op: FUNC, Out: nn, Inputs: renamedInputs}}
		return nn
	}

	return circ.Name + variableIndicationSign + constraint

}

//func (circ *Circuit) renameInputs(inputs []string) {
//	if len(inputs) != len(circ.Inputs) {
//		panic("given inputs != circuit.Inputs")
//	}
//	mapping := make(map[string]string)
//	for i := 0; i < len(inputs); i++ {
//		if _, ex := circ.gateMap[inputs[i]]; ex {
//
//			//this is a tricky part. So we replace former inputs with the new ones, thereby
//			//it might be, that the new input name has already been used for some output inside the function
//			//currently I dont know an elegant way how to handle this renaming issue
//			if circ.gateMap[inputs[i]].value.Op != IN {
//				panic(fmt.Sprintf("renaming collsion with %s", inputs[i]))
//			}
//
//		}
//		mapping[circ.Inputs[i]] = inputs[i]
//	}
//	//fmt.Println(mapping)
//	//circ.Inputs = inputs
//	permute := func(in string) string {
//		if out, ex := mapping[in]; ex {
//			return out
//		}
//		return in
//	}
//
//	permuteListe := func(in []string) []string {
//		for i := 0; i < len(in); i++ {
//			in[i] = permute(in[i])
//		}
//		return in
//	}
//
//	for _, constraint := range circ.gateMap {
//
//		if constraint.value.Op == IN {
//			constraint.value.Out = permute(constraint.value.Out)
//			continue
//		}
//
//		if b, n, in := isFunction(constraint.value.Out); b {
//			constraint.value.Out = composeNewFunction(n, permuteListe(in))
//			constraint.value.Inputs = permuteListe(in)
//		}
//		if b, n, in := isFunction(constraint.value.V1); b {
//			constraint.value.V1 = composeNewFunction(n, permuteListe(in))
//			constraint.value.Inputs = permuteListe(in)
//		}
//		if b, n, in := isFunction(constraint.value.V2); b {
//			constraint.value.V2 = composeNewFunction(n, permuteListe(in))
//			constraint.value.Inputs = permuteListe(in)
//		}
//
//		constraint.value.V1 = permute(constraint.value.V1)
//		constraint.value.V2 = permute(constraint.value.V2)
//
//	}
//	return
//}

func getContextFromVariable(in string) string {
	//if strings.Contains(in, variableIndicationSign) {
	//	return strings.Split(in, variableIndicationSign)[0]
	//}
	//return ""
	return strings.Split(in, variableIndicationSign)[0]
}

func composeNewFunction(fname string, inputs []string) string {
	builder := strings.Builder{}
	builder.WriteString(fname)
	builder.WriteRune('(')
	for i := 0; i < len(inputs); i++ {
		builder.WriteString(inputs[i])
		if i < len(inputs)-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(')')
	return builder.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func TreeDepth(g *gate) int {
	return printDepth(g, 0)
}

func printDepth(g *gate, d int) int {
	d = d + 1
	if g.left != nil && g.right != nil {
		return max(printDepth(g.left, d), printDepth(g.right, d))
	} else if g.left != nil {
		return printDepth(g.left, d)
	} else if g.right != nil {
		return printDepth(g.right, d)
	}
	return d
}
func CountMultiplicationGates(g *gate) int {
	if g == nil {
		return 0
	}
	if len(g.rightIns) > 0 || len(g.leftIns) > 0 {
		return 1 + CountMultiplicationGates(g.left) + CountMultiplicationGates(g.right)
	} else {
		return CountMultiplicationGates(g.left) + CountMultiplicationGates(g.right)
	}
	return 0
}

//TODO avoid printing multiple times in case of loops
func PrintTree(g *gate) {
	printTree(g, 0)
}
func printTree(g *gate, d int) {
	d += 1

	if g.leftIns == nil || g.rightIns == nil {
		fmt.Printf("Depth: %v - %s \t \t \t \t \n", d, g.value)
	} else {
		fmt.Printf("Depth: %v - %s \t \t \t \t with  l %v  and r %v\n", d, g.value, g.leftIns, g.rightIns)
	}
	if g.funcInputs != nil {
		for _, v := range g.funcInputs {
			printTree(v, d)
		}
	}

	if g.left != nil {
		printTree(g.left, d)
	}
	if g.right != nil {
		printTree(g.right, d)
	}
}

func Xor(a, b bool) bool {
	return (a && !b) || (!a && b)
}

func (g *gate) ExtractValues(in []int) (er error) {
	if b, v1 := isValue(g.value.V1); b {
		if b2, v2 := isValue(g.value.V2); b2 {
			in = append(in, v1, v2)
			return nil
		}
	}
	return errors.New(fmt.Sprintf("Gate \"%s\" has no int values", g.value))
}

func (g *gate) OperationType() Token {
	return g.value.Op
}

//returns index of e if its in arr
//return -1 if e not in arr
func indexInArray(arr []string, e string) int {
	for i, a := range arr {
		if a == e {
			return i
		}
	}
	panic("lul")
	return -1
}
func isValue(a string) (bool, int) {
	v, err := strconv.Atoi(a)
	if err != nil {
		return false, 0
	}
	return true, v
}
func isFunction(a string) (tf bool, name string, inputs []string) {

	if !strings.ContainsRune(a, '(') && !strings.ContainsRune(a, ')') {
		return false, "", nil
	}
	name = strings.Split(a, "(")[0]

	// read string inside ( )
	rgx := regexp.MustCompile(`\((.*?)\)`)
	insideParenthesis := rgx.FindStringSubmatch(a)
	varsString := strings.Replace(insideParenthesis[1], " ", "", -1)
	inputs = strings.Split(varsString, ",")

	return true, name, inputs
}

type Inputs struct {
	Private []*big.Int
	Publics []*big.Int
}
