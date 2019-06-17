package main

import (
	"fmt"
	"log"

	"github.com/arnaucube/go-snark/circuit"
	"github.com/urfave/cli"
)

func verify(context *cli.Context) error {
	// load circuit
	var cir circuit.Circuit
	if err := loadFromFile(compiledFileName, &cir); err != nil {
		return err
	}
	log.Printf("circuit: %v\n", cir)

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
	log.Printf("setup: %v\n", setup)

	// load proof
	proof, err := newProof()
	if err != nil {
		return err
	}
	if err := loadFromFile(proofFileName, proof); err != nil {
		return err
	}

	// verify proof
	if ok := setup.Verify(cir, proof, inputs.Public, true); !ok {
		return fmt.Errorf("verif KO")
	}
	log.Printf("verif OK\n")
	return nil
}
