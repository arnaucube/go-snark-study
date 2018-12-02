package fields

import (
	"bytes"
	"math/big"
)

// Fq6 is Field 6
type Fq6 struct {
	F          Fq2
	NonResidue [2]*big.Int
}

// NewFq6 generates a new Fq6
func NewFq6(f Fq2, nonResidue [2]*big.Int) Fq6 {
	fq6 := Fq6{
		f,
		nonResidue,
	}
	return fq6
}

// Zero returns a Zero value on the Fq6
func (fq6 Fq6) Zero() [3][2]*big.Int {
	return [3][2]*big.Int{fq6.F.Zero(), fq6.F.Zero(), fq6.F.Zero()}
}

// One returns a One value on the Fq6
func (fq6 Fq6) One() [3][2]*big.Int {
	return [3][2]*big.Int{fq6.F.One(), fq6.F.Zero(), fq6.F.Zero()}
}

func (fq6 Fq6) mulByNonResidue(a [2]*big.Int) [2]*big.Int {
	return fq6.F.Mul(fq6.NonResidue, a)
}

// Add performs an addition on the Fq6
func (fq6 Fq6) Add(a, b [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		fq6.F.Add(a[0], b[0]),
		fq6.F.Add(a[1], b[1]),
		fq6.F.Add(a[2], b[2]),
	}
}

func (fq6 Fq6) Double(a [3][2]*big.Int) [3][2]*big.Int {
	return fq6.Add(a, a)
}

// Sub performs a subtraction on the Fq6
func (fq6 Fq6) Sub(a, b [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		fq6.F.Sub(a[0], b[0]),
		fq6.F.Sub(a[1], b[1]),
		fq6.F.Sub(a[2], b[2]),
	}
}

// Neg performs a negation on the Fq6
func (fq6 Fq6) Neg(a [3][2]*big.Int) [3][2]*big.Int {
	return fq6.Sub(fq6.Zero(), a)
}

// Mul performs a multiplication on the Fq6
func (fq6 Fq6) Mul(a, b [3][2]*big.Int) [3][2]*big.Int {
	v0 := fq6.F.Mul(a[0], b[0])
	v1 := fq6.F.Mul(a[1], b[1])
	v2 := fq6.F.Mul(a[2], b[2])
	return [3][2]*big.Int{
		fq6.F.Add(
			v0,
			fq6.mulByNonResidue(
				fq6.F.Sub(
					fq6.F.Mul(
						fq6.F.Add(a[1], a[2]),
						fq6.F.Add(b[1], b[2])),
					fq6.F.Add(v1, v2)))),

		fq6.F.Add(
			fq6.F.Sub(
				fq6.F.Mul(
					fq6.F.Add(a[0], a[1]),
					fq6.F.Add(b[0], b[1])),
				fq6.F.Add(v0, v1)),
			fq6.mulByNonResidue(v2)),

		fq6.F.Add(
			fq6.F.Sub(
				fq6.F.Mul(
					fq6.F.Add(a[0], a[2]),
					fq6.F.Add(b[0], b[2])),
				fq6.F.Add(v0, v2)),
			v1),
	}
}

func (fq6 Fq6) MulScalar(base [3][2]*big.Int, e *big.Int) [3][2]*big.Int {
	// for more possible implementations see g2.go file, at the function g2.MulScalar()

	res := fq6.Zero()
	rem := e
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		// if rem % 2 == 1
		if bytes.Equal(new(big.Int).Rem(rem, big.NewInt(int64(2))).Bytes(), big.NewInt(int64(1)).Bytes()) {
			res = fq6.Add(res, exp)
		}
		exp = fq6.Double(exp)
		rem = rem.Rsh(rem, 1) // rem = rem >> 1
	}
	return res
}

// Inverse returns the inverse on the Fq6
func (fq6 Fq6) Inverse(a [3][2]*big.Int) [3][2]*big.Int {
	t0 := fq6.F.Square(a[0])
	t1 := fq6.F.Square(a[1])
	t2 := fq6.F.Square(a[2])
	t3 := fq6.F.Mul(a[0], a[1])
	t4 := fq6.F.Mul(a[0], a[2])
	t5 := fq6.F.Mul(a[1], a[2])

	c0 := fq6.F.Sub(t0, fq6.mulByNonResidue(t5))
	c1 := fq6.F.Sub(fq6.mulByNonResidue(t2), t3)
	c2 := fq6.F.Sub(t1, t4)

	t6 := fq6.F.Inverse(
		fq6.F.Add(
			fq6.F.Mul(a[0], c0),
			fq6.mulByNonResidue(
				fq6.F.Add(
					fq6.F.Mul(a[2], c1),
					fq6.F.Mul(a[1], c2)))))
	return [3][2]*big.Int{
		fq6.F.Mul(t6, c0),
		fq6.F.Mul(t6, c1),
		fq6.F.Mul(t6, c2),
	}
}

// Div performs a division on the Fq6
func (fq6 Fq6) Div(a, b [3][2]*big.Int) [3][2]*big.Int {
	return fq6.Mul(a, fq6.Inverse(b))
}

// Square performs a square operation on the Fq6
func (fq6 Fq6) Square(a [3][2]*big.Int) [3][2]*big.Int {
	s0 := fq6.F.Square(a[0])
	ab := fq6.F.Mul(a[0], a[1])
	s1 := fq6.F.Add(ab, ab)
	s2 := fq6.F.Square(
		fq6.F.Add(
			fq6.F.Sub(a[0], a[1]),
			a[2]))
	bc := fq6.F.Mul(a[1], a[2])
	s3 := fq6.F.Add(bc, bc)
	s4 := fq6.F.Square(a[2])

	return [3][2]*big.Int{
		fq6.F.Add(
			s0,
			fq6.mulByNonResidue(s3)),
		fq6.F.Add(
			s1,
			fq6.mulByNonResidue(s4)),
		fq6.F.Sub(
			fq6.F.Add(
				fq6.F.Add(s1, s2),
				s3),
			fq6.F.Add(s0, s4)),
	}
}

func (fq6 Fq6) Affine(a [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		fq6.F.Affine(a[0]),
		fq6.F.Affine(a[1]),
		fq6.F.Affine(a[2]),
	}
}
func (fq6 Fq6) Equal(a, b [3][2]*big.Int) bool {
	return fq6.F.Equal(a[0], b[0]) && fq6.F.Equal(a[1], b[1]) && fq6.F.Equal(a[2], b[2])
}

func (fq6 Fq6) Copy(a [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		fq6.F.Copy(a[0]),
		fq6.F.Copy(a[1]),
		fq6.F.Copy(a[2]),
	}
}
