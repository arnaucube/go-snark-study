package circuitcompiler

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircuitParser(t *testing.T) {
	// y = x^3 + x + 5
	flat := `
	func main(private s0, public s1):
		s2 = s0 * s0
		s3 = s2 * s0
		s4 = s3 + s0
		s5 = s4 + 5
		equals(s1, s5)
		out = 1 * 1
	`
	parser := NewParser(strings.NewReader(flat))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	// flat code to R1CS
	a, b, c := circuit.GenerateR1CS()
	assert.Equal(t, "s0", circuit.PrivateInputs[0])
	assert.Equal(t, "s1", circuit.PublicInputs[0])

	assert.Equal(t, []string{"one", "s1", "s0", "s2", "s3", "s4", "s5", "out"}, circuit.Signals)

	// expected result
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b5 := big.NewInt(int64(5))
	aExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b1, b0, b0, b0},
		[]*big.Int{b5, b0, b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
	}
	bExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
	}
	cExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b1, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b0, b1},
	}

	assert.Equal(t, aExpected, a)
	assert.Equal(t, bExpected, b)
	assert.Equal(t, cExpected, c)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicInputs := []*big.Int{b35}
	// Calculate Witness
	w, err := circuit.CalculateWitness(privateInputs, publicInputs)
	assert.Nil(t, err)
	b9 := big.NewInt(int64(9))
	b27 := big.NewInt(int64(27))
	b30 := big.NewInt(int64(30))
	wExpected := []*big.Int{b1, b35, b3, b9, b27, b30, b35, b1}
	assert.Equal(t, wExpected, w)

	// circuitJson, _ := json.Marshal(circuit)
	// fmt.Println("circuit:", string(circuitJson))

	assert.Equal(t, circuit.NPublic, 1)
	assert.Equal(t, len(circuit.PublicInputs), 1)
	assert.Equal(t, len(circuit.PrivateInputs), 1)
}

func TestCircuitWithFuncCallsParser(t *testing.T) {
	// y = x^3 + x + 5
	code := `
		func exp3(private a):
			b = a * a
			c = a * b
			return c
		func sum(private a, private b):
			c = a + b
			return c

		func main(private s0, public s1):
			s3 = exp3(s0)
			s4 = sum(s3, s0)
			s5 = s4 + 5
			equals(s1, s5)
			out = 1 * 1
	`
	parser := NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	// flat code to R1CS
	a, b, c := circuit.GenerateR1CS()
	assert.Equal(t, "s0", circuit.PrivateInputs[0])
	assert.Equal(t, "s1", circuit.PublicInputs[0])

	assert.Equal(t, []string{"one", "s1", "s0", "b0", "s3", "s4", "s5", "out"}, circuit.Signals)

	// expected result
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b5 := big.NewInt(int64(5))
	aExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b1, b0, b0, b0},
		[]*big.Int{b5, b0, b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
	}
	bExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0, b0, b0},
	}
	cExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b1, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b0, b0, b1},
	}

	assert.Equal(t, aExpected, a)
	assert.Equal(t, bExpected, b)
	assert.Equal(t, cExpected, c)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicInputs := []*big.Int{b35}
	// Calculate Witness
	w, err := circuit.CalculateWitness(privateInputs, publicInputs)
	assert.Nil(t, err)
	b9 := big.NewInt(int64(9))
	b27 := big.NewInt(int64(27))
	b30 := big.NewInt(int64(30))
	wExpected := []*big.Int{b1, b35, b3, b9, b27, b30, b35, b1}
	assert.Equal(t, wExpected, w)

	// circuitJson, _ := json.Marshal(circuit)
	// fmt.Println("circuit:", string(circuitJson))

	assert.Equal(t, circuit.NPublic, 1)
	assert.Equal(t, len(circuit.PublicInputs), 1)
	assert.Equal(t, len(circuit.PrivateInputs), 1)
}
