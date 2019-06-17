package circuitcompiler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"math/rand"
	"strings"
	"testing"
)

//factors are essential to identify, if a specific gate has been computed already
//eg. if we can extract a factor from a gate that is independent of commutativity, multiplicativitz we will do much better, in finding and reusing old outputs do
//minimize the multiplication gate number
// for example the gate a*b == gate b*a hence, we only need to compute one of both.

func TestFactorSignature(t *testing.T) {
	facNeutral := factors{&factor{multiplicative: [2]int{1, 1}}}

	//dont let the random number be to big, cuz of overflow
	r1, r2 := rand.Intn(1<<16), rand.Intn(1<<16)
	fmt.Println(r1, r2)
	equalityGroups := [][]factors{
		[]factors{ //test sign and gcd
			{&factor{multiplicative: [2]int{r1 * 2, -r2 * 2}}},
			{&factor{multiplicative: [2]int{-r1, r2}}},
			{&factor{multiplicative: [2]int{r1, -r2}}},
			{&factor{multiplicative: [2]int{r1 * 3, -r2 * 3}}},
			{&factor{multiplicative: [2]int{r1 * r1, -r2 * r1}}},
			{&factor{multiplicative: [2]int{r1 * r2, -r2 * r2}}},
		}, []factors{ //test kommutativity
			{&factor{multiplicative: [2]int{r1, -r2}}, &factor{multiplicative: [2]int{13, 27}}},
			{&factor{multiplicative: [2]int{13, 27}}, &factor{multiplicative: [2]int{-r1, r2}}},
		},
	}

	for _, equalityGroup := range equalityGroups {
		for i := 0; i < len(equalityGroup)-1; i++ {
			sig1, _, _ := factorsSignature(facNeutral, equalityGroup[i])
			sig2, _, _ := factorsSignature(facNeutral, equalityGroup[i+1])
			assert.Equal(t, sig1, sig2)
			sig1, _, _ = factorsSignature(equalityGroup[i], facNeutral)
			sig2, _, _ = factorsSignature(facNeutral, equalityGroup[i+1])
			assert.Equal(t, sig1, sig2)

			sig1, _, _ = factorsSignature(facNeutral, equalityGroup[i])
			sig2, _, _ = factorsSignature(equalityGroup[i+1], facNeutral)
			assert.Equal(t, sig1, sig2)

			sig1, _, _ = factorsSignature(equalityGroup[i], facNeutral)
			sig2, _, _ = factorsSignature(equalityGroup[i+1], facNeutral)
			assert.Equal(t, sig1, sig2)
		}
	}

}

func TestGate_ExtractValues(t *testing.T) {
	facNeutral := factors{&factor{multiplicative: [2]int{8, 7}}, &factor{multiplicative: [2]int{9, 3}}}
	facNeutral2 := factors{&factor{multiplicative: [2]int{9, 1}}, &factor{multiplicative: [2]int{13, 7}}}
	fmt.Println(factorsSignature(facNeutral, facNeutral2))
	f, fc := extractFactor(facNeutral)
	fmt.Println(f)
	fmt.Println(fc)

	f2, _ := extractFactor(facNeutral2)
	fmt.Println(f)
	fmt.Println(fc)
	fmt.Println(factorsSignature(facNeutral, facNeutral2))
	fmt.Println(factorsSignature(f, f2))
}

func TestGCD(t *testing.T) {
	fmt.Println(LCM(10, 15))
	fmt.Println(LCM(10, 15, 20))
	fmt.Println(LCM(1, 2, 3, 4, 5, 6, 7, 8, 9, 10))
}

var correctnesTest2 = []TraceCorrectnessTest{
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

func TestCorrectness2(t *testing.T) {

	for _, test := range correctnesTest2 {
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
