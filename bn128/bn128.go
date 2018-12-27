package bn128

import (
	"errors"
	"math/big"

	"github.com/arnaucube/go-snark/fields"
)

// Bn128 is the data structure of the BN128
type Bn128 struct {
	Q             *big.Int
	R             *big.Int
	Gg1           [2]*big.Int
	Gg2           [2][2]*big.Int
	NonResidueFq2 *big.Int
	NonResidueFq6 [2]*big.Int
	Fq1           fields.Fq
	Fq2           fields.Fq2
	Fq6           fields.Fq6
	Fq12          fields.Fq12
	G1            G1
	G2            G2
	LoopCount     *big.Int
	LoopCountNeg  bool

	TwoInv             *big.Int
	CoefB              *big.Int
	TwistCoefB         [2]*big.Int
	Twist              [2]*big.Int
	FrobeniusCoeffsC11 *big.Int
	TwistMulByQX       [2]*big.Int
	TwistMulByQY       [2]*big.Int
	FinalExp           *big.Int
}

// NewBn128 returns the BN128
func NewBn128() (Bn128, error) {
	var b Bn128
	q, ok := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)
	if !ok {
		return b, errors.New("err with q")
	}
	b.Q = q

	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	if !ok {
		return b, errors.New("err with r")
	}
	b.R = r

	b.Gg1 = [2]*big.Int{
		big.NewInt(int64(1)),
		big.NewInt(int64(2)),
	}

	g2_00, ok := new(big.Int).SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781", 10)
	if !ok {
		return b, errors.New("err with g2_00")
	}
	g2_01, ok := new(big.Int).SetString("11559732032986387107991004021392285783925812861821192530917403151452391805634", 10)
	if !ok {
		return b, errors.New("err with g2_00")
	}
	g2_10, ok := new(big.Int).SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930", 10)
	if !ok {
		return b, errors.New("err with g2_00")
	}
	g2_11, ok := new(big.Int).SetString("4082367875863433681332203403145435568316851327593401208105741076214120093531", 10)
	if !ok {
		return b, errors.New("err with g2_00")
	}

	b.Gg2 = [2][2]*big.Int{
		[2]*big.Int{
			g2_00,
			g2_01,
		},
		[2]*big.Int{
			g2_10,
			g2_11,
		},
	}

	b.Fq1 = fields.NewFq(q)
	b.NonResidueFq2, ok = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208582", 10) // i
	if !ok {
		return b, errors.New("err with nonResidueFq2")
	}
	b.NonResidueFq6 = [2]*big.Int{
		big.NewInt(int64(9)),
		big.NewInt(int64(1)),
	}

	b.Fq2 = fields.NewFq2(b.Fq1, b.NonResidueFq2)
	b.Fq6 = fields.NewFq6(b.Fq2, b.NonResidueFq6)
	b.Fq12 = fields.NewFq12(b.Fq6, b.Fq2, b.NonResidueFq6)

	b.G1 = NewG1(b.Fq1, b.Gg1)
	b.G2 = NewG2(b.Fq2, b.Gg2)

	err := b.preparePairing()
	if err != nil {
		return b, err
	}

	return b, nil
}

// NewFqR returns a new Finite Field over R
func NewFqR() (fields.Fq, error) {
	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	if !ok {
		return fields.Fq{}, errors.New("err parsing R")
	}
	fqR := fields.NewFq(r)
	return fqR, nil
}

func (bn128 *Bn128) preparePairing() error {
	var ok bool
	bn128.LoopCount, ok = new(big.Int).SetString("29793968203157093288", 10)
	if !ok {
		return errors.New("err with LoopCount from string")
	}

	bn128.LoopCountNeg = false

	bn128.TwoInv = bn128.Fq1.Inverse(big.NewInt(int64(2)))

	bn128.CoefB = big.NewInt(int64(3))
	bn128.Twist = [2]*big.Int{
		big.NewInt(int64(9)),
		big.NewInt(int64(1)),
	}
	bn128.TwistCoefB = bn128.Fq2.MulScalar(bn128.Fq2.Inverse(bn128.Twist), bn128.CoefB)

	bn128.FrobeniusCoeffsC11, ok = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208582", 10)
	if !ok {
		return errors.New("error parsing frobeniusCoeffsC11")
	}

	a, ok := new(big.Int).SetString("21575463638280843010398324269430826099269044274347216827212613867836435027261", 10)
	if !ok {
		return errors.New("error parsing a")
	}
	b, ok := new(big.Int).SetString("10307601595873709700152284273816112264069230130616436755625194854815875713954", 10)
	if !ok {
		return errors.New("error parsing b")
	}
	bn128.TwistMulByQX = [2]*big.Int{
		a,
		b,
	}

	a, ok = new(big.Int).SetString("2821565182194536844548159561693502659359617185244120367078079554186484126554", 10)
	if !ok {
		return errors.New("error parsing a")
	}
	b, ok = new(big.Int).SetString("3505843767911556378687030309984248845540243509899259641013678093033130930403", 10)
	if !ok {
		return errors.New("error parsing b")
	}
	bn128.TwistMulByQY = [2]*big.Int{
		a,
		b,
	}

	bn128.FinalExp, ok = new(big.Int).SetString("552484233613224096312617126783173147097382103762957654188882734314196910839907541213974502761540629817009608548654680343627701153829446747810907373256841551006201639677726139946029199968412598804882391702273019083653272047566316584365559776493027495458238373902875937659943504873220554161550525926302303331747463515644711876653177129578303191095900909191624817826566688241804408081892785725967931714097716709526092261278071952560171111444072049229123565057483750161460024353346284167282452756217662335528813519139808291170539072125381230815729071544861602750936964829313608137325426383735122175229541155376346436093930287402089517426973178917569713384748081827255472576937471496195752727188261435633271238710131736096299798168852925540549342330775279877006784354801422249722573783561685179618816480037695005515426162362431072245638324744480", 10)
	if !ok {
		return errors.New("error parsing finalExp")
	}

	return nil

}

// Pairing calculates the BN128 Pairing of two given values
func (bn128 Bn128) Pairing(p1 [3]*big.Int, p2 [3][2]*big.Int) [2][3][2]*big.Int {
	pre1 := bn128.preComputeG1(p1)
	pre2 := bn128.preComputeG2(p2)

	r1 := bn128.MillerLoop(pre1, pre2)
	res := bn128.finalExponentiation(r1)
	return res
}

type AteG1Precomp struct {
	Px *big.Int
	Py *big.Int
}

func (bn128 Bn128) preComputeG1(p [3]*big.Int) AteG1Precomp {
	pCopy := bn128.G1.Affine(p)
	res := AteG1Precomp{
		Px: pCopy[0],
		Py: pCopy[1],
	}
	return res
}

type EllCoeffs struct {
	Ell0  [2]*big.Int
	EllVW [2]*big.Int
	EllVV [2]*big.Int
}
type AteG2Precomp struct {
	Qx     [2]*big.Int
	Qy     [2]*big.Int
	Coeffs []EllCoeffs
}

func (bn128 Bn128) preComputeG2(p [3][2]*big.Int) AteG2Precomp {
	qCopy := bn128.G2.Affine(p)
	res := AteG2Precomp{
		qCopy[0],
		qCopy[1],
		[]EllCoeffs{},
	}
	r := [3][2]*big.Int{
		bn128.Fq2.Copy(qCopy[0]),
		bn128.Fq2.Copy(qCopy[1]),
		bn128.Fq2.One(),
	}
	var c EllCoeffs
	for i := bn128.LoopCount.BitLen() - 2; i >= 0; i-- {
		bit := bn128.LoopCount.Bit(i)

		c, r = bn128.doublingStep(r)
		res.Coeffs = append(res.Coeffs, c)
		if bit == 1 {
			c, r = bn128.mixedAdditionStep(qCopy, r)
			res.Coeffs = append(res.Coeffs, c)
		}
	}

	q1 := bn128.G2.Affine(bn128.g2MulByQ(qCopy))
	if !bn128.Fq2.Equal(q1[2], bn128.Fq2.One()) {
		// return res, errors.New("q1[2] != Fq2.One")
		panic(errors.New("q1[2] != Fq2.One()"))
	}
	q2 := bn128.G2.Affine(bn128.g2MulByQ(q1))
	if !bn128.Fq2.Equal(q2[2], bn128.Fq2.One()) {
		// return res, errors.New("q2[2] != Fq2.One")
		panic(errors.New("q2[2] != Fq2.One()"))
	}

	if bn128.LoopCountNeg {
		r[1] = bn128.Fq2.Neg(r[1])
	}
	q2[1] = bn128.Fq2.Neg(q2[1])

	c, r = bn128.mixedAdditionStep(q1, r)
	res.Coeffs = append(res.Coeffs, c)

	c, r = bn128.mixedAdditionStep(q2, r)
	res.Coeffs = append(res.Coeffs, c)

	return res
}

func (bn128 Bn128) doublingStep(current [3][2]*big.Int) (EllCoeffs, [3][2]*big.Int) {
	x := current[0]
	y := current[1]
	z := current[2]

	a := bn128.Fq2.MulScalar(bn128.Fq2.Mul(x, y), bn128.TwoInv)
	b := bn128.Fq2.Square(y)
	c := bn128.Fq2.Square(z)
	d := bn128.Fq2.Add(c, bn128.Fq2.Add(c, c))
	e := bn128.Fq2.Mul(bn128.TwistCoefB, d)
	f := bn128.Fq2.Add(e, bn128.Fq2.Add(e, e))
	g := bn128.Fq2.MulScalar(bn128.Fq2.Add(b, f), bn128.TwoInv)
	h := bn128.Fq2.Sub(
		bn128.Fq2.Square(bn128.Fq2.Add(y, z)),
		bn128.Fq2.Add(b, c))
	i := bn128.Fq2.Sub(e, b)
	j := bn128.Fq2.Square(x)
	eSqr := bn128.Fq2.Square(e)
	current[0] = bn128.Fq2.Mul(a, bn128.Fq2.Sub(b, f))
	current[1] = bn128.Fq2.Sub(bn128.Fq2.Sub(bn128.Fq2.Square(g), eSqr),
		bn128.Fq2.Add(eSqr, eSqr))
	current[2] = bn128.Fq2.Mul(b, h)
	res := EllCoeffs{
		Ell0:  bn128.Fq2.Mul(i, bn128.Twist),
		EllVW: bn128.Fq2.Neg(h),
		EllVV: bn128.Fq2.Add(j, bn128.Fq2.Add(j, j)),
	}

	return res, current
}

func (bn128 Bn128) mixedAdditionStep(base, current [3][2]*big.Int) (EllCoeffs, [3][2]*big.Int) {
	x1 := current[0]
	y1 := current[1]
	z1 := current[2]
	x2 := base[0]
	y2 := base[1]

	d := bn128.Fq2.Sub(x1, bn128.Fq2.Mul(x2, z1))
	e := bn128.Fq2.Sub(y1, bn128.Fq2.Mul(y2, z1))
	f := bn128.Fq2.Square(d)
	g := bn128.Fq2.Square(e)
	h := bn128.Fq2.Mul(d, f)
	i := bn128.Fq2.Mul(x1, f)
	j := bn128.Fq2.Sub(
		bn128.Fq2.Add(h, bn128.Fq2.Mul(z1, g)),
		bn128.Fq2.Add(i, i))

	current[0] = bn128.Fq2.Mul(d, j)
	current[1] = bn128.Fq2.Sub(
		bn128.Fq2.Mul(e, bn128.Fq2.Sub(i, j)),
		bn128.Fq2.Mul(h, y1))
	current[2] = bn128.Fq2.Mul(z1, h)

	coef := EllCoeffs{
		Ell0: bn128.Fq2.Mul(
			bn128.Twist,
			bn128.Fq2.Sub(
				bn128.Fq2.Mul(e, x2),
				bn128.Fq2.Mul(d, y2))),
		EllVW: d,
		EllVV: bn128.Fq2.Neg(e),
	}
	return coef, current
}
func (bn128 Bn128) g2MulByQ(p [3][2]*big.Int) [3][2]*big.Int {
	fmx := [2]*big.Int{
		p[0][0],
		bn128.Fq1.Mul(p[0][1], bn128.Fq1.Copy(bn128.FrobeniusCoeffsC11)),
	}
	fmy := [2]*big.Int{
		p[1][0],
		bn128.Fq1.Mul(p[1][1], bn128.Fq1.Copy(bn128.FrobeniusCoeffsC11)),
	}
	fmz := [2]*big.Int{
		p[2][0],
		bn128.Fq1.Mul(p[2][1], bn128.Fq1.Copy(bn128.FrobeniusCoeffsC11)),
	}

	return [3][2]*big.Int{
		bn128.Fq2.Mul(bn128.TwistMulByQX, fmx),
		bn128.Fq2.Mul(bn128.TwistMulByQY, fmy),
		fmz,
	}
}

func (bn128 Bn128) MillerLoop(pre1 AteG1Precomp, pre2 AteG2Precomp) [2][3][2]*big.Int {
	// https://cryptojedi.org/papers/dclxvi-20100714.pdf
	// https://eprint.iacr.org/2008/096.pdf

	idx := 0
	var c EllCoeffs
	f := bn128.Fq12.One()

	for i := bn128.LoopCount.BitLen() - 2; i >= 0; i-- {
		bit := bn128.LoopCount.Bit(i)

		c = pre2.Coeffs[idx]
		idx++
		f = bn128.Fq12.Square(f)

		f = bn128.mulBy024(f,
			c.Ell0,
			bn128.Fq2.MulScalar(c.EllVW, pre1.Py),
			bn128.Fq2.MulScalar(c.EllVV, pre1.Px))

		if bit == 1 {
			c = pre2.Coeffs[idx]
			idx++
			f = bn128.mulBy024(
				f,
				c.Ell0,
				bn128.Fq2.MulScalar(c.EllVW, pre1.Py),
				bn128.Fq2.MulScalar(c.EllVV, pre1.Px))
		}
	}
	if bn128.LoopCountNeg {
		f = bn128.Fq12.Inverse(f)
	}

	c = pre2.Coeffs[idx]
	idx++
	f = bn128.mulBy024(
		f,
		c.Ell0,
		bn128.Fq2.MulScalar(c.EllVW, pre1.Py),
		bn128.Fq2.MulScalar(c.EllVV, pre1.Px))

	c = pre2.Coeffs[idx]
	idx++

	f = bn128.mulBy024(
		f,
		c.Ell0,
		bn128.Fq2.MulScalar(c.EllVW, pre1.Py),
		bn128.Fq2.MulScalar(c.EllVV, pre1.Px))

	return f
}

func (bn128 Bn128) mulBy024(a [2][3][2]*big.Int, ell0, ellVW, ellVV [2]*big.Int) [2][3][2]*big.Int {
	b := [2][3][2]*big.Int{
		[3][2]*big.Int{
			ell0,
			bn128.Fq2.Zero(),
			ellVV,
		},
		[3][2]*big.Int{
			bn128.Fq2.Zero(),
			ellVW,
			bn128.Fq2.Zero(),
		},
	}
	return bn128.Fq12.Mul(a, b)
}

func (bn128 Bn128) finalExponentiation(r [2][3][2]*big.Int) [2][3][2]*big.Int {
	res := bn128.Fq12.Exp(r, bn128.FinalExp)
	return res
}
