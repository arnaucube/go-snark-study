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
	raw := `
	y = x^x
	z = x + y
	out = z + 5
	`
	parser := NewParser(strings.NewReader(raw))
	res, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println(res)

	// flat code
	// flat code to R1CS
}
