package circuitcompiler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

type InOut struct {
	inputs []*big.Int
	result *big.Int
}

type TraceCorrectnessTest struct {
	code string
	io   []InOut
}

var bigNumberResult1, _ = new(big.Int).SetString("2297704271284150716235246193843898764109352875", 10)
var bigNumberResult2, _ = new(big.Int).SetString("75263346540254220740876250", 10)

var correctnesTest = []TraceCorrectnessTest{
	{
		io: []InOut{{
			inputs: []*big.Int{big.NewInt(int64(7)), big.NewInt(int64(11))},
			result: big.NewInt(int64(1729500084900343)),
		}, {
			inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},

			result: bigNumberResult1,
		}},
		code: `
	def main( x  ,  z ) :
		out = do(z) + add(x,x)

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
	

	def mul(a,b):
		out = a * b
	`,
	},
	{io: []InOut{{
		inputs: []*big.Int{big.NewInt(int64(7))},
		result: big.NewInt(int64(4)),
	}},
		code: `
	def mul(a,b):
		out = a * b
	
	def main(a):
		b = a * a
		c = 4 - b
		d = 5 * c
		out =  mul(d,c) /  mul(b,b)
	`,
	},
	{io: []InOut{{
		inputs: []*big.Int{big.NewInt(int64(7)), big.NewInt(int64(11))},
		result: big.NewInt(int64(22638)),
	}, {
		inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},
		result: bigNumberResult2,
	}},
		code: `
	def main(a,b):
		d = b + b
		c = a * d
		e = c - a
		out = e * c
	`,
	},
	{
		io: []InOut{{
			inputs: []*big.Int{big.NewInt(int64(643)), big.NewInt(int64(76548465))},
			result: big.NewInt(int64(98441327276)),
		}, {
			inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},
			result: big.NewInt(int64(8675445947220)),
		}},
		code: `
	def main(a,b):
		c = a + b
		e = c - a
		f = e + b
		g = f + 2
		out = g * a
	`,
	},
	{
		io: []InOut{{
			inputs: []*big.Int{big.NewInt(int64(3)), big.NewInt(int64(5)), big.NewInt(int64(7)), big.NewInt(int64(11))},
			result: big.NewInt(int64(444675)),
		}},
		code: `
	def main(a,b,c,d):
		e = a * b
		f = c * d
		g = e * f
		h = g / e
		i = h * 5
		out = g * i
	`,
	},
	{
		io: []InOut{{
			inputs: []*big.Int{big.NewInt(int64(3)), big.NewInt(int64(5)), big.NewInt(int64(7)), big.NewInt(int64(11))},
			result: big.NewInt(int64(264)),
		}},
		code: `
	def main(a,b,c,d):
		e = a * 3
		f = b * 7
		g = c * 11
		h = d * 13
		i = e + f
		j = g + h
		k = i + j
		out = k * 1
	`,
	},
}

func TestCorrectness(t *testing.T) {

	for _, test := range correctnesTest {
		parser := NewParser(strings.NewReader(test.code))
		program, err := parser.Parse()

		if err != nil {
			panic(err)
		}
		fmt.Println("\n unreduced")
		fmt.Println(test.code)

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

		fmt.Println("\n generating R1CS")
		r1cs := program.GenerateReducedR1CS(gates)
		fmt.Println(r1cs.A)
		fmt.Println(r1cs.B)
		fmt.Println(r1cs.C)

		for _, io := range test.io {
			inputs := io.inputs
			fmt.Println("input")
			fmt.Println(inputs)
			w := CalculateWitness(inputs, r1cs)
			fmt.Println("witness")
			fmt.Println(w)
			assert.Equal(t, io.result, w[program.GlobalInputCount()])
		}

	}

}

//test to check gate optimisation
//mess around the code s.t. results is unchanged. number of gates should remain the same in any case
func TestGateOptimisation(t *testing.T) {

	io := InOut{
		inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},

		result: bigNumberResult1,
	}

	equalCodes := []string{
		`
	def main( x  ,  z ) :
		out = do(z) + add(x,x)

	def do(x):
		e = x * 5
		b = e * 6
		c = 7 * b
		f = c * 1
		d = f * c
		out = d * mul(d,e)
	
	def add(x ,k):
		z = k * x
		out = do(x) + mul(x,z)
	

	def mul(a,b):
		out = b * a
	`, //switching order
		`
	def main( x  ,  z ) :
		out = do(z) + add(x,x)

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
	

	def mul(a,b):
		out = a * b
	`, //switching order
		`
	def main( x  ,  z ) :
		out = do(z) + add(x,x)

	def do(x):
		e = x * 5
		j = e * 3
		k = e * 3
		b = j+k
		c = b * 7
		f = c * 1
		d = c * f
		g = d * 1
		out = g * mul(d,e)
	
	def add(k ,x):
		z = k * x
		out = do(x) + mul(x,z)
	

	def mul(b,a):
		out = a * b
	`, `
	def main( x  ,  z ) :
		out =  add(x,x)+do(z)

	def do(x):
		e = x * 5
		j = 3 * e
		k = e * 3
		b = j+k
		c = b * 7
		f = c * 1
		d = c * f
		g = d * 1
		out = mul(d,e) * g
	
	def add(k ,x):
		z = k * x
		out =  mul(x,z) + do(x)
	

	def mul(b,a):
		out = a * b
	`,
	}

	var r1css = make([]R1CS, len(equalCodes))

	for i, c := range equalCodes {
		parser := NewParser(strings.NewReader(c))
		program, err := parser.Parse()

		if err != nil {
			panic(err)
		}
		program.BuildConstraintTrees()

		gates := program.ReduceCombinedTree()
		for _, g := range gates {
			fmt.Printf("\n %v", g)
		}
		fmt.Println("\n generating R1CS")
		r1cs := program.GenerateReducedR1CS(gates)
		r1css[i] = r1cs
		fmt.Println(r1cs.A)
		fmt.Println(r1cs.B)
		fmt.Println(r1cs.C)
	}

	for i := 0; i < len(equalCodes)-1; i++ {
		assert.Equal(t, len(r1css[i].A), len(r1css[i+1].A))
	}

	for i := 0; i < len(equalCodes); i++ {
		//assert.Equal(t, len(r1css[i].A), len(r1css[i+1].A))
		w := CalculateWitness(io.inputs, r1css[i])
		fmt.Println("witness")
		fmt.Println(w)
		assert.Equal(t, io.result, w[3])

	}

}
