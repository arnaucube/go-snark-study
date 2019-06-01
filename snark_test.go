package snark

import (
	"fmt"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

func TestNewProgramm(t *testing.T) {

	flat := `
	def main(a,b,c,d):
		e = a * b
		f = c * d
		g = e * f
		h = g / e
		i = h * 5
		out = g * i
	`

	parser := circuitcompiler.NewParser(strings.NewReader(flat))
	program, err := parser.Parse()

	if err != nil {
		panic(err)
	}
	fmt.Println("\n unreduced")
	fmt.Println(flat)

	program.BuildConstraintTrees()
	program.PrintContraintTrees()
	fmt.Println("\nReduced gates")
	//PrintTree(froots["mul"])
	gates := program.ReduceCombinedTree()
	for _, g := range gates {
		fmt.Println(g)
	}

	fmt.Println("generating R1CS")
	r1cs := program.GenerateReducedR1CS(gates)
	//[[0 1 0 0 0 0 0 0 0 0] [0 0 0 1 0 0 0 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 1 0 0]]
	//[[0 0 1 0 0 0 0 0 0 0] [0 0 0 0 1 0 0 0 0 0] [0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 0 0 5 0]]
	//[[0 0 0 0 0 1 0 0 0 0] [0 0 0 0 0 0 1 0 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 1 0 0] [0 0 0 0 0 0 0 0 0 1]]
	a1 := big.NewInt(int64(6))
	a2 := big.NewInt(int64(5))
	inputs := []*big.Int{a1, a2, a1, a2}
	w := circuitcompiler.CalculateWitness(inputs, r1cs)

	r1csReordered, wReordered := RelocateOutput(program.GlobalInputCount(), r1cs, w)

	a, b, c := r1csReordered.A, r1csReordered.B, r1csReordered.C
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

	ax, bx, cx, px := Utils.PF.CombinePolynomials(wReordered, alphas, betas, gammas)
	fmt.Println("ax length", len(ax))
	fmt.Println("bx length", len(bx))
	fmt.Println("cx length", len(cx))
	fmt.Println("px length", len(px))

	hxQAP := Utils.PF.DivisorPolynomial(px, domain)
	fmt.Println("hx length", hxQAP)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hxQAP, domain))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)

	div, rem := Utils.PF.Div(px, domain)
	assert.Equal(t, hxQAP, div) //not necessary, since DivisorPolynomial is Div, just discarding 'rem'
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(len(px)-len(domain)))

	//calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), alphas, betas, gammas)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)
	//
	//// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
	//assert.Equal(t, domain, setup.Pk.Z)

	fmt.Println("hx pk.z", hxQAP)
	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	fmt.Println("hx pk.z", hx)
	// assert.Equal(t, hxQAP, hx)
	assert.Equal(t, px, Utils.PF.Mul(hxQAP, domain))
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(domain)+1)
	// fmt.Println("pk.Z", len(setup.Pk.Z))
	// fmt.Println("zxQAP", len(zxQAP))

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(setup, program.GlobalInputCount(), wReordered, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nwitness", w)
	fmt.Println("\nwitness Reordered ", wReordered)
	// b1 := big.NewInt(int64(1))
	//b35 := big.NewInt(int64(35))
	//// publicSignals := []*big.Int{b1, b35}
	//publicSignals := []*big.Int{b35}
	//before := time.Now()
	assert.True(t, VerifyProof(setup, proof, wReordered[:program.GlobalInputCount()], true))
	//fmt.Println("verify proof time elapsed:", time.Since(before))
}
