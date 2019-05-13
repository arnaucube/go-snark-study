package bn128

import (
	"math/big"

	"github.com/mottla/go-snark/fields"
)

type G2 struct {
	F fields.Fq2
	G [3][2]*big.Int
}

func NewG2(f fields.Fq2, g [2][2]*big.Int) G2 {
	var g2 G2
	g2.F = f
	g2.G = [3][2]*big.Int{
		g[0],
		g[1],
		g2.F.One(),
	}
	return g2
}

func (g2 G2) Zero() [3][2]*big.Int {
	return [3][2]*big.Int{g2.F.Zero(), g2.F.One(), g2.F.Zero()}
}
func (g2 G2) IsZero(p [3][2]*big.Int) bool {
	return g2.F.IsZero(p[2])
}

func (g2 G2) Add(p1, p2 [3][2]*big.Int) [3][2]*big.Int {

	// https://en.wikibooks.org/wiki/Cryptography/Prime_Curve/Jacobian_Coordinates
	// https://github.com/zcash/zcash/blob/master/src/snark/libsnark/algebra/curves/alt_bn128/alt_bn128_g2.cpp#L208
	// http://hyperelliptic.org/EFD/g2p/auto-code/shortw/jacobian-0/addition/add-2007-bl.op3

	if g2.IsZero(p1) {
		return p2
	}
	if g2.IsZero(p2) {
		return p1
	}

	x1 := p1[0]
	y1 := p1[1]
	z1 := p1[2]
	x2 := p2[0]
	y2 := p2[1]
	z2 := p2[2]

	z1z1 := g2.F.Square(z1)
	z2z2 := g2.F.Square(z2)

	u1 := g2.F.Mul(x1, z2z2)
	u2 := g2.F.Mul(x2, z1z1)

	t0 := g2.F.Mul(z2, z2z2)
	s1 := g2.F.Mul(y1, t0)

	t1 := g2.F.Mul(z1, z1z1)
	s2 := g2.F.Mul(y2, t1)

	h := g2.F.Sub(u2, u1)
	t2 := g2.F.Add(h, h)
	i := g2.F.Square(t2)
	j := g2.F.Mul(h, i)
	t3 := g2.F.Sub(s2, s1)
	r := g2.F.Add(t3, t3)
	v := g2.F.Mul(u1, i)
	t4 := g2.F.Square(r)
	t5 := g2.F.Add(v, v)
	t6 := g2.F.Sub(t4, j)
	x3 := g2.F.Sub(t6, t5)
	t7 := g2.F.Sub(v, x3)
	t8 := g2.F.Mul(s1, j)
	t9 := g2.F.Add(t8, t8)
	t10 := g2.F.Mul(r, t7)

	y3 := g2.F.Sub(t10, t9)

	t11 := g2.F.Add(z1, z2)
	t12 := g2.F.Square(t11)
	t13 := g2.F.Sub(t12, z1z1)
	t14 := g2.F.Sub(t13, z2z2)
	z3 := g2.F.Mul(t14, h)

	return [3][2]*big.Int{x3, y3, z3}
}

func (g2 G2) Neg(p [3][2]*big.Int) [3][2]*big.Int {
	return [3][2]*big.Int{
		p[0],
		g2.F.Neg(p[1]),
		p[2],
	}
}

func (g2 G2) Sub(a, b [3][2]*big.Int) [3][2]*big.Int {
	return g2.Add(a, g2.Neg(b))
}

func (g2 G2) Double(p [3][2]*big.Int) [3][2]*big.Int {

	// https://en.wikibooks.org/wiki/Cryptography/Prime_Curve/Jacobian_Coordinates
	// http://hyperelliptic.org/EFD/g2p/auto-code/shortw/jacobian-0/doubling/dbl-2009-l.op3
	// https://github.com/zcash/zcash/blob/master/src/snark/libsnark/algebra/curves/alt_bn128/alt_bn128_g2.cpp#L325

	if g2.IsZero(p) {
		return p
	}

	a := g2.F.Square(p[0])
	b := g2.F.Square(p[1])
	c := g2.F.Square(b)

	t0 := g2.F.Add(p[0], b)
	t1 := g2.F.Square(t0)
	t2 := g2.F.Sub(t1, a)
	t3 := g2.F.Sub(t2, c)

	d := g2.F.Double(t3)
	e := g2.F.Add(g2.F.Add(a, a), a)
	f := g2.F.Square(e)

	t4 := g2.F.Double(d)
	x3 := g2.F.Sub(f, t4)

	t5 := g2.F.Sub(d, x3)
	twoC := g2.F.Add(c, c)
	fourC := g2.F.Add(twoC, twoC)
	t6 := g2.F.Add(fourC, fourC)
	t7 := g2.F.Mul(e, t5)
	y3 := g2.F.Sub(t7, t6)

	t8 := g2.F.Mul(p[1], p[2])
	z3 := g2.F.Double(t8)

	return [3][2]*big.Int{x3, y3, z3}
}

func (g2 G2) MulScalar(p [3][2]*big.Int, e *big.Int) [3][2]*big.Int {
	// https://en.wikipedia.org/wiki/Elliptic_curve_point_multiplication#Double-and-add

	q := [3][2]*big.Int{g2.F.Zero(), g2.F.Zero(), g2.F.Zero()}
	d := g2.F.F.Copy(e) // d := e
	r := p

	/*
		here are three possible implementations:
	*/

	/* index decreasing: */
	for i := d.BitLen() - 1; i >= 0; i-- {
		q = g2.Double(q)
		if d.Bit(i) == 1 {
			q = g2.Add(q, r)
		}
	}

	/* index increasing: */
	// for i := 0; i <= d.BitLen(); i++ {
	// 	if d.Bit(i) == 1 {
	// 		q = g2.Add(q, r)
	// 	}
	// 	r = g2.Double(r)
	// }

	// foundone := false
	// for i := d.BitLen(); i >= 0; i-- {
	// 	if foundone {
	// 		q = g2.Double(q)
	// 	}
	// 	if d.Bit(i) == 1 {
	// 		foundone = true
	// 		q = g2.Add(q, r)
	// 	}
	// }

	return q
}

func (g2 G2) Affine(p [3][2]*big.Int) [3][2]*big.Int {
	if g2.IsZero(p) {
		return g2.Zero()
	}

	zinv := g2.F.Inverse(p[2])
	zinv2 := g2.F.Square(zinv)
	x := g2.F.Mul(p[0], zinv2)

	zinv3 := g2.F.Mul(zinv2, zinv)
	y := g2.F.Mul(p[1], zinv3)

	return [3][2]*big.Int{
		g2.F.Affine(x),
		g2.F.Affine(y),
		g2.F.One(),
	}
}

func (g2 G2) Equal(p1, p2 [3][2]*big.Int) bool {
	if g2.IsZero(p1) {
		return g2.IsZero(p2)
	}
	if g2.IsZero(p2) {
		return g2.IsZero(p1)
	}

	z1z1 := g2.F.Square(p1[2])
	z2z2 := g2.F.Square(p2[2])

	u1 := g2.F.Mul(p1[0], z2z2)
	u2 := g2.F.Mul(p2[0], z1z1)

	z1cub := g2.F.Mul(p1[2], z1z1)
	z2cub := g2.F.Mul(p2[2], z2z2)

	s1 := g2.F.Mul(p1[1], z2cub)
	s2 := g2.F.Mul(p2[1], z1cub)

	return g2.F.Equal(u1, u2) && g2.F.Equal(s1, s2)
}
