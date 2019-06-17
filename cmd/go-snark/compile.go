package main

import (
	"bytes"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/urfave/cli"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/proof"
	"github.com/arnaucube/go-snark/r1csqap"
)

func compile(context *cli.Context) error {
	circuitPath := context.Args().Get(0)

	// load compiled
	circuitFile, err := os.Open(circuitPath)
	if err != nil {
		return err
	}
	parser := circuit.NewParser(circuitFile)
	cir, err := parser.Parse()
	if err != nil {
		return err
	}
	log.Printf("circuit: %v\n", cir)

	// load inputs
	var inputs circuit.Inputs
	if err := loadFromFile(privateFileName, &inputs.Private); err != nil {
		return err
	}
	if err := loadFromFile(publicFileName, &inputs.Public); err != nil {
		return err
	}

	// calculate witness
	w, err := cir.CalculateWitness(inputs.Private, inputs.Public)
	if err != nil {
		return err
	}
	log.Printf("w: %v\n", w)

	// flat code to R1CS
	a, b, c := cir.GenerateR1CS()
	log.Printf("a: %v\n", a)
	log.Printf("b: %v\n", b)
	log.Printf("c: %v\n", c)

	// R1CS to QAP
	alphas, betas, gammas, zx := proof.Utils.PF.R1CSToQAP(a, b, c)
	log.Printf("alphas: %v\n", alphas)
	log.Printf("betas: %v\n", betas)
	log.Printf("gammas: %v\n", gammas)
	log.Printf("zx: %v\n", zx)

	ax, bx, cx, px := proof.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := proof.Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	// assert.Equal(t, px, snark.Utils.PF.Mul(hx, zx))
	if !r1csqap.BigArraysEqual(px, proof.Utils.PF.Mul(hx, zx)) {
		return fmt.Errorf("px != hx*zx")
	}

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := proof.Utils.PF.Sub(proof.Utils.PF.Mul(ax, bx), cx)
	// assert.Equal(t, abc, px)
	if !r1csqap.BigArraysEqual(abc, px) {
		return fmt.Errorf("abc != px")
	}

	hz := proof.Utils.PF.Mul(hx, zx)
	if !r1csqap.BigArraysEqual(abc, hz) {
		return fmt.Errorf("abc != hz")
	}
	// assert.Equal(t, abc, hz)

	div, rem := proof.Utils.PF.Div(px, zx)
	if !r1csqap.BigArraysEqual(hx, div) {
		return fmt.Errorf("hx != div")
	}
	// assert.Equal(t, hx, div)
	// assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))
	for _, r := range rem {
		if !bytes.Equal(r.Bytes(), big.NewInt(int64(0)).Bytes()) {
			return fmt.Errorf("error:error:  px/zx rem not equal to zeros")
		}
	}

	// save circuit
	return saveToFile(compiledFileName, cir)
}
