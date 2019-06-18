// implementation of https://eprint.iacr.org/2016/260.pdf

package proof

import (
	"fmt"
	"math/big"

	"github.com/arnaucube/go-snark/circuit"
)

// Groth16Setup is Groth16 system setup structure
type Groth16Setup struct {
	Toxic struct {
		T      *big.Int // trusted setup secret
		Kalpha *big.Int
		Kbeta  *big.Int
		Kgamma *big.Int
		Kdelta *big.Int
	} `json:"-"`

	// public
	Pk struct { // Proving Key
		BACDelta [][3]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to l
		Z        []*big.Int
		G1       struct {
			Alpha    [3]*big.Int
			Beta     [3]*big.Int
			Delta    [3]*big.Int
			At       [][3]*big.Int // {a(τ)} from 0 to m
			BACGamma [][3]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m
		}
		G2 struct {
			Beta     [3][2]*big.Int
			Gamma    [3][2]*big.Int
			Delta    [3][2]*big.Int
			BACGamma [][3][2]*big.Int // {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m
		}
		PowersTauDelta [][3]*big.Int // powers of τ encrypted in G1 curve, divided by δ
	}
	Vk struct {
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
}

// Groth16Proof is Groth16 proof structure
type Groth16Proof struct {
	PiA [3]*big.Int
	PiB [3][2]*big.Int
	PiC [3]*big.Int
}

// Z is ...
func (setup *Groth16Setup) Z() []*big.Int {
	return setup.Pk.Z
}

// Init setups the trusted setup from a compiled circuit
func (setup *Groth16Setup) Init(cir *circuit.Circuit, alphas, betas, gammas [][]*big.Int) error {
	var err error

	setup.Toxic.T, err = Utils.FqR.Rand()
	if err != nil {
		return err
	}

	setup.Toxic.Kalpha, err = Utils.FqR.Rand()
	if err != nil {
		return err
	}
	setup.Toxic.Kbeta, err = Utils.FqR.Rand()
	if err != nil {
		return err
	}
	setup.Toxic.Kgamma, err = Utils.FqR.Rand()
	if err != nil {
		return err
	}
	setup.Toxic.Kdelta, err = Utils.FqR.Rand()
	if err != nil {
		return err
	}

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

	for i := 0; i < len(cir.Signals); i++ {
		// Pk.G1.At: {a(τ)} from 0 to m
		at := Utils.PF.Eval(alphas[i], setup.Toxic.T)
		a := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, at)
		setup.Pk.G1.At = append(setup.Pk.G1.At, a)

		bt := Utils.PF.Eval(betas[i], setup.Toxic.T)
		g1bt := Utils.Bn.G1.MulScalar(Utils.Bn.G1.G, bt)
		g2bt := Utils.Bn.G2.MulScalar(Utils.Bn.G2.G, bt)
		// G1.BACGamma: {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m in G1
		setup.Pk.G1.BACGamma = append(setup.Pk.G1.BACGamma, g1bt)
		// G2.BACGamma: {( βui(x)+αvi(x)+wi(x) ) / δ } from l+1 to m in G2
		setup.Pk.G2.BACGamma = append(setup.Pk.G2.BACGamma, g2bt)
	}

	zero3 := [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}
	for i := 0; i < cir.NPublic+1; i++ {
		setup.Pk.BACDelta = append(setup.Pk.BACDelta, zero3)
	}
	for i := cir.NPublic + 1; i < cir.NVars; i++ {
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

		// Pk.BACDelta: {( βui(x)+αvi(x)+wi(x) ) / γ } from 0 to l
		setup.Pk.BACDelta = append(setup.Pk.BACDelta, g1c)
	}

	for i := 0; i <= cir.NPublic; i++ {
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

	return nil
}

// Generate generates Pinocchio proof
func (setup Groth16Setup) Generate(cir *circuit.Circuit, w []*big.Int, px []*big.Int) (Proof, error) {
	proof := &Groth16Proof{}
	proof.PiA = [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}
	proof.PiB = Utils.Bn.Fq6.Zero()
	proof.PiC = [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}

	r, err := Utils.FqR.Rand()
	if err != nil {
		return &Groth16Proof{}, err
	}
	s, err := Utils.FqR.Rand()
	if err != nil {
		return &Groth16Proof{}, err
	}

	// piBG1 will hold all the same than proof.PiB but in G1 curve
	piBG1 := [3]*big.Int{Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero(), Utils.Bn.G1.F.Zero()}

	for i := 0; i < cir.NVars; i++ {
		proof.PiA = Utils.Bn.G1.Add(proof.PiA, Utils.Bn.G1.MulScalar(setup.Pk.G1.At[i], w[i]))
		piBG1 = Utils.Bn.G1.Add(piBG1, Utils.Bn.G1.MulScalar(setup.Pk.G1.BACGamma[i], w[i]))
		proof.PiB = Utils.Bn.G2.Add(proof.PiB, Utils.Bn.G2.MulScalar(setup.Pk.G2.BACGamma[i], w[i]))
	}
	for i := cir.NPublic + 1; i < cir.NVars; i++ {
		proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(setup.Pk.BACDelta[i], w[i]))
	}

	// piA = (Σ from 0 to m (pk.A * w[i])) + pk.Alpha1 + r * δ
	proof.PiA = Utils.Bn.G1.Add(proof.PiA, setup.Pk.G1.Alpha)
	deltaR := Utils.Bn.G1.MulScalar(setup.Pk.G1.Delta, r)
	proof.PiA = Utils.Bn.G1.Add(proof.PiA, deltaR)

	// piBG1 = (Σ from 0 to m (pk.B1 * w[i])) + pk.g1.Beta + s * δ
	// piB = piB2 = (Σ from 0 to m (pk.B2 * w[i])) + pk.g2.Beta + s * δ
	piBG1 = Utils.Bn.G1.Add(piBG1, setup.Pk.G1.Beta)
	proof.PiB = Utils.Bn.G2.Add(proof.PiB, setup.Pk.G2.Beta)
	deltaSG1 := Utils.Bn.G1.MulScalar(setup.Pk.G1.Delta, s)
	piBG1 = Utils.Bn.G1.Add(piBG1, deltaSG1)
	deltaSG2 := Utils.Bn.G2.MulScalar(setup.Pk.G2.Delta, s)
	proof.PiB = Utils.Bn.G2.Add(proof.PiB, deltaSG2)

	hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z) // maybe move this calculation to a previous step

	// piC = (Σ from l+1 to m (w[i] * (pk.g1.Beta + pk.g1.Alpha + pk.C)) + h(tau)) / δ) + piA*s + r*piB - r*s*δ
	for i := 0; i < len(hx); i++ {
		proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(setup.Pk.PowersTauDelta[i], hx[i]))
	}
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(proof.PiA, s))
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(piBG1, r))
	negRS := Utils.FqR.Neg(Utils.FqR.Mul(r, s))
	proof.PiC = Utils.Bn.G1.Add(proof.PiC, Utils.Bn.G1.MulScalar(setup.Pk.G1.Delta, negRS))

	return proof, nil
}

// Verify verifies over the BN128 the Pairings of the Proof
func (setup Groth16Setup) Verify(proof Proof, publicSignals []*big.Int) (bool, error) {
	pproof, ok := proof.(*Groth16Proof)
	if !ok {
		return false, fmt.Errorf("bad proof type")
	}

	icPubl := setup.Vk.IC[0]
	for i := 0; i < len(publicSignals); i++ {
		icPubl = Utils.Bn.G1.Add(icPubl, Utils.Bn.G1.MulScalar(setup.Vk.IC[i+1], publicSignals[i]))
	}

	if !Utils.Bn.Fq12.Equal(
		Utils.Bn.Pairing(pproof.PiA, pproof.PiB),
		Utils.Bn.Fq12.Mul(
			Utils.Bn.Pairing(setup.Vk.G1.Alpha, setup.Vk.G2.Beta),
			Utils.Bn.Fq12.Mul(
				Utils.Bn.Pairing(icPubl, setup.Vk.G2.Gamma),
				Utils.Bn.Pairing(pproof.PiC, setup.Vk.G2.Delta),
			),
		)) {
		return false, nil
	}

	return true, nil
}
