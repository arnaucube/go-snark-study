package r1csqap

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/arnaucube/go-snark/fields"
	"github.com/stretchr/testify/assert"
)

func TestTranspose(t *testing.T) {
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	bFive := big.NewInt(int64(5))
	a := [][]*big.Int{
		[]*big.Int{b0, b1, b0, b0, b0, b0},
		[]*big.Int{b0, b0, b0, b1, b0, b0},
		[]*big.Int{b0, b1, b0, b0, b1, b0},
		[]*big.Int{bFive, b0, b0, b0, b0, b1},
	}
	aT := Transpose(a)
	assert.Equal(t, aT, [][]*big.Int{
		[]*big.Int{b0, b0, b0, bFive},
		[]*big.Int{b1, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b0},
		[]*big.Int{b0, b1, b0, b0},
		[]*big.Int{b0, b0, b1, b0},
		[]*big.Int{b0, b0, b0, b1},
	})
}

func neg(a *big.Int) *big.Int {
	return new(big.Int).Neg(a)
}

func TestPol(t *testing.T) {
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b2 := big.NewInt(int64(2))
	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	b5 := big.NewInt(int64(5))
	b6 := big.NewInt(int64(6))
	b16 := big.NewInt(int64(16))

	a := []*big.Int{b1, b0, b5}
	b := []*big.Int{b3, b0, b1}

	// new Finite Field
	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	assert.True(nil, ok)
	f := fields.NewFq(r)

	// new Polynomial Field
	pf := NewPolynomialField(f)

	// polynomial multiplication
	o := pf.Mul(a, b)
	assert.Equal(t, o, []*big.Int{b3, b0, b16, b0, b5})

	// polynomial division
	quo, rem := pf.Div(a, b)
	assert.Equal(t, quo[0].Int64(), int64(5))
	assert.Equal(t, new(big.Int).Sub(rem[0], r).Int64(), int64(-14)) // check the rem result without modulo

	c := []*big.Int{neg(b4), b0, neg(b2), b1}
	d := []*big.Int{neg(b3), b1}
	quo2, rem2 := pf.Div(c, d)
	assert.Equal(t, quo2, []*big.Int{b3, b1, b1})
	assert.Equal(t, rem2[0].Int64(), int64(5))

	// polynomial addition
	o = pf.Add(a, b)
	assert.Equal(t, o, []*big.Int{b4, b0, b6})

	// polynomial subtraction
	o1 := pf.Sub(a, b)
	o2 := pf.Sub(b, a)
	o = pf.Add(o1, o2)
	assert.True(t, bytes.Equal(b0.Bytes(), o[0].Bytes()))
	assert.True(t, bytes.Equal(b0.Bytes(), o[1].Bytes()))
	assert.True(t, bytes.Equal(b0.Bytes(), o[2].Bytes()))

	c = []*big.Int{b5, b6, b1}
	d = []*big.Int{b1, b3}
	o = pf.Sub(c, d)
	assert.Equal(t, o, []*big.Int{b4, b3, b1})

	// NewPolZeroAt
	o = pf.NewPolZeroAt(3, 4, b4)
	assert.Equal(t, pf.Eval(o, big.NewInt(3)), b4)
	o = pf.NewPolZeroAt(2, 4, b3)
	assert.Equal(t, pf.Eval(o, big.NewInt(2)), b3)
}

func TestLagrangeInterpolation(t *testing.T) {
	// new Finite Field
	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	assert.True(nil, ok)
	f := fields.NewFq(r)
	// new Polynomial Field
	pf := NewPolynomialField(f)

	b0 := big.NewInt(int64(0))
	b5 := big.NewInt(int64(5))
	a := []*big.Int{b0, b0, b0, b5}
	alpha := pf.LagrangeInterpolation(a)

	assert.Equal(t, pf.Eval(alpha, big.NewInt(int64(4))), b5)
	aux := pf.Eval(alpha, big.NewInt(int64(3))).Int64()
	assert.Equal(t, aux, int64(0))

}
