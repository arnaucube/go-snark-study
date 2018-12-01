package bn128

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func iToBig(a int) *big.Int {
	return big.NewInt(int64(a))
}

func iiToBig(a, b int) [2]*big.Int {
	return [2]*big.Int{iToBig(a), iToBig(b)}
}

func iiiToBig(a, b int) [2]*big.Int {
	return [2]*big.Int{iToBig(a), iToBig(b)}
}

func TestFq1(t *testing.T) {
	fq1 := NewFq(iToBig(7))

	res := fq1.Add(iToBig(4), iToBig(4))
	assert.Equal(t, iToBig(1), fq1.Affine(res))

	res = fq1.Double(iToBig(5))
	assert.Equal(t, iToBig(3), fq1.Affine(res))

	res = fq1.Sub(iToBig(5), iToBig(7))
	assert.Equal(t, iToBig(5), fq1.Affine(res))

	res = fq1.Neg(iToBig(5))
	assert.Equal(t, iToBig(2), fq1.Affine(res))

	res = fq1.Mul(iToBig(5), iToBig(11))
	assert.Equal(t, iToBig(6), fq1.Affine(res))

	res = fq1.Inverse(iToBig(4))
	assert.Equal(t, iToBig(2), res)

	res = fq1.Square(iToBig(5))
	assert.Equal(t, iToBig(4), res)
}

func TestFq2(t *testing.T) {
	fq1 := NewFq(iToBig(7))
	nonResidueFq2str := "-1" // i/j
	nonResidueFq2, ok := new(big.Int).SetString(nonResidueFq2str, 10)
	assert.True(t, ok)
	assert.Equal(t, nonResidueFq2.String(), nonResidueFq2str)

	fq2 := Fq2{fq1, nonResidueFq2}

	res := fq2.Add(iiToBig(4, 4), iiToBig(3, 4))
	assert.Equal(t, iiToBig(0, 1), fq2.Affine(res))

	res = fq2.Double(iiToBig(5, 3))
	assert.Equal(t, iiToBig(3, 6), fq2.Affine(res))

	res = fq2.Sub(iiToBig(5, 3), iiToBig(7, 2))
	assert.Equal(t, iiToBig(5, 1), fq2.Affine(res))

	res = fq2.Neg(iiToBig(4, 4))
	assert.Equal(t, iiToBig(3, 3), fq2.Affine(res))

	res = fq2.Mul(iiToBig(4, 4), iiToBig(3, 4))
	assert.Equal(t, iiToBig(3, 0), fq2.Affine(res))

	res = fq2.Inverse(iiToBig(4, 4))
	assert.Equal(t, iiToBig(1, 6), fq2.Affine(res))

	res = fq2.Square(iiToBig(4, 4))
	assert.Equal(t, iiToBig(0, 4), fq2.Affine(res))
	res2 := fq2.Mul(iiToBig(4, 4), iiToBig(4, 4))
	assert.Equal(t, fq2.Affine(res), fq2.Affine(res2))
	assert.True(t, fq2.Equal(res, res2))

	res = fq2.Square(iiToBig(3, 5))
	assert.Equal(t, iiToBig(5, 2), fq2.Affine(res))
	res2 = fq2.Mul(iiToBig(3, 5), iiToBig(3, 5))
	assert.Equal(t, fq2.Affine(res), fq2.Affine(res2))
}

func TestFq6(t *testing.T) {
	bn128, err := NewBn128()
	assert.Nil(t, err)

	a := [3][2]*big.Int{
		iiToBig(1, 2),
		iiToBig(3, 4),
		iiToBig(5, 6)}
	b := [3][2]*big.Int{
		iiToBig(12, 11),
		iiToBig(10, 9),
		iiToBig(8, 7)}

	mulRes := bn128.Fq6.Mul(a, b)
	divRes := bn128.Fq6.Div(mulRes, b)
	assert.Equal(t, bn128.Fq6.Affine(a), bn128.Fq6.Affine(divRes))
}

func TestFq12(t *testing.T) {
	q, ok := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10) // i
	assert.True(t, ok)
	fq1 := NewFq(q)
	nonResidueFq2, ok := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208582", 10) // i
	assert.True(t, ok)
	nonResidueFq6 := iiToBig(9, 1)

	fq2 := Fq2{fq1, nonResidueFq2}
	fq6 := Fq6{fq2, nonResidueFq6}
	fq12 := Fq12{fq6, fq2, nonResidueFq6}

	a := [2][3][2]*big.Int{
		{
			iiToBig(1, 2),
			iiToBig(3, 4),
			iiToBig(5, 6),
		},
		{
			iiToBig(7, 8),
			iiToBig(9, 10),
			iiToBig(11, 12),
		},
	}
	b := [2][3][2]*big.Int{
		{
			iiToBig(12, 11),
			iiToBig(10, 9),
			iiToBig(8, 7),
		},
		{
			iiToBig(6, 5),
			iiToBig(4, 3),
			iiToBig(2, 1),
		},
	}

	res := fq12.Add(a, b)
	assert.Equal(t,
		[2][3][2]*big.Int{
			{
				iiToBig(13, 13),
				iiToBig(13, 13),
				iiToBig(13, 13),
			},
			{
				iiToBig(13, 13),
				iiToBig(13, 13),
				iiToBig(13, 13),
			},
		},
		res)

	mulRes := fq12.Mul(a, b)
	divRes := fq12.Div(mulRes, b)
	assert.Equal(t, fq12.Affine(a), fq12.Affine(divRes))
}
