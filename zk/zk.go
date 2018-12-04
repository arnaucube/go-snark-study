package zk

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/arnaucube/go-snark/bn128"
)

const bits = 512

func GenerateTrustedSetup(bn bn128.Bn128, pollength int) ([][3]*big.Int, [][3][2]*big.Int, error) {
	// generate random t value
	t, err := rand.Prime(rand.Reader, bits)
	if err != nil {
		return [][3]*big.Int{}, [][3][2]*big.Int{}, err
	}
	fmt.Print("trusted t: ")
	fmt.Println(t)

	// encrypt t values with curve generators
	var gt1 [][3]*big.Int
	var gt2 [][3][2]*big.Int
	for i := 0; i < pollength; i++ {
		tPow := bn.Fq1.Exp(t, big.NewInt(int64(i)))
		tEncr1 := bn.G1.MulScalar(bn.G1.G, tPow)
		gt1 = append(gt1, tEncr1)
		tEncr2 := bn.G2.MulScalar(bn.G2.G, tPow)
		gt2 = append(gt2, tEncr2)
	}
	// gt1: g1, g1*t, g1*t^2, g1*t^3, ...
	// gt2: g2, g2*t, g2*t^2, ...
	return gt1, gt2, nil
}
func GenerateProofs(bn bn128.Bn128, gt1 [][3]*big.Int, gt2 [][3][2]*big.Int, ax, bx, cx, hx, zx []*big.Int) ([3]*big.Int, [3][2]*big.Int, [3]*big.Int, [3]*big.Int, [3][2]*big.Int) {

	// multiply g1*A(x), g2*B(x), g1*C(x), g1*H(x)

	// g1*A(x)
	g1At := [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(ax); i++ {
		m := bn.G1.MulScalar(gt1[i], ax[i])
		g1At = bn.G1.Add(g1At, m)
	}
	g2Bt := bn.Fq6.Zero()
	for i := 0; i < len(bx); i++ {
		m := bn.G2.MulScalar(gt2[i], bx[i])
		g2Bt = bn.G2.Add(g2Bt, m)
	}

	g1Ct := [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(cx); i++ {
		m := bn.G1.MulScalar(gt1[i], cx[i])
		g1Ct = bn.G1.Add(g1Ct, m)
	}
	g1Ht := [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(hx); i++ {
		m := bn.G1.MulScalar(gt1[i], hx[i])
		g1Ht = bn.G1.Add(g1Ht, m)
	}
	g2Zt := bn.Fq6.Zero()
	for i := 0; i < len(bx); i++ {
		m := bn.G2.MulScalar(gt2[i], zx[i])
		g2Zt = bn.G2.Add(g2Zt, m)
	}

	return g1At, g2Bt, g1Ct, g1Ht, g2Zt
}
