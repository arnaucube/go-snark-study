package snark

import (
	"fmt"
	"github.com/mottla/go-snark/circuitcompiler"
	"github.com/mottla/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
)

func TestNewProgramm(t *testing.T) {

	flat := `

	func add(x ,k):
		z = k * x
		out = x + mul(x,z)
	
	func main(a,b):
		out = add(a,b) * a

	func mul(a,b):
		out = a * b	
	`

	parser := circuitcompiler.NewParser(strings.NewReader(flat))
	program, err := parser.Parse()

	if err != nil {
		panic(err)
	}
	fmt.Println("\n unreduced")
	fmt.Println(flat)

	program.BuildConstraintTrees()
	program.PrintConstraintTrees()
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

	// R1CS to QAP
	alphas, betas, gammas, zxQAP := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println("alphas", len(alphas))
	fmt.Println("alphas", alphas)
	fmt.Println("betas", len(betas))
	fmt.Println("gammas", len(gammas))
	fmt.Println("zx length", len(zxQAP))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	fmt.Println("ax length", len(ax))
	fmt.Println("bx length", len(bx))
	fmt.Println("cx length", len(cx))
	fmt.Println("px length", len(px))

	hxQAP := Utils.PF.DivisorPolynomial(px, zxQAP)
	fmt.Println("hx length", len(hxQAP))

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hzQAP := Utils.PF.Mul(hxQAP, zxQAP)
	assert.Equal(t, abc, hzQAP)

	div, rem := Utils.PF.Div(px, zxQAP)
	assert.Equal(t, hxQAP, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	// calculate trusted setup
	//setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	//assert.Nil(t, err)
	//fmt.Println("\nt:", setup.Toxic.T)
	//
	//// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
	//// assert.Equal(t, zxQAP, setup.Pk.Z)
	//
	//fmt.Println("hx pk.z", hxQAP)
	//hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	//fmt.Println("hx pk.z", hx)
	//// assert.Equal(t, hxQAP, hx)
	//assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	//assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))
	//
	//assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	//assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)
	//// fmt.Println("pk.Z", len(setup.Pk.Z))
	//// fmt.Println("zxQAP", len(zxQAP))
	//
	//// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	//proof, err := GenerateProofs(*circuit, setup, w, px)
	//assert.Nil(t, err)
	//
	//// fmt.Println("\n proofs:")
	//// fmt.Println(proof)
	//
	//// fmt.Println("public signals:", proof.PublicSignals)
	//fmt.Println("\nwitness", w)
	//// b1 := big.NewInt(int64(1))
	//b35 := big.NewInt(int64(35))
	//// publicSignals := []*big.Int{b1, b35}
	//publicSignals := []*big.Int{b35}
	//before := time.Now()
	//assert.True(t, VerifyProof(*circuit, setup, proof, publicSignals, true))
	//fmt.Println("verify proof time elapsed:", time.Since(before))
}
