package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/urfave/cli"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/proof"
)

func test(context *cli.Context) error {
	// load circuit
	cir := &circuit.Circuit{}
	if err := loadFromFile(compiledFileName, cir); err != nil {
		return err
	}

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

	// R1CS to QAP
	alphas, betas, gammas, zx := proof.R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C,
	)

	// px == ax * bx - cx
	ax, bx, cx, px := proof.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	// hx == px / zx
	hx := proof.Utils.PF.DivisorPolynomial(px, zx)
	if !fields.BigArraysEqual(px, proof.Utils.PF.Mul(hx, zx)) {
		return fmt.Errorf("px != hx * zx")
	}

	// ax * bx - cx == px
	abc := proof.Utils.PF.Sub(proof.Utils.PF.Mul(ax, bx), cx)
	if !fields.BigArraysEqual(abc, px) {
		return fmt.Errorf("ax * bx - cx != px")
	}

	// hx * zx == ax * bx - cx
	hz := proof.Utils.PF.Mul(hx, zx)
	if !fields.BigArraysEqual(hz, abc) {
		return fmt.Errorf("hx * zx != ax * bx - cx")
	}

	// dx == px / zx + rx
	dx, rx := proof.Utils.PF.Div(px, zx)
	if !fields.BigArraysEqual(dx, hx) {
		return fmt.Errorf("dx != hx")
	}
	for _, r := range rx {
		if !bytes.Equal(r.Bytes(), big.NewInt(int64(0)).Bytes()) {
			return fmt.Errorf("rx != 0")
		}
	}

	return nil
}
