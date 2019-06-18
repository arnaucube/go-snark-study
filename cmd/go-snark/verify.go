package main

import (
	"fmt"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/urfave/cli"
)

func verify(context *cli.Context) error {
	// load circuit
	cir := &circuit.Circuit{}
	if err := loadFromFile(compiledFileName, cir); err != nil {
		return err
	}

	// load inputs
	var inputs circuit.Inputs
	if err := loadFromFile(publicFileName, &inputs.Public); err != nil {
		return err
	}

	// load setup
	setup, err := newSetup()
	if err != nil {
		return err
	}
	if err := loadFromFile(setupFileName, setup); err != nil {
		return err
	}

	// load proof
	proof, err := newProof()
	if err != nil {
		return err
	}
	if err := loadFromFile(proofFileName, proof); err != nil {
		return err
	}

	// verify proof
	ok, err := setup.Verify(proof, inputs.Public)
	if err != nil {
		return err
	}
	if ok {
		fmt.Println("OK")
	} else {
		fmt.Println("KO")
	}
	return nil
}
