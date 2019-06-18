package proof

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arnaucube/go-snark/fields"
)

func TestR1CSToQAP(t *testing.T) {
	// new Finite Field
	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	assert.True(nil, ok)
	f := fields.NewFq(r)
	Utils.FqR = f
	// new Polynomial Field
	Utils.PF = fields.NewPF(f)

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
	alphas, betas, gammas, zx := R1CSToQAP(a, b, c)
	// fmt.Println(alphas)
	// fmt.Println(betas)
	// fmt.Println(gammas)
	// fmt.Print("Z(x): ")
	// fmt.Println(zx)

	w := []*big.Int{b1, b3, b35, b9, b27, b30}
	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	// fmt.Println(ax)
	// fmt.Println(bx)
	// fmt.Println(cx)
	// fmt.Println(px)

	hx := Utils.PF.DivisorPolynomial(px, zx)
	// fmt.Println(hx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := Utils.PF.Mul(hx, zx)
	assert.Equal(t, abc, hz)

}
