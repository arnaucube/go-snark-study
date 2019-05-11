package snark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
)

func TestZkFromFlatCircuitCode(t *testing.T) {
	// compile circuit and get the R1CS

	// circuit function
	// y = x^3 + x + 5
	flatCode := `
	func test(private s0, public s1):
		s2 = s0 * s0
		s3 = s2 * s0
		s4 = s3 + s0
		s5 = s4 + 5
		equals(s1, s5)
		out = 1 * 1
	`
	fmt.Print("\nflat code of the circuit:")
	fmt.Println(flatCode)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println("\ncircuit data:", circuit)
	circuitJson, _ := json.Marshal(circuit)
	fmt.Println("circuit:", string(circuitJson))

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)
	fmt.Println("\n", circuit.Signals)
	fmt.Println("witness", w)

	// flat code to R1CS
	fmt.Println("\ngenerating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, bad calculated. TODO remove
	alphas, betas, gammas, zxQAP := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println("alphas", len(alphas))
	fmt.Println("alphas[1]", alphas[1])
	fmt.Println("betas", len(betas))
	fmt.Println("gammas", len(gammas))
	fmt.Println("zx length", len(zxQAP))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	fmt.Println("ax length", len(ax))
	fmt.Println("bx length", len(bx))
	fmt.Println("cx length", len(cx))
	fmt.Println("px length", len(px))
	fmt.Println("px[last]", px[0])

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
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(6))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)

	// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
	// assert.Equal(t, zxQAP, setup.Pk.Z)

	fmt.Println("hx pk.z", hxQAP)
	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	fmt.Println("hx pk.z", hx)
	// assert.Equal(t, hxQAP, hx)
	div, rem = Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(6))

	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)

	// fmt.Println("pk.Z", len(setup.Pk.Z))
	// fmt.Println("zxQAP", len(zxQAP))

	proof, err := GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\n", circuit.Signals)
	fmt.Println("\nwitness", w)
	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, true))
}

func TestZkMultiplication(t *testing.T) {
	flatCode := `
	func test(private a, private b, public c):
		d = a * b
		equals(c, d)
		out = 1 * 1
	`
	fmt.Print("\nflat code of the circuit:")
	fmt.Println(flatCode)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println("\ncircuit data:", circuit)
	circuitJson, _ := json.Marshal(circuit)
	fmt.Println("circuit:", string(circuitJson))

	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	privateInputs := []*big.Int{b3, b4}
	b12 := big.NewInt(int64(12))
	publicSignals := []*big.Int{b12}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)
	fmt.Println("\n", circuit.Signals)
	fmt.Println("witness", w)

	// flat code to R1CS
	fmt.Println("\ngenerating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, bad calculated. TODO remove
	alphas, betas, gammas, zxQAP := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println("alphas", len(alphas))
	fmt.Println("alphas[1]", alphas[1])
	fmt.Println("betas", len(betas))
	fmt.Println("gammas", len(gammas))
	fmt.Println("zx length", len(zxQAP))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	fmt.Println("ax length", len(ax))
	fmt.Println("bx length", len(bx))
	fmt.Println("cx length", len(cx))
	fmt.Println("px length", len(px))
	fmt.Println("px[last]", px[0])

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
	setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)

	// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
	// assert.Equal(t, zxQAP, setup.Pk.Z)

	fmt.Println("hx pk.z", hxQAP)
	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	fmt.Println("hx pk.z", hx)
	// assert.Equal(t, hxQAP, hx)
	div, rem = Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)

	// fmt.Println("pk.Z", len(setup.Pk.Z))
	// fmt.Println("zxQAP", len(zxQAP))

	proof, err := GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\n", circuit.Signals)
	fmt.Println("\nwitness", w)
	b12Verif := big.NewInt(int64(12))
	publicSignalsVerif := []*big.Int{b12Verif}
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(11))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, true))
}

/*
func TestZkFromHardcodedR1CS(t *testing.T) {
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b3 := big.NewInt(int64(3))
	b5 := big.NewInt(int64(5))
	b9 := big.NewInt(int64(9))
	b27 := big.NewInt(int64(27))
	b30 := big.NewInt(int64(30))
	b35 := big.NewInt(int64(35))
	a := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b1, b0},
		[]*big.Int{b5, b0, b0, b0, b0, b1},
	}
	b := [][]*big.Int{
		[]*big.Int{b0, b0, b1, b0, b0, b0},
		[]*big.Int{b0, b0, b1, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
	}
	c := [][]*big.Int{
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b1},
		[]*big.Int{b0, b1, b0, b0, b0, b0},
	}
	alphas, betas, gammas, zx := Utils.PF.R1CSToQAP(a, b, c)

	// wittness = 1, 35, 3, 9, 27, 30
	w := []*big.Int{b1, b35, b3, b9, b27, b30}
	circuit := circuitcompiler.Circuit{
		NVars:    6,
		NPublic:  1,
		NSignals: len(w),
	}
	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := Utils.PF.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	div, rem := Utils.PF.Div(px, zx)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas, zx)
	assert.Nil(t, err)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(circuit, setup, hx, w)
	assert.Nil(t, err)

	// assert.True(t, VerifyProof(circuit, setup, proof, true))
	publicSignals := []*big.Int{b35}
	assert.True(t, VerifyProof(circuit, setup, proof, publicSignals, true))
}

func TestZkMultiplication(t *testing.T) {

	// compile circuit and get the R1CS
	flatCode := `
	func test(a, b):
		out = a * b
	`

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	inputs := []*big.Int{b3, b4}
	// wittness
	w, err := circuit.CalculateWitness(inputs)
	assert.Nil(t, err)

	// flat code to R1CS
	a, b, c := circuit.GenerateR1CS()

	// R1CS to QAP
	alphas, betas, gammas, zx := Utils.PF.R1CSToQAP(a, b, c)

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := Utils.PF.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	div, rem := Utils.PF.Div(px, zx)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(1))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas, zx)
	assert.Nil(t, err)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(*circuit, setup, hx, w)
	assert.Nil(t, err)

	// assert.True(t, VerifyProof(*circuit, setup, proof, false))
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignals, true))
}
*/
