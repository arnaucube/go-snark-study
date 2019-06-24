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
	"github.com/arnaucube/go-snark/groth16"
	"github.com/arnaucube/go-snark/r1csqap"
	"github.com/arnaucube/go-snark/wasm/utils"
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
	{
		Name:    "groth16",
		Aliases: []string{},
		Usage:   "use groth16 protocol",
		Subcommands: []cli.Command{
			{
				Name:    "trustedsetup",
				Aliases: []string{},
				Usage:   "generate trusted setup for a circuit",
				Action:  Groth16TrustedSetup,
			},
			{
				Name:    "genproofs",
				Aliases: []string{},
				Usage:   "generate the snark proofs",
				Action:  Groth16GenerateProofs,
			},
			{
				Name:    "verify",
				Aliases: []string{},
				Usage:   "verify the snark proofs",
				Action:  Groth16VerifyProofs,
			},
		},
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "go-snarks-cli"
	app.Version = "0.0.3-alpha"
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

	wasmFlag := false
	if context.Args().Get(1) == "wasm" {
		wasmFlag = true
	}

	// read circuit file
	circuitFile, err := os.Open(circuitPath)
	panicErr(err)

	// parse circuit code
	parser := circuitcompiler.NewParser(bufio.NewReader(circuitFile))
	circuit, err := parser.Parse()
	panicErr(err)
	fmt.Println("\ncircuit data:", circuit)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile("privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
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

	if wasmFlag {
		circuitString := utils.CircuitToString(*circuit)
		jsonData, err := json.Marshal(circuitString)
		panicErr(err)
		// store setup into file
		jsonFile, err := os.Create("compiledcircuitString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}

	// store px
	jsonData, err = json.Marshal(px)
	panicErr(err)
	// store setup into file
	jsonFile, err = os.Create("px.json")
	panicErr(err)
	defer jsonFile.Close()
	jsonFile.Write(jsonData)
	jsonFile.Close()
	fmt.Println("Px data written to ", jsonFile.Name())
	if wasmFlag {
		pxString := utils.ArrayBigIntToString(px)
		jsonData, err = json.Marshal(pxString)
		panicErr(err)
		// store setup into file
		jsonFile, err = os.Create("pxString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}

	return nil
}

func TrustedSetup(context *cli.Context) error {
	wasmFlag := false
	if context.Args().Get(0) == "wasm" {
		wasmFlag = true
	}
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile("privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)

	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(circuit.R1CS.A, circuit.R1CS.B, circuit.R1CS.C)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	// calculate trusted setup
	setup, err := snark.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas)
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
	if wasmFlag {
		tsetupString := utils.SetupToString(tsetup)
		jsonData, err := json.Marshal(tsetupString)
		panicErr(err)
		// store setup into file
		jsonFile, err := os.Create("trustedsetupString.json")
		panicErr(err)
		defer jsonFile.Close()
		jsonFile.Write(jsonData)
		jsonFile.Close()
	}
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

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile("privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)
	fmt.Println("witness", w)

	// flat code to R1CS
	a := circuit.R1CS.A
	b := circuit.R1CS.B
	c := circuit.R1CS.C
	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)
	_, _, _, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := snark.Utils.PF.DivisorPolynomial(px, trustedsetup.Pk.Z)

	fmt.Println(circuit)
	fmt.Println(trustedsetup.G1T)
	fmt.Println(hx)
	fmt.Println(w)
	proof, err := snark.GenerateProofs(circuit, trustedsetup, w, px)
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

	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)
	var publicSignals []*big.Int
	err = json.Unmarshal([]byte(string(publicInputsFile)), &publicSignals)
	panicErr(err)

	verified := snark.VerifyProof(circuit, trustedsetup, proof, publicSignals, true)
	if !verified {
		fmt.Println("ERROR: proofs not verified")
	} else {
		fmt.Println("Proofs verified")
	}
	return nil
}

func Groth16TrustedSetup(context *cli.Context) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile("privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)

	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)

	// R1CS to QAP
	alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(circuit.R1CS.A, circuit.R1CS.B, circuit.R1CS.C)
	fmt.Println("qap")
	fmt.Println(alphas)
	fmt.Println(betas)
	fmt.Println(gammas)

	// calculate trusted setup
	setup, err := groth16.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas)
	panicErr(err)
	fmt.Println("\nt:", setup.Toxic.T)

	// remove setup.Toxic
	var tsetup groth16.Setup
	tsetup.Pk = setup.Pk
	tsetup.Vk = setup.Vk

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

func Groth16GenerateProofs(context *cli.Context) error {
	// open compiledcircuit.json
	compiledcircuitFile, err := ioutil.ReadFile("compiledcircuit.json")
	panicErr(err)
	var circuit circuitcompiler.Circuit
	json.Unmarshal([]byte(string(compiledcircuitFile)), &circuit)
	panicErr(err)

	// open trustedsetup.json
	trustedsetupFile, err := ioutil.ReadFile("trustedsetup.json")
	panicErr(err)
	var trustedsetup groth16.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read privateInputs file
	privateInputsFile, err := ioutil.ReadFile("privateInputs.json")
	panicErr(err)
	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)
	// parse inputs from inputsFile
	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(string(privateInputsFile)), &inputs.Private)
	panicErr(err)
	err = json.Unmarshal([]byte(string(publicInputsFile)), &inputs.Public)
	panicErr(err)

	// calculate wittness
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)
	panicErr(err)
	fmt.Println("witness", w)

	// flat code to R1CS
	a := circuit.R1CS.A
	b := circuit.R1CS.B
	c := circuit.R1CS.C
	// R1CS to QAP
	alphas, betas, gammas, _ := groth16.Utils.PF.R1CSToQAP(a, b, c)
	_, _, _, px := groth16.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)
	hx := groth16.Utils.PF.DivisorPolynomial(px, trustedsetup.Pk.Z)

	fmt.Println(circuit)
	fmt.Println(trustedsetup.Pk.PowersTauDelta)
	fmt.Println(hx)
	fmt.Println(w)
	proof, err := groth16.GenerateProofs(circuit, trustedsetup, w, px)
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

func Groth16VerifyProofs(context *cli.Context) error {
	// open proofs.json
	proofsFile, err := ioutil.ReadFile("proofs.json")
	panicErr(err)
	var proof groth16.Proof
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
	var trustedsetup groth16.Setup
	json.Unmarshal([]byte(string(trustedsetupFile)), &trustedsetup)
	panicErr(err)

	// read publicInputs file
	publicInputsFile, err := ioutil.ReadFile("publicInputs.json")
	panicErr(err)
	var publicSignals []*big.Int
	err = json.Unmarshal([]byte(string(publicInputsFile)), &publicSignals)
	panicErr(err)

	verified := groth16.VerifyProof(circuit, trustedsetup, proof, publicSignals, true)
	if !verified {
		fmt.Println("ERROR: proofs not verified")
	} else {
		fmt.Println("Proofs verified")
	}
	return nil
}
