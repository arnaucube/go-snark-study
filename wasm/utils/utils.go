package utils

import (
	"errors"
	"math/big"

	snark "github.com/arnaucube/go-snark"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/groth16"
)

// []*big.Int
func ArrayBigIntToString(b []*big.Int) []string {
	var o []string
	for i := 0; i < len(b); i++ {
		o = append(o, b[i].String())
	}
	return o
}
func ArrayStringToBigInt(s []string) ([]*big.Int, error) {
	var px []*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 10)
		if !ok {
			return nil, errors.New("error parsing px from pxString")
		}
		px = append(px, param)
	}
	return px, nil
}

// [3]*big.Int
func String3ToBigInt(s [3]string) ([3]*big.Int, error) {
	var o [3]*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 10)
		if !ok {
			return o, errors.New("error parsing [3]*big.Int from [3]string")
		}
		o[i] = param
	}
	return o, nil
}
func BigInt3ToString(b [3]*big.Int) [3]string {
	var o [3]string
	o[0] = b[0].String()
	o[1] = b[1].String()
	o[2] = b[2].String()
	return o
}

// [][3]*big.Int
func Array3StringToBigInt(s [][3]string) ([][3]*big.Int, error) {
	var o [][3]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := String3ToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func Array3BigIntToString(b [][3]*big.Int) [][3]string {
	var o [][3]string
	for i := 0; i < len(b); i++ {
		o = append(o, BigInt3ToString(b[i]))
	}
	return o
}

func String2ToBigInt(s [2]string) ([2]*big.Int, error) {
	var o [2]*big.Int
	for i := 0; i < len(s); i++ {
		param, ok := new(big.Int).SetString(s[i], 10)
		if !ok {
			return o, errors.New("error parsing [2]*big.Int from [2]string")
		}
		o[i] = param
	}
	return o, nil
}

// [3][2]*big.Int
func String32ToBigInt(s [3][2]string) ([3][2]*big.Int, error) {
	var o [3][2]*big.Int
	var err error

	o[0], err = String2ToBigInt(s[0])
	if err != nil {
		return o, err
	}
	o[1], err = String2ToBigInt(s[1])
	if err != nil {
		return o, err
	}
	o[2], err = String2ToBigInt(s[2])
	if err != nil {
		return o, err
	}
	return o, nil
}
func BigInt32ToString(b [3][2]*big.Int) [3][2]string {
	var o [3][2]string
	o[0][0] = b[0][0].String()
	o[0][1] = b[0][1].String()
	o[1][0] = b[1][0].String()
	o[1][1] = b[1][1].String()
	o[2][0] = b[2][0].String()
	o[2][1] = b[2][1].String()
	return o
}

// [][3][2]*big.Int
func Array32StringToBigInt(s [][3][2]string) ([][3][2]*big.Int, error) {
	var o [][3][2]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := String32ToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func Array32BigIntToString(b [][3][2]*big.Int) [][3][2]string {
	var o [][3][2]string
	for i := 0; i < len(b); i++ {
		o = append(o, BigInt32ToString(b[i]))
	}
	return o
}

// Setup
type SetupString struct {
	// public
	G1T [][3]string
	G2T [][3][2]string
	Pk  struct {
		A  [][3]string
		B  [][3][2]string
		C  [][3]string
		Kp [][3]string
		Ap [][3]string
		Bp [][3]string
		Cp [][3]string
		Z  []string
	}
	Vk struct {
		Vka   [3][2]string
		Vkb   [3]string
		Vkc   [3][2]string
		IC    [][3]string
		G1Kbg [3]string
		G2Kbg [3][2]string
		G2Kg  [3][2]string
		Vkz   [3][2]string
	}
}

func SetupToString(setup snark.Setup) SetupString {
	var s SetupString
	s.G1T = Array3BigIntToString(setup.G1T)
	s.G2T = Array32BigIntToString(setup.G2T)
	s.Pk.A = Array3BigIntToString(setup.Pk.A)
	s.Pk.B = Array32BigIntToString(setup.Pk.B)
	s.Pk.C = Array3BigIntToString(setup.Pk.C)
	s.Pk.Kp = Array3BigIntToString(setup.Pk.Kp)
	s.Pk.Ap = Array3BigIntToString(setup.Pk.Ap)
	s.Pk.Bp = Array3BigIntToString(setup.Pk.Bp)
	s.Pk.Cp = Array3BigIntToString(setup.Pk.Cp)
	s.Pk.Z = ArrayBigIntToString(setup.Pk.Z)
	s.Vk.Vka = BigInt32ToString(setup.Vk.Vka)
	s.Vk.Vkb = BigInt3ToString(setup.Vk.Vkb)
	s.Vk.Vkc = BigInt32ToString(setup.Vk.Vkc)
	s.Vk.IC = Array3BigIntToString(setup.Vk.IC)
	s.Vk.G1Kbg = BigInt3ToString(setup.Vk.G1Kbg)
	s.Vk.G2Kbg = BigInt32ToString(setup.Vk.G2Kbg)
	s.Vk.G2Kg = BigInt32ToString(setup.Vk.G2Kg)
	s.Vk.Vkz = BigInt32ToString(setup.Vk.Vkz)
	return s
}
func SetupFromString(s SetupString) (snark.Setup, error) {
	var o snark.Setup
	var err error
	o.G1T, err = Array3StringToBigInt(s.G1T)
	if err != nil {
		return o, err
	}
	o.G2T, err = Array32StringToBigInt(s.G2T)
	if err != nil {
		return o, err
	}
	o.Pk.A, err = Array3StringToBigInt(s.Pk.A)
	if err != nil {
		return o, err
	}
	o.Pk.B, err = Array32StringToBigInt(s.Pk.B)
	if err != nil {
		return o, err
	}
	o.Pk.C, err = Array3StringToBigInt(s.Pk.C)
	if err != nil {
		return o, err
	}
	o.Pk.Kp, err = Array3StringToBigInt(s.Pk.Kp)
	if err != nil {
		return o, err
	}
	o.Pk.Ap, err = Array3StringToBigInt(s.Pk.Ap)
	if err != nil {
		return o, err
	}
	o.Pk.Bp, err = Array3StringToBigInt(s.Pk.Bp)
	if err != nil {
		return o, err
	}
	o.Pk.Cp, err = Array3StringToBigInt(s.Pk.Cp)
	if err != nil {
		return o, err
	}
	o.Pk.Z, err = ArrayStringToBigInt(s.Pk.Z)
	if err != nil {
		return o, err
	}

	o.Vk.Vka, err = String32ToBigInt(s.Vk.Vka)
	if err != nil {
		return o, err
	}
	o.Vk.Vkb, err = String3ToBigInt(s.Vk.Vkb)
	if err != nil {
		return o, err
	}
	o.Vk.Vkc, err = String32ToBigInt(s.Vk.Vkc)
	if err != nil {
		return o, err
	}
	o.Vk.IC, err = Array3StringToBigInt(s.Vk.IC)
	if err != nil {
		return o, err
	}
	o.Vk.G1Kbg, err = String3ToBigInt(s.Vk.G1Kbg)
	if err != nil {
		return o, err
	}
	o.Vk.G2Kbg, err = String32ToBigInt(s.Vk.G2Kbg)
	if err != nil {
		return o, err
	}
	o.Vk.G2Kg, err = String32ToBigInt(s.Vk.G2Kg)
	if err != nil {
		return o, err
	}
	o.Vk.Vkz, err = String32ToBigInt(s.Vk.Vkz)
	if err != nil {
		return o, err
	}

	return o, nil

}

// circuit
type CircuitString struct {
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

func ArrayArrayBigIntToString(b [][]*big.Int) [][]string {
	var o [][]string
	for i := 0; i < len(b); i++ {
		o = append(o, ArrayBigIntToString(b[i]))
	}
	return o
}
func ArrayArrayStringToBigInt(s [][]string) ([][]*big.Int, error) {
	var o [][]*big.Int
	for i := 0; i < len(s); i++ {
		parsed, err := ArrayStringToBigInt(s[i])
		if err != nil {
			return o, err
		}
		o = append(o, parsed)
	}
	return o, nil
}
func CircuitToString(c circuitcompiler.Circuit) CircuitString {
	var cs CircuitString
	cs.NVars = c.NVars
	cs.NPublic = c.NPublic
	cs.NSignals = c.NSignals
	cs.PrivateInputs = c.PrivateInputs
	cs.PublicInputs = c.PublicInputs
	cs.Signals = c.Signals
	cs.Witness = ArrayBigIntToString(c.Witness)
	cs.Constraints = c.Constraints
	cs.R1CS.A = ArrayArrayBigIntToString(c.R1CS.A)
	cs.R1CS.B = ArrayArrayBigIntToString(c.R1CS.B)
	cs.R1CS.C = ArrayArrayBigIntToString(c.R1CS.C)
	return cs
}
func CircuitFromString(cs CircuitString) (circuitcompiler.Circuit, error) {
	var c circuitcompiler.Circuit
	var err error
	c.NVars = cs.NVars
	c.NPublic = cs.NPublic
	c.NSignals = cs.NSignals
	c.PrivateInputs = cs.PrivateInputs
	c.PublicInputs = cs.PublicInputs
	c.Signals = cs.Signals
	c.Witness, err = ArrayStringToBigInt(cs.Witness)
	if err != nil {
		return c, err
	}
	c.Constraints = cs.Constraints
	c.R1CS.A, err = ArrayArrayStringToBigInt(cs.R1CS.A)
	if err != nil {
		return c, err
	}
	c.R1CS.B, err = ArrayArrayStringToBigInt(cs.R1CS.B)
	if err != nil {
		return c, err
	}
	c.R1CS.C, err = ArrayArrayStringToBigInt(cs.R1CS.C)
	if err != nil {
		return c, err
	}
	return c, nil
}

// Proof
type ProofString struct {
	PiA  [3]string
	PiAp [3]string
	PiB  [3][2]string
	PiBp [3]string
	PiC  [3]string
	PiCp [3]string
	PiH  [3]string
	PiKp [3]string
}

func ProofToString(p snark.Proof) ProofString {
	var s ProofString
	s.PiA = BigInt3ToString(p.PiA)
	s.PiAp = BigInt3ToString(p.PiAp)
	s.PiB = BigInt32ToString(p.PiB)
	s.PiBp = BigInt3ToString(p.PiBp)
	s.PiC = BigInt3ToString(p.PiC)
	s.PiCp = BigInt3ToString(p.PiCp)
	s.PiH = BigInt3ToString(p.PiH)
	s.PiKp = BigInt3ToString(p.PiKp)
	return s
}
func ProofFromString(s ProofString) (snark.Proof, error) {
	var p snark.Proof
	var err error

	p.PiA, err = String3ToBigInt(s.PiA)
	if err != nil {
		return p, err
	}
	p.PiAp, err = String3ToBigInt(s.PiAp)
	if err != nil {
		return p, err
	}
	p.PiB, err = String32ToBigInt(s.PiB)
	if err != nil {
		return p, err
	}
	p.PiBp, err = String3ToBigInt(s.PiBp)
	if err != nil {
		return p, err
	}
	p.PiC, err = String3ToBigInt(s.PiC)
	if err != nil {
		return p, err
	}
	p.PiCp, err = String3ToBigInt(s.PiCp)
	if err != nil {
		return p, err
	}
	p.PiH, err = String3ToBigInt(s.PiH)
	if err != nil {
		return p, err
	}
	p.PiKp, err = String3ToBigInt(s.PiKp)
	if err != nil {
		return p, err
	}
	return p, nil
}

// groth
type GrothSetupString struct {
	Pk struct { // Proving Key
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
	Vk struct {
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
}

func GrothSetupToString(setup groth16.Setup) GrothSetupString {
	var s GrothSetupString
	s.Pk.BACDelta = Array3BigIntToString(setup.Pk.BACDelta)
	s.Pk.Z = ArrayBigIntToString(setup.Pk.Z)
	s.Pk.G1.Alpha = BigInt3ToString(setup.Pk.G1.Alpha)
	s.Pk.G1.Beta = BigInt3ToString(setup.Pk.G1.Beta)
	s.Pk.G1.Delta = BigInt3ToString(setup.Pk.G1.Delta)
	s.Pk.G1.At = Array3BigIntToString(setup.Pk.G1.At)
	s.Pk.G1.BACGamma = Array3BigIntToString(setup.Pk.G1.BACGamma)
	s.Pk.G2.Beta = BigInt32ToString(setup.Pk.G2.Beta)
	s.Pk.G2.Gamma = BigInt32ToString(setup.Pk.G2.Gamma)
	s.Pk.G2.Delta = BigInt32ToString(setup.Pk.G2.Delta)
	s.Pk.G2.BACGamma = Array32BigIntToString(setup.Pk.G2.BACGamma)
	s.Pk.PowersTauDelta = Array3BigIntToString(setup.Pk.PowersTauDelta)
	s.Vk.IC = Array3BigIntToString(setup.Vk.IC)
	s.Vk.G1.Alpha = BigInt3ToString(setup.Vk.G1.Alpha)
	s.Vk.G2.Beta = BigInt32ToString(setup.Vk.G2.Beta)
	s.Vk.G2.Gamma = BigInt32ToString(setup.Vk.G2.Gamma)
	s.Vk.G2.Delta = BigInt32ToString(setup.Vk.G2.Delta)
	return s
}
func GrothSetupFromString(s GrothSetupString) (groth16.Setup, error) {
	var o groth16.Setup
	var err error
	o.Pk.BACDelta, err = Array3StringToBigInt(s.Pk.BACDelta)
	if err != nil {
		return o, err
	}
	o.Pk.Z, err = ArrayStringToBigInt(s.Pk.Z)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Alpha, err = String3ToBigInt(s.Pk.G1.Alpha)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Beta, err = String3ToBigInt(s.Pk.G1.Beta)
	if err != nil {
		return o, err
	}
	o.Pk.G1.Delta, err = String3ToBigInt(s.Pk.G1.Delta)
	if err != nil {
		return o, err
	}
	o.Pk.G1.At, err = Array3StringToBigInt(s.Pk.G1.At)
	if err != nil {
		return o, err
	}
	o.Pk.G1.BACGamma, err = Array3StringToBigInt(s.Pk.G1.BACGamma)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Beta, err = String32ToBigInt(s.Pk.G2.Beta)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Gamma, err = String32ToBigInt(s.Pk.G2.Gamma)
	if err != nil {
		return o, err
	}
	o.Pk.G2.Delta, err = String32ToBigInt(s.Pk.G2.Delta)
	if err != nil {
		return o, err
	}
	o.Pk.G2.BACGamma, err = Array32StringToBigInt(s.Pk.G2.BACGamma)
	if err != nil {
		return o, err
	}
	o.Pk.PowersTauDelta, err = Array3StringToBigInt(s.Pk.PowersTauDelta)
	if err != nil {
		return o, err
	}
	o.Vk.IC, err = Array3StringToBigInt(s.Vk.IC)
	if err != nil {
		return o, err
	}
	o.Vk.G1.Alpha, err = String3ToBigInt(s.Vk.G1.Alpha)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Beta, err = String32ToBigInt(s.Vk.G2.Beta)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Gamma, err = String32ToBigInt(s.Vk.G2.Gamma)
	if err != nil {
		return o, err
	}
	o.Vk.G2.Delta, err = String32ToBigInt(s.Vk.G2.Delta)
	if err != nil {
		return o, err
	}
	return o, nil
}

type GrothProofString struct {
	PiA [3]string
	PiB [3][2]string
	PiC [3]string
}

func GrothProofToString(p groth16.Proof) GrothProofString {
	var s GrothProofString
	s.PiA = BigInt3ToString(p.PiA)
	s.PiB = BigInt32ToString(p.PiB)
	s.PiC = BigInt3ToString(p.PiC)
	return s
}
func GrothProofFromString(s GrothProofString) (groth16.Proof, error) {
	var p groth16.Proof
	var err error

	p.PiA, err = String3ToBigInt(s.PiA)
	if err != nil {
		return p, err
	}
	p.PiB, err = String32ToBigInt(s.PiB)
	if err != nil {
		return p, err
	}
	p.PiC, err = String3ToBigInt(s.PiC)
	if err != nil {
		return p, err
	}
	return p, nil
}
