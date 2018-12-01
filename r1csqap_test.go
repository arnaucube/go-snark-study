package sn

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranspose(t *testing.T) {
	b0 := big.NewFloat(float64(0))
	b1 := big.NewFloat(float64(1))
	bFive := big.NewFloat(float64(5))
	a := [][]*big.Float{
		[]*big.Float{b0, b1, b0, b0, b0, b0},
		[]*big.Float{b0, b0, b0, b1, b0, b0},
		[]*big.Float{b0, b1, b0, b0, b1, b0},
		[]*big.Float{bFive, b0, b0, b0, b0, b1},
	}
	aT := Transpose(a)
	assert.Equal(t, aT, [][]*big.Float{
		[]*big.Float{b0, b0, b0, bFive},
		[]*big.Float{b1, b0, b1, b0},
		[]*big.Float{b0, b0, b0, b0},
		[]*big.Float{b0, b1, b0, b0},
		[]*big.Float{b0, b0, b1, b0},
		[]*big.Float{b0, b0, b0, b1},
	})
}

func TestPol(t *testing.T) {
	b0 := big.NewFloat(float64(0))
	b1 := big.NewFloat(float64(1))
	// b1neg := big.NewFloat(float64(-1))
	// b2 := big.NewFloat(float64(2))
	b2neg := big.NewFloat(float64(-2))
	b3 := big.NewFloat(float64(3))
	b4 := big.NewFloat(float64(4))
	b5 := big.NewFloat(float64(5))
	b6 := big.NewFloat(float64(6))
	b16 := big.NewFloat(float64(16))

	a := []*big.Float{b1, b0, b5}
	b := []*big.Float{b3, b0, b1}

	// polynomial multiplication
	c := PolMul(a, b)
	assert.Equal(t, c, []*big.Float{b3, b0, b16, b0, b5})

	// polynomial addition
	c = PolAdd(a, b)
	assert.Equal(t, c, []*big.Float{b4, b0, b6})

	// polynomial substraction
	c = PolSub(a, b)
	assert.Equal(t, c, []*big.Float{b2neg, b0, b4})

	// FloatPow
	p := FloatPow(big.NewFloat(float64(5)), 3)
	assert.Equal(t, p, big.NewFloat(float64(125)))
	p = FloatPow(big.NewFloat(float64(5)), 0)
	assert.Equal(t, p, big.NewFloat(float64(1)))

	// NewPolZeroAt
	r := NewPolZeroAt(3, 4, b4)
	assert.Equal(t, PolEval(r, big.NewFloat(3)), b4)
	r = NewPolZeroAt(2, 4, b3)
	assert.Equal(t, PolEval(r, big.NewFloat(2)), b3)
}

func TestLagrangeInterpolation(t *testing.T) {
	b0 := big.NewFloat(float64(0))
	b5 := big.NewFloat(float64(5))
	a := []*big.Float{b0, b0, b0, b5}
	alpha := LagrangeInterpolation(a)

	assert.Equal(t, PolEval(alpha, big.NewFloat(4)), b5)
	aux, _ := PolEval(alpha, big.NewFloat(3)).Int64()
	assert.Equal(t, aux, int64(0))

}

func TestR1CSToQAP(t *testing.T) {
	b0 := big.NewFloat(float64(0))
	b1 := big.NewFloat(float64(1))
	b5 := big.NewFloat(float64(5))
	a := [][]*big.Float{
		[]*big.Float{b0, b1, b0, b0, b0, b0},
		[]*big.Float{b0, b0, b0, b1, b0, b0},
		[]*big.Float{b0, b1, b0, b0, b1, b0},
		[]*big.Float{b5, b0, b0, b0, b0, b1},
	}
	b := [][]*big.Float{
		[]*big.Float{b0, b1, b0, b0, b0, b0},
		[]*big.Float{b0, b1, b0, b0, b0, b0},
		[]*big.Float{b1, b0, b0, b0, b0, b0},
		[]*big.Float{b1, b0, b0, b0, b0, b0},
	}
	c := [][]*big.Float{
		[]*big.Float{b0, b0, b0, b1, b0, b0},
		[]*big.Float{b0, b0, b0, b0, b1, b0},
		[]*big.Float{b0, b0, b0, b0, b0, b1},
		[]*big.Float{b0, b0, b1, b0, b0, b0},
	}
	alpha, beta, gamma, z := R1CSToQAP(a, b, c)
	fmt.Println(alpha)
	fmt.Println(beta)
	fmt.Println(gamma)
	fmt.Println(z)

}
