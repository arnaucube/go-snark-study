package snark

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
)

func TestZk(t *testing.T) {
	bn, err := bn128.NewBn128()
	assert.Nil(t, err)

	// new Finite Field
	fqR := fields.NewFq(bn.R)

	// new Polynomial Field
	pf := r1csqap.NewPolynomialField(fqR)

	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b3 := big.NewInt(int64(3))
	b5 := big.NewInt(int64(5))
	b9 := big.NewInt(int64(9))
	b27 := big.NewInt(int64(27))
	b30 := big.NewInt(int64(30))
	b35 := big.NewInt(int64(35))
	a := [][]*big.Int{
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b1, b0, b0, b1, b0},
		[]*big.Int{b5, b0, b0, b0, b0, b1},
	}
	b := [][]*big.Int{
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
		[]*big.Int{b1, b0, b0, b0, b0, b0},
	}
	c := [][]*big.Int{
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b0, b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0, b0, b1},
		[]*big.Int{b0, b0, b1, b0, b0, b0},
	}
	alphas, betas, gammas, zx := pf.R1CSToQAP(a, b, c)

	// wittness = 1, 3, 35, 9, 27, 30
	w := []*big.Int{b1, b3, b35, b9, b27, b30}
	circuit := circuitcompiler.Circuit{
		NVars:    6,
		NPublic:  0,
		NSignals: len(w),
	}
	ax, bx, cx, px := pf.CombinePolynomials(w, alphas, betas, gammas)

	hx := pf.DivisorPolinomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, pf.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := pf.Sub(pf.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := pf.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	div, rem := pf.Div(px, zx)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(bn, fqR, pf, len(w), circuit, alphas, betas, gammas, zx)
	assert.Nil(t, err)
	fmt.Println("t", setup.Toxic.T)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(bn, fqR, circuit, setup, hx, w)
	assert.Nil(t, err)

	assert.True(t, VerifyProof(bn, circuit, setup, proof))
}
