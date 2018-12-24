package circuitcompiler

import (
	"fmt"
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
	circuit.GenerateR1CS()
	fmt.Println(circuit.Inputs)
}
