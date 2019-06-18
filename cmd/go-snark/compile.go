package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/arnaucube/go-snark/circuit"
)

func compile(context *cli.Context) error {
	circuitPath := context.Args().Get(0)

	// load circuit
	circuitFile, err := os.Open(circuitPath)
	if err != nil {
		return err
	}
	parser := circuit.NewParser(circuitFile)

	// parse circuit
	cir, err := parser.Parse()
	if err != nil {
		return err
	}

	// flat code to R1CS
	cir.GenerateR1CS()

	// save circuit
	return saveToFile(compiledFileName, cir)
}
