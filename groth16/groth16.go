// implementation of https://eprint.iacr.org/2016/260.pdf

package groth16

import (
	"fmt"
	"math/big"

	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/circuitcompiler"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/r1csqap"
)

type Pk struct { // Proving Key
	BACDelta [][3]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m
	Z        []*big.Int
	G1       struct {
		Alpha    [3]*big.Int
		Beta     [3]*big.Int
		Delta    [3]*big.Int
		At       [][3]*big.Int // {a(τ)} from 0 to m
		BACGamma [][3]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to m
	}
	G2 struct {
		Beta     [3][2]*big.Int
		Gamma    [3][2]*big.Int
		Delta    [3][2]*big.Int
		BACGamma [][3][2]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to m
	}
	PowersTauDelta [][3]*big.Int // powers of τ encrypted in G1 curve, divided by δ
}
type Vk struct {
	IC [][3]*big.Int
	G1 struct {
		Alpha [3]*big.Int
	}
	G2 struct {
		Beta  [3][2]*big.Int
		Gamma [3][2]*big.Int
		Delta [3][2]*big.Int
	}
}

// Setup is the data structure holding the Trusted Setup data. The Setup.Toxic sub struct must be destroyed after the GenerateTrustedSetup function is completed
type Setup struct {
	Toxic struct {
		T      *big.Int // trusted setup secret
		Kalpha *big.Int
		Kbeta  *big.Int
		Kgamma *big.Int
		Kdelta *big.Int
	}

	// public
	Pk Pk
	Vk Vk
}

// Proof contains the parameters to proof the zkSNARK
type Proof struct {
	PiA [3]*big.Int
	PiB [3][2]*big.Int
	PiC [3]*big.Int
}

type utils struct {
	Bn  bn128.Bn128
	FqR fields.Fq
	PF  r1csqap.PolynomialField
}

// Utils is the data structure holding the BN128, FqR Finite Field over R, PolynomialField, that will be used inside the snarks operations
var Utils = prepareUtils()

func prepareUtils() utils {
	bn, err := bn128.NewBn128()
	if err != nil {
		panic(err)
	}
	// new Finite Field
	fqR := fields.NewFq(bn.R)
	// new Polynomial Field
	pf := r1csqap.NewPolynomialField(fqR)

	return utils{
		Bn:  bn,
		FqR: fqR,
		PF:  pf,
	}
}

// GenerateTrustedSetup generates the Trusted Setup from a compiled Circuit. The Setup.Toxic sub data structure must be destroyed
func GenerateTrustedSetup(witnessLength int, circuit circuitcompiler.Circuit, alphas, betas, gammas [][]*big.Int) (Setup, error) {
	var setup Setup
	var err error

	// generate random t value
	setup.Toxic.T, err = Utils.FqR.Rand()
	if err != nil {
		return Setup{}, err
	}

	setup.Toxic.Kalpha, err = Utils.FqR.Rand()
	if err != nil {
		return Setup{}, err
	}
	setup.Toxic.Kbeta, err = Utils.FqR.Rand()
	if err != nil {
		return Setup{}, err
	}
	setup.Toxic.Kgamma, err = Utils.FqR.Rand()
	if err != nil {
		return Setup{}, err
	}
	setup.Toxic.Kdelta, err = Utils.FqR.Rand()
	if err != nil {
		return Setup{}, err
	}

	// z pol
	zpol := []*big.Int{big.NewInt(int64(1))}
	for i := 1; i < len(alphas)-1; i++ {
		zpol = Utils.PF.Mul(
			zpol,
			[]*big.Int{
				Utils.FqR.Neg(
					big.NewInt(int64(i))),
				big.NewInt(int64(1)),
			})
	}
	setup.Pk.Z = zpol
	zt := Utils.PF.Eval(zpol, setup.Toxic.T)
	invDelta := Utils.FqR.Inverse(setup.Toxic.Kdelta)
	ztinvDelta := Utils.FqR.Mul(invDelta, zt)

	// encrypt t values with curve generators
	// powers of tau divided by delta
	var ptd [][3]*big.Int
	ini := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, ztinvDelta)
	ptd = append(ptd, ini)
	tEncr := setup.Toxic.T
	for i := 1; i < len(zpol); i++ {
		ptd = append(ptd, Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, Utils.FqR.Mul(tEncr, ztinvDelta)))
		tEncr = Utils.FqR.Mul(tEncr, setup.Toxic.T)
	}
	// powers of τ encrypted in G1 curve, divided by δ
	// (G1 * τ) / δ
	setup.Pk.PowersTauDelta = ptd

	setup.Pk.G1.Alpha = Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, setup.Toxic.Kalpha)
	setup.Pk.G1.Beta = Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, setup.Toxic.Kbeta)
	setup.Pk.G1.Delta = Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, setup.Toxic.Kdelta)
	setup.Pk.G2.Beta = Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, setup.Toxic.Kbeta)
	setup.Pk.G2.Delta = Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, setup.Toxic.Kdelta)

	setup.Vk.G1.Alpha = Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, setup.Toxic.Kalpha)
	setup.Vk.G2.Beta = Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, setup.Toxic.Kbeta)
	setup.Vk.G2.Gamma = Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, setup.Toxic.Kgamma)
	setup.Vk.G2.Delta = Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, setup.Toxic.Kdelta)

	for i := 0; i < len(circuit.Signals); i++ {
		// Pk.G1.At: {a(τ)} from 0 to m
		at := Utils.PF.Eval(alphas[i], setup.Toxic.T)
		a := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, at)
		setup.Pk.G1.At = append(setup.Pk.G1.At, a)

		bt := Utils.PF.Eval(betas[i], setup.Toxic.T)
		g1bt := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, bt)
		g2bt := Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, bt)
		// G1.BACGamma: {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to m in G1
		setup.Pk.G1.BACGamma = append(setup.Pk.G1.BACGamma, g1bt)
		// G2.BACGamma: {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to m in G2
		setup.Pk.G2.BACGamma = append(setup.Pk.G2.BACGamma, g2bt)
	}

	zero3 := [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}
	for i := 0; i < circuit.NPublic+1; i++ {
		setup.Pk.BACDelta = append(setup.Pk.BACDelta, zero3)
	}
	for i := circuit.NPublic + 1; i < circuit.NVars; i++ {
		// TODO calculate all at, bt, ct outside, to avoid repeating calculations
		at := Utils.PF.Eval(alphas[i], setup.Toxic.T)
		bt := Utils.PF.Eval(betas[i], setup.Toxic.T)
		ct := Utils.PF.Eval(gammas[i], setup.Toxic.T)
		c := Utils.FqR.Mul(
			invDelta,
			Utils.FqR.Add(
				Utils.FqR.Add(
					Utils.FqR.Mul(at, setup.Toxic.Kbeta),
					Utils.FqR.Mul(bt, setup.Toxic.Kalpha),
				),
				ct,
			),
		)
		g1c := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, c)

		// Pk.BACDelta: {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m
		setup.Pk.BACDelta = append(setup.Pk.BACDelta, g1c)
	}

	for i := 0; i <= circuit.NPublic; i++ {
		at := Utils.PF.Eval(alphas[i], setup.Toxic.T)
		bt := Utils.PF.Eval(betas[i], setup.Toxic.T)
		ct := Utils.PF.Eval(gammas[i], setup.Toxic.T)
		ic := Utils.FqR.Mul(
			Utils.FqR.Inverse(setup.Toxic.Kgamma),
			Utils.FqR.Add(
				Utils.FqR.Add(
					Utils.FqR.Mul(at, setup.Toxic.Kbeta),
					Utils.FqR.Mul(bt, setup.Toxic.Kalpha),
				),
				ct,
			),
		)
		g1ic := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, ic)
		// used in verifier
		setup.Vk.IC = append(setup.Vk.IC, g1ic)
	}

	return setup, nil
}

// GenerateProofs generates all the parameters to proof the zkSNARK from the Circuit, Setup and the Witness
func GenerateProofs(circuit circuitcompiler.Circuit, pk Pk, w []*big.Int, px []*big.Int) (Proof, error) {
	var proof Proof
	proof.PiA = [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}
	proof.PiB = Utils.Bn.Fq6.Zero()
	proof.PiC = [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}

	r, err := Utils.FqR.Rand()
	if err != nil {
		return Proof{}, err
	}
	s, err := Utils.FqR.Rand()
	if err != nil {
		return Proof{}, err
	}

	// piBG1 will hold all the same than proof.PiB but in G1 curve
	piBG1 := [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}

	for i := 0; i < circuit.NVars; i++ {
		proof.PiA = Utils.Bn.G1.Add(proof.PiA, Utils.Bn.G1.MulScalar(pk.G1.At[i], w[i]))
		piBG1 = Utils.Bn.G1.Add(piBG1, Utils.Bn.G1.MulScalar(pk.G1.BACGamma[i], w[i]))
		proof.PiB = Utils.Bn.G2.Add(proof.PiB, Utils.Bn.G2.MulScalar(pk.G2.BACGamma[i], w[i]))
	}
	for i := circuit.NPublic + 1; i < circuit.NVars; i++ {
		proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(pk.BACDelta[i], w[i]))
	}

	// piA = (Σ from 0 to m (pk.A * w[i])) + pk.Alpha1 + r * δ
	proof.PiA = Utils.Bn.G1.Add(proof.PiA, pk.G1.Alpha)
	deltaR := Utils.Bn.G1.MulScalar(pk.G1.Delta, r)
	proof.PiA = Utils.Bn.G1.Add(proof.PiA, deltaR)

	// piBG1 = (Σ from 0 to m (pk.B1 * w[i])) + pk.g1.Beta + s * δ
	// piB = piB2 = (Σ from 0 to m (pk.B2 * w[i])) + pk.g2.Beta + s * δ
	piBG1 = Utils.Bn.G1.Add(piBG1, pk.G1.Beta)
	proof.PiB = Utils.Bn.G2.Add(proof.PiB, pk.G2.Beta)
	deltaSG1 := Utils.Bn.G1.MulScalar(pk.G1.Delta, s)
	piBG1 = Utils.Bn.G1.Add(piBG1, deltaSG1)
	deltaSG2 := Utils.Bn.G2.MulScalar(pk.G2.Delta, s)
	proof.PiB = Utils.Bn.G2.Add(proof.PiB, deltaSG2)

	hx := Utils.PF.DivisorPolynomial(px, pk.Z) // maybe move this calculation to a previous step

	// piC = (Σ from l+1 to m (w[i] * (pk.g1.Beta + pk.g1.Alpha + pk.C)) + h(tau)) / δ) + piA*s + r*piB - r*s*δ
	for i := 0; i < len(hx); i++ {
		proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(pk.PowersTauDelta[i], hx[i]))
	}
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(proof.PiA, s))
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(piBG1, r))
	negRS := Utils.FqR.Neg(Utils.FqR.Mul(r, s))
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(pk.G1.Delta, negRS))

	return proof, nil
}

// VerifyProof verifies over the BN128 the Pairings of the Proof
func VerifyProof(vk Vk, proof Proof, publicSignals []*big.Int, debug bool) bool {

	icPubl := vk.IC[0]
	for i := 0; i < len(publicSignals); i++ {
		icPubl = Utils.Bn.G1.Add(icPubl, Utils.Bn.G1.MulScalar(vk.IC[i+1], publicSignals[i]))
	}

	if !Utils.Bn.Fq12.Equal(
		Utils.Bn.Pairing(proof.PiA, proof.PiB),
		Utils.Bn.Fq12.Mul(
			Utils.Bn.Pairing(vk.G1.Alpha, vk.G2.Beta),
			Utils.Bn.Fq12.Mul(
				Utils.Bn.Pairing(icPubl, vk.G2.Gamma),
				Utils.Bn.Pairing(proof.PiC, vk.G2.Delta)))) {
		if debug {
			fmt.Println("❌ groth16 verification not passed")
		}
		return false
	}
	if debug {
		fmt.Println("✓ groth16 verification passed")
	}

	return true
}
