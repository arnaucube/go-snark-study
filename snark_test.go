package snark

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/stretchr/testify/assert"
)

func TestZkFromFlatCircuitCode(t *testing.T) {

	// compile circuit and get the R1CS
	flatCode := `
	func test(x):
		aux = x*x
		y = aux*x
		z = x + y
		out = z + 5
	`
	fmt.Print("\nflat code of the circuit:")
	fmt.Println(flatCode)

	// parse the code
	parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
	circuit, err := parser.Parse()
	assert.Nil(t, err)
	fmt.Println("\ncircuit data:", circuit)

	b3 := big.NewInt(int64(3))
	inputs := []*big.Int{b3}
	// wittness
	w := circuit.CalculateWitness(inputs)
	fmt.Println("\nwitness", w)

	// flat code to R1CS
	fmt.Println("\ngenerating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	alphas, betas, gammas, zx := Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := Utils.PF.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	div, rem := Utils.PF.Div(px, zx)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas, zx)
	assert.Nil(t, err)
	fmt.Println("\nt:", setup.Toxic.T)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(*circuit, setup, hx, w)
	assert.Nil(t, err)

	fmt.Println("\n proofs:")
	fmt.Println(proof)
	before := time.Now()
	assert.True(t, VerifyProof(*circuit, setup, proof))
	fmt.Println("verify proof time elapsed:", time.Since(before))
}

func TestZkFromHardcodedR1CS(t *testing.T) {

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
	alphas, betas, gammas, zx := Utils.PF.R1CSToQAP(a, b, c)

	// wittness = 1, 3, 35, 9, 27, 30
	w := []*big.Int{b1, b3, b35, b9, b27, b30}
	circuit := circuitcompiler.Circuit{
		NVars:    6,
		NPublic:  0,
		NSignals: len(w),
	}
	ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	assert.Equal(t, px, Utils.PF.Mul(hx, zx))

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := Utils.PF.Sub(Utils.PF.Mul(ax, bx), cx)
	assert.Equal(t, abc, px)
	hz := Utils.PF.Mul(hx, zx)
	assert.Equal(t, abc, hz)

	div, rem := Utils.PF.Div(px, zx)
	assert.Equal(t, hx, div)
	assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

	// calculate trusted setup
	setup, err := GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas, zx)
	assert.Nil(t, err)
	fmt.Println("t", setup.Toxic.T)

	// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
	proof, err := GenerateProofs(circuit, setup, hx, w)
	assert.Nil(t, err)

	assert.True(t, VerifyProof(circuit, setup, proof))
}
