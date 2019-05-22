package circuitcompiler

import (
	"fmt"
	"github.com/mottla/go-snark/bn128"
	"github.com/mottla/go-snark/fields"
	"github.com/mottla/go-snark/r1csqap"
	"math/big"
)

type utils struct {
	Bn  bn128.Bn128
	FqR fields.Fq
	PF  r1csqap.PolynomialField
}

type Program struct {
	functions               map[string]*Circuit
	globalInputs            []Constraint
	arithmeticEnvironment   utils //find a better name
	extendedFunctionRenamer func(context *Circuit, c Constraint) (newContext *Circuit)
	R1CS                    struct {
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
		root := circuit.gateMap[fName]
		functionRootMap[fName] = root
		circuit.root = root
		circuit.buildTree(root)
	}

	return

}

func (c *Circuit) buildTree(g *gate) {
	if _, ex := c.gateMap[g.value.Out]; ex {
		if g.OperationType()&(IN|CONST) != 0 {
			return
		}
	} else {
		panic(fmt.Sprintf("undefined variable %s", g.value.Out))
	}
	if g.OperationType() == FUNC {
		//g.funcInputs = []*gate{}
		for _, in := range g.value.Inputs {
			if gate, ex := c.gateMap[in]; ex {
				//sadf

				g.funcInputs = append(g.funcInputs, gate)
				//note that we do repeated work here. the argument
				c.buildTree(gate)
			} else {
				panic(fmt.Sprintf("undefined argument %s", g.value.V1))
			}
		}
		return
	}
	if constr, ex := c.gateMap[g.value.V1]; ex {
		g.left = constr
		c.buildTree(g.left)
	} else {
		panic(fmt.Sprintf("undefined value %s", g.value.V1))
	}

	if constr, ex := c.gateMap[g.value.V2]; ex {
		g.right = constr
		c.buildTree(g.right)
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

	p.extendedFunctionRenamer = func(context *Circuit, c Constraint) (nextContext *Circuit) {
		if c.Op != FUNC {
			panic("not a function")
		}
		if _, ex := context.gateMap[c.Out]; !ex {
			panic("constraint mus be within the context circuit")
		}

		if b, name, in := isFunction(c.Out); b {
			if newContext, v := p.functions[name]; v {
				//fmt.Println("unrenamed thing")
				//PrintTree(k.root)
				for i, argument := range in {
					if gate, ex := context.gateMap[argument]; ex {
						oldGate := newContext.gateMap[newContext.Inputs[i]]
						//we take the old gate which was nothing but a input
						//and link this input to its constituents comming from the calling context.
						//i think this is pretty neat
						oldGate.value = gate.value
						oldGate.right = gate.right
						oldGate.left = gate.left

					} else {
						panic("not expected")
					}
				}

				newContext.renameInputs(in)

				//fmt.Println("renamed thing")
				//PrintTree(k.root)
				return newContext
			}
		}
		panic("not a function dude")
		return nil
	}
	//traverseCombinedMultiplicationGates(p.getMainCircut().root, mGatesUsed, &orderedmGates, functionRootMap, functionRenamer, false, false)

	//markMgates(p.getMainCircut().root, mGatesUsed, &orderedmGates, functionRenamer, false, false)
	p.markMgates2(p.getMainCircut(), p.getMainCircut().root, mGatesUsed, &orderedmGates, false, false)
	return orderedmGates
}

func (p *Program) markMgates2(contextCircut *Circuit, root *gate, mGatesUsed map[string]bool, orderedmGates *[]gate, negate bool, inverse bool) (isConstant bool) {

	if root.OperationType() == IN {
		return false
	}

	if root.OperationType() == CONST {
		return true
	}

	if root.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, root.value)
		isConstant = p.markMgates2(nextContext, nextContext.root, mGatesUsed, orderedmGates, negate, inverse)
	} else {
		if _, alreadyComputed := mGatesUsed[root.value.V1]; !alreadyComputed {
			isConstant = p.markMgates2(contextCircut, root.left, mGatesUsed, orderedmGates, negate, inverse)
		}

		if _, alreadyComputed := mGatesUsed[root.value.V2]; !alreadyComputed {
			cons := p.markMgates2(contextCircut, root.right, mGatesUsed, orderedmGates, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
			isConstant = isConstant || cons
		}
	}

	if root.OperationType() == MULTIPLY {

		_, n, _ := isFunction(root.value.Out)
		if isConstant && !root.value.invert && n != "main" {
			return false
		}
		root.leftIns = p.collectAtomsInSubtree2(contextCircut, root.left, mGatesUsed, false, false)
		//if root.left.value.Out== root.right.value.Out{
		//	//note this is not a full copy, but shouldnt be a problem
		//	root.rightIns= root.leftIns
		//}else{
		//	collectAtomsInSubtree(root.right, mGatesUsed, 1, root.rightIns, functionRootMap, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		//}
		//root.rightIns = collectAtomsInSubtree3(root.right, mGatesUsed, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		root.rightIns = p.collectAtomsInSubtree2(contextCircut, root.right, mGatesUsed, false, false)
		root.index = len(mGatesUsed)
		mGatesUsed[root.value.Out] = true
		rootGate := cloneGate(root)
		*orderedmGates = append(*orderedmGates, *rootGate)

	}

	return isConstant
	//TODO optimize if output is not a multipication gate
}

type factor struct {
	typ            Token
	name           string
	invert, negate bool
	multiplicative [2]int
}

func (f factor) String() string {
	if f.typ == CONST {
		return fmt.Sprintf("(const fac: %v)", f.multiplicative)
	}
	str := f.name
	if f.invert {
		str += "^-1"
	}
	if f.negate {
		str = "-" + str
	}
	return fmt.Sprintf("(\"%s\"  fac: %v)", str, f.multiplicative)
}

func mul2DVector(a, b [2]int) [2]int {
	return [2]int{a[0] * b[0], a[1] * b[1]}
}

func mulFactors(leftFactors, rightFactors []factor) (result []factor) {

	for _, facLeft := range leftFactors {

		for i, facRight := range rightFactors {
			if facLeft.typ == CONST && facRight.typ == IN {
				rightFactors[i] = factor{typ: IN, name: facRight.name, negate: Xor(facLeft.negate, facRight.negate), invert: facRight.invert, multiplicative: mul2DVector(facRight.multiplicative, facLeft.multiplicative)}
				continue
			}
			if facRight.typ == CONST && facLeft.typ == IN {
				rightFactors[i] = factor{typ: IN, name: facLeft.name, negate: Xor(facLeft.negate, facRight.negate), invert: facLeft.invert, multiplicative: mul2DVector(facRight.multiplicative, facLeft.multiplicative)}
				continue
			}

			if facRight.typ&facLeft.typ == CONST {
				rightFactors[i] = factor{typ: CONST, negate: Xor(facRight.negate, facLeft.negate), multiplicative: mul2DVector(facRight.multiplicative, facLeft.multiplicative)}
				continue

			}
			//tricky part here
			//this one should only be reached, after a true mgate had its left and right braches computed. here we
			//a factor can appear at most in quadratic form. we reduce terms a*a^-1 here.
			if facRight.typ&facLeft.typ == IN {
				//if facRight.n
				//rightFactors[i] = factor{typ: CONST, negate: Xor(facRight.negate, facLeft.negate), multiplicative: mul2DVector(facRight.multiplicative, facLeft.multiplicative)}
				//continue

			}
			panic("unexpected")

		}

	}

	return rightFactors
}

//returns the absolute value of a signed int and a flag telling if the input was positive or not
//this implementation is awesome and fast (see Henry S Warren, Hackers's Delight)
func abs(n int) (val int, positive bool) {
	y := n >> 63
	return (n ^ y) - y, y == 0
}

//returns the reduced sum of two input factor arrays
//if no reduction was done (worst case), it returns the concatenation of the input arrays
func addFactors(leftFactors, rightFactors []factor) (res []factor) {
	var found bool
	for _, facLeft := range leftFactors {

		for i, facRight := range rightFactors {

			if facLeft.typ&facRight.typ == CONST {
				var a0, b0 = facLeft.multiplicative[0], facRight.multiplicative[0]
				if facLeft.negate {
					a0 *= -1
				}
				if facRight.negate {
					b0 *= -1
				}
				absValue, negate := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)
				rightFactors[i] = factor{typ: CONST, negate: negate, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}
				found = true
				//res = append(res, factor{typ: CONST, negate: negate, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}})
				break
			}
			if facLeft.typ&facRight.typ == IN && facLeft.invert == facRight.invert && facLeft.name == facRight.name {
				var a0, b0 = facLeft.multiplicative[0], facRight.multiplicative[0]
				if facLeft.negate {
					a0 *= -1
				}
				if facRight.negate {
					b0 *= -1
				}
				absValue, negate := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)
				rightFactors[i] = factor{typ: IN, invert: facRight.invert, name: facRight.name, negate: negate, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}
				found = true
				//res = append(res, factor{typ: CONST, negate: negate, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}})
				break
			}
		}
		if !found {
			res = append(res, facLeft)
			found = false
		}
	}
	return append(res, rightFactors...)
}

func (p *Program) collectAtomsInSubtree2(contextCircut *Circuit, g *gate, mGatesUsed map[string]bool, negate bool, invert bool) []factor {

	if _, ex := mGatesUsed[g.value.Out]; ex {
		return []factor{{typ: IN, name: g.value.Out, invert: invert, negate: negate, multiplicative: [2]int{1, 1}}}
	}

	if g.OperationType() == IN {
		return []factor{{typ: IN, name: g.value.Out, invert: invert, negate: negate, multiplicative: [2]int{1, 1}}}
	}
	if g.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, g.value)
		return p.collectAtomsInSubtree2(nextContext, nextContext.root, mGatesUsed, negate, invert)
	}

	if g.OperationType() == CONST {
		b1, v1 := isValue(g.value.Out)
		if !b1 {
			panic("not a constant")
		}
		if invert {
			return []factor{{typ: CONST, negate: negate, multiplicative: [2]int{1, v1}}}
		}
		return []factor{{typ: CONST, negate: negate, multiplicative: [2]int{v1, 1}}}
	}

	var leftFactors, rightFactors []factor
	if g.left.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, g.left.value)
		leftFactors = p.collectAtomsInSubtree2(nextContext, nextContext.root, mGatesUsed, negate, invert)
	} else {
		leftFactors = p.collectAtomsInSubtree2(contextCircut, g.left, mGatesUsed, negate, invert)
	}

	if g.right.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, g.right.value)
		rightFactors = p.collectAtomsInSubtree2(nextContext, nextContext.root, mGatesUsed, Xor(negate, g.value.negate), Xor(invert, g.value.invert))
	} else {
		rightFactors = p.collectAtomsInSubtree2(contextCircut, g.right, mGatesUsed, Xor(negate, g.value.negate), Xor(invert, g.value.invert))
	}

	switch g.OperationType() {
	case MULTIPLY:
		return mulFactors(leftFactors, rightFactors)
	case PLUS:
		return addFactors(leftFactors, rightFactors)
	default:
		panic("unexpected gate")
	}

}

//copies a gate neglecting its references to other gates
func cloneGate(in *gate) (out *gate) {
	constr := Constraint{Inputs: in.value.Inputs, Out: in.value.Out, Op: in.value.Op, invert: in.value.invert, negate: in.value.negate, V2: in.value.V2, V1: in.value.V1}
	nRightins := make([]factor, len(in.rightIns))
	nLeftInst := make([]factor, len(in.leftIns))
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

func (p *Program) addGlobalInput(c Constraint) {
	p.globalInputs = append(p.globalInputs, c)
}

func prepareUtils() utils {
	bn, err := bn128.NewBn128()
	if err != nil {
		panic(err)
	}
	// new Finite Field
	fqR := fields.NewFq(bn.R)
	// new Polynomial Field
	pf := r1csqap.NewPolynomialField(fqR)

	return utils{
		Bn:  bn,
		FqR: fqR,
		PF:  pf,
	}
}
func NewProgramm() *Program {

	//return &Program{functions: map[string]*Circuit{}, signals: []string{}, globalInputs: []*Constraint{{Op: PLUS, V1:"1",V2:"0", Out: "one"}}}
	return &Program{functions: map[string]*Circuit{}, globalInputs: []Constraint{{Op: IN, Out: "one"}}, arithmeticEnvironment: prepareUtils()}
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

	//I need the inputs to be defined as input constraints for each function for later renaming conventions
	//if constraint.Literal == "main" {
	for _, in := range constraint.Inputs {
		newConstr := Constraint{
			Op:  IN,
			Out: in,
		}
		if name == "main" {
			p.addGlobalInput(newConstr)
		}
		c.addConstraint(newConstr)
	}
	//}

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

	for i, v := range p.globalInputs {
		indexMap[v.Out] = i

	}
	for i, v := range mGates {
		indexMap[v.value.Out] = i + offset
	}

	for _, gate := range mGates {

		if gate.OperationType() == MULTIPLY {
			aConstraint := r1csqap.ArrayOfBigZeros(size)
			bConstraint := r1csqap.ArrayOfBigZeros(size)
			cConstraint := r1csqap.ArrayOfBigZeros(size)

			for _, val := range gate.leftIns {
				convertAndInsertFactorAt(aConstraint, val, indexMap[val.name])
			}

			for _, val := range gate.rightIns {
				convertAndInsertFactorAt(bConstraint, val, indexMap[val.name])
			}

			cConstraint[indexMap[gate.value.Out]] = big.NewInt(int64(1))

			if gate.value.invert {
				tmp := aConstraint
				aConstraint = cConstraint
				cConstraint = tmp
			}
			a = append(a, aConstraint)
			b = append(b, bConstraint)
			c = append(c, cConstraint)

		} else {
			panic("not a m gate")
		}
	}
	p.R1CS.A = a
	p.R1CS.B = b
	p.R1CS.C = c
	return a, b, c
}

var Utils = prepareUtils()

func fractionToField(in [2]int) *big.Int {
	return Utils.FqR.Mul(big.NewInt(int64(in[0])), Utils.FqR.Inverse(big.NewInt(int64(in[1]))))

}

func convertAndInsertFactorAt(arr []*big.Int, val factor, index int) {
	if val.typ == CONST {
		arr[0] = new(big.Int).Add(arr[0], fractionToField(val.multiplicative))
		return
	}
	arr[index] = new(big.Int).Add(arr[index], fractionToField(val.multiplicative))

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
