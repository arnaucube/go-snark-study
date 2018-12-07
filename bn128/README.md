## Bn128
[![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/bn128?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/bn128) bn128
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

- [x] Fq, Fq2, Fq6, Fq12 operations
- [x] G1, G2 operations
- [x] preparePairing
- [x] PreComupteG1, PreComupteG2
- [x] DoubleStep, AddStep
- [x] MillerLoop
- [x] Pairing

### Installation
```
go get github.com/arnaucube/bn128
```

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

#### Test
```
go test -v
```

##### Internal operations more deeply

First let's assume that we have these three basic functions to convert integer compositions to big integer compositions:
```go
func iToBig(a int) *big.Int {
	return big.NewInt(int64(a))
}

func iiToBig(a, b int) [2]*big.Int {
	return [2]*big.Int{iToBig(a), iToBig(b)}
}

func iiiToBig(a, b int) [2]*big.Int {
	return [2]*big.Int{iToBig(a), iToBig(b)}
}
```
- Finite Fields (1, 2, 6, 12) operations
```go
// new finite field of order 1
fq1 := NewFq(iToBig(7))

// basic operations of finite field 1
res := fq1.Add(iToBig(4), iToBig(4))
res = fq1.Double(iToBig(5))
res = fq1.Sub(iToBig(5), iToBig(7))
res = fq1.Neg(iToBig(5))
res = fq1.Mul(iToBig(5), iToBig(11))
res = fq1.Inverse(iToBig(4))
res = fq1.Square(iToBig(5))

// new finite field of order 2
nonResidueFq2str := "-1" // i/j
nonResidueFq2, ok := new(big.Int).SetString(nonResidueFq2str, 10)
fq2 := Fq2{fq1, nonResidueFq2}
nonResidueFq6 := iiToBig(9, 1)

// basic operations of finite field of order 2
res := fq2.Add(iiToBig(4, 4), iiToBig(3, 4))
res = fq2.Double(iiToBig(5, 3))
res = fq2.Sub(iiToBig(5, 3), iiToBig(7, 2))
res = fq2.Neg(iiToBig(4, 4))
res = fq2.Mul(iiToBig(4, 4), iiToBig(3, 4))
res = fq2.Inverse(iiToBig(4, 4))
res = fq2.Div(iiToBig(4, 4), iiToBig(3, 4))
res = fq2.Square(iiToBig(4, 4))


// new finite field of order 6
nonResidueFq6 := iiToBig(9, 1) // TODO
fq6 := Fq6{fq2, nonResidueFq6}

// define two new values of Finite Field 6, in order to be able to perform the operations
a := [3][2]*big.Int{
	iiToBig(1, 2),
	iiToBig(3, 4),
	iiToBig(5, 6)}
b := [3][2]*big.Int{
	iiToBig(12, 11),
	iiToBig(10, 9),
	iiToBig(8, 7)}

// basic operations of finite field order 6
res := fq6.Add(a, b)
res = fq6.Sub(a, b)
res = fq6.Mul(a, b)
divRes := fq6.Div(mulRes, b)


// new finite field of order 12
q, ok := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10) // i
if !ok {
	fmt.Println("error parsing string to big integer")
}

fq1 := NewFq(q)
nonResidueFq2, ok := new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208582", 10) // i
assert.True(t, ok)
nonResidueFq6 := iiToBig(9, 1)

fq2 := Fq2{fq1, nonResidueFq2}
fq6 := Fq6{fq2, nonResidueFq6}
fq12 := Fq12{fq6, fq2, nonResidueFq6}

```

- G1 operations
```go
bn128, err := NewBn128()
assert.Nil(t, err)

r1 := big.NewInt(int64(33))
r2 := big.NewInt(int64(44))

gr1 := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(r1))
gr2 := bn128.G1.MulScalar(bn128.G1.G, bn128.Fq1.Copy(r2))

grsum1 := bn128.G1.Add(gr1, gr2)
r1r2 := bn128.Fq1.Add(r1, r2)
grsum2 := bn128.G1.MulScalar(bn128.G1.G, r1r2)

a := bn128.G1.Affine(grsum1)
b := bn128.G1.Affine(grsum2)
assert.Equal(t, a, b)
assert.Equal(t, "0x2f978c0ab89ebaa576866706b14787f360c4d6c3869efe5a72f7c3651a72ff00", utils.BytesToHex(a[0].Bytes()))
assert.Equal(t, "0x12e4ba7f0edca8b4fa668fe153aebd908d322dc26ad964d4cd314795844b62b2", utils.BytesToHex(a[1].Bytes()))
```

- G2 operations
```go
bn128, err := NewBn128()
assert.Nil(t, err)

r1 := big.NewInt(int64(33))
r2 := big.NewInt(int64(44))

gr1 := bn128.G2.MulScalar(bn128.G2.G, bn128.Fq1.Copy(r1))
gr2 := bn128.G2.MulScalar(bn128.G2.G, bn128.Fq1.Copy(r2))

grsum1 := bn128.G2.Add(gr1, gr2)
r1r2 := bn128.Fq1.Add(r1, r2)
grsum2 := bn128.G2.MulScalar(bn128.G2.G, r1r2)

a := bn128.G2.Affine(grsum1)
b := bn128.G2.Affine(grsum2)
assert.Equal(t, a, b)
```
