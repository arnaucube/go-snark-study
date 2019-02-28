package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	snark "github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/urfave/cli"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

var commands = []cli.Command{
	{
		Name:    "compile",
		Aliases: []string{},
		Usage:   "compile a circuit",
		Action:  CompileCircuit,
	},
	{
		Name:    "trustedsetup",
		Aliases: []string{},
		Usage:   "generate trusted setup for a circuit",
		Action:  TrustedSetup,
	},
	{
		Name:    "genproofs",
		Aliases: []string{},
		Usage:   "generate the snark proofs",
		Action:  GenerateProofs,
	},
	{
		Name:    "verify",
		Aliases: []string{},
		Usage:   "verify the snark proofs",
		Action:  VerifyProofs,
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "go-snarks-cli"
	app.Version = "0.1.0-alpha"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "config"},
	}
	app.Commands = commands

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func CompileCircuit(context *cli.Context) error {
	fmt.Println("cli")

	circuitPath := context.Args().Get(0)

	// read circuit file
	circuitFile, err := os.Open(circuitPath)
	panicErr(err)

	// parse circuit code
	parser := circuitcompiler.NewParser(bufio.NewReader(circuitFile))
	circuit, err := parser.Parse()
	panicErr(err)
	fmt.Println("\ncircuit data:", circuit)

	// read inputs file
	inputsFile, err := ioutil.ReadFile("inputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	// var inputs []*big.Int
	var inputs circuitcompiler.Inputs
	json.Unmarshal([]byte(string(inputsFile)), &inputs)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private)
	panicErr(err)
	fmt.Println("\nwitness", w)

	// flat code to R1CS
	fmt.Println("\ngenerating R1CS from flat code")
	a, b, c := circuit.GenerateR1CS()
	fmt.Println("\nR1CS:")
	fmt.Println("a:", a)
	fmt.Println("b:", b)
	fmt.Println("c:", c)

	// R1CS to QAP
	alphas, betas, gammas, zx := snark.Utils.PF.R1CSToQAP(a, b, c)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	ax, bx, cx, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

	hx := snark.Utils.PF.DivisorPolynomial(px, zx)

	// hx==px/zx so px==hx*zx
	// assert.Equal(t, px, snark.Utils.PF.Mul(hx, zx))
	if !r1csqap.BigArraysEqual(px, snark.Utils.PF.Mul(hx, zx)) {
		panic(errors.New("px != hx*zx"))
	}

	// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
	abc := snark.Utils.PF.Sub(snark.Utils.PF.Mul(ax, bx), cx)
	// assert.Equal(t, abc, px)
	if !r1csqap.BigArraysEqual(abc, px) {
		panic(errors.New("abc != px"))
	}
	hz := snark.Utils.PF.Mul(hx, zx)
	if !r1csqap.BigArraysEqual(abc, hz) {
		panic(errors.New("abc != hz"))
	}
	// assert.Equal(t, abc, hz)

	div, rem := snark.Utils.PF.Div(px, zx)
	if !r1csqap.BigArraysEqual(hx, div) {
		panic(errors.New("hx != div"))
	}
	// assert.Equal(t, hx, div)
	// assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))
	for _, r := range rem {
		if !bytes.Equal(r.Bytes(), big.NewInt(int64(0)).Bytes()) {
			panic(errors.New("error:error:  px/zx rem not equal to zeros"))
		}
	}

	// store circuit to json
	jsonData, err := json.Marshal(circuit)
	panicErr(err)
	// store setup into file
	jsonFile, err := os.Create("compiledcircuit.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Compiled Circuit data written to ", jsonFile.Name())

	return nil
}

func TrustedSetup(context *cli.Context) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// read inputs file
	inputsFile, err := ioutil.ReadFile("inputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	// var inputs []*big.Int
	var inputs circuitcompiler.Inputs
	json.Unmarshal([]byte(string(inputsFile)), &inputs)
	panicErr(err)
	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private)
	panicErr(err)

	// R1CS to QAP
	alphas, betas, gammas, zx := snark.Utils.PF.R1CSToQAP(circuit.R1CS.A, circuit.R1CS.B, circuit.R1CS.C)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	// calculate trusted setup
	setup, err := snark.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas, zx)
	panicErr(err)
	fmt.Println("\nt:", setup.Toxic.T)

	// remove setup.Toxic
	var tsetup snark.Setup
	tsetup.Pk = setup.Pk
	tsetup.Vk = setup.Vk
	tsetup.G1T = setup.G1T
	tsetup.G2T = setup.G2T

	// store setup to json
	jsonData, err := json.Marshal(tsetup)
	panicErr(err)
	// store setup into file
	jsonFile, err := os.Create("trustedsetup.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Trusted Setup data written to ", jsonFile.Name())
	return nil
}

func GenerateProofs(context *cli.Context) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile("trustedsetup.json")
	panicErr(err)
	var trustedsetup snark.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read inputs file
	inputsFile, err := ioutil.ReadFile("inputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	// var inputs []*big.Int
	var inputs circuitcompiler.Inputs
	json.Unmarshal([]byte(string(inputsFile)), &inputs)
	panicErr(err)
	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private)
	panicErr(err)
	fmt.Println("\nwitness", w)

	// flat code to R1CS
	// a, b, c := circuit.GenerateR1CS()
	a := circuit.R1CS.A
	b := circuit.R1CS.B
	c := circuit.R1CS.C
	// R1CS to QAP
	alphas, betas, gammas, zx := snark.Utils.PF.R1CSToQAP(a, b, c)
	_, _, _, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := snark.Utils.PF.DivisorPolynomial(px, zx)

	fmt.Println(circuit)
	fmt.Println(trustedsetup.G1T)
	fmt.Println(hx)
	fmt.Println(w)
	proof, err := snark.GenerateProofs(circuit, trustedsetup, hx, w)
	panicErr(err)

	fmt.Println("\n proofs:")
	fmt.Println(proof)

	// store proofs to json
	jsonData, err := json.Marshal(proof)
	panicErr(err)
	// store proof into file
	jsonFile, err := os.Create("proofs.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Proofs data written to ", jsonFile.Name())
	return nil
}

func VerifyProofs(context *cli.Context) error {
	// open proofs.json
	proofsFile, err := ioutil.ReadFile("proofs.json")
	panicErr(err)
	var proof snark.Proof
	json.Unmarshal([]byte(string(proofsFile)), &proof)
	panicErr(err)

	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile("trustedsetup.json")
	panicErr(err)
	var trustedsetup snark.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// TODO read publicSignals from file
	verified := snark.VerifyProof(circuit, trustedsetup, proof, publicSignals, true)
	if !verified {
		fmt.Println("ERROR: proofs not verified")
	} else {
		fmt.Println("Proofs verified")
	}
	return nil
}
