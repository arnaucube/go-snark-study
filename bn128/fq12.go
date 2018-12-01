package bn128

import (
	"bytes"
	"math/big"
)

// Fq12 uses the same algorithms than Fq2, but with [2][3][2]*big.Int data structure

// Fq12 is Field 12
type Fq12 struct {
	F          Fq6
	Fq2        Fq2
	NonResidue [2]*big.Int
}

// NewFq12 generates a new Fq12
func NewFq12(f Fq6, fq2 Fq2, nonResidue [2]*big.Int) Fq12 {
	fq12 := Fq12{
		f,
		fq2,
		nonResidue,
	}
	return fq12
}

// Zero returns a Zero value on the Fq12
func (fq12 Fq12) Zero() [2][3][2]*big.Int {
	return [2][3][2]*big.Int{fq12.F.Zero(), fq12.F.Zero()}
}

// One returns a One value on the Fq12
func (fq12 Fq12) One() [2][3][2]*big.Int {
	return [2][3][2]*big.Int{fq12.F.One(), fq12.F.Zero()}
}

func (fq12 Fq12) mulByNonResidue(a [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		fq12.Fq2.Mul(fq12.NonResidue, a[2]),
		a[0],
		a[1],
	}
}

// Add performs an addition on the Fq12
func (fq12 Fq12) Add(a, b [2][3][2]*big.Int) [2][3][2]*big.Int {
	return [2][3][2]*big.Int{
		fq12.F.Add(a[0], b[0]),
		fq12.F.Add(a[1], b[1]),
	}
}

// Double performs a doubling on the Fq12
func (fq12 Fq12) Double(a [2][3][2]*big.Int) [2][3][2]*big.Int {
	return fq12.Add(a, a)
}

// Sub performs a subtraction on the Fq12
func (fq12 Fq12) Sub(a, b [2][3][2]*big.Int) [2][3][2]*big.Int {
	return [2][3][2]*big.Int{
		fq12.F.Sub(a[0], b[0]),
		fq12.F.Sub(a[1], b[1]),
	}
}

// Neg performs a negation on the Fq12
func (fq12 Fq12) Neg(a [2][3][2]*big.Int) [2][3][2]*big.Int {
	return fq12.Sub(fq12.Zero(), a)
}

// Mul performs a multiplication on the Fq12
func (fq12 Fq12) Mul(a, b [2][3][2]*big.Int) [2][3][2]*big.Int {
	// Multiplication and Squaring on Pairing-Friendly .pdf; Section 3 (Karatsuba)
	v0 := fq12.F.Mul(a[0], b[0])
	v1 := fq12.F.Mul(a[1], b[1])
	return [2][3][2]*big.Int{
		fq12.F.Add(v0, fq12.mulByNonResidue(v1)),
		fq12.F.Sub(
			fq12.F.Mul(
				fq12.F.Add(a[0], a[1]),
				fq12.F.Add(b[0], b[1])),
			fq12.F.Add(v0, v1)),
	}
}

func (fq12 Fq12) MulScalar(base [2][3][2]*big.Int, e *big.Int) [2][3][2]*big.Int {
	// for more possible implementations see g2.go file, at the function g2.MulScalar()

	res := fq12.Zero()
	rem := e
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		// if rem % 2 == 1
		if bytes.Equal(new(big.Int).Rem(rem, big.NewInt(int64(2))).Bytes(), big.NewInt(int64(1)).Bytes()) {
			res = fq12.Add(res, exp)
		}
		exp = fq12.Double(exp)
		rem = rem.Rsh(rem, 1) // rem = rem >> 1
	}
	return res
}

// Inverse returns the inverse on the Fq12
func (fq12 Fq12) Inverse(a [2][3][2]*big.Int) [2][3][2]*big.Int {
	t0 := fq12.F.Square(a[0])
	t1 := fq12.F.Square(a[1])
	t2 := fq12.F.Sub(t0, fq12.mulByNonResidue(t1))
	t3 := fq12.F.Inverse(t2)
	return [2][3][2]*big.Int{
		fq12.F.Mul(a[0], t3),
		fq12.F.Neg(fq12.F.Mul(a[1], t3)),
	}
}

// Div performs a division on the Fq12
func (fq12 Fq12) Div(a, b [2][3][2]*big.Int) [2][3][2]*big.Int {
	return fq12.Mul(a, fq12.Inverse(b))
}

// Square performs a square operation on the Fq12
func (fq12 Fq12) Square(a [2][3][2]*big.Int) [2][3][2]*big.Int {
	ab := fq12.F.Mul(a[0], a[1])

	return [2][3][2]*big.Int{
		fq12.F.Sub(
			fq12.F.Mul(
				fq12.F.Add(a[0], a[1]),
				fq12.F.Add(
					a[0],
					fq12.mulByNonResidue(a[1]))),
			fq12.F.Add(
				ab,
				fq12.mulByNonResidue(ab))),
		fq12.F.Add(ab, ab),
	}
}

func (fq12 Fq12) Exp(base [2][3][2]*big.Int, e *big.Int) [2][3][2]*big.Int {
	res := fq12.One()
	rem := fq12.Fq2.F.Copy(e)
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		if BigIsOdd(rem) {
			res = fq12.Mul(res, exp)
		}
		exp = fq12.Square(exp)
		rem = new(big.Int).Rsh(rem, 1)
	}
	return res
}
func (fq12 Fq12) Affine(a [2][3][2]*big.Int) [2][3][2]*big.Int {
	return [2][3][2]*big.Int{
		fq12.F.Affine(a[0]),
		fq12.F.Affine(a[1]),
	}
}
func (fq12 Fq12) Equal(a, b [2][3][2]*big.Int) bool {
	return fq12.F.Equal(a[0], b[0]) && fq12.F.Equal(a[1], b[1])
}
