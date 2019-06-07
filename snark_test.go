package snark

import (
	"fmt"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
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
	//{
	//	io: []InOut{{
	//		inputs: []*big.Int{big.NewInt(int64(7)), big.NewInt(int64(11))},
	//		result: big.NewInt(int64(1729500084900343)),
	//	}, {
	//		inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},
	//
	//		result: bigNumberResult1,
	//	}},
	//	code: `
	//def main( x  ,  z ) :
	//	out = do(z) + add(x,x)
	//
	//def do(x):
	//	e = x * 5
	//	b = e * 6
	//	c = b * 7
	//	f = c * 1
	//	d = c * f
	//	out = d * mul(d,e)
	//
	//def add(x ,k):
	//	z = k * x
	//	out = do(x) + mul(x,z)
	//
	//
	//def mul(a,b):
	//	out = a * b
	//`,
	//},
	//{io: []InOut{{
	//	inputs: []*big.Int{big.NewInt(int64(7))},
	//	result: big.NewInt(int64(4)),
	//}},
	//	code: `
	//def mul(a,b):
	//	out = a * b
	//
	//def main(a):
	//	b = a * a
	//	c = 4 - b
	//	d = 5 * c
	//	out =  mul(d,c) /  mul(b,b)
	//`,
	//},
	//{io: []InOut{{
	//	inputs: []*big.Int{big.NewInt(int64(7)), big.NewInt(int64(11))},
	//	result: big.NewInt(int64(22638)),
	//}, {
	//	inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},
	//	result: bigNumberResult2,
	//}},
	//	code: `
	//def main(a,b):
	//	d = b + b
	//	c = a * d
	//	e = c - a
	//	out = e * c
	//`,
	//},
	//{
	//	io: []InOut{{
	//		inputs: []*big.Int{big.NewInt(int64(643)), big.NewInt(int64(76548465))},
	//		result: big.NewInt(int64(98441327276)),
	//	}, {
	//		inputs: []*big.Int{big.NewInt(int64(365235)), big.NewInt(int64(11876525))},
	//		result: big.NewInt(int64(8675445947220)),
	//	}},
	//	code: `
	//def main(a,b):
	//	c = a + b
	//	e = c - a
	//	f = e + b
	//	g = f + 2
	//	out = g * a
	//`,
	//},
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
}

func TestGenerateAndVerifyProof(t *testing.T) {

	for _, test := range correctnesTest {

		parser := circuitcompiler.NewParser(strings.NewReader(test.code))
		program, err := parser.Parse()

		if err != nil {
			panic(err)
		}
		fmt.Println("\n unreduced")
		fmt.Println(test.code)

		program.BuildConstraintTrees()
		program.PrintContraintTrees()
		fmt.Println("\nReduced gates")
		//PrintTree(froots["mul"])
		gates := program.ReduceCombinedTree()
		for _, g := range gates {
			fmt.Println(g)
		}

		fmt.Println("generating R1CS")
		//NOTE MOVE DOES NOTHING CURRENTLY
		r1cs := program.GenerateReducedR1CS(gates)
		//[[0 1 0 0 0 0 0 0 0 0] [0 0 0 1 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 1 0 0]]
		//[[0 0 1 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 5 0]]
		//[[0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 0 0 1]]

		a, b, c := r1cs.A, r1cs.B, r1cs.C
		fmt.Println(a)
		fmt.Println(b)
		fmt.Println(c)

		// R1CS to QAP
		alphas, betas, gammas, domain := Utils.PF.R1CSToQAP(a, b, c)
		fmt.Println("QAP array lengths")
		fmt.Println("alphas", len(alphas))
		fmt.Println("betas", len(betas))
		fmt.Println("gammas", len(gammas))
		fmt.Println("domain polynomial ", len(domain))

		before := time.Now()
		//calculate trusted setup
		setup, err := GenerateTrustedSetup(program.GlobalInputCount()+program.GlobalOutputCount(), alphas, betas, gammas)
		fmt.Println("Generate CRS time elapsed:", time.Since(before))
		assert.Nil(t, err)
		fmt.Println("\nt:", setup.Toxic.T)

		for _, io := range test.io {

			inputs := io.inputs
			fmt.Println("input")
			fmt.Println(inputs)
			w := circuitcompiler.CalculateWitness(inputs, r1cs)
			fmt.Println("\nwitness", w)

			assert.Equal(t, io.result, w[program.GlobalInputCount()])

			ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
			fmt.Println("ax length", len(ax))
			fmt.Println("bx length", len(bx))
			fmt.Println("cx length", len(cx))
			fmt.Println("px length", len(px))

			hxQAP := Utils.PF.DivisorPolynomial(px, domain)
			fmt.Println("hx length", len(hxQAP))

			// hx==px/zx so px==hx*zx
			assert.Equal(t, px, Utils.PF.Mul(hxQAP, domain))

			// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
			abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
			assert.Equal(t, abc, px)

			div, rem := Utils.PF.Div(px, domain)
			assert.Equal(t, hxQAP, div) //not necessary, since DivisorPolynomial is Div, just discarding 'rem'
			assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(len(px)-len(domain)))

			//// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
			//assert.Equal(t, domain, setup.Pk.Z)

			hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)

			// assert.Equal(t, hxQAP, hx)
			assert.Equal(t, px, Utils.PF.Mul(hxQAP, domain))
			assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))
			assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
			assert.Equal(t, len(hxQAP), len(px)-len(domain)+1)

			before := time.Now()
			// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
			proof, err := GenerateProofs(setup, program.GlobalInputCount()+program.GlobalOutputCount(), w, px)
			fmt.Println("proof generation time elapsed:", time.Since(before))
			assert.Nil(t, err)
			fmt.Println(program.GlobalInputCount() + program.GlobalOutputCount())
			before = time.Now()
			Signals := w[:program.GlobalInputCount()+program.GlobalOutputCount()]
			assert.True(t, VerifyProof(setup, proof, Signals, true))
			fmt.Println("verify proof time elapsed:", time.Since(before))

		}

	}

}
