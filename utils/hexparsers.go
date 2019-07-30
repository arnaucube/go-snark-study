package utils

import (
	"errors"
	"fmt"
	"math/big"

	snark "github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/groth16"
)

// []*big.Int
func ArrayBigIntToHex(b []*big.Int) []string {
	var o []string
	for i := 0; i < len(b); i++ {
		o = append(o, fmt.Sprintf("%x", b[i]))
	}
	return o
}
func ArrayHexToBigInt(s []string) ([]*big.Int, error) {
	var px []*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 16)
		if !ok {
			return nil, errors.New("error parsing px from pxHex")
		}
		px = append(px, param)
	}
	return px, nil
}

// [3]*big.Int
func Hex3ToBigInt(s [3]string) ([3]*big.Int, error) {
	var o [3]*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 16)
		if !ok {
			return o, errors.New("error parsing [3]*big.Int from [3]string")
		}
		o[i] = param
	}
	return o, nil
}
func BigInt3ToHex(b [3]*big.Int) [3]string {
	var o [3]string
	o[0] = fmt.Sprintf("%x", b[0])
	o[1] = fmt.Sprintf("%x", b[1])
	o[2] = fmt.Sprintf("%x", b[2])
	return o
}

// [][3]*big.Int
func Array3HexToBigInt(s [][3]string) ([][3]*big.Int, error) {
	var o [][3]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := Hex3ToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func Array3BigIntToHex(b [][3]*big.Int) [][3]string {
	var o [][3]string
	for i := 0; i < len(b); i++ {
		o = append(o, BigInt3ToHex(b[i]))
	}
	return o
}

func Hex2ToBigInt(s [2]string) ([2]*big.Int, error) {
	var o [2]*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 16)
		if !ok {
			return o, errors.New("error parsing [2]*big.Int from [2]string")
		}
		o[i] = param
	}
	return o, nil
}

// [3][2]*big.Int
func Hex32ToBigInt(s [3][2]string) ([3][2]*big.Int, error) {
	var o [3][2]*big.Int
	var err error

	o[0], err = Hex2ToBigInt(s[0])
	if err != nil {
		return o, err
	}
	o[1], err = Hex2ToBigInt(s[1])
	if err != nil {
		return o, err
	}
	o[2], err = Hex2ToBigInt(s[2])
	if err != nil {
		return o, err
	}
	return o, nil
}
func BigInt32ToHex(b [3][2]*big.Int) [3][2]string {
	var o [3][2]string
	o[0][0] = fmt.Sprintf("%x", b[0][0])
	o[0][1] = fmt.Sprintf("%x", b[0][1])
	o[1][0] = fmt.Sprintf("%x", b[1][0])
	o[1][1] = fmt.Sprintf("%x", b[1][1])
	o[2][0] = fmt.Sprintf("%x", b[2][0])
	o[2][1] = fmt.Sprintf("%x", b[2][1])
	return o
}

// [][3][2]*big.Int
func Array32HexToBigInt(s [][3][2]string) ([][3][2]*big.Int, error) {
	var o [][3][2]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := Hex32ToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func Array32BigIntToHex(b [][3][2]*big.Int) [][3][2]string {
	var o [][3][2]string
	for i := 0; i < len(b); i++ {
		o = append(o, BigInt32ToHex(b[i]))
	}
	return o
}

// Setup
type PkHex struct {
	G1T [][3]string
	A   [][3]string
	B   [][3][2]string
	C   [][3]string
	Kp  [][3]string
	Ap  [][3]string
	Bp  [][3]string
	Cp  [][3]string
	Z   []string
}
type VkHex struct {
	Vka   [3][2]string
	Vkb   [3]string
	Vkc   [3][2]string
	IC    [][3]string
	G1Kbg [3]string
	G2Kbg [3][2]string
	G2Kg  [3][2]string
	Vkz   [3][2]string
}
type SetupHex struct {
	Pk PkHex
	Vk VkHex
}

func SetupToHex(setup snark.Setup) SetupHex {
	var s SetupHex
	s.Pk.G1T = Array3BigIntToHex(setup.Pk.G1T)
	s.Pk.A = Array3BigIntToHex(setup.Pk.A)
	s.Pk.B = Array32BigIntToHex(setup.Pk.B)
	s.Pk.C = Array3BigIntToHex(setup.Pk.C)
	s.Pk.Kp = Array3BigIntToHex(setup.Pk.Kp)
	s.Pk.Ap = Array3BigIntToHex(setup.Pk.Ap)
	s.Pk.Bp = Array3BigIntToHex(setup.Pk.Bp)
	s.Pk.Cp = Array3BigIntToHex(setup.Pk.Cp)
	s.Pk.Z = ArrayBigIntToHex(setup.Pk.Z)
	s.Vk.Vka = BigInt32ToHex(setup.Vk.Vka)
	s.Vk.Vkb = BigInt3ToHex(setup.Vk.Vkb)
	s.Vk.Vkc = BigInt32ToHex(setup.Vk.Vkc)
	s.Vk.IC = Array3BigIntToHex(setup.Vk.IC)
	s.Vk.G1Kbg = BigInt3ToHex(setup.Vk.G1Kbg)
	s.Vk.G2Kbg = BigInt32ToHex(setup.Vk.G2Kbg)
	s.Vk.G2Kg = BigInt32ToHex(setup.Vk.G2Kg)
	s.Vk.Vkz = BigInt32ToHex(setup.Vk.Vkz)
	return s
}
func SetupFromHex(s SetupHex) (snark.Setup, error) {
	var o snark.Setup
	var err error
	o.Pk.G1T, err = Array3HexToBigInt(s.Pk.G1T)
	if err != nil {
		return o, err
	}
	o.Pk.A, err = Array3HexToBigInt(s.Pk.A)
	if err != nil {
		return o, err
	}
	o.Pk.B, err = Array32HexToBigInt(s.Pk.B)
	if err != nil {
		return o, err
	}
	o.Pk.C, err = Array3HexToBigInt(s.Pk.C)
	if err != nil {
		return o, err
	}
	o.Pk.Kp, err = Array3HexToBigInt(s.Pk.Kp)
	if err != nil {
		return o, err
	}
	o.Pk.Ap, err = Array3HexToBigInt(s.Pk.Ap)
	if err != nil {
		return o, err
	}
	o.Pk.Bp, err = Array3HexToBigInt(s.Pk.Bp)
	if err != nil {
		return o, err
	}
	o.Pk.Cp, err = Array3HexToBigInt(s.Pk.Cp)
	if err != nil {
		return o, err
	}
	o.Pk.Z, err = ArrayHexToBigInt(s.Pk.Z)
	if err != nil {
		return o, err
	}

	o.Vk.Vka, err = Hex32ToBigInt(s.Vk.Vka)
	if err != nil {
		return o, err
	}
	o.Vk.Vkb, err = Hex3ToBigInt(s.Vk.Vkb)
	if err != nil {
		return o, err
	}
	o.Vk.Vkc, err = Hex32ToBigInt(s.Vk.Vkc)
	if err != nil {
		return o, err
	}
	o.Vk.IC, err = Array3HexToBigInt(s.Vk.IC)
	if err != nil {
		return o, err
	}
	o.Vk.G1Kbg, err = Hex3ToBigInt(s.Vk.G1Kbg)
	if err != nil {
		return o, err
	}
	o.Vk.G2Kbg, err = Hex32ToBigInt(s.Vk.G2Kbg)
	if err != nil {
		return o, err
	}
	o.Vk.G2Kg, err = Hex32ToBigInt(s.Vk.G2Kg)
	if err != nil {
		return o, err
	}
	o.Vk.Vkz, err = Hex32ToBigInt(s.Vk.Vkz)
	if err != nil {
		return o, err
	}

	return o, nil

}

// circuit
type CircuitHex struct {
	NVars         int
	NPublic       int
	NSignals      int
	PrivateInputs []string
	PublicInputs  []string
	Signals       []string
	Witness       []string
	Constraints   []circuitcompiler.Constraint
	R1CS          struct {
		A [][]string
		B [][]string
		C [][]string
	}
}

func ArrayArrayBigIntToHex(b [][]*big.Int) [][]string {
	var o [][]string
	for i := 0; i < len(b); i++ {
		o = append(o, ArrayBigIntToHex(b[i]))
	}
	return o
}
func ArrayArrayHexToBigInt(s [][]string) ([][]*big.Int, error) {
	var o [][]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := ArrayHexToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func CircuitToHex(c circuitcompiler.Circuit) CircuitHex {
	var cs CircuitHex
	cs.NVars = c.NVars
	cs.NPublic = c.NPublic
	cs.NSignals = c.NSignals
	cs.PrivateInputs = c.PrivateInputs
	cs.PublicInputs = c.PublicInputs
	cs.Signals = c.Signals
	cs.Witness = ArrayBigIntToHex(c.Witness)
	cs.Constraints = c.Constraints
	cs.R1CS.A = ArrayArrayBigIntToHex(c.R1CS.A)
	cs.R1CS.B = ArrayArrayBigIntToHex(c.R1CS.B)
	cs.R1CS.C = ArrayArrayBigIntToHex(c.R1CS.C)
	return cs
}
func CircuitFromHex(cs CircuitHex) (circuitcompiler.Circuit, error) {
	var c circuitcompiler.Circuit
	var err error
	c.NVars = cs.NVars
	c.NPublic = cs.NPublic
	c.NSignals = cs.NSignals
	c.PrivateInputs = cs.PrivateInputs
	c.PublicInputs = cs.PublicInputs
	c.Signals = cs.Signals
	c.Witness, err = ArrayHexToBigInt(cs.Witness)
	if err != nil {
		return c, err
	}
	c.Constraints = cs.Constraints
	c.R1CS.A, err = ArrayArrayHexToBigInt(cs.R1CS.A)
	if err != nil {
		return c, err
	}
	c.R1CS.B, err = ArrayArrayHexToBigInt(cs.R1CS.B)
	if err != nil {
		return c, err
	}
	c.R1CS.C, err = ArrayArrayHexToBigInt(cs.R1CS.C)
	if err != nil {
		return c, err
	}
	return c, nil
}

// Proof
type ProofHex struct {
	PiA  [3]string
	PiAp [3]string
	PiB  [3][2]string
	PiBp [3]string
	PiC  [3]string
	PiCp [3]string
	PiH  [3]string
	PiKp [3]string
}

func ProofToHex(p snark.Proof) ProofHex {
	var s ProofHex
	s.PiA = BigInt3ToHex(p.PiA)
	s.PiAp = BigInt3ToHex(p.PiAp)
	s.PiB = BigInt32ToHex(p.PiB)
	s.PiBp = BigInt3ToHex(p.PiBp)
	s.PiC = BigInt3ToHex(p.PiC)
	s.PiCp = BigInt3ToHex(p.PiCp)
	s.PiH = BigInt3ToHex(p.PiH)
	s.PiKp = BigInt3ToHex(p.PiKp)
	return s
}
func ProofFromHex(s ProofHex) (snark.Proof, error) {
	var p snark.Proof
	var err error

	p.PiA, err = Hex3ToBigInt(s.PiA)
	if err != nil {
		return p, err
	}
	p.PiAp, err = Hex3ToBigInt(s.PiAp)
	if err != nil {
		return p, err
	}
	p.PiB, err = Hex32ToBigInt(s.PiB)
	if err != nil {
		return p, err
	}
	p.PiBp, err = Hex3ToBigInt(s.PiBp)
	if err != nil {
		return p, err
	}
	p.PiC, err = Hex3ToBigInt(s.PiC)
	if err != nil {
		return p, err
	}
	p.PiCp, err = Hex3ToBigInt(s.PiCp)
	if err != nil {
		return p, err
	}
	p.PiH, err = Hex3ToBigInt(s.PiH)
	if err != nil {
		return p, err
	}
	p.PiKp, err = Hex3ToBigInt(s.PiKp)
	if err != nil {
		return p, err
	}
	return p, nil
}

// groth
type GrothPkHex struct { // Proving Key
	BACDelta [][3]string
	Z        []string
	G1       struct {
		Alpha    [3]string
		Beta     [3]string
		Delta    [3]string
		At       [][3]string
		BACGamma [][3]string
	}
	G2 struct {
		Beta     [3][2]string
		Gamma    [3][2]string
		Delta    [3][2]string
		BACGamma [][3][2]string
	}
	PowersTauDelta [][3]string
}
type GrothVkHex struct {
	IC [][3]string
	G1 struct {
		Alpha [3]string
	}
	G2 struct {
		Beta  [3][2]string
		Gamma [3][2]string
		Delta [3][2]string
	}
}
type GrothSetupHex struct {
	Pk GrothPkHex
	Vk GrothVkHex
}

func GrothSetupToHex(setup groth16.Setup) GrothSetupHex {
	var s GrothSetupHex
	s.Pk.BACDelta = Array3BigIntToHex(setup.Pk.BACDelta)
	s.Pk.Z = ArrayBigIntToHex(setup.Pk.Z)
	s.Pk.G1.Alpha = BigInt3ToHex(setup.Pk.G1.Alpha)
	s.Pk.G1.Beta = BigInt3ToHex(setup.Pk.G1.Beta)
	s.Pk.G1.Delta = BigInt3ToHex(setup.Pk.G1.Delta)
	s.Pk.G1.At = Array3BigIntToHex(setup.Pk.G1.At)
	s.Pk.G1.BACGamma = Array3BigIntToHex(setup.Pk.G1.BACGamma)
	s.Pk.G2.Beta = BigInt32ToHex(setup.Pk.G2.Beta)
	s.Pk.G2.Gamma = BigInt32ToHex(setup.Pk.G2.Gamma)
	s.Pk.G2.Delta = BigInt32ToHex(setup.Pk.G2.Delta)
	s.Pk.G2.BACGamma = Array32BigIntToHex(setup.Pk.G2.BACGamma)
	s.Pk.PowersTauDelta = Array3BigIntToHex(setup.Pk.PowersTauDelta)
	s.Vk.IC = Array3BigIntToHex(setup.Vk.IC)
	s.Vk.G1.Alpha = BigInt3ToHex(setup.Vk.G1.Alpha)
	s.Vk.G2.Beta = BigInt32ToHex(setup.Vk.G2.Beta)
	s.Vk.G2.Gamma = BigInt32ToHex(setup.Vk.G2.Gamma)
	s.Vk.G2.Delta = BigInt32ToHex(setup.Vk.G2.Delta)
	return s
}
func GrothSetupFromHex(s GrothSetupHex) (groth16.Setup, error) {
	var o groth16.Setup
	var err error
	o.Pk.BACDelta, err = Array3HexToBigInt(s.Pk.BACDelta)
	if err != nil {
		return o, err
	}
	o.Pk.Z, err = ArrayHexToBigInt(s.Pk.Z)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Alpha, err = Hex3ToBigInt(s.Pk.G1.Alpha)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Beta, err = Hex3ToBigInt(s.Pk.G1.Beta)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Delta, err = Hex3ToBigInt(s.Pk.G1.Delta)
	if err != nil {
		return o, err
	}
	o.Pk.G1.At, err = Array3HexToBigInt(s.Pk.G1.At)
	if err != nil {
		return o, err
	}
	o.Pk.G1.BACGamma, err = Array3HexToBigInt(s.Pk.G1.BACGamma)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Beta, err = Hex32ToBigInt(s.Pk.G2.Beta)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Gamma, err = Hex32ToBigInt(s.Pk.G2.Gamma)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Delta, err = Hex32ToBigInt(s.Pk.G2.Delta)
	if err != nil {
		return o, err
	}
	o.Pk.G2.BACGamma, err = Array32HexToBigInt(s.Pk.G2.BACGamma)
	if err != nil {
		return o, err
	}
	o.Pk.PowersTauDelta, err = Array3HexToBigInt(s.Pk.PowersTauDelta)
	if err != nil {
		return o, err
	}
	o.Vk.IC, err = Array3HexToBigInt(s.Vk.IC)
	if err != nil {
		return o, err
	}
	o.Vk.G1.Alpha, err = Hex3ToBigInt(s.Vk.G1.Alpha)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Beta, err = Hex32ToBigInt(s.Vk.G2.Beta)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Gamma, err = Hex32ToBigInt(s.Vk.G2.Gamma)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Delta, err = Hex32ToBigInt(s.Vk.G2.Delta)
	if err != nil {
		return o, err
	}
	return o, nil
}

type GrothProofHex struct {
	PiA [3]string
	PiB [3][2]string
	PiC [3]string
}

func GrothProofToHex(p groth16.Proof) GrothProofHex {
	var s GrothProofHex
	s.PiA = BigInt3ToHex(p.PiA)
	s.PiB = BigInt32ToHex(p.PiB)
	s.PiC = BigInt3ToHex(p.PiC)
	return s
}
func GrothProofFromHex(s GrothProofHex) (groth16.Proof, error) {
	var p groth16.Proof
	var err error

	p.PiA, err = Hex3ToBigInt(s.PiA)
	if err != nil {
		return p, err
	}
	p.PiB, err = Hex32ToBigInt(s.PiB)
	if err != nil {
		return p, err
	}
	p.PiC, err = Hex3ToBigInt(s.PiC)
	if err != nil {
		return p, err
	}
	return p, nil
}
