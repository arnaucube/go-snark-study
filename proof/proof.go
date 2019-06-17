package proof

import (
	"math/big"

	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/fields"
	"github.com/arnaucube/go-snark/r1csqap"
)

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

// Proof is
type Proof interface{}

// Setup is
type Setup interface {
	Z() []*big.Int
	Init(witnessLength int, circuit circuit.Circuit, alphas, betas, gammas [][]*big.Int) error
	Generate(circuit circuit.Circuit, w []*big.Int, px []*big.Int) (Proof, error)
	Verify(circuit circuit.Circuit, proof Proof, publicSignals []*big.Int, debug bool) bool
}
