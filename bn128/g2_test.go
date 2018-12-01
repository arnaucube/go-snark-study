package bn128

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestG2(t *testing.T) {
	bn128, err := NewBn128()
	assert.Nil(t, err)

	r1 := big.NewInt(int64(33))
	r2 := big.NewInt(int64(44))

	gr1 := bn128.G2.Affine(bn128.G2.MulScalar(bn128.G2.G, r1))
	gr2 := bn128.G2.Affine(bn128.G2.MulScalar(bn128.G2.G, r2))

	grsum1 := bn128.G2.Affine(bn128.G2.Add(gr1, gr2))
	r1r2 := bn128.Fq1.Affine(bn128.Fq1.Add(r1, r2))
	grsum2 := bn128.G2.Affine(bn128.G2.MulScalar(bn128.G2.G, r1r2))
	assert.True(t, bn128.G2.Equal(grsum1, grsum2))
}
