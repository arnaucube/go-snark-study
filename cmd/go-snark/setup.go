package main

import (
	"log"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/proof"
	"github.com/urfave/cli"
)

func setup(context *cli.Context) error {
	// load circuit
	var cir circuit.Circuit
	if err := loadFromFile(compiledFileName, &cir); err != nil {
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

	// R1CS to QAP
	alphas, betas, gammas, _ := proof.Utils.PF.R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C)

	// calculate trusted setup
	setup, err := newSetup()
	if err != nil {
		return err
	}
	if err := setup.Init(len(w), cir, alphas, betas, gammas); err != nil {
		return err
	}
	log.Printf("setup: %v\n", setup)

	// save setup
	return saveToFile(setupFileName, setup)
}
