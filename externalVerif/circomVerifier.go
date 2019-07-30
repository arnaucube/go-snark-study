package externalVerif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/arnaucube/go-snark/groth16"
	"github.com/arnaucube/go-snark/utils"
)

type CircomProof struct {
	PiA [3]string    `json:"pi_a"`
	PiB [3][2]string `json:"pi_b"`
	PiC [3]string    `json:"pi_c"`
}
type CircomVk struct {
	IC          [][3]string     `json:"IC"`
	Alpha1      [3]string       `json:"vk_alfa_1"`
	Beta2       [3][2]string    `json:"vk_beta_2"`
	Gamma2      [3][2]string    `json:"vk_gamma_2"`
	Delta2      [3][2]string    `json:"vk_delta_2"`
	AlphaBeta12 [2][3][2]string `json:"vk_alpfabeta_12"` // not really used, for the moment in go-snarks calculed in verification time
}

func VerifyFromCircom(vkPath, proofPath, publicSignalsPath string) (bool, error) {
	// open verification_key.json
	vkFile, err := ioutil.ReadFile(vkPath)
	if err != nil {
		return false, err
	}
	var circomVk CircomVk
	json.Unmarshal([]byte(string(vkFile)), &circomVk)
	if err != nil {
		return false, err
	}

	var strVk utils.GrothVkString
	strVk.IC = circomVk.IC
	strVk.G1.Alpha = circomVk.Alpha1
	strVk.G2.Beta = circomVk.Beta2
	strVk.G2.Gamma = circomVk.Gamma2
	strVk.G2.Delta = circomVk.Delta2
	vk, err := utils.GrothVkFromString(strVk)
	if err != nil {
		return false, err
	}
	fmt.Println("vk parsed:", vk)

	// open proof.json
	proofsFile, err := ioutil.ReadFile(proofPath)
	if err != nil {
		return false, err
	}
	var circomProof CircomProof
	json.Unmarshal([]byte(string(proofsFile)), &circomProof)
	if err != nil {
		return false, err
	}

	strProof := utils.GrothProofString{
		PiA: circomProof.PiA,
		PiB: circomProof.PiB,
		PiC: circomProof.PiC,
	}
	proof, err := utils.GrothProofFromString(strProof)
	if err != nil {
		return false, err
	}
	fmt.Println("proof parsed:", proof)

	// open public.json
	publicFile, err := ioutil.ReadFile(publicSignalsPath)
	if err != nil {
		return false, err
	}
	var publicStr []string
	json.Unmarshal([]byte(string(publicFile)), &publicStr)
	if err != nil {
		return false, err
	}
	publicSignals, err := utils.ArrayStringToBigInt(publicStr)
	if err != nil {
		return false, err
	}
	fmt.Println("publicSignals parsed:", publicSignals)

	verified := groth16.VerifyProof(vk, proof, publicSignals, true)
	return verified, nil
}
