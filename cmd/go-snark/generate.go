package main

import (
	"log"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/proof"
	"github.com/urfave/cli"
)

func generate(context *cli.Context) error {
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

	// load setup
	setup, err := newSetup()
	if err != nil {
		return err
	}
	if err := loadFromFile(setupFileName, setup); err != nil {
		return err
	}
	log.Printf("setup: %v\n", setup)

	// R1CS to QAP
	alphas, betas, gammas, _ := proof.Utils.PF.R1CSToQAP(
		cir.R1CS.A,
		cir.R1CS.B,
		cir.R1CS.C)
	_, _, _, px := proof.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := proof.Utils.PF.DivisorPolynomial(px, setup.Z())
	log.Printf("hx: %v\n", hx)

	// generate proof
	proof, err := setup.Generate(cir, w, px)
	if err != nil {
		return err
	}
	log.Printf("proof: %v\n", proof)

	// save proof
	return saveToFile(proofFileName, proof)
}
