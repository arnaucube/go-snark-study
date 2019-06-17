package fields

import (
	"math/big"
)

// Fq2 is Field 2
type Fq2 struct {
	F          Fq
	NonResidue *big.Int
}

// NewFq2 generates a new Fq2
func NewFq2(f Fq, nonResidue *big.Int) Fq2 {
	fq2 := Fq2{
		f,
		nonResidue,
	}
	return fq2
}

// Zero returns a Zero value on the Fq2
func (fq2 Fq2) Zero() [2]*big.Int {
	return [2]*big.Int{fq2.F.Zero(), fq2.F.Zero()}
}

// One returns a One value on the Fq2
func (fq2 Fq2) One() [2]*big.Int {
	return [2]*big.Int{fq2.F.One(), fq2.F.Zero()}
}

func (fq2 Fq2) mulByNonResidue(a *big.Int) *big.Int {
	return fq2.F.Mul(fq2.NonResidue, a)
}

// Add performs an addition on the Fq2
func (fq2 Fq2) Add(a, b [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{
		fq2.F.Add(a[0], b[0]),
		fq2.F.Add(a[1], b[1]),
	}
}

// Double performs a doubling on the Fq2
func (fq2 Fq2) Double(a [2]*big.Int) [2]*big.Int {
	return fq2.Add(a, a)
}

// Sub performs a subtraction on the Fq2
func (fq2 Fq2) Sub(a, b [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{
		fq2.F.Sub(a[0], b[0]),
		fq2.F.Sub(a[1], b[1]),
	}
}

// Neg performs a negation on the Fq2
func (fq2 Fq2) Neg(a [2]*big.Int) [2]*big.Int {
	return fq2.Sub(fq2.Zero(), a)
}

// Mul performs a multiplication on the Fq2
func (fq2 Fq2) Mul(a, b [2]*big.Int) [2]*big.Int {
	// Multiplication and Squaring on Pairing-Friendly.pdf; Section 3 (Karatsuba)
	// https://pdfs.semanticscholar.org/3e01/de88d7428076b2547b60072088507d881bf1.pdf
	v0 := fq2.F.Mul(a[0], b[0])
	v1 := fq2.F.Mul(a[1], b[1])
	return [2]*big.Int{
		fq2.F.Add(v0, fq2.mulByNonResidue(v1)),
		fq2.F.Sub(
			fq2.F.Mul(
				fq2.F.Add(a[0], a[1]),
				fq2.F.Add(b[0], b[1])),
			fq2.F.Add(v0, v1)),
	}
}

// MulScalar is ...
func (fq2 Fq2) MulScalar(p [2]*big.Int, e *big.Int) [2]*big.Int {
	// for more possible implementations see g2.go file, at the function g2.MulScalar()

	q := fq2.Zero()
	d := fq2.F.Copy(e)
	r := p

	foundone := false
	for i := d.BitLen(); i >= 0; i-- {
		if foundone {
			q = fq2.Double(q)
		}
		if d.Bit(i) == 1 {
			foundone = true
			q = fq2.Add(q, r)
		}
	}
	return q
}

// Inverse returns the inverse on the Fq2
func (fq2 Fq2) Inverse(a [2]*big.Int) [2]*big.Int {
	// High-Speed Software Implementation of the Optimal Ate Pairing over Barreto–Naehrig Curves .pdf
	// https://eprint.iacr.org/2010/354.pdf , algorithm 8
	t0 := fq2.F.Square(a[0])
	t1 := fq2.F.Square(a[1])
	t2 := fq2.F.Sub(t0, fq2.mulByNonResidue(t1))
	t3 := fq2.F.Inverse(t2)
	return [2]*big.Int{
		fq2.F.Mul(a[0], t3),
		fq2.F.Neg(fq2.F.Mul(a[1], t3)),
	}
}

// Div performs a division on the Fq2
func (fq2 Fq2) Div(a, b [2]*big.Int) [2]*big.Int {
	return fq2.Mul(a, fq2.Inverse(b))
}

// Square performs a square operation on the Fq2
func (fq2 Fq2) Square(a [2]*big.Int) [2]*big.Int {
	// https://pdfs.semanticscholar.org/3e01/de88d7428076b2547b60072088507d881bf1.pdf , complex squaring
	ab := fq2.F.Mul(a[0], a[1])
	return [2]*big.Int{
		fq2.F.Sub(
			fq2.F.Mul(
				fq2.F.Add(a[0], a[1]),
				fq2.F.Add(
					a[0],
					fq2.mulByNonResidue(a[1]))),
			fq2.F.Add(
				ab,
				fq2.mulByNonResidue(ab))),
		fq2.F.Add(ab, ab),
	}
}

// IsZero is ...
func (fq2 Fq2) IsZero(a [2]*big.Int) bool {
	return fq2.F.IsZero(a[0]) && fq2.F.IsZero(a[1])
}

// Affine is ...
func (fq2 Fq2) Affine(a [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{
		fq2.F.Affine(a[0]),
		fq2.F.Affine(a[1]),
	}
}

// Equal is ...
func (fq2 Fq2) Equal(a, b [2]*big.Int) bool {
	return fq2.F.Equal(a[0], b[0]) && fq2.F.Equal(a[1], b[1])
}

// Copy is ...
func (fq2 Fq2) Copy(a [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{
		fq2.F.Copy(a[0]),
		fq2.F.Copy(a[1]),
	}
}
