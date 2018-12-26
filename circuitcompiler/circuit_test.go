package circuitcompiler

import (
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

	// flat code
	flat := `
	func test(x):
		aux = x*x
		y = aux*x
		z = x + y
		out = z + 5
	`
	parser := NewParser(strings.NewReader(flat))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println(circuit)

	// flat code to R1CS
	fmt.Println("generating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Print("function with inputs: ")
	fmt.Println(circuit.Inputs)

	fmt.Print("signals: ")
	fmt.Println(circuit.Signals)

	// expected result
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b5 := big.NewInt(int64(5))
	aExpected := [][]*big.Int{
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b1, b0, b0, b1, b0},
		[]*big.Int{b5, b0, b0, b0, b0, b1},
	}
	bExpected := [][]*big.Int{
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
	}
	cExpected := [][]*big.Int{
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b1},
		[]*big.Int{b0, b0, b1, b0, b0, b0},
	}

	assert.Equal(t, aExpected, a)
	assert.Equal(t, bExpected, b)
	assert.Equal(t, cExpected, c)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
