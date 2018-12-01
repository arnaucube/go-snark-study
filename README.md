# go-snark [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/go-snark)](https://goreportcard.com/report/github.com/arnaucube/go-snark)

Not finished, work in progress (doing this in my free time, so I don't have much time).



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
b0 := big.NewFloat(float64(0))
b1 := big.NewFloat(float64(1))
b5 := big.NewFloat(float64(5))
a := [][]*big.Float{
  []*big.Float{b0, b1, b0, b0, b0, b0},
  []*big.Float{b0, b0, b0, b1, b0, b0},
  []*big.Float{b0, b1, b0, b0, b1, b0},
  []*big.Float{b5, b0, b0, b0, b0, b1},
}
b := [][]*big.Float{
  []*big.Float{b0, b1, b0, b0, b0, b0},
  []*big.Float{b0, b1, b0, b0, b0, b0},
  []*big.Float{b1, b0, b0, b0, b0, b0},
  []*big.Float{b1, b0, b0, b0, b0, b0},
}
c := [][]*big.Float{
  []*big.Float{b0, b0, b0, b1, b0, b0},
  []*big.Float{b0, b0, b0, b0, b1, b0},
  []*big.Float{b0, b0, b0, b0, b0, b1},
  []*big.Float{b0, b0, b1, b0, b0, b0},
}
alpha, beta, gamma, z := R1CSToQAP(a, b, c)
fmt.Println(alpha)
fmt.Println(beta)
fmt.Println(gamma)
fmt.Println(z)
/*
out:
alpha: [[-5 9.166666666666666 -5 0.8333333333333334] [8 -11.333333333333332 5 -0.6666666666666666] [0 0 0 0] [-6 9.5 -4 0.5] [4 -7 3.5 -0.5] [-1 1.8333333333333333 -1 0.16666666666666666]]
beta: [[3 -5.166666666666667 2.5 -0.33333333333333337] [-2 5.166666666666667 -2.5 0.33333333333333337] [0 0 0 0] [0 0 0 0] [0 0 0 0] [0 0 0 0]]
gamma: [[0 0 0 0] [0 0 0 0] [-1 1.8333333333333333 -1 0.16666666666666666] [4 -4.333333333333333 1.5 -0.16666666666666666] [-6 9.5 -4 0.5] [4 -7 3.5 -0.5]]
z: [24 -50 35 -10 1]
*/
```

## Bn128
Implementation of the bn128 pairing.


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
