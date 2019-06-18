package proof

import (
	"math/big"

	"github.com/arnaucube/go-snark/bn128"
	"github.com/arnaucube/go-snark/circuit"
	"github.com/arnaucube/go-snark/fields"
)

// Proof is ...
type Proof interface{}

// Setup is ...
type Setup interface {
	Z() []*big.Int
	Init(cir *circuit.Circuit, alphas, betas, gammas [][]*big.Int) error
	Generate(cir *circuit.Circuit, w []*big.Int, px []*big.Int) (Proof, error)
	Verify(p Proof, publicSignals []*big.Int) (bool, error)
}

// Utils is ...
var Utils struct {
	Bn  bn128.Bn128
	FqR fields.Fq
	PF  fields.PF
}

func init() {
	var err error
	if Utils.Bn, err = bn128.NewBn128(); err != nil {
		panic(err)
	}
	Utils.FqR = fields.NewFq(Utils.Bn.R)
	Utils.PF = fields.NewPF(Utils.FqR)
}
