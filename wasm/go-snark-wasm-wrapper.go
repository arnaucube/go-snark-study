package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/wasm/utils"
)

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}

func registerCallbacks() {
	js.Global().Set("generateProofs", js.FuncOf(generateProofs))
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
