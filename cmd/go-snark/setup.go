package main

import (
	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/proof"
	"github.com/urfave/cli"
)

func setup(context *cli.Context) error {
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

	// R1CS to QAP
	alphas, betas, gammas, _ := proof.R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C)

	// calculate trusted setup
	setup, err := newSetup()
	if err != nil {
		return err
	}
	if err := setup.Init(cir, alphas, betas, gammas); err != nil {
		return err
	}

	// save setup
	return saveToFile(setupFileName, setup)
}
