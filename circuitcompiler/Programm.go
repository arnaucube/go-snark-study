package circuitcompiler

import (
	"crypto/sha256"
	"fmt"
	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/r1csqap"
	"math/big"
	"sync"
)

type utils struct {
	Bn  bn128.Bn128
	FqR fields.Fq
	PF  r1csqap.PolynomialField
}

type R1CS struct {
	A [][]*big.Int
	B [][]*big.Int
	C [][]*big.Int
}

type MultiplicationGateSignature struct {
	identifier      string
	commonExtracted [2]int //if the mgate had a extractable factor, it will be stored here
}

type Program struct {
	functions             map[string]*Circuit
	globalInputs          []string
	globalOutput          map[string]bool
	arithmeticEnvironment utils //find a better name

	//key 1: the hash chain indicating from where the variable is called H( H(main(a,b)) , doSomething(x,z) ), where H is a hash function.
	//value 1 : map
	//			with key variable name
	//			with value variable name + hash Chain
	//this datastructure is nice but maybe ill replace it later with something less confusing
	//it serves the elementary purpose of not computing a variable a second time.
	//it boosts parse time
	computedInContext map[string]map[string]MultiplicationGateSignature

	//to reduce the number of multiplication gates, we store each factor signature, and the variable name,
	//so each time a variable is computed, that happens to have the very same factors, we reuse the former
	//it boost setup and proof time
	computedFactors map[string]MultiplicationGateSignature
}

//returns the cardinality of all main inputs + 1 for the "one" signal
func (p *Program) GlobalInputCount() int {
	return len(p.globalInputs)
}

//returns the cardinaltiy of the output signals. Current only 1 output possible
func (p *Program) GlobalOutputCount() int {
	return len(p.globalOutput)
}

func (p *Program) PrintContraintTrees() {
	for k, v := range p.functions {
		fmt.Println(k)
		PrintTree(v.root)
	}
}

func (p *Program) BuildConstraintTrees() {

	mainRoot := p.getMainCircuit().root

	//if our programs last operation is not a multiplication gate, we need to introduce on
	if mainRoot.value.Op&(MINUS|PLUS) != 0 {
		newOut := Constraint{Out: "out", V1: "1", V2: "out2", Op: MULTIPLY}
		p.getMainCircuit().addConstraint(&newOut)
		mainRoot.value.Out = "main@out2"
		p.getMainCircuit().gateMap[mainRoot.value.Out] = mainRoot
	}

	for _, in := range p.getMainCircuit().Inputs {
		p.globalInputs = append(p.globalInputs, in)
	}

	var wg = sync.WaitGroup{}

	//we build the parse trees concurrently! because we can! go rocks
	for _, circuit := range p.functions {
		wg.Add(1)
		//interesting: if circuit is not passed as argument, the program fails. duno why..
		go func(c *Circuit) {
			c.buildTree(c.root)
			wg.Done()
		}(circuit)

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

		for _, in := range g.value.Inputs {
			if gate, ex := c.gateMap[in]; ex {
				g.funcInputs = append(g.funcInputs, gate)
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
	orderedmGates = []gate{}
	p.computedInContext = make(map[string]map[string]MultiplicationGateSignature)
	p.computedFactors = make(map[string]MultiplicationGateSignature)
	rootHash := make([]byte, 10)
	p.computedInContext[string(rootHash)] = make(map[string]MultiplicationGateSignature)
	p.r1CSRecursiveBuild(p.getMainCircuit(), p.getMainCircuit().root, rootHash, &orderedmGates, false, false)
	return orderedmGates
}

//recursively walks through the parse tree to create a list of all
//multiplication gates needed for the QAP construction
//Takes into account, that multiplication with constants and addition (= substraction) can be reduced, and does so
func (p *Program) r1CSRecursiveBuild(currentCircuit *Circuit, node *gate, hashTraceBuildup []byte, orderedmGates *[]gate, negate bool, invert bool) (facs factors, hashTraceResult []byte, variableEnd bool) {

	if node.OperationType() == CONST {
		b1, v1 := isValue(node.value.Out)
		if !b1 {
			panic("not a constant")
		}
		mul := [2]int{v1, 1}
		if invert {
			mul = [2]int{1, v1}

		}
		return factors{{typ: CONST, negate: negate, multiplicative: mul}}, hashTraceBuildup, false
	}

	if node.OperationType() == FUNC {
		nextContext := p.extendedFunctionRenamer(currentCircuit, node.value)
		currentCircuit = nextContext
		node = nextContext.root
		hashTraceBuildup = hashTogether(hashTraceBuildup, []byte(currentCircuit.currentOutputName()))
		if _, ex := p.computedInContext[string(hashTraceBuildup)]; !ex {
			p.computedInContext[string(hashTraceBuildup)] = make(map[string]MultiplicationGateSignature)
		}

	}

	if node.OperationType() == IN {
		fac := &factor{typ: IN, name: node.value.Out, invert: invert, negate: negate, multiplicative: [2]int{1, 1}}
		return factors{fac}, hashTraceBuildup, true
	}

	if out, ex := p.computedInContext[string(hashTraceBuildup)][node.value.Out]; ex {
		fac := &factor{typ: IN, name: out.identifier, invert: invert, negate: negate, multiplicative: out.commonExtracted}
		return factors{fac}, hashTraceBuildup, true
	}

	leftFactors, leftHash, variableEnd := p.r1CSRecursiveBuild(currentCircuit, node.left, hashTraceBuildup, orderedmGates, negate, invert)

	rightFactors, rightHash, cons := p.r1CSRecursiveBuild(currentCircuit, node.right, hashTraceBuildup, orderedmGates, Xor(negate, node.value.negate), Xor(invert, node.value.invert))

	if node.OperationType() == MULTIPLY {

		if !(variableEnd && cons) && !node.value.invert && node != p.getMainCircuit().root {
			return mulFactors(leftFactors, rightFactors), hashTraceBuildup, variableEnd || cons

		}
		sig, newLef, newRigh := factorsSignature(leftFactors, rightFactors)
		if out, ex := p.computedFactors[sig.identifier]; ex {
			return factors{{typ: IN, name: out.identifier, invert: invert, negate: negate, multiplicative: sig.commonExtracted}}, hashTraceBuildup, true

		}

		rootGate := cloneGate(node)
		//rootGate := node
		rootGate.index = len(*orderedmGates)
		if p.getMainCircuit().root == node {
			newLef = mulFactors(newLef, factors{&factor{typ: CONST, multiplicative: sig.commonExtracted}})
		}

		rootGate.leftIns = newLef
		rootGate.rightIns = newRigh

		out := hashTogether(leftHash, rightHash)
		rootGate.value.V1 = rootGate.value.V1 + string(leftHash[:10])
		rootGate.value.V2 = rootGate.value.V2 + string(rightHash[:10])

		//note we only check for existence, but not for truth.
		//global outputs do not require a hash identifier, since they are unique
		if _, ex := p.globalOutput[rootGate.value.Out]; !ex {
			rootGate.value.Out = rootGate.value.Out + string(out[:10])
		}

		p.computedInContext[string(hashTraceBuildup)][node.value.Out] = MultiplicationGateSignature{identifier: rootGate.value.Out, commonExtracted: sig.commonExtracted}

		p.computedFactors[sig.identifier] = MultiplicationGateSignature{identifier: rootGate.value.Out, commonExtracted: sig.commonExtracted}
		*orderedmGates = append(*orderedmGates, *rootGate)

		return factors{{typ: IN, name: rootGate.value.Out, invert: invert, negate: negate, multiplicative: sig.commonExtracted}}, hashTraceBuildup, true
	}

	switch node.OperationType() {
	case PLUS:
		return addFactors(leftFactors, rightFactors), hashTraceBuildup, variableEnd || cons
	default:
		panic("unexpected gate")
	}

}

//copies a gate neglecting its references to other gates
func cloneGate(in *gate) (out *gate) {
	constr := &Constraint{Inputs: in.value.Inputs, Out: in.value.Out, Op: in.value.Op, invert: in.value.invert, negate: in.value.negate, V2: in.value.V2, V1: in.value.V1}
	nRightins := in.rightIns.clone()
	nLeftInst := in.leftIns.clone()
	return &gate{value: constr, leftIns: nLeftInst, rightIns: nRightins, index: in.index}
}

func (p *Program) getMainCircuit() *Circuit {
	return p.functions["main"]
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
	b, n, _ := isFunction(constraint.Out)
	if !b {
		panic("not expected")
	}
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
			//first we get the circuit in which the argument was created
			inputOriginCircuit := p.functions[getContextFromVariable(argument)]

			//we pick the gate that has the argument as output
			if gate, ex := inputOriginCircuit.gateMap[argument]; ex {
				//we pick the old circuit inputs and let them now reference the same as the argument gate did,
				oldGate := newContext.gateMap[newContext.Inputs[i]]
				//we take the old gate which was nothing but a input
				//and link this input to its constituents coming from the calling contextCircuit.
				//i think this is pretty neat
				oldGate.value = gate.value
				oldGate.right = gate.right
				oldGate.left = gate.left

			} else {
				panic("not expected")
			}
		}
		//newContext.renameInputs(constraint.Inputs)
		return newContext
	}

	return nil
}

func NewProgram() (p *Program) {
	p = &Program{
		functions:             map[string]*Circuit{},
		globalInputs:          []string{"one"},
		globalOutput:          map[string]bool{"main": true},
		arithmeticEnvironment: prepareUtils(),
	}
	return
}

// GenerateR1CS generates the R1CS polynomials from the Circuit
func (p *Program) GenerateReducedR1CS(mGates []gate) (r1CS R1CS) {
	// from flat code to R1CS

	offset := len(p.globalInputs)
	//  one + in1 +in2+... + gate1 + gate2 .. + out
	size := offset + len(mGates)
	indexMap := make(map[string]int)

	for i, v := range p.globalInputs {
		indexMap[v] = i
	}
	for k, _ := range p.globalOutput {
		indexMap[k] = len(indexMap)
	}
	for _, v := range mGates {
		if _, ex := indexMap[v.value.Out]; !ex {
			indexMap[v.value.Out] = len(indexMap)
		}

	}

	for _, g := range mGates {

		if g.OperationType() == MULTIPLY {
			aConstraint := r1csqap.ArrayOfBigZeros(size)
			bConstraint := r1csqap.ArrayOfBigZeros(size)
			cConstraint := r1csqap.ArrayOfBigZeros(size)

			insertValue := func(val *factor, arr []*big.Int) {
				if val.typ != CONST {
					if _, ex := indexMap[val.name]; !ex {
						panic(fmt.Sprintf("%v index not found!!!", val.name))
					}
				}
				value := new(big.Int).Add(new(big.Int), fractionToField(val.multiplicative))
				if val.negate {
					value.Neg(value)
				}
				//not that index is 0 if its a constant, since 0 is the map default if no entry was found
				arr[indexMap[val.name]] = value
			}

			for _, val := range g.leftIns {
				insertValue(val, aConstraint)
			}

			for _, val := range g.rightIns {
				insertValue(val, bConstraint)
			}

			cConstraint[indexMap[g.value.Out]] = big.NewInt(int64(1))

			if g.value.invert {
				tmp := aConstraint
				aConstraint = cConstraint
				cConstraint = tmp
			}
			r1CS.A = append(r1CS.A, aConstraint)
			r1CS.B = append(r1CS.B, bConstraint)
			r1CS.C = append(r1CS.C, cConstraint)

		} else {
			panic("not a m gate")
		}
	}

	return
}

var Utils = prepareUtils()

func fractionToField(in [2]int) *big.Int {
	return Utils.FqR.Mul(big.NewInt(int64(in[0])), Utils.FqR.Inverse(big.NewInt(int64(in[1]))))

}

//Calculates the witness (program trace) given some input
//asserts that R1CS has been computed and is stored in the program p memory calling this function
func CalculateWitness(input []*big.Int, r1cs R1CS) (witness []*big.Int) {

	witness = r1csqap.ArrayOfBigZeros(len(r1cs.A[0]))
	set := make([]bool, len(witness))
	witness[0] = big.NewInt(int64(1))
	set[0] = true

	for i := range input {
		witness[i+1] = input[i]
		set[i+1] = true
	}

	zero := big.NewInt(int64(0))

	for i := 0; i < len(r1cs.A); i++ {
		gatesLeftInputs := r1cs.A[i]
		gatesRightInputs := r1cs.B[i]
		gatesOutputs := r1cs.C[i]

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
			//TODO replace with proper multiplication of b^-1 within the finite field
			witness[index] = big.NewInt(c / b)
			//Utils.FqR.Mul(sumOut, Utils.FqR.Inverse(sumRight))
		}

	}

	return
}

var hasher = sha256.New()

func hashTogether(a, b []byte) []byte {
	hasher.Reset()
	hasher.Write(a)
	hasher.Write(b)
	return hasher.Sum(nil)
}
