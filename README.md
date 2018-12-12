# go-snark [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/go-snark)](https://goreportcard.com/report/github.com/arnaucube/go-snark)

zkSNARK library implementation in Go



`Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture`, Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, Madars Virza https://eprint.iacr.org/2013/879.pdf

### Usage
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark?status.svg)](https://godoc.org/github.com/arnaucube/go-snark) zkSnark
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/bn128?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/bn128) bn128 (more details: https://github.com/arnaucube/go-snark/tree/master/bn128)
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/fields?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/fields) Finite Fields operations
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/r1csqap?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/r1csqap) R1CS to QAP (more details: https://github.com/arnaucube/go-snark/tree/master/r1csqap)

Example:
```go
bn, err := bn128.NewBn128()
assert.Nil(t, err)

// new Finite Field
f := fields.NewFq(bn.R)

// new Polynomial Field
pf := r1csqap.NewPolynomialField(f)

/*
suppose that we have the following variables with *big.Int elements:
a = [[0 1 0 0 0 0] [0 0 0 1 0 0] [0 1 0 0 1 0] [5 0 0 0 0 1]]
b = [[0 1 0 0 0 0] [0 1 0 0 0 0] [1 0 0 0 0 0] [1 0 0 0 0 0]]
c = [[0 0 0 1 0 0] [0 0 0 0 1 0] [0 0 0 0 0 1] [0 0 1 0 0 0]]

w = [1, 3, 35, 9, 27, 30]
*/

alphas, betas, gammas, zx := pf.R1CSToQAP(a, b, c)

ax, bx, cx, px := pf.CombinePolynomials(w, alphas, betas, gammas)

hx := pf.DivisorPolinomial(px, zx)

// hx==px/zx so px==hx*zx
assert.Equal(t, px, pf.Mul(hx, zx))

// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
abc := pf.Sub(pf.Mul(ax, bx), cx)
assert.Equal(t, abc, px)
hz := pf.Mul(hx, zx)
assert.Equal(t, abc, hz)

// calculate trusted setup
setup, err := GenerateTrustedSetup(bn, len(ax))
assert.Nil(t, err)
fmt.Println("trusted setup:")
fmt.Println(setup.G1T)
fmt.Println(setup.G2T)

// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
proof, err := GenerateProofs(bn, f, setup, ax, bx, cx, hx, zx)
assert.Nil(t, err)


// verify the proofs with the bn128 pairing
verified := VerifyProof(bn, publicSetup, proof)
assert.True(t, verified)
```

### Test
```
go test ./... -v
```

---

## Caution
Not finished, work in progress (implementing this in my free time to understand it better, so I don't have much time).

Thanks to [@jbaylina](https://github.com/jbaylina), [@bellesmarta](https://github.com/bellesmarta), [@adriamb](https://github.com/adriamb) for their explanations that helped to understand this a little bit.
