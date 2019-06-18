package proof

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/fields"
)

func TestGroth16MinimalFlow(t *testing.T) {
	code := `
	func main(private s0, public s1):
		s2 = s0 * s0
		s3 = s2 * s0
		s4 = s3 + s0
		s5 = s4 + 5
		equals(s1, s5)
		out = 1 * 1
	`

	parser := circuit.NewParser(strings.NewReader(code))
	cir, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	privateInputs := []*big.Int{b3}
	b35 := big.NewInt(int64(35))
	publicSignals := []*big.Int{b35}

	w, err := cir.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	cir.GenerateR1CS()

	// TODO zxQAP is not used and is an old impl
	// TODO remove
	alphas, betas, gammas, _ := R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C,
	)
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.Equal(t, 8, len(alphas))
	assert.True(t, !bytes.Equal(alphas[1][1].Bytes(), big.NewInt(int64(0)).Bytes()))

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	assert.Equal(t, 7, len(ax))
	assert.Equal(t, 7, len(bx))
	assert.Equal(t, 7, len(cx))
	assert.Equal(t, 13, len(px))

	setup := &Groth16Setup{}
	err = setup.Init(cir, alphas, betas, gammas)
	assert.Nil(t, err)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	div, rem := Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, fields.ArrayOfBigZeros(6))

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)

	proof, err := setup.Generate(cir, w, px)
	assert.Nil(t, err)

	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	{
		r, err := setup.Verify(proof, publicSignalsVerif)
		assert.Nil(t, err)
		assert.True(t, r)
	}

	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	{
		r, err := setup.Verify(proof, wrongPublicSignalsVerif)
		assert.Nil(t, err)
		assert.False(t, r)
	}
}
