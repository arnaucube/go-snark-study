# go-snark [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/go-snark)](https://goreportcard.com/report/github.com/arnaucube/go-snark)

zk-SNARK library implementation in Go


#### Test
```
go test ./... -v
```


## R1CS to Quadratic Arithmetic Program
- `Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture`, Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, Madars Virza https://eprint.iacr.org/2013/879.pdf
- Vitalik Buterin blog post about QAP https://medium.com/@VitalikButerin/quadratic-arithmetic-programs-from-zero-to-hero-f6d558cea649
- Ariel Gabizon in Zcash blog https://z.cash/blog/snark-explain5
- Lagrange polynomial Wikipedia article https://en.wikipedia.org/wiki/Lagrange_polynomial

#### Usage
- R1CS to QAP
```go
pf := NewPolynomialField(f)

b0 := big.NewInt(int64(0))
b1 := big.NewInt(int64(1))
b3 := big.NewInt(int64(3))
b5 := big.NewInt(int64(5))
b9 := big.NewInt(int64(9))
b27 := big.NewInt(int64(27))
b30 := big.NewInt(int64(30))
b35 := big.NewInt(int64(35))
a := [][]*big.Int{
  []*big.Int{b0, b1, b0, b0, b0, b0},
  []*big.Int{b0, b0, b0, b1, b0, b0},
  []*big.Int{b0, b1, b0, b0, b1, b0},
  []*big.Int{b5, b0, b0, b0, b0, b1},
}
b := [][]*big.Int{
  []*big.Int{b0, b1, b0, b0, b0, b0},
  []*big.Int{b0, b1, b0, b0, b0, b0},
  []*big.Int{b1, b0, b0, b0, b0, b0},
  []*big.Int{b1, b0, b0, b0, b0, b0},
}
c := [][]*big.Int{
  []*big.Int{b0, b0, b0, b1, b0, b0},
  []*big.Int{b0, b0, b0, b0, b1, b0},
  []*big.Int{b0, b0, b0, b0, b0, b1},
  []*big.Int{b0, b0, b1, b0, b0, b0},
}
alphas, betas, gammas, zx := pf.R1CSToQAP(a, b, c)
fmt.Println(alphas)
fmt.Println(betas)
fmt.Println(gammas)
fmt.Println(z)

w := []*big.Int{b1, b3, b35, b9, b27, b30}
ax, bx, cx, px := pf.CombinePolynomials(w, alphas, betas, gammas)
fmt.Println(ax)
fmt.Println(bx)
fmt.Println(cx)
fmt.Println(px)

hx := pf.DivisorPolinomial(px, zx)
fmt.Println(hx)
```

## Bn128
Implementation of the bn128 pairing in Go.


Implementation followng the information and the implementations from:
- `Multiplication and Squaring on Pairing-Friendly
Fields`, Augusto Jun Devegili, Colm Ó hÉigeartaigh, Michael Scott, and Ricardo Dahab https://pdfs.semanticscholar.org/3e01/de88d7428076b2547b60072088507d881bf1.pdf
- `Optimal Pairings`, Frederik Vercauteren https://www.cosic.esat.kuleuven.be/bcrypt/optimal.pdf , https://eprint.iacr.org/2008/096.pdf
- `Double-and-Add with Relative Jacobian
Coordinates`, Björn Fay https://eprint.iacr.org/2014/1014.pdf
- `Fast and Regular Algorithms for Scalar Multiplication
over Elliptic Curves`, Matthieu Rivain https://eprint.iacr.org/2011/338.pdf
- `High-Speed Software Implementation of the Optimal Ate Pairing over Barreto–Naehrig Curves`,  Jean-Luc Beuchat, Jorge E. González-Díaz, Shigeo Mitsunari, Eiji Okamoto, Francisco Rodríguez-Henríquez, and Tadanori Teruya https://eprint.iacr.org/2010/354.pdf
- `New software speed records for cryptographic pairings`, Michael Naehrig, Ruben Niederhagen, Peter Schwabe https://cryptojedi.org/papers/dclxvi-20100714.pdf
- `Implementing Cryptographic Pairings over Barreto-Naehrig Curves`, Augusto Jun Devegili, Michael Scott, Ricardo Dahab https://eprint.iacr.org/2007/390.pdf
- https://github.com/zcash/zcash/tree/master/src/snark
- https://github.com/iden3/snarkjs
- https://github.com/ethereum/py_ecc/tree/master/py_ecc/bn128


#### Usage

- Pairing
```go
bn128, err := NewBn128()
assert.Nil(t, err)

big25 := big.NewInt(int64(25))
big30 := big.NewInt(int64(30))

g1a := bn128.G1.MulScalar(bn128.G1.G, big25)
g2a := bn128.G2.MulScalar(bn128.G2.G, big30)

g1b := bn128.G1.MulScalar(bn128.G1.G, big30)
g2b := bn128.G2.MulScalar(bn128.G2.G, big25)

pA, err := bn128.Pairing(g1a, g2a)
assert.Nil(t, err)
pB, err := bn128.Pairing(g1b, g2b)
assert.Nil(t, err)
assert.True(t, bn128.Fq12.Equal(pA, pB))
```


---

## Caution
Not finished, work in progress (implementing this in my free time to understand it better, so I don't have much time).

Thanks to [@jbaylina](https://github.com/jbaylina), [@bellesmarta](https://github.com/bellesmarta), [@adriamb](https://github.com/adriamb) for their explanations that helped to understand this a little bit.
