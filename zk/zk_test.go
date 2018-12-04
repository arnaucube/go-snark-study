package zk

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
)

func TestZk(t *testing.T) {
	bn, err := bn128.NewBn128()
	assert.Nil(t, err)

	// new Finite Field
	f := fields.NewFq(bn.R)

	// new Polynomial Field
	pf := r1csqap.NewPolynomialField(f)

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

	w := []*big.Int{b1, b3, b35, b9, b27, b30}
	ax, bx, cx, px := pf.CombinePolynomials(w, alphas, betas, gammas)

	hx := pf.DivisorPolinomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, pf.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := pf.Sub(pf.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := pf.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	// calculate trusted setup
	gt1, gt2, err := GenerateTrustedSetup(bn, len(ax))
	assert.Nil(t, err)
	fmt.Println("trusted setup:")
	fmt.Println(gt1)
	fmt.Println(gt2)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	piA, piB, piC, piH, piZ := GenerateProofs(bn, gt1, gt2, ax, bx, cx, hx, zx)
	fmt.Println("proofs:")
	fmt.Println(piA)
	fmt.Println(piB)
	fmt.Println(piC)
	fmt.Println(piH)
	fmt.Println(piZ)

	// pairing
	fmt.Println("pairing")
	pairingAB, err := bn.Pairing(piA, piB)
	assert.Nil(t, err)
	pairingCg2, err := bn.Pairing(piC, bn.G2.G)
	assert.Nil(t, err)
	pairingLeft := bn.Fq12.Div(pairingAB, pairingCg2)
	pairingHg2Z, err := bn.Pairing(piH, piZ)
	assert.Nil(t, err)

	fmt.Println(bn.Fq12.Affine(pairingLeft))
	fmt.Println(bn.Fq12.Affine(pairingHg2Z))

	assert.True(t, bn.Fq12.Equal(pairingLeft, pairingHg2Z))
}
