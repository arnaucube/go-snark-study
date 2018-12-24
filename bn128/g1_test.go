package bn128

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestG1(t *testing.T) {
	bn128, err := NewBn128()
	assert.Nil(t, err)

	r1 := big.NewInt(int64(33))
	r2 := big.NewInt(int64(44))

	gr1 := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(r1))
	gr2 := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(r2))

	grsum1 := bn128.G1.Add(gr1, gr2)               // g*33 + g*44
	r1r2 := bn128.Fq1.Add(r1, r2)                  // 33 + 44
	grsum2 := bn128.G1.MulScalar(bn128.G1.G, r1r2) // g * (33+44)

	assert.True(t, bn128.G1.Equal(grsum1, grsum2))
	a := bn128.G1.Affine(grsum1)
	b := bn128.G1.Affine(grsum2)
	assert.Equal(t, a, b)
	assert.Equal(t, "2f978c0ab89ebaa576866706b14787f360c4d6c3869efe5a72f7c3651a72ff00", hex.EncodeToString(a[0].Bytes()))
	assert.Equal(t, "12e4ba7f0edca8b4fa668fe153aebd908d322dc26ad964d4cd314795844b62b2", hex.EncodeToString(a[1].Bytes()))
}
