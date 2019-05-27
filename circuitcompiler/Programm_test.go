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
	//flat := `
	//
	//def add(x ,k):
	//	z = k * x
	//	out = 6 + mul(x,z)
	//
	//def main(x,z):
	//	out = mul(x,z) - add(x,x)
	//
	//def mul(a,b):
	//	out = a * b
	//`
	flat := `
	def do(x):
		e = x * 5
		b = e * 6
		c = b * 7
		f = c * 1
		d = c * f
		out = d * mul(d,e)
	
	def add(x ,k):
		z = k * x
		out = do(x) + mul(x,z)
	
	def main(x,z):
		out = do(z) + add(x,x)
	
	def mul(a,b):
		out = a * b
	`
	//flat := `
	//func mul(a,b):
	//	out = a * b
	//
	//func main(a,aa):
	//	b = a * a
	//	c = 4 - b
	//	d = 5 * c
	//	out =  mul(d,c) /  mul(b,b)
	//`
	//flat := `
	//func main(a,b):
	//	c = a + b
	//	e = c - a
	//	f = e + b
	//	g = f + 2
	//	out = g * a
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
		fmt.Printf("\n %v", g)
	}

	fmt.Println("generating R1CS")
	a, b, c := program.GenerateReducedR1CS(gates)
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	a1 := big.NewInt(int64(7))
	a2 := big.NewInt(int64(11))
	inputs := []*big.Int{a1, a2}
	w := program.CalculateWitness(inputs)
	fmt.Println("witness")
	fmt.Println(w)

}
