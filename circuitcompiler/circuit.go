package circuitcompiler

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// Circuit is the data structure of the compiled circuit
type Circuit struct {
	NVars         int
	NPublic       int
	NSignals      int
	Inputs        []string
	Signals       []string
	PublicSignals []string
	Witness       []*big.Int
	Name          string
	root          *gate
	//after reducing
	constraintMap map[string]*Constraint
	//used          map[string]bool
	R1CS struct {
		A [][]*big.Int
		B [][]*big.Int
		C [][]*big.Int
	}
}

type gate struct {
	index      int
	left       *gate
	right      *gate
	funcInputs []*gate
	value      *Constraint
	leftIns    map[string]int //leftIns and RightIns after addition gates have been reduced. only multiplication gates remain
	rightIns   map[string]int
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
	return &Circuit{Name: name, constraintMap: make(map[string]*Constraint)}
}

func (g *gate) addLeft(c *Constraint) {
	if g.left != nil {
		panic("already set left gate")
	}
	g.left = &gate{value: c}
}
func (g *gate) addRight(c *Constraint) {
	if g.right != nil {
		panic("already set left gate")
	}
	g.right = &gate{value: c}
}

func (circ *Circuit) addConstraint(constraint *Constraint) {
	if _, ex := circ.constraintMap[constraint.Out]; ex {
		panic("already used FlatConstraint")
	}

	if constraint.Op == DIVIDE {
		constraint.Op = MULTIPLY
		constraint.invert = true
	} else if constraint.Op == MINUS {
		constraint.Op = PLUS
		constraint.negate = true
	}

	//todo this is dangerous.. if someone would use out as variable name, things would be fucked
	if constraint.Out == "out" {
		constraint.Out = composeNewFunction(circ.Name, circ.Inputs)
		if circ.Name == "main" {
			//the main functions output must be a multiplication gate
			//if its not, then we simple create one where outNew = 1 * outOld
			if constraint.Op&(MINUS|PLUS) != 0 {
				newOut := &Constraint{Out: constraint.Out, V1: "1", V2: "out2", Op: MULTIPLY}
				//TODO reachable?
				delete(circ.constraintMap, constraint.Out)
				circ.addConstraint(newOut)
				constraint.Out = "out2"
				circ.addConstraint(constraint)
			}
		}

	}

	addConstantsAndFunctions := func(constraint string) {
		if b, _ := isValue(constraint); b {
			circ.constraintMap[constraint] = &Constraint{Op: CONST, Out: constraint}
		} else if b, _, inputs := isFunction(constraint); b {

			//check if function input is a constant like foo(a,4)
			for _, in := range inputs {
				if b, _ := isValue(in); b {
					circ.constraintMap[in] = &Constraint{Op: CONST, Out: in}
				}
			}
			circ.constraintMap[constraint] = &Constraint{Op: FUNC, Out: constraint, Inputs: inputs}
		}
	}

	addConstantsAndFunctions(constraint.V1)
	addConstantsAndFunctions(constraint.V2)

	circ.constraintMap[constraint.Out] = constraint
}

func (circ *Circuit) renameInputs(inputs []string) {
	if len(inputs) != len(circ.Inputs) {
		panic("given inputs != circuit.Inputs")
	}
	mapping := make(map[string]string)
	for i := 0; i < len(inputs); i++ {
		if _, ex := circ.constraintMap[inputs[i]]; ex {

			//this is a tricky part. So we replace former inputs with the new ones, thereby
			//it might be, that the new input name has already been used for some output inside the function
			//currently I dont know an elegant way how to handle this renaming issue
			if circ.constraintMap[inputs[i]].Op != IN {
				panic(fmt.Sprintf("renaming collsion with %s", inputs[i]))
			}

		}
		mapping[circ.Inputs[i]] = inputs[i]
	}
	//fmt.Println(mapping)
	circ.Inputs = inputs
	permute := func(in string) string {
		if out, ex := mapping[in]; ex {
			return out
		}
		return in
	}

	permuteListe := func(in []string) []string {
		for i := 0; i < len(in); i++ {
			in[i] = permute(in[i])
		}
		return in
	}

	for _, constraint := range circ.constraintMap {

		if constraint.Op == IN {
			constraint.Out = permute(constraint.Out)
			continue
		}

		if b, n, in := isFunction(constraint.Out); b {
			constraint.Out = composeNewFunction(n, permuteListe(in))
			constraint.Inputs = permuteListe(in)
		}
		if b, n, in := isFunction(constraint.V1); b {
			constraint.V1 = composeNewFunction(n, permuteListe(in))
			constraint.Inputs = permuteListe(in)
		}
		if b, n, in := isFunction(constraint.V2); b {
			constraint.V2 = composeNewFunction(n, permuteListe(in))
			constraint.Inputs = permuteListe(in)
		}

		constraint.V1 = permute(constraint.V1)
		constraint.V2 = permute(constraint.V2)

	}
	return
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
