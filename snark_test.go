package snark

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/groth16"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
)

func TestGroth16MinimalFlow(t *testing.T) {
	fmt.Println("testing Groth16 minimal flow")
	// circuit function
	// y = x^3 + x + 5
	code := `
	func main(private s0, public s1):
		s2 = s0 * s0
		s3 = s2 * s0
		s4 = s3 + s0
		s5 = s4 + 5
		equals(s1, s5)
		out = 1 * 1
	`
	fmt.Print("\ncode of the circuit:")

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	// code to R1CS
	fmt.Println("\ngenerating R1CS from code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, TODO remove
	alphas, betas, gammas, _ := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	assert.Equal(t, 7, len(ax))
	assert.Equal(t, 7, len(bx))
	assert.Equal(t, 7, len(cx))
	assert.Equal(t, 13, len(px))

	// ---
	// from here is the GROTH16
	// ---
	// calculate trusted setup
	fmt.Println("groth")
	setup, err := groth16.GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	div, rem := Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(6))

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)

	proof, err := groth16.GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nsignals:", circuit.Signals)
	fmt.Println("witness:", w)
	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	before := time.Now()
	assert.True(t, groth16.VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !groth16.VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, false))
}

func TestZkFromFlatCircuitCode(t *testing.T) {
	// compile circuit and get the R1CS

	// circuit function
	// y = x^3 + x + 5
	code := `
		func exp3(private a):
			b = a * a
			c = a * b
			return c
		func sum(private a, private b):
			c = a + b
			return c

		func main(private s0, public s1):
			s3 = exp3(s0)
			s4 = sum(s3, s0)
			s5 = s4 + 5
			equals(s1, s5)
			out = 1 * 1
	`
	// the same code without the functions calling, all in one func
	// code := `
	// func test(private s0, public s1):
	//         s2 = s0 * s0
	//         s3 = s2 * s0
	//         s4 = s3 + s0
	//         s5 = s4 + 5
	//         equals(s1, s5)
	//         out = 1 * 1
	// `
	fmt.Print("\ncode of the circuit:")
	fmt.Println(code)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	// fmt.Println("\ncircuit data:", circuit)
	// circuitJson, _ := json.Marshal(circuit)
	// fmt.Println("circuit:", string(circuitJson))

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	// code to R1CS
	fmt.Println("\ngenerating R1CS from code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, TODO remove
	alphas, betas, gammas, zxQAP := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 7, len(zxQAP))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	assert.Equal(t, 7, len(ax))
	assert.Equal(t, 7, len(bx))
	assert.Equal(t, 7, len(cx))
	assert.Equal(t, 13, len(px))

	hxQAP := Utils.PF.DivisorPolynomial(px, zxQAP)
	assert.Equal(t, 7, len(hxQAP))

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
	assert.Equal(t, zxQAP, setup.Pk.Z)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	assert.Equal(t, hx, hxQAP)
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

	proof, err := GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nsignals:", circuit.Signals)
	fmt.Println("witness:", w)
	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, false))
}

func TestZkMultiplication(t *testing.T) {
	code := `
	func main(private a, private b, public c):
		d = a * b
		equals(c, d)
		out = 1 * 1
	`
	fmt.Println("code", code)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	privateInputs := []*big.Int{b3, b4}
	b12 := big.NewInt(int64(12))
	publicSignals := []*big.Int{b12}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	// code to R1CS
	fmt.Println("\ngenerating R1CS from code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl. TODO remove
	alphas, betas, gammas, zxQAP := Utils.PF.R1CSToQAP(a, b, c)
	assert.Equal(t, 6, len(alphas))
	assert.Equal(t, 6, len(betas))
	assert.Equal(t, 6, len(betas))
	assert.Equal(t, 5, len(zxQAP))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	assert.Equal(t, 4, len(ax))
	assert.Equal(t, 4, len(bx))
	assert.Equal(t, 4, len(cx))
	assert.Equal(t, 7, len(px))

	hxQAP := Utils.PF.DivisorPolynomial(px, zxQAP)
	assert.Equal(t, 3, len(hxQAP))

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
	// fmt.Println("\nt:", setup.Toxic.T)

	// zx and setup.Pk.Z should be the same (currently not, the correct one is the calculation used inside GenerateTrustedSetup function), the calculation is repeated. TODO avoid repeating calculation
	assert.Equal(t, zxQAP, setup.Pk.Z)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	assert.Equal(t, 3, len(hx))
	assert.Equal(t, hx, hxQAP)

	div, rem = Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)

	proof, err := GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\n", circuit.Signals)
	fmt.Println("witness", w)
	b12Verif := big.NewInt(int64(12))
	publicSignalsVerif := []*big.Int{b12Verif}
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(11))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, false))
}

func TestMinimalFlow(t *testing.T) {
	// circuit function
	// y = x^3 + x + 5
	code := `
	func main(private s0, public s1):
		s2 = s0 * s0
		s3 = s2 * s0
		s4 = s3 + s0
		s5 = s4 + 5
		equals(s1, s5)
		out = 1 * 1
	`
	fmt.Print("\ncode of the circuit:")
	fmt.Println(code)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(code))
	circuit, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	// wittness
	w, err := circuit.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	// code to R1CS
	fmt.Println("\ngenerating R1CS from code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, TODO remove
	alphas, betas, gammas, _ := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	assert.Equal(t, 7, len(ax))
	assert.Equal(t, 7, len(bx))
	assert.Equal(t, 7, len(cx))
	assert.Equal(t, 13, len(px))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	div, rem := Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(6))

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)

	proof, err := GenerateProofs(*circuit, setup, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nsignals:", circuit.Signals)
	fmt.Println("witness:", w)
	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !VerifyProof(*circuit, setup, proof, wrongPublicSignalsVerif, false))
}
