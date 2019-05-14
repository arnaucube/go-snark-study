package circuitcompiler

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestProgramm_BuildConstraintTree(t *testing.T) {
	line := "asdf asfd"
	line = strings.TrimFunc(line, func(i rune) bool { return isWhitespace(i) })
	fmt.Println(line)
}

func TestNewProgramm(t *testing.T) {

	flat := `
	func do(x):
		e = x * 5
		b = e * 6
		c = b * 7
		f = c * 1
		d = c / f
		out = d * 1
	
	func add(x ,k):
		z = k * x
		out = do(x) + mul(x,z)
	
	func main(x,z):	
		out = do(x) + 4
	
	func mul(a,b):
		out = a * b
	`

	//flat := `
	//func main(a,b):
	//	e = 4 * a
	//	c = a * e
	//	d = c * 70
	//	out = a * d
	//`
	parser := NewParser(strings.NewReader(flat))
	program, err := parser.Parse()

	if err != nil {
		panic(err)
	}
	fmt.Println("\n unreduced")
	fmt.Println(flat)

	program.BuildConstraintTrees()
	for k, v := range program.functions {
		fmt.Println(k)
		PrintTree(v.root)
	}

	fmt.Println("\nReduced gates")
	//PrintTree(froots["mul"])
	gates := program.ReduceCombinedTree()
	for _, g := range gates {
		fmt.Println(g)
	}

	fmt.Println("generating R1CS")
	a, b, c := program.GenerateReducedR1CS(gates)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	a1 := big.NewInt(int64(6))
	a2 := big.NewInt(int64(5))
	inputs := []*big.Int{a1, a2}
	w := program.CalculateWitness(inputs)
	fmt.Println("witness")
	fmt.Println(w)

}
