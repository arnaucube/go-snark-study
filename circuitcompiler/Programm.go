package circuitcompiler

import (
	"fmt"
	"github.com/mottla/go-snark/bn128"
	"github.com/mottla/go-snark/fields"
	"github.com/mottla/go-snark/r1csqap"
	"math/big"
	"sync"
)

type utils struct {
	Bn  bn128.Bn128
	FqR fields.Fq
	PF  r1csqap.PolynomialField
}

type Program struct {
	functions             map[string]*Circuit
	globalInputs          []Constraint
	arithmeticEnvironment utils //find a better name

	R1CS struct {
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

	mainRoot := p.getMainCircuit().root

	if mainRoot.value.Op&(MINUS|PLUS) != 0 {
		newOut := Constraint{Out: "out", V1: "1", V2: "out2", Op: MULTIPLY}
		p.getMainCircuit().addConstraint(&newOut)
		mainRoot.value.Out = "main@out2"
		p.getMainCircuit().gateMap[mainRoot.value.Out] = mainRoot
	}

	var wg = sync.WaitGroup{}

	for _, circuit := range p.functions {
		wg.Add(1)
		func() {
			circuit.buildTree(circuit.root)
			wg.Done()
		}()

	}
	wg.Wait()
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
	p.r1CSRecursiveBuild(p.getMainCircuit(), p.getMainCircuit().root, mGatesUsed, &orderedmGates, false, false)
	return orderedmGates
}

func (p *Program) r1CSRecursiveBuild(currentCircuit *Circuit, root *gate, mGatesUsed map[string]bool, orderedmGates *[]gate, negate bool, inverse bool) (isConstant bool) {

	if root.OperationType() == IN {
		return false
	}

	if root.OperationType() == CONST {
		return true
	}

	if _, alreadyComputed := mGatesUsed[root.value.Out]; alreadyComputed {
		return false
	}

	if root.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(currentCircuit, root.value)
		isConstant = p.r1CSRecursiveBuild(nextContext, nextContext.root, mGatesUsed, orderedmGates, negate, inverse)
		return isConstant
	}

	if _, alreadyComputed := mGatesUsed[root.value.V1]; !alreadyComputed {
		isConstant = p.r1CSRecursiveBuild(currentCircuit, root.left, mGatesUsed, orderedmGates, negate, inverse)
	}

	if _, alreadyComputed := mGatesUsed[root.value.V2]; !alreadyComputed {
		cons := p.r1CSRecursiveBuild(currentCircuit, root.right, mGatesUsed, orderedmGates, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		isConstant = isConstant || cons
	}

	if root.OperationType() == MULTIPLY {

		if isConstant && !root.value.invert && root != p.getMainCircuit().root {
			return false
		}
		root.leftIns = p.collectFactors(currentCircuit, root.left, mGatesUsed, false, false)
		//if root.left.value.Out== root.right.value.Out{
		//	//note this is not a full copy, but shouldnt be a problem
		//	root.rightIns= root.leftIns
		//}else{
		//	collectAtomsInSubtree(root.right, mGatesUsed, 1, root.rightIns, functionRootMap, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		//}
		//root.rightIns = collectAtomsInSubtree3(root.right, mGatesUsed, Xor(negate, root.value.negate), Xor(inverse, root.value.invert))
		root.rightIns = p.collectFactors(currentCircuit, root.right, mGatesUsed, false, false)
		root.index = len(mGatesUsed)
		var nn = root.value.Out
		//if _, ex := p.functions[nn]; ex {
		//	nn = composeNewFunction(root.value.Out, currentCircuit.Inputs)
		//}

		mGatesUsed[nn] = true
		rootGate := cloneGate(root)
		rootGate.value.Out = nn
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
func addFactors(leftFactors, rightFactors []factor) []factor {
	var found bool
	res := make([]factor, 0, len(leftFactors)+len(rightFactors))
	for _, facLeft := range leftFactors {

		found = false
		for i, facRight := range rightFactors {

			if facLeft.typ&facRight.typ == CONST {
				var a0, b0 = facLeft.multiplicative[0], facRight.multiplicative[0]
				if facLeft.negate {
					a0 *= -1
				}
				if facRight.negate {
					b0 *= -1
				}
				absValue, positive := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)

				rightFactors[i] = factor{typ: CONST, negate: !positive, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}

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
				absValue, positive := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)

				rightFactors[i] = factor{typ: IN, invert: facRight.invert, name: facRight.name, negate: !positive, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}

				found = true
				//res = append(res, factor{typ: CONST, negate: negate, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}})
				break
			}
		}
		if !found {
			res = append(res, facLeft)
		}
	}

	for _, val := range rightFactors {
		if val.multiplicative[0] != 0 {
			res = append(res, val)
		}
	}

	return res
}

func (p *Program) collectFactors(contextCircut *Circuit, g *gate, mGatesUsed map[string]bool, negate bool, invert bool) []factor {

	if _, ex := mGatesUsed[g.value.Out]; ex {
		return []factor{{typ: IN, name: g.value.Out, invert: invert, negate: negate, multiplicative: [2]int{1, 1}}}
	}

	if g.OperationType() == IN {
		return []factor{{typ: IN, name: g.value.Out, invert: invert, negate: negate, multiplicative: [2]int{1, 1}}}
	}
	if g.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, g.value)
		return p.collectFactors(nextContext, nextContext.root, mGatesUsed, negate, invert)
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
		leftFactors = p.collectFactors(nextContext, nextContext.root, mGatesUsed, negate, invert)
	} else {
		leftFactors = p.collectFactors(contextCircut, g.left, mGatesUsed, negate, invert)
	}

	if g.right.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(contextCircut, g.right.value)
		rightFactors = p.collectFactors(nextContext, nextContext.root, mGatesUsed, Xor(negate, g.value.negate), Xor(invert, g.value.invert))
	} else {
		rightFactors = p.collectFactors(contextCircut, g.right, mGatesUsed, Xor(negate, g.value.negate), Xor(invert, g.value.invert))
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
	constr := &Constraint{Inputs: in.value.Inputs, Out: in.value.Out, Op: in.value.Op, invert: in.value.invert, negate: in.value.negate, V2: in.value.V2, V1: in.value.V1}
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

func (p *Program) getMainCircuit() *Circuit {
	return p.functions["main"]
}

func (p *Program) addGlobalInput(c Constraint) {
	c.Out = "main@" + c.Out
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

func (p *Program) extendedFunctionRenamer(contextCircuit *Circuit, constraint *Constraint) (nextContext *Circuit) {

	if constraint.Op != FUNC {
		panic("not a function")
	}
	//if _, ex := contextCircuit.gateMap[constraint.Out]; !ex {
	//	panic("constraint must be within the contextCircuit circuit")
	//}
	if b, n, _ := isFunction(constraint.Out); b {
		if newContext, v := p.functions[n]; v {
			//am i certain that constraint.inputs is alwazs equal to n??? me dont like it
			for i, argument := range constraint.Inputs {
				isConst, _ := isValue(argument)
				if isConst {
					continue
				}
				isFunc, _, _ := isFunction(argument)
				if isFunc {
					panic("functions as arguments no supported yet")
					//p.extendedFunctionRenamer(contextCircuit,)
				}
				//at this point I assert that argument is a variable. This can become troublesome later
				inputOriginCircuit := p.functions[getContextFromVariable(argument)]
				if gate, ex := inputOriginCircuit.gateMap[argument]; ex {
					oldGate := newContext.gateMap[newContext.Inputs[i]]
					//we take the old gate which was nothing but a input
					//and link this input to its constituents comming from the calling contextCircuit.
					//i think this is pretty neat
					oldGate.value = gate.value
					oldGate.right = gate.right
					oldGate.left = gate.left

				} else {
					panic("not expected")
				}
			}
			newContext.renameInputs(constraint.Inputs)
			return newContext
		}
	} else {
		panic("not expected")
	}

	return nil
}

func NewProgram() (p *Program) {
	p = &Program{functions: map[string]*Circuit{}, globalInputs: []Constraint{{Op: IN, Out: "one"}}, arithmeticEnvironment: prepareUtils()}
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
	value := new(big.Int).Add(new(big.Int), fractionToField(val.multiplicative))

	if val.negate {
		value.Neg(value)
	}

	if val.typ == CONST {
		arr[0] = value
	} else {
		arr[index] = value
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
