package fields

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

// Fq is the Z field over modulus Q
type Fq struct {
	Q *big.Int // Q
}

// NewFq generates a new Fq
func NewFq(q *big.Int) Fq {
	return Fq{
		q,
	}
}

// Zero returns a Zero value on the Fq
func (fq Fq) Zero() *big.Int {
	return big.NewInt(int64(0))
}

// One returns a One value on the Fq
func (fq Fq) One() *big.Int {
	return big.NewInt(int64(1))
}

// Add performs an addition on the Fq
func (fq Fq) Add(a, b *big.Int) *big.Int {
	r := new(big.Int).Add(a, b)
	return new(big.Int).Mod(r, fq.Q)
}

// Double performs a doubling on the Fq
func (fq Fq) Double(a *big.Int) *big.Int {
	r := new(big.Int).Add(a, a)
	return new(big.Int).Mod(r, fq.Q)
}

// Sub performs a subtraction on the Fq
func (fq Fq) Sub(a, b *big.Int) *big.Int {
	r := new(big.Int).Sub(a, b)
	return new(big.Int).Mod(r, fq.Q)
}

// Neg performs a negation on the Fq
func (fq Fq) Neg(a *big.Int) *big.Int {
	m := new(big.Int).Neg(a)
	return new(big.Int).Mod(m, fq.Q)
}

// Mul performs a multiplication on the Fq
func (fq Fq) Mul(a, b *big.Int) *big.Int {
	m := new(big.Int).Mul(a, b)
	return new(big.Int).Mod(m, fq.Q)
}

// MulScalar is ...
func (fq Fq) MulScalar(base, e *big.Int) *big.Int {
	return fq.Mul(base, e)
}

// Inverse returns the inverse on the Fq
func (fq Fq) Inverse(a *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, fq.Q)
	// q := bigCopy(fq.Q)
	// t := big.NewInt(int64(0))
	// r := fq.Q
	// newt := big.NewInt(int64(0))
	// newr := fq.Affine(a)
	// for !bytes.Equal(newr.Bytes(), big.NewInt(int64(0)).Bytes()) {
	// 	q := new(big.Int).Div(bigCopy(r), bigCopy(newr))
	//
	// 	t = bigCopy(newt)
	// 	newt = fq.Sub(t, fq.Mul(q, newt))
	//
	// 	r = bigCopy(newr)
	// 	newr = fq.Sub(r, fq.Mul(q, newr))
	// }
	// if t.Cmp(big.NewInt(0)) == -1 { // t< 0
	// 	t = fq.Add(t, q)
	// }
	// return t
}

// Div performs the division over the finite field
func (fq Fq) Div(a, b *big.Int) *big.Int {
	d := fq.Mul(a, fq.Inverse(b))
	return new(big.Int).Mod(d, fq.Q)
}

// Square performs a square operation on the Fq
func (fq Fq) Square(a *big.Int) *big.Int {
	m := new(big.Int).Mul(a, a)
	return new(big.Int).Mod(m, fq.Q)
}

// Exp performs the exponential over Fq
func (fq Fq) Exp(base *big.Int, e *big.Int) *big.Int {
	res := fq.One()
	rem := fq.Copy(e)
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		if BigIsOdd(rem) {
			res = fq.Mul(res, exp)
		}
		exp = fq.Square(exp)
		rem = new(big.Int).Rsh(rem, 1)
	}
	return res
}

// Rand is ...
func (fq Fq) Rand() (*big.Int, error) {

	// twoexp := new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(maxbits)), nil)
	// max := new(big.Int).Sub(twoexp, big.NewInt(1))

	maxbits := fq.Q.BitLen()
	b := make([]byte, (maxbits/8)-1)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(b)
	rq := new(big.Int).Mod(r, fq.Q)

	// r over q, nil
	return rq, nil
}

// IsZero is ...
func (fq Fq) IsZero(a *big.Int) bool {
	return bytes.Equal(a.Bytes(), fq.Zero().Bytes())
}

// Copy is ...
func (fq Fq) Copy(a *big.Int) *big.Int {
	return new(big.Int).SetBytes(a.Bytes())
}

// Affine is ...
func (fq Fq) Affine(a *big.Int) *big.Int {
	nq := fq.Neg(fq.Q)

	aux := a
	if aux.Cmp(big.NewInt(int64(0))) == -1 { // negative value
		if aux.Cmp(nq) != 1 { // aux less or equal nq
			aux = new(big.Int).Mod(aux, fq.Q)
		}
		if aux.Cmp(big.NewInt(int64(0))) == -1 { // negative value
			aux = new(big.Int).Add(aux, fq.Q)
		}
	} else {
		if aux.Cmp(fq.Q) != -1 { // aux greater or equal nq
			aux = new(big.Int).Mod(aux, fq.Q)
		}
	}
	return aux
}

// Equal is ...
func (fq Fq) Equal(a, b *big.Int) bool {
	aAff := fq.Affine(a)
	bAff := fq.Affine(b)
	return bytes.Equal(aAff.Bytes(), bAff.Bytes())
}

// BigIsOdd is ...
func BigIsOdd(n *big.Int) bool {
	one := big.NewInt(int64(1))
	and := new(big.Int).And(n, one)
	return bytes.Equal(and.Bytes(), big.NewInt(int64(1)).Bytes())
}
