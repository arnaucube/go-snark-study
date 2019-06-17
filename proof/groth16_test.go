package proof

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/arnaucube/go-snark/circuit"
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
	parser := circuit.NewParser(strings.NewReader(code))
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
	setup := &Groth16Setup{}
	err = setup.Init(len(w), *circuit, alphas, betas, gammas)
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

	proof, err := setup.Generate(*circuit, w, px)
	assert.Nil(t, err)

	// fmt.Println("\n proofs:")
	// fmt.Println(proof)

	// fmt.Println("public signals:", proof.PublicSignals)
	fmt.Println("\nsignals:", circuit.Signals)
	fmt.Println("witness:", w)
	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	before := time.Now()
	assert.True(t, setup.Verify(*circuit, proof, publicSignalsVerif, true))
	fmt.Println("verify proof time elapsed:", time.Since(before))

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	assert.True(t, !setup.Verify(*circuit, proof, wrongPublicSignalsVerif, false))
}
