package circuitcompiler

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircuitParser(t *testing.T) {
	/*
		input:
		def test():
			y = x**3
			return x + y + 5

		flattened:
			m1 = s1 * s1
			m2 = m1 * s1
			m3 = m2 + s1
			out = m3 + 5

	*/

	// flat code, where er is expected_result
	// equals(s5, s1)
	// s1 = s5 * 1
	flat := `
	func test(private s0, public s1):
		s2 = s0*s0
		s3 = s2*s0
		s4 = s0 + s3
		s5 = s4 + 5
		s5 = s1 * one
		out = 1 * 1
	`
	parser := NewParser(strings.NewReader(flat))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println("circuit parsed: ", circuit)

	// flat code to R1CS
	fmt.Println("generating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("private inputs: ", circuit.PrivateInputs)
	fmt.Println("public inputs: ", circuit.PublicInputs)

	fmt.Println("signals:", circuit.Signals)

	// expected result
	// b0 := big.NewInt(int64(0))
	// b1 := big.NewInt(int64(1))
	// b5 := big.NewInt(int64(5))
	// aExpected := [][]*big.Int{
	//         []*big.Int{b0, b0, b1, b0, b0, b0},
	//         []*big.Int{b0, b0, b0, b1, b0, b0},
	//         []*big.Int{b0, b0, b1, b0, b1, b0},
	//         []*big.Int{b5, b0, b0, b0, b0, b1},
	// }
	// bExpected := [][]*big.Int{
	//         []*big.Int{b0, b0, b1, b0, b0, b0},
	//         []*big.Int{b0, b0, b1, b0, b0, b0},
	//         []*big.Int{b1, b0, b0, b0, b0, b0},
	//         []*big.Int{b1, b0, b0, b0, b0, b0},
	// }
	// cExpected := [][]*big.Int{
	//         []*big.Int{b0, b0, b0, b1, b0, b0},
	//         []*big.Int{b0, b0, b0, b0, b1, b0},
	//         []*big.Int{b0, b0, b0, b0, b0, b1},
	//         []*big.Int{b0, b1, b0, b0, b0, b0},
	// }
	//
	// assert.Equal(t, aExpected, a)
	// assert.Equal(t, bExpected, b)
	// assert.Equal(t, cExpected, c)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicInputs := []*big.Int{b35}
	// Calculate Witness
	w, err := circuit.CalculateWitness(privateInputs, publicInputs)
	assert.Nil(t, err)
	fmt.Println("w", w)

	circuitJson, _ := json.Marshal(circuit)
	fmt.Println("circuit:", string(circuitJson))

	assert.Equal(t, circuit.NPublic, 1)
	assert.Equal(t, len(circuit.PublicInputs), 1)
	assert.Equal(t, len(circuit.PrivateInputs), 1)
}
