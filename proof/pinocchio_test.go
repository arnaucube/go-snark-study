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

func TestZkFromFlatCircuitCode(t *testing.T) {
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

	alphas, betas, gammas, zxQAP := R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C,
	)
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
	assert.Equal(t, rem, fields.ArrayOfBigZeros(6))

	// calculate trusted setup
	setup := &PinocchioSetup{}
	err = setup.Init(cir, alphas, betas, gammas)
	assert.Nil(t, err)

	// zx and setup.Pk.Z should be the same
	// currently not, the correct one is the calculation used inside GenerateTrustedSetup function
	// the calculation is repeated. TODO avoid repeating calculation
	assert.Equal(t, zxQAP, setup.Pk.Z)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	assert.Equal(t, hx, hxQAP)
	div, rem = Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, fields.ArrayOfBigZeros(6))

	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)

	proof, err := setup.Generate(cir, w, px)
	assert.Nil(t, err)

	b35Verif := big.NewInt(int64(35))
	publicSignalsVerif := []*big.Int{b35Verif}
	{
		r, err := setup.Verify(proof, publicSignalsVerif)
		assert.Nil(t, err)
		assert.True(t, r)
	}

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	{
		r, err := setup.Verify(proof, wrongPublicSignalsVerif)
		assert.Nil(t, err)
		assert.False(t, r)
	}
}

func TestZkMultiplication(t *testing.T) {
	code := `
	func main(private a, private b, public c):
		d = a * b
		equals(c, d)
		out = 1 * 1
	`

	parser := circuit.NewParser(strings.NewReader(code))
	cir, err := parser.Parse()
	assert.Nil(t, err)

	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	privateInputs := []*big.Int{b3, b4}
	b12 := big.NewInt(int64(12))
	publicSignals := []*big.Int{b12}

	w, err := cir.CalculateWitness(privateInputs, publicSignals)
	assert.Nil(t, err)

	cir.GenerateR1CS()

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl.
	// TODO remove
	alphas, betas, gammas, zxQAP := R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C,
	)
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
	assert.Equal(t, rem, fields.ArrayOfBigZeros(4))

	setup := &PinocchioSetup{}
	err = setup.Init(cir, alphas, betas, gammas)
	assert.Nil(t, err)

	// zx and setup.Pk.Z should be the same
	// currently not, the correct one is the calculation used inside GenerateTrustedSetup function
	// the calculation is repeated. TODO avoid repeating calculation
	assert.Equal(t, zxQAP, setup.Pk.Z)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)
	assert.Equal(t, 3, len(hx))
	assert.Equal(t, hx, hxQAP)

	div, rem = Utils.PF.Div(px, setup.Pk.Z)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, fields.ArrayOfBigZeros(4))

	assert.Equal(t, px, Utils.PF.Mul(hxQAP, zxQAP))
	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, setup.Pk.Z))

	// check length of polynomials H(x) and Z(x)
	assert.Equal(t, len(hx), len(px)-len(setup.Pk.Z)+1)
	assert.Equal(t, len(hxQAP), len(px)-len(zxQAP)+1)

	proof, err := setup.Generate(cir, w, px)
	assert.Nil(t, err)

	b12Verif := big.NewInt(int64(12))
	publicSignalsVerif := []*big.Int{b12Verif}
	{
		r, err := setup.Verify(proof, publicSignalsVerif)
		assert.Nil(t, err)
		assert.True(t, r)
	}

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(11))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	{
		r, err := setup.Verify(proof, wrongPublicSignalsVerif)
		assert.Nil(t, err)
		assert.False(t, r)
	}
}

func TestMinimalFlow(t *testing.T) {
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

	// R1CS to QAP
	// TODO zxQAP is not used and is an old impl, TODO remove
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

	// calculate trusted setup
	setup := &PinocchioSetup{}
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

	// check that with another public input the verification returns false
	bOtherWrongPublic := big.NewInt(int64(34))
	wrongPublicSignalsVerif := []*big.Int{bOtherWrongPublic}
	{
		r, err := setup.Verify(proof, wrongPublicSignalsVerif)
		assert.Nil(t, err)
		assert.False(t, r)
	}
}
