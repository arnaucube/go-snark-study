package circuitcompiler

import (
	"fmt"
	"github.com/mottla/go-snark/r1csqap"
	"math/big"
)

type Program struct {
	functions    map[string]*Circuit
	signals      []string
	globalInputs []*Constraint
	R1CS         struct {
		A [][]*big.Int
		B [][]*big.Int
		C [][]*big.Int
	}
}

func (p *Program) PrintContraintTrees() {
	for k, v := range p.functions {
		fmt.Println(k)
		PrintTree(v.root)
	}
}

func (p *Program) BuildConstraintTrees() {

	functionRootMap := make(map[string]*gate)
	for _, circuit := range p.functions {
		//circuit.addConstraint(p.oneConstraint())
		fName := composeNewFunction(circuit.Name, circuit.Inputs)
		root := &gate{value: circuit.constraintMap[fName]}
		functionRootMap[fName] = root
		circuit.root = root
	}

	for _, circuit := range p.functions {

		buildTree(circuit.constraintMap, circuit.root)

	}

	return

}

func buildTree(con map[string]*Constraint, g *gate) {
	if _, ex := con[g.value.Out]; ex {
		if g.OperationType()&(IN|CONST) != 0 {
			return
		}
	} else {
		panic(fmt.Sprintf("undefined variable %s", g.value.Out))
	}
	if g.OperationType() == FUNC {
		g.funcInputs = []*gate{}
		for _, in := range g.value.Inputs {
			if constr, ex := con[in]; ex {
				newGate := &gate{value: constr}
				g.funcInputs = append(g.funcInputs, newGate)
				buildTree(con, newGate)
			} else {
				panic(fmt.Sprintf("undefined value %s", g.value.V1))
			}
		}
		return
	}
	if constr, ex := con[g.value.V1]; ex {
		g.addLeft(constr)
		buildTree(con, g.left)
	} else {
		panic(fmt.Sprintf("undefined value %s", g.value.V1))
	}

	if constr, ex := con[g.value.V2]; ex {
		g.addRight(constr)
		buildTree(con, g.right)
	} else {
		panic(fmt.Sprintf("undefined value %s", g.value.V2))
	}
}

func (p *Program) ReduceCombinedTree() (orderedmGates []gate) {
	mGatesUsed := make(map[string]bool)
	orderedmGates = []gate{}
	functionRootMap := make(map[string]*gate)
	for k, v := range p.functions {
		functionRootMap[k] = v.root
	}

	functionRenamer := func(c *Constraint) *gate {

		if c.Op != FUNC {
			panic("not a function")
		}
		if b, name, in := isFunction(c.Out); b {

			if k, v := p.functions[name]; v {
				//fmt.Println("unrenamed thing")
				//PrintTree(k.root)
				k.renameInputs(in)
				//fmt.Println("renamed thing")
				//PrintTree(k.root)
				return k.root
			}
		} else {
			panic("not a function dude")
		}
		return nil
	}

	traverseCombinedMultiplicationGates(p.getMainCircut().root, mGatesUsed, &orderedmGates, functionRootMap, functionRenamer, false, false)

	//for _, g := range mGates {
	//	orderedmGates[len(orderedmGates)-1-g.index] = g
	//}

	return orderedmGates
}

func traverseCombinedMultiplicationGates(root *gate, mGatesUsed map[string]bool, orderedmGates *[]gate, functionRootMap map[string]*gate, functionRenamer func(c *Constraint) *gate, negate bool, inverse bool) {
	//if root == nil {
	//	return
	//}
	//fmt.Printf("\n%p",mGatesUsed)
	if root.OperationType() == FUNC {
		//if a input has already been built, we let this subroutine know
		//newMap := make(map[string]bool)
		for _, in := range root.funcInputs {

			if _, ex := mGatesUsed[in.value.Out]; ex {
				//newMap[in.value.Out] = true
			} else {
				traverseCombinedMultiplicationGates(in, mGatesUsed, orderedmGates, functionRootMap, functionRenamer, negate, inverse)
			}
		}
		//mGatesUsed[root.value.Out] = true
		traverseCombinedMultiplicationGates(functionRenamer(root.value), mGatesUsed, orderedmGates, functionRootMap, functionRenamer, negate, inverse)
	} else {
		if _, alreadyComputed := mGatesUsed[root.value.V1]; !alreadyComputed && root.OperationType()&(IN|CONST) == 0 {
			traverseCombinedMultiplicationGates(root.left, mGatesUsed, orderedmGates, functionRootMap, functionRenamer, negate, inverse)
		}

		if _, alreadyComputed := mGatesUsed[root.value.V2]; !alreadyComputed && root.OperationType()&(IN|CONST) == 0 {
			traverseCombinedMultiplicationGates(root.right, mGatesUsed, orderedmGates, functionRootMap, functionRenamer, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		}
	}

	if root.OperationType() == MULTIPLY {

		_, n, _ := isFunction(root.value.Out)
		if (root.left.OperationType()|root.right.OperationType())&CONST != 0 && n != "main" {
			return
		}

		root.leftIns = make(map[string]int)
		collectAtomsInSubtree(root.left, mGatesUsed, 1, root.leftIns, functionRootMap, negate, inverse)
		root.rightIns = make(map[string]int)
		//if root.left.value.Out== root.right.value.Out{
		//	//note this is not a full copy, but shouldnt be a problem
		//	root.rightIns= root.leftIns
		//}else{
		//	collectAtomsInSubtree(root.right, mGatesUsed, 1, root.rightIns, functionRootMap, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		//}
		collectAtomsInSubtree(root.right, mGatesUsed, 1, root.rightIns, functionRootMap, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		root.index = len(mGatesUsed)
		mGatesUsed[root.value.Out] = true

		rootGate := cloneGate(root)
		*orderedmGates = append(*orderedmGates, *rootGate)
	}

	//TODO optimize if output is not a multipication gate
}

func collectAtomsInSubtree(g *gate, mGatesUsed map[string]bool, multiplicative int, in map[string]int, functionRootMap map[string]*gate, negate bool, invert bool) {
	if g == nil {
		return
	}
	if _, ex := mGatesUsed[g.value.Out]; ex {
		addToMap(g.value.Out, multiplicative, in, negate)
		return
	}

	if g.OperationType()&(IN|CONST) != 0 {
		addToMap(g.value.Out, multiplicative, in, negate)
		return
	}

	if g.OperationType()&(MULTIPLY) != 0 {
		b1, v1 := isValue(g.value.V1)
		b2, v2 := isValue(g.value.V2)

		if b1 && !b2 {
			multiplicative *= v1
			collectAtomsInSubtree(g.right, mGatesUsed, multiplicative, in, functionRootMap, Xor(negate, g.value.negate), invert)
			return
		} else if !b1 && b2 {
			multiplicative *= v2
			collectAtomsInSubtree(g.left, mGatesUsed, multiplicative, in, functionRootMap, negate, invert)
			return
		} else if b1 && b2 {
			panic("multiply constants not supported yet")
		} else {
			panic("werird")
		}
	}
	if g.OperationType() == FUNC {
		if b, name, _ := isFunction(g.value.Out); b {
			collectAtomsInSubtree(functionRootMap[name], mGatesUsed, multiplicative, in, functionRootMap, negate, invert)

		} else {
			panic("function expected")
		}

	}
	collectAtomsInSubtree(g.left, mGatesUsed, multiplicative, in, functionRootMap, negate, invert)
	collectAtomsInSubtree(g.right, mGatesUsed, multiplicative, in, functionRootMap, Xor(negate, g.value.negate), invert)

}

func addOneToMap(value string, in map[string]int, negate bool) {
	addToMap(value, 1, in, negate)
}
func addToMap(value string, val int, in map[string]int, negate bool) {
	if negate {
		in[value] = (in[value] - 1) * val
	} else {
		in[value] = (in[value] + 1) * val
	}
}

//copies a gate neglecting its references to other gates
func cloneGate(in *gate) (out *gate) {
	constr := &Constraint{Inputs: in.value.Inputs, Out: in.value.Out, Op: in.value.Op, invert: in.value.invert, negate: in.value.negate, V2: in.value.V2, V1: in.value.V1}
	nRightins := make(map[string]int)
	nLeftInst := make(map[string]int)
	for k, v := range in.rightIns {
		nRightins[k] = v
	}
	for k, v := range in.leftIns {
		nLeftInst[k] = v
	}
	return &gate{value: constr, leftIns: nLeftInst, rightIns: nRightins, index: in.index}
}

func (p *Program) getMainCircut() *Circuit {
	return p.functions["main"]
}

func (p *Program) addGlobalInput(c *Constraint) {
	p.globalInputs = append(p.globalInputs, c)
}

func NewProgramm() *Program {
	//return &Program{functions: map[string]*Circuit{}, signals: []string{}, globalInputs: []*Constraint{{Op: PLUS, V1:"1",V2:"0", Out: "one"}}}
	return &Program{functions: map[string]*Circuit{}, signals: []string{}, globalInputs: []*Constraint{{Op: IN, Out: "one"}}}
}

//func (p *Program) oneConstraint() *Constraint {
//	if p.globalInputs[0].Out != "one" {
//		panic("'one' should be first global input")
//	}
//	return p.globalInputs[0]
//}

func (p *Program) addSignal(name string) {
	p.signals = append(p.signals, name)
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

	//if constraint.Literal == "main" {
	for _, in := range constraint.Inputs {
		newConstr := &Constraint{
			Op:  IN,
			Out: in,
		}
		if name == "main" {
			p.addGlobalInput(newConstr)
		}
		c.addConstraint(newConstr)
	}

	c.Inputs = constraint.Inputs
	return

}

// GenerateR1CS generates the R1CS polynomials from the Circuit
func (p *Program) GenerateReducedR1CS(mGates []gate) (a, b, c [][]*big.Int) {
	// from flat code to R1CS

	offset := len(p.globalInputs)
	//  one + in1 +in2+... + gate1 + gate2 .. + out
	size := offset + len(mGates)
	indexMap := make(map[string]int)

	//circ.Signals = []string{"one"}
	for i, v := range p.globalInputs {
		indexMap[v.Out] = i
		//circ.Signals = append(circ.Signals, v)

	}
	for i, v := range mGates {
		indexMap[v.value.Out] = i + offset
		//circ.Signals = append(circ.Signals, v.value.Out)
	}
	//circ.NVars = len(circ.Signals)
	//circ.NSignals = len(circ.Signals)

	for _, gate := range mGates {

		if gate.OperationType() == MULTIPLY {
			aConstraint := r1csqap.ArrayOfBigZeros(size)
			bConstraint := r1csqap.ArrayOfBigZeros(size)
			cConstraint := r1csqap.ArrayOfBigZeros(size)

			//if len(gate.leftIns)>=len(gate.rightIns){
			//	for leftInput, _ := range gate.leftIns {
			//		if v, ex := gate.rightIns[leftInput]; ex {
			//			gate.leftIns[leftInput] *= v
			//			gate.rightIns[leftInput] = 1
			//
			//		}
			//	}
			//}else{
			//	for rightInput, _ := range gate.rightIns {
			//		if v, ex := gate.leftIns[rightInput]; ex {
			//			gate.rightIns[rightInput] *= v
			//			gate.leftIns[rightInput] = 1
			//		}
			//	}
			//}

			for leftInput, val := range gate.leftIns {

				insertVar3(aConstraint, val, leftInput, indexMap[leftInput])
			}
			for rightInput, val := range gate.rightIns {
				insertVar3(bConstraint, val, rightInput, indexMap[rightInput])
			}
			cConstraint[indexMap[gate.value.Out]] = big.NewInt(int64(1))

			if gate.value.invert {
				a = append(a, cConstraint)
				b = append(b, bConstraint)
				c = append(c, aConstraint)
			} else {
				a = append(a, aConstraint)
				b = append(b, bConstraint)
				c = append(c, cConstraint)
			}

		} else {
			panic("not a m gate")
		}
	}
	p.R1CS.A = a
	p.R1CS.B = b
	p.R1CS.C = c
	return a, b, c
}

func insertVar3(arr []*big.Int, val int, input string, index int) {
	isVal, value := isValue(input)
	var valueBigInt *big.Int
	if isVal {
		valueBigInt = big.NewInt(int64(value))
		arr[0] = new(big.Int).Add(arr[0], valueBigInt)
	} else {
		//if !indexMap[leftInput] {
		//	panic(errors.New("using variable before it's set"))
		//}
		valueBigInt = big.NewInt(int64(val))
		arr[index] = new(big.Int).Add(arr[index], valueBigInt)
	}

}

func (p *Program) CalculateWitness(input []*big.Int) (witness []*big.Int) {

	if len(p.globalInputs)-1 != len(input) {
		panic("input do not match the required inputs")
	}

	witness = r1csqap.ArrayOfBigZeros(len(p.R1CS.A[0]))
	set := make([]bool, len(witness))
	witness[0] = big.NewInt(int64(1))
	set[0] = true

	for i := range input {
		witness[i+1] = input[i]
		set[i+1] = true
	}

	zero := big.NewInt(int64(0))

	for i := 0; i < len(p.R1CS.A); i++ {
		gatesLeftInputs := p.R1CS.A[i]
		gatesRightInputs := p.R1CS.B[i]
		gatesOutputs := p.R1CS.C[i]

		sumLeft := big.NewInt(int64(0))
		sumRight := big.NewInt(int64(0))
		sumOut := big.NewInt(int64(0))

		index := -1
		division := false
		for j, val := range gatesLeftInputs {
			if val.Cmp(zero) != 0 {
				if !set[j] {
					index = j
					division = true
					break
				}
				sumLeft.Add(sumLeft, new(big.Int).Mul(val, witness[j]))
			}
		}
		for j, val := range gatesRightInputs {
			if val.Cmp(zero) != 0 {
				sumRight.Add(sumRight, new(big.Int).Mul(val, witness[j]))
			}
		}

		for j, val := range gatesOutputs {
			if val.Cmp(zero) != 0 {
				if !set[j] {
					if index != -1 {
						panic("invalid R1CS form")
					}

					index = j
					break
				}
				sumOut.Add(sumOut, new(big.Int).Mul(val, witness[j]))
			}
		}

		if !division {
			set[index] = true
			witness[index] = new(big.Int).Mul(sumLeft, sumRight)

		} else {
			b := sumRight.Int64()
			c := sumOut.Int64()
			set[index] = true
			witness[index] = big.NewInt(c / b)
		}

	}

	return
}
