package bn128

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBN128(t *testing.T) {
	bn128, err := NewBn128()
	assert.Nil(t, err)

	big40 := big.NewInt(int64(40))
	big75 := big.NewInt(int64(75))

	g1a := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(big40))
	g2a := bn128.G2.MulScalar(bn128.G2.G, bn128.Fq1.Copy(big75))

	g1b := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(big75))
	g2b := bn128.G2.MulScalar(bn128.G2.G, bn128.Fq1.Copy(big40))

	pre1a := bn128.preComputeG1(g1a)
	pre2a := bn128.preComputeG2(g2a)
	assert.Nil(t, err)
	pre1b := bn128.preComputeG1(g1b)
	pre2b := bn128.preComputeG2(g2b)
	assert.Nil(t, err)

	r1 := bn128.MillerLoop(pre1a, pre2a)
	r2 := bn128.MillerLoop(pre1b, pre2b)

	rbe := bn128.Fq12.Mul(r1, bn128.Fq12.Inverse(r2))

	res := bn128.finalExponentiation(rbe)

	a := bn128.Fq12.Affine(res)
	b := bn128.Fq12.Affine(bn128.Fq12.One())

	assert.True(t, bn128.Fq12.Equal(a, b))
	assert.True(t, bn128.Fq12.Equal(res, bn128.Fq12.One()))
}

func TestBN128Pairing(t *testing.T) {
	bn128, err := NewBn128()
	assert.Nil(t, err)

	big25 := big.NewInt(int64(25))
	big30 := big.NewInt(int64(30))

	g1a := bn128.G1.MulScalar(bn128.G1.G, big25)
	g2a := bn128.G2.MulScalar(bn128.G2.G, big30)

	g1b := bn128.G1.MulScalar(bn128.G1.G, big30)
	g2b := bn128.G2.MulScalar(bn128.G2.G, big25)

	pA := bn128.Pairing(g1a, g2a)
	pB := bn128.Pairing(g1b, g2b)

	assert.True(t, bn128.Fq12.Equal(pA, pB))

	assert.True(t, !bytes.Equal(pA[0][0][0].Bytes(), big.NewInt(int64(0)).Bytes()))
	assert.True(t, !bytes.Equal(bn128.Fq12.Affine(pA)[0][0][0].Bytes(), big.NewInt(int64(0)).Bytes()))
	// assert.Equal(t, pA[0][0][0].String(), "73680848340331011700282047627232219336104151861349893575958589557226556635706")
	// assert.Equal(t, bn128.Fq12.Affine(pA)[0][0][0].String(), "8016119724813186033542830391460394070015218389456422587891475873290878009957")
}

func TestBN128Pairing2(t *testing.T) {
	// test idea from https://bplib.readthedocs.io/en/latest/ by George Danezis
	bn, err := NewBn128()
	assert.Nil(t, err)

	gt := bn.Pairing(bn.G1.G, bn.G2.G)

	gt6 := bn.Fq12.Exp(gt, big.NewInt(int64(6)))

	// e(g1, g2)^6 == e(g1, 6*g2)
	assert.True(t, bn.Fq12.Equal(gt6, bn.Pairing(bn.G1.G, bn.G2.MulScalar(bn.G2.G, big.NewInt(int64(6))))))

	// e(g1, g2)^6 == e(6* g1, g2)
	assert.True(t, bn.Fq12.Equal(gt6, bn.Pairing(bn.G1.MulScalar(bn.G1.G, big.NewInt(int64(6))), bn.G2.G)))
	// e(g1, g2)^6 == e(3*g1, 2*g2)
	assert.True(t, bn.Fq12.Equal(gt6, bn.Pairing(bn.G1.MulScalar(bn.G1.G, big.NewInt(int64(3))), bn.G2.MulScalar(bn.G2.G, big.NewInt(int64(2))))))
	// e(g1, g2)^6 == e(2*g1, 3*g2)
	assert.True(t, bn.Fq12.Equal(gt6, bn.Pairing(bn.G1.MulScalar(bn.G1.G, big.NewInt(int64(2))), bn.G2.MulScalar(bn.G2.G, big.NewInt(int64(3))))))

}
