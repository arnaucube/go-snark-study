package main

import (
	"encoding/json"
	"math/big"
	"syscall/js"

	"github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/wasm/utils"
)

func main() {
	c := make(chan struct{}, 0)
	println("WASM Go Initialized")
	registerCallbacks()
	<-c
}

func registerCallbacks() {
	js.Global().Set("generateProofs", js.FuncOf(generateProofs))
	js.Global().Set("verifyProofs", js.FuncOf(verifyProofs))
}

func generateProofs(this js.Value, i []js.Value) interface{} {
	var circuitStr utils.CircuitString
	err := json.Unmarshal([]byte(i[0].String()), &circuitStr)
	if err != nil {
		println(i[0].String())
		println("error parsing circuit from stringified json")
	}
	circuit, err := utils.CircuitFromString(circuitStr)
	if err != nil {
		println("error " + err.Error())
	}
	sj, err := json.Marshal(circuit)
	if err != nil {
		println("error " + err.Error())
	}
	println("circuit", string(sj))

	var setupStr utils.SetupString
	println(i[1].String())
	err = json.Unmarshal([]byte(i[1].String()), &setupStr)
	if err != nil {
		println("error parsing setup from stringified json")
	}
	setup, err := utils.SetupFromString(setupStr)
	if err != nil {
		println("error " + err.Error())
	}
	sj, err = json.Marshal(setup)
	if err != nil {
		println("error " + err.Error())
	}
	println("set", string(sj))

	var pxStr []string
	err = json.Unmarshal([]byte(i[2].String()), &pxStr)
	if err != nil {
		println("error parsing pxStr from stringified json")
	}
	px, err := utils.ArrayStringToBigInt(pxStr)
	if err != nil {
		println(err.Error())
	}
	sj, err = json.Marshal(px)
	if err != nil {
		println("error " + err.Error())
	}
	println("px", string(sj))

	var inputs circuitcompiler.Inputs
	err = json.Unmarshal([]byte(i[3].String()), &inputs)
	if err != nil {
		println("error parsing inputs from stringified json")
	}
	w, err := circuit.CalculateWitness(inputs.Private, inputs.Public)

	proof, err := snark.GenerateProofs(circuit, setup, w, px)
	if err != nil {
		println("error generating proof", err)
	}
	proofString := utils.ProofToString(proof)
	proofJson, err := json.Marshal(proofString)
	if err != nil {
		println("error marshal proof to json", err)
	}
	println("proofJson", string(proofJson))
	return js.ValueOf(string(proofJson))
}

func verifyProofs(this js.Value, i []js.Value) interface{} {
	var circuitStr utils.CircuitString
	err := json.Unmarshal([]byte(i[0].String()), &circuitStr)
	if err != nil {
		println(i[0].String())
		println("error parsing circuit from stringified json")
	}
	circuit, err := utils.CircuitFromString(circuitStr)
	if err != nil {
		println("error " + err.Error())
	}

	var setupStr utils.SetupString
	println(i[1].String())
	err = json.Unmarshal([]byte(i[1].String()), &setupStr)
	if err != nil {
		println("error parsing setup from stringified json")
	}
	setup, err := utils.SetupFromString(setupStr)
	if err != nil {
		println("error " + err.Error())
	}

	var proofStr utils.ProofString
	err = json.Unmarshal([]byte(i[2].String()), &proofStr)
	if err != nil {
		println(i[0].String())
		println("error parsing proof from stringified json")
	}
	proof, err := utils.ProofFromString(proofStr)
	if err != nil {
		println("error " + err.Error())
	}

	var publicInputs []*big.Int
	err = json.Unmarshal([]byte(i[3].String()), &publicInputs)
	if err != nil {
		println(i[0].String())
		println("error parsing publicInputs from stringified json")
	}

	verified := snark.VerifyProof(circuit, setup, proof, publicInputs, false)
	if err != nil {
		println("error verifiyng proof", err)
	}
	verifiedJson, err := json.Marshal(verified)
	if err != nil {
		println("error marshal verified to json", err)
	}
	println("verifiedJson", string(verifiedJson))
	return js.ValueOf(string(verifiedJson))
}
