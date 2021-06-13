## R1CS to Quadratic Arithmetic Program
[![GoDoc](https://godoc.org/github.com/arnaucube/go-snark-study/r1csqap?status.svg)](https://godoc.org/github.com/arnaucube/go-snark-study/r1csqap) R1CS to QAP
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
