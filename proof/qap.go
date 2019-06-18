package proof

import (
	"math/big"

	"github.com/arnaucube/go-snark/fields"
)

// R1CSToQAP converts the R1CS values to the QAP values
func R1CSToQAP(a, b, c [][]*big.Int) ([][]*big.Int, [][]*big.Int, [][]*big.Int, []*big.Int) {
	aT := fields.Transpose(a)
	bT := fields.Transpose(b)
	cT := fields.Transpose(c)
	var alphas [][]*big.Int
	for i := 0; i < len(aT); i++ {
		alphas = append(alphas, Utils.PF.LagrangeInterpolation(aT[i]))
	}
	var betas [][]*big.Int
	for i := 0; i < len(bT); i++ {
		betas = append(betas, Utils.PF.LagrangeInterpolation(bT[i]))
	}
	var gammas [][]*big.Int
	for i := 0; i < len(cT); i++ {
		gammas = append(gammas, Utils.PF.LagrangeInterpolation(cT[i]))
	}
	z := []*big.Int{big.NewInt(int64(1))}
	for i := 1; i < len(alphas)-1; i++ {
		z = Utils.PF.Mul(
			z,
			[]*big.Int{
				Utils.PF.F.Neg(
					big.NewInt(int64(i))),
				big.NewInt(int64(1)),
			})
	}
	return alphas, betas, gammas, z
}
