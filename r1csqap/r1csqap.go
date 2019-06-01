package r1csqap

import (
	"bytes"
	"math/big"

	"github.com/arnaucube/go-snark/fields"
)

// Transpose transposes the *big.Int matrix
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

// ArrayOfBigZeros creates a *big.Int array with n elements to zero
func ArrayOfBigZeros(num int) []*big.Int {
	bigZero := big.NewInt(int64(0))
	var r = make([]*big.Int, num, num)
	for i := 0; i < num; i++ {
		r[i] = bigZero
	}
	return r
}
func BigArraysEqual(a, b []*big.Int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !bytes.Equal(a[i].Bytes(), b[i].Bytes()) {
			return false
		}
	}
	return true
}

// PolynomialField is the Polynomial over a Finite Field where the polynomial operations are performed
type PolynomialField struct {
	F fields.Fq
}

// NewPolynomialField creates a new PolynomialField with the given FiniteField
func NewPolynomialField(f fields.Fq) PolynomialField {
	return PolynomialField{
		f,
	}
}

// Mul multiplies two polinomials over the Finite Field
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

// Div divides two polinomials over the Finite Field, returning the result and the remainder
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

// Add adds two polinomials over the Finite Field
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

// Sub subtracts two polinomials over the Finite Field
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

// Eval evaluates the polinomial over the Finite Field at the given value x
func (pf PolynomialField) Eval(v []*big.Int, x *big.Int) *big.Int {
	r := big.NewInt(int64(0))
	for i := 0; i < len(v); i++ {
		xi := pf.F.Exp(x, big.NewInt(int64(i)))
		elem := pf.F.Mul(v[i], xi)
		r = pf.F.Add(r, elem)
	}
	return r
}

// NewPolZeroAt generates a new polynomial that has value zero at the given value
func (pf PolynomialField) NewPolZeroAt(pointPos, totalPoints int, height *big.Int) []*big.Int {
	//todo note that this will blow up. big may be necessary
	fac := 1
	//(xj-x0)(xj-x1)..(xj-x_j-1)(xj-x_j+1)..(x_j-x_k)
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
			//is b1 necessary?
			b1 := big.NewInt(int64(1))
			r = pf.Mul(r, []*big.Int{ineg, b1})
		}
	}
	return r
}

// LagrangeInterpolation performs the Lagrange Interpolation / Lagrange Polynomials operation
func (pf PolynomialField) LagrangeInterpolation(v []*big.Int) []*big.Int {
	// https://en.wikipedia.org/wiki/Lagrange_polynomial
	var r []*big.Int
	for i := 0; i < len(v); i++ {
		r = pf.Add(r, pf.NewPolZeroAt(i+1, len(v), v[i]))
		//r = pf.Mul(v[i], pf.NewPolZeroAt(i+1, len(v), v[i]))
	}
	//
	return r
}

// R1CSToQAP converts the R1CS values to the QAP values
//it uses Lagrange interpolation to to fit a polynomial through each slice. The x coordinate
//is simply a linear increment starting at 1
//within this process, the polynomial is evaluated at position 0
//so an alpha/beta/gamma value is the polynomial evaluated at 0
// the domain polynomial therefor is (-1+x)(-2+x)...(-n+x)
func (pf PolynomialField) R1CSToQAP(a, b, c [][]*big.Int) (alphas [][]*big.Int, betas [][]*big.Int, gammas [][]*big.Int, domain []*big.Int) {
	aT := Transpose(a)
	bT := Transpose(b)
	cT := Transpose(c)

	for i := 0; i < len(aT); i++ {
		alphas = append(alphas, pf.LagrangeInterpolation(aT[i]))
	}

	for i := 0; i < len(bT); i++ {
		betas = append(betas, pf.LagrangeInterpolation(bT[i]))
	}

	for i := 0; i < len(cT); i++ {
		gammas = append(gammas, pf.LagrangeInterpolation(cT[i]))
	}
	//it used to range till len(alphas)-1, but this was wrong.
	z := []*big.Int{big.NewInt(int64(1))}
	for i := 1; i < len(a); i++ {
		z = pf.Mul(
			z,
			[]*big.Int{
				pf.F.Neg(
					big.NewInt(int64(i))),
				big.NewInt(int64(1)),
			})
	}
	return alphas, betas, gammas, z
}

// CombinePolynomials combine the given polynomials arrays into one, also returns the P(x)
func (pf PolynomialField) CombinePolynomials(r []*big.Int, ap, bp, cp [][]*big.Int) ([]*big.Int, []*big.Int, []*big.Int, []*big.Int) {
	var ax []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, ap[i])
		ax = pf.Add(ax, m)
	}
	var bx []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, bp[i])
		bx = pf.Add(bx, m)
	}
	var cx []*big.Int
	for i := 0; i < len(r); i++ {
		m := pf.Mul([]*big.Int{r[i]}, cp[i])
		cx = pf.Add(cx, m)
	}

	px := pf.Sub(pf.Mul(ax, bx), cx)
	return ax, bx, cx, px
}

// DivisorPolynomial returns the divisor polynomial given two polynomials
func (pf PolynomialField) DivisorPolynomial(px, z []*big.Int) []*big.Int {
	quo, _ := pf.Div(px, z)
	return quo
}
