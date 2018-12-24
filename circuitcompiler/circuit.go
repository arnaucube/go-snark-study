package circuitcompiler

import (
	"fmt"
	"math/big"
)

type Circuit struct {
	NVars       int
	NPublic     int
	NSignals    int
	Inputs      []string
	Signals     []string
	Witness     []*big.Int
	Constraints []Constraint
	R1CS        struct {
		A [][]*big.Int
		B [][]*big.Int
		C [][]*big.Int
	}
}

func (c *Circuit) GenerateR1CS() {
	fmt.Print("function with inputs: ")
	fmt.Println(c.Inputs)
	fmt.Print("signals: ")
	fmt.Println(c.Signals)
	for _, constraint := range c.Constraints {
		fmt.Println(constraint.Literal)

	}
}
