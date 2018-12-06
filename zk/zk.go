package zk

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/arnaucube/go-snark/bn128"
)

type Setup struct {
	T  *big.Int // trusted setup secret
	Ka *big.Int // trusted setup
	Kb *big.Int // trusted setup
	Kc *big.Int // trusted setup

	// public
	G1T [][3]*big.Int    // t encrypted in G1 curve
	G2T [][3][2]*big.Int // t encrypted in G2 curve
}
type Proof struct {
	PiA  [3]*big.Int
	PiAp [3]*big.Int
	PiB  [3][2]*big.Int
	PiBp [3][2]*big.Int
	PiC  [3]*big.Int
	PiCp [3]*big.Int
	PiH  [3]*big.Int
	Va   [3][2]*big.Int
	Vb   [3][2]*big.Int
	Vc   [3][2]*big.Int
	Vz   [3][2]*big.Int
}

const bits = 512

func GenerateTrustedSetup(bn bn128.Bn128, pollength int) (Setup, error) {
	var setup Setup
	var err error
	// generate random t value
	setup.T, err = rand.Prime(rand.Reader, bits)
	if err != nil {
		return Setup{}, err
	}
	fmt.Print("trusted t: ")
	fmt.Println(setup.T)

	// encrypt t values with curve generators
	var gt1 [][3]*big.Int
	var gt2 [][3][2]*big.Int
	for i := 0; i < pollength; i++ {
		tPow := bn.Fq1.Exp(setup.T, big.NewInt(int64(i)))
		tEncr1 := bn.G1.MulScalar(bn.G1.G, tPow)
		gt1 = append(gt1, tEncr1)
		tEncr2 := bn.G2.MulScalar(bn.G2.G, tPow)
		gt2 = append(gt2, tEncr2)
	}
	// gt1: g1, g1*t, g1*t^2, g1*t^3, ...
	// gt2: g2, g2*t, g2*t^2, ...
	setup.G1T = gt1
	setup.G2T = gt2

	// k for pi'
	setup.Ka, err = rand.Prime(rand.Reader, bits)
	if err != nil {
		return Setup{}, err
	}
	setup.Kb, err = rand.Prime(rand.Reader, bits)
	if err != nil {
		return Setup{}, err
	}
	setup.Kc, err = rand.Prime(rand.Reader, bits)
	if err != nil {
		return Setup{}, err
	}

	return setup, nil
}

func GenerateProofs(bn bn128.Bn128, setup Setup, ax, bx, cx, hx, zx []*big.Int) Proof {
	var proof Proof

	// g1*A(x)
	proof.PiA = [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(ax); i++ {
		m := bn.G1.MulScalar(setup.G1T[i], ax[i])
		proof.PiA = bn.G1.Add(proof.PiA, m)
	}
	proof.PiAp = bn.G1.MulScalar(proof.PiA, setup.Ka)

	// g1*B(x)
	proof.PiB = bn.Fq6.Zero()
	for i := 0; i < len(bx); i++ {
		m := bn.G2.MulScalar(setup.G2T[i], bx[i])
		proof.PiB = bn.G2.Add(proof.PiB, m)
	}
	proof.PiBp = bn.G2.MulScalar(proof.PiB, setup.Kb)

	// g1*C(x)
	proof.PiC = [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(cx); i++ {
		m := bn.G1.MulScalar(setup.G1T[i], cx[i])
		proof.PiC = bn.G1.Add(proof.PiC, m)
	}
	proof.PiCp = bn.G1.MulScalar(proof.PiC, setup.Kc)

	g1Ht := [3]*big.Int{bn.G1.F.Zero(), bn.G1.F.Zero(), bn.G1.F.Zero()}
	for i := 0; i < len(hx); i++ {
		m := bn.G1.MulScalar(setup.G1T[i], hx[i])
		g1Ht = bn.G1.Add(g1Ht, m)
	}
	g2Zt := bn.Fq6.Zero()
	for i := 0; i < len(bx); i++ {
		m := bn.G2.MulScalar(setup.G2T[i], zx[i])
		g2Zt = bn.G2.Add(g2Zt, m)
	}
	proof.PiH = g1Ht
	proof.Vz = g2Zt
	proof.Va = bn.G2.MulScalar(bn.G2.G, setup.Ka)
	proof.Vb = bn.G2.MulScalar(bn.G2.G, setup.Kb)
	proof.Vc = bn.G2.MulScalar(bn.G2.G, setup.Kc)

	return proof
}

func VerifyProof(bn bn128.Bn128, setup Setup, proof Proof) bool {

	// e(piA, Va) == e(piA', g2)
	pairingPiaVa, err := bn.Pairing(proof.PiA, proof.Va)
	if err != nil {
		return false
	}
	pairingPiapG2, err := bn.Pairing(proof.PiAp, bn.G2.G)
	if err != nil {
		return false
	}
	if !bn.Fq12.Equal(pairingPiaVa, pairingPiapG2) {
		return false
	}

	// e(piB, Vb) == e(piB', g2)

	// e(piC, Vc) == e(piC', g2)
	pairingPicVc, err := bn.Pairing(proof.PiC, proof.Vc)
	if err != nil {
		return false
	}
	pairingPicpG2, err := bn.Pairing(proof.PiCp, bn.G2.G)
	if err != nil {
		return false
	}
	if !bn.Fq12.Equal(pairingPicVc, pairingPicpG2) {
		return false
	}

	//

	return true
}
