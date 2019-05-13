package bn128

import (
	"math/big"

	"github.com/mottla/go-snark/fields"
)

type G1 struct {
	F fields.Fq
	G [3]*big.Int
}

func NewG1(f fields.Fq, g [2]*big.Int) G1 {
	var g1 G1
	g1.F = f
	g1.G = [3]*big.Int{
		g[0],
		g[1],
		g1.F.One(),
	}
	return g1
}

func (g1 G1) Zero() [2]*big.Int {
	return [2]*big.Int{g1.F.Zero(), g1.F.Zero()}
}
func (g1 G1) IsZero(p [3]*big.Int) bool {
	return g1.F.IsZero(p[2])
}

func (g1 G1) Add(p1, p2 [3]*big.Int) [3]*big.Int {

	// https://en.wikibooks.org/wiki/Cryptography/Prime_Curve/Jacobian_Coordinates
	// https://github.com/zcash/zcash/blob/master/src/snark/libsnark/algebra/curves/alt_bn128/alt_bn128_g1.cpp#L208
	// http://hyperelliptic.org/EFD/g1p/auto-code/shortw/jacobian-0/addition/add-2007-bl.op3

	if g1.IsZero(p1) {
		return p2
	}
	if g1.IsZero(p2) {
		return p1
	}

	x1 := p1[0]
	y1 := p1[1]
	z1 := p1[2]
	x2 := p2[0]
	y2 := p2[1]
	z2 := p2[2]

	z1z1 := g1.F.Square(z1)
	z2z2 := g1.F.Square(z2)

	u1 := g1.F.Mul(x1, z2z2)
	u2 := g1.F.Mul(x2, z1z1)

	t0 := g1.F.Mul(z2, z2z2)
	s1 := g1.F.Mul(y1, t0)

	t1 := g1.F.Mul(z1, z1z1)
	s2 := g1.F.Mul(y2, t1)

	h := g1.F.Sub(u2, u1)
	t2 := g1.F.Add(h, h)
	i := g1.F.Square(t2)
	j := g1.F.Mul(h, i)
	t3 := g1.F.Sub(s2, s1)
	r := g1.F.Add(t3, t3)
	v := g1.F.Mul(u1, i)
	t4 := g1.F.Square(r)
	t5 := g1.F.Add(v, v)
	t6 := g1.F.Sub(t4, j)
	x3 := g1.F.Sub(t6, t5)
	t7 := g1.F.Sub(v, x3)
	t8 := g1.F.Mul(s1, j)
	t9 := g1.F.Add(t8, t8)
	t10 := g1.F.Mul(r, t7)

	y3 := g1.F.Sub(t10, t9)

	t11 := g1.F.Add(z1, z2)
	t12 := g1.F.Square(t11)
	t13 := g1.F.Sub(t12, z1z1)
	t14 := g1.F.Sub(t13, z2z2)
	z3 := g1.F.Mul(t14, h)

	return [3]*big.Int{x3, y3, z3}
}

func (g1 G1) Neg(p [3]*big.Int) [3]*big.Int {
	return [3]*big.Int{
		p[0],
		g1.F.Neg(p[1]),
		p[2],
	}
}
func (g1 G1) Sub(a, b [3]*big.Int) [3]*big.Int {
	return g1.Add(a, g1.Neg(b))
}
func (g1 G1) Double(p [3]*big.Int) [3]*big.Int {

	// https://en.wikibooks.org/wiki/Cryptography/Prime_Curve/Jacobian_Coordinates
	// http://hyperelliptic.org/EFD/g1p/auto-code/shortw/jacobian-0/doubling/dbl-2009-l.op3
	// https://github.com/zcash/zcash/blob/master/src/snark/libsnark/algebra/curves/alt_bn128/alt_bn128_g1.cpp#L325

	if g1.IsZero(p) {
		return p
	}

	a := g1.F.Square(p[0])
	b := g1.F.Square(p[1])
	c := g1.F.Square(b)

	t0 := g1.F.Add(p[0], b)
	t1 := g1.F.Square(t0)
	t2 := g1.F.Sub(t1, a)
	t3 := g1.F.Sub(t2, c)

	d := g1.F.Double(t3)
	e := g1.F.Add(g1.F.Add(a, a), a)
	f := g1.F.Square(e)

	t4 := g1.F.Double(d)
	x3 := g1.F.Sub(f, t4)

	t5 := g1.F.Sub(d, x3)
	twoC := g1.F.Add(c, c)
	fourC := g1.F.Add(twoC, twoC)
	t6 := g1.F.Add(fourC, fourC)
	t7 := g1.F.Mul(e, t5)
	y3 := g1.F.Sub(t7, t6)

	t8 := g1.F.Mul(p[1], p[2])
	z3 := g1.F.Double(t8)

	return [3]*big.Int{x3, y3, z3}
}

func (g1 G1) MulScalar(p [3]*big.Int, e *big.Int) [3]*big.Int {
	// https://en.wikipedia.org/wiki/Elliptic_curve_point_multiplication#Double-and-add
	// for more possible implementations see g2.go file, at the function g2.MulScalar()

	q := [3]*big.Int{g1.F.Zero(), g1.F.Zero(), g1.F.Zero()}
	d := g1.F.Copy(e)
	r := p
	for i := d.BitLen() - 1; i >= 0; i-- {
		q = g1.Double(q)
		if d.Bit(i) == 1 {
			q = g1.Add(q, r)
		}
	}

	return q
}

func (g1 G1) Affine(p [3]*big.Int) [2]*big.Int {
	if g1.IsZero(p) {
		return g1.Zero()
	}

	zinv := g1.F.Inverse(p[2])
	zinv2 := g1.F.Square(zinv)
	x := g1.F.Mul(p[0], zinv2)

	zinv3 := g1.F.Mul(zinv2, zinv)
	y := g1.F.Mul(p[1], zinv3)

	return [2]*big.Int{x, y}
}

func (g1 G1) Equal(p1, p2 [3]*big.Int) bool {
	if g1.IsZero(p1) {
		return g1.IsZero(p2)
	}
	if g1.IsZero(p2) {
		return g1.IsZero(p1)
	}

	z1z1 := g1.F.Square(p1[2])
	z2z2 := g1.F.Square(p2[2])

	u1 := g1.F.Mul(p1[0], z2z2)
	u2 := g1.F.Mul(p2[0], z1z1)

	z1cub := g1.F.Mul(p1[2], z1z1)
	z2cub := g1.F.Mul(p2[2], z2z2)

	s1 := g1.F.Mul(p1[1], z2cub)
	s2 := g1.F.Mul(p2[1], z1cub)

	return g1.F.Equal(u1, u2) && g1.F.Equal(s1, s2)
}
