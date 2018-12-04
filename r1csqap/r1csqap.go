package r1csqap

import (
	"math/big"

	"github.com/arnaucube/go-snark/fields"
)

func Transpose(matrix [][]*big.Int) [][]*big.Int {
	var r [][]*big.Int
	for i := 0; i < len(matrix[0]); i++ {
		var row []*big.Int
		for j := 0; j < len(matrix); j++ {
			row = append(row, matrix[j][i])
		}
		r = append(r, row)
	}
	return r
}

func ArrayOfBigZeros(num int) []*big.Int {
	bigZero := big.NewInt(int64(0))
	var r []*big.Int
	for i := 0; i < num; i++ {
		r = append(r, bigZero)
	}
	return r
}

type PolynomialField struct {
	F fields.Fq
}

func NewPolynomialField(f fields.Fq) PolynomialField {
	return PolynomialField{
		f,
	}
}
func (pf PolynomialField) Mul(a, b []*big.Int) []*big.Int {
	r := ArrayOfBigZeros(len(a) + len(b) - 1)
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			r[i+j] = pf.F.Add(
				r[i+j],
				pf.F.Mul(a[i], b[j]))
		}
	}
	return r
}
func (pf PolynomialField) Div(a, b []*big.Int) ([]*big.Int, []*big.Int) {
	// https://en.wikipedia.org/wiki/Division_algorithm
	r := ArrayOfBigZeros(len(a) - len(b) + 1)
	rem := a
	for len(rem) >= len(b) {
		l := pf.F.Div(rem[len(rem)-1], b[len(b)-1])
		pos := len(rem) - len(b)
		r[pos] = l
		aux := ArrayOfBigZeros(pos)
		aux1 := append(aux, l)
		aux2 := pf.Sub(rem, pf.Mul(b, aux1))
		rem = aux2[:len(aux2)-1]
	}
	return r, rem
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (pf PolynomialField) Add(a, b []*big.Int) []*big.Int {
	r := ArrayOfBigZeros(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = pf.F.Add(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		r[i] = pf.F.Add(r[i], b[i])
	}
	return r
}

func (pf PolynomialField) Sub(a, b []*big.Int) []*big.Int {
	r := ArrayOfBigZeros(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = pf.F.Add(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		r[i] = pf.F.Sub(r[i], b[i])
	}
	return r
}

func (pf PolynomialField) Eval(v []*big.Int, x *big.Int) *big.Int {
	r := big.NewInt(int64(0))
	for i := 0; i < len(v); i++ {
		// xi := FloatPow(x, i)
		xi := pf.F.Exp(x, big.NewInt(int64(i)))
		elem := pf.F.Mul(v[i], xi)
		r = pf.F.Add(r, elem)
	}
	return r
}

func (pf PolynomialField) NewPolZeroAt(pointPos, totalPoints int, height *big.Int) []*big.Int {
	fac := 1
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			fac = fac * (pointPos - i)
		}
	}
	facBig := big.NewInt(int64(fac))
	hf := pf.F.Div(height, facBig)
	r := []*big.Int{hf}
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			ineg := big.NewInt(int64(-i))
			b1 := big.NewInt(int64(1))
			r = pf.Mul(r, []*big.Int{ineg, b1})
		}
	}
	return r
}

func (pf PolynomialField) LagrangeInterpolation(v []*big.Int) []*big.Int {
	// https://en.wikipedia.org/wiki/Lagrange_polynomial
	var r []*big.Int
	for i := 0; i < len(v); i++ {
		r = pf.Add(r, pf.NewPolZeroAt(i+1, len(v), v[i]))
	}
	//
	return r
}

func (pf PolynomialField) R1CSToQAP(a, b, c [][]*big.Int) ([][]*big.Int, [][]*big.Int, [][]*big.Int, []*big.Int) {
	aT := Transpose(a)
	bT := Transpose(b)
	cT := Transpose(c)
	var alphas [][]*big.Int
	for i := 0; i < len(aT); i++ {
		alphas = append(alphas, pf.LagrangeInterpolation(aT[i]))
	}
	var betas [][]*big.Int
	for i := 0; i < len(bT); i++ {
		betas = append(betas, pf.LagrangeInterpolation(bT[i]))
	}
	var gammas [][]*big.Int
	for i := 0; i < len(cT); i++ {
		gammas = append(gammas, pf.LagrangeInterpolation(cT[i]))
	}
	z := []*big.Int{big.NewInt(int64(1))}
	for i := 1; i < len(aT[0])+1; i++ {
		ineg := big.NewInt(int64(-i))
		b1 := big.NewInt(int64(1))
		z = pf.Mul(z, []*big.Int{ineg, b1})
	}
	return alphas, betas, gammas, z
}

func (pf PolynomialField) CombinePolynomials(r []*big.Int, ap, bp, cp [][]*big.Int) ([]*big.Int, []*big.Int, []*big.Int, []*big.Int) {
	var alpha []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, ap[i])
		alpha = pf.Add(alpha, m)
	}
	var beta []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, bp[i])
		beta = pf.Add(beta, m)
	}
	var gamma []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, cp[i])
		gamma = pf.Add(gamma, m)
	}

	px := pf.Sub(pf.Mul(alpha, beta), gamma)
	return alpha, beta, gamma, px
}

func (pf PolynomialField) DivisorPolinomial(px, z []*big.Int) []*big.Int {
	quo, _ := pf.Div(px, z)
	return quo
}
