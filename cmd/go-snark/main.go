package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"github.com/arnaucube/go-snark/proof"
)

const (
	compiledFileName = "compiled.json"
	setupFileName    = "setup.json"
	privateFileName  = "private.json"
	publicFileName   = "public.json"
	proofFileName    = "proof.json"
)

const (
	proofSystemPinocchio = iota
	proofSystemGroth16
)

var proofSystem int

var commands = []cli.Command{
	{
		Name:    "compile",
		Aliases: []string{},
		Usage:   "compile a circuit",
		Action:  compile,
	},
	{
		Name:    "test",
		Aliases: []string{},
		Usage:   "test a circuit",
		Action:  test,
	},
	{
		Name:    "setup",
		Aliases: []string{},
		Usage:   "generate trusted setup for a circuit",
		Action:  setup,
	},
	{
		Name:    "generate",
		Aliases: []string{},
		Usage:   "generate the snark proofs",
		Action:  generate,
	},
	{
		Name:    "verify",
		Aliases: []string{},
		Usage:   "verify the snark proofs",
		Action:  verify,
	},
}

func initProofSystem() error {
	switch p := os.Getenv("PROOF_SYSTEM"); p {
	case "", "PINOCCHIO":
		proofSystem = proofSystemPinocchio
	case "GROTH16":
		proofSystem = proofSystemGroth16
	default:
		return fmt.Errorf("proof system not supported: %v", p)
	}
	return nil
}

func newSetup() (proof.Setup, error) {
	var s proof.Setup
	switch proofSystem {
	case proofSystemPinocchio:
		s = &proof.PinocchioSetup{}
	case proofSystemGroth16:
		s = &proof.Groth16Setup{}
	default:
		return nil, fmt.Errorf("proof system not supported: %v", proofSystem)
	}
	return s, nil
}

func newProof() (proof.Proof, error) {
	var p proof.Proof
	switch proofSystem {
	case proofSystemPinocchio:
		p = &proof.PinocchioProof{}
	case proofSystemGroth16:
		p = &proof.Groth16Proof{}
	default:
		return nil, fmt.Errorf("proof system not supported: %v", proofSystem)
	}
	return p, nil
}

func main() {
	if err := initProofSystem(); err != nil {
		panic(err)
	}
	app := cli.NewApp()
	app.Name = "go-snark"
	app.Version = "0.0.3-alpha"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "config"},
	}
	app.Commands = commands
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
