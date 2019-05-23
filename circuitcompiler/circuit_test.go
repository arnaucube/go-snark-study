package circuitcompiler

import (
	//"fmt"
	////"math/big"
	//"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXor(t *testing.T) {
	assert.Equal(t, false, Xor(true, true))
	assert.Equal(t, true, Xor(true, false))
	assert.Equal(t, true, Xor(false, true))
	assert.Equal(t, false, Xor(false, false))

}

func TestCircuitParser(t *testing.T) {

	//flat := `
	//func main(a,b):
	//	c = a / b
	//	d = c * b
	//	e = d - c
	//	f = c * 55
	//	out = f / e
	//`
	//flat := `
	//func test(a,b):
	//	d = a / b
	//	c = a + d
	//	f = a * 55
	//	g = a / d
	//	h = g + f
	//	i = c + h
	//	out = i / c
	//`
	//parser := NewParser(strings.NewReader(flat))
	//programm, err := parser.Parse()
	//circuit := programm.getMainCircuit()
	//assert.Nil(t, err)
	//fmt.Println("\n unreduced")
	//fmt.Println(flat)

	//fmt.Println("generating R1CS from flat code")
	//a, b, c := circuit.GenerateR1CS()
	//fmt.Println(a)
	//fmt.Println(b)
	//fmt.Println(c)
	//
	//a1 := big.NewInt(int64(6))
	//a2 := big.NewInt(int64(5))
	//inputs := []*big.Int{a1, a2}
	//// Calculate Witness
	//w, err := circuit.CalculateWitness(inputs)
	//assert.Nil(t, err)
	//fmt.Println("w", w)
	//fmt.Printf("inputs %s", circuit.Inputs)
	//fmt.Printf("signals %s", circuit.Signals)

	//
	//fmt.Println("Reduced Tree Parsing")
	//r := circuit.BuildConstraintTree()
	//constraintReduced := ReduceTree(r)
	//fmt.Printf("depth %v, mGates %v \n", printDepth(r, 0), CountMultiplicationGates(r))
	//printTree(r, 0)
	//
	//a,b,c,w = circuit.GenerateReducedR1CSandWitness(inputs,constraintReduced)
	//fmt.Println("\ngenerating R1CS from reduced flat code")
	//fmt.Println("\nR1CS:")
	//fmt.Println("a:", a)
	//fmt.Println("b:", b)
	//fmt.Println("c:", c)
	//fmt.Println("w", w)

	//// R1CS to QAP
	//alphas, betas, gammas, zxQAP := fields.R1CSToQAP(a, b, c)
	//fmt.Println("qap")
	//fmt.Println("alphas", len(alphas))
	//fmt.Println("alphas", alphas[0])
	//fmt.Println("betas", len(betas))
	//fmt.Println("gammas", len(gammas))
	//fmt.Println("zx length", len(zxQAP))
	//circuit.reduceAdditionGates()
	//
	//fmt.Println(circuit)
	//fmt.Println("generating R1CS from flat code")
	//a, b, c = circuit.GenerateR1CS()
	//
	//fmt.Println(a)
	//fmt.Println(b)
	//fmt.Println(c)
	//
	//// Calculate Witness
	//w, err = circuit.CalculateWitness(inputs)
	//assert.Nil(t, err)
	//fmt.Println("w", w)

	//// expected result
	//b0 := big.NewInt(int64(0))
	//b1 := big.NewInt(int64(1))
	//b5 := big.NewInt(int64(5))
	//aExpected := [][]*big.Int{
	//	[]*big.Int{b0, b0, b1, b0, b0, b0},
	//	[]*big.Int{b0, b0, b0, b1, b0, b0},
	//	[]*big.Int{b0, b0, b1, b0, b1, b0},
	//	[]*big.Int{b5, b0, b0, b0, b0, b1},
	//}
	//bExpected := [][]*big.Int{
	//	[]*big.Int{b0, b0, b1, b0, b0, b0},
	//	[]*big.Int{b0, b0, b1, b0, b0, b0},
	//	[]*big.Int{b1, b0, b0, b0, b0, b0},
	//	[]*big.Int{b1, b0, b0, b0, b0, b0},
	//}
	//cExpected := [][]*big.Int{
	//	[]*big.Int{b0, b0, b0, b1, b0, b0},
	//	[]*big.Int{b0, b0, b0, b0, b1, b0},
	//	[]*big.Int{b0, b0, b0, b0, b0, b1},
	//	[]*big.Int{b0, b1, b0, b0, b0, b0},
	//}

	//assert.Equal(t, aExpected, a)
	//assert.Equal(t, bExpected, b)
	//assert.Equal(t, cExpected, c)

}
