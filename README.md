# go-snark [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/go-snark)](https://goreportcard.com/report/github.com/arnaucube/go-snark) [![Build Status](https://travis-ci.org/arnaucube/go-snark.svg?branch=master)](https://travis-ci.org/arnaucube/go-snark) [![Gitter](https://badges.gitter.im/go-snark/community.svg)](https://gitter.im/go-snark/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

zkSNARK library implementation in Go


- `Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture`, Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, Madars Virza https://eprint.iacr.org/2013/879.pdf
- `Pinocchio: Nearly practical verifiable computation`, Bryan Parno, Craig Gentry, Jon Howell, Mariana Raykova https://eprint.iacr.org/2013/279.pdf

## Caution & Warning
Implementation of the zkSNARK [Pinocchio protocol](https://eprint.iacr.org/2013/279.pdf) from scratch in Go to understand the concepts. Do not use in production.

Not finished, implementing this in my free time to understand it better, so I don't have much time.

Currently allows to do the complete path with [Pinocchio protocol](https://eprint.iacr.org/2013/279.pdf) :
1. compile circuuit
2. generate trusted setup
3. calculate witness
4. generate proofs
5. verify proofs

Minimal complete flow implementation:
- [x] Finite Fields (1, 2, 6, 12) operations
- [x] G1 and G2 curve operations
- [x] BN128 Pairing
- [x] circuit flat code compiler
- [x] circuit to R1CS
- [x] polynomial operations
- [x] R1CS to QAP
- [x] generate trusted setup
- [x] generate proofs
- [x] verify proofs with BN128 pairing

Improvements from the minimal implementation:
- [x] allow to call functions in circuits language
- [x] allow `import` in circuits language
- [ ] allow `for` in circuits language
- [ ] move witness values calculation outside the setup phase
- [ ] Groth16
- [ ] multiple optimizations


## Usage
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark?status.svg)](https://godoc.org/github.com/arnaucube/go-snark) zkSnark
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/bn128?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/bn128) bn128 (more details: https://github.com/arnaucube/go-snark/tree/master/bn128)
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/fields?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/fields) Finite Fields operations
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/r1csqap?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/r1csqap) R1CS to QAP (more details: https://github.com/arnaucube/go-snark/tree/master/r1csqap)
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/circuitcompiler?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/circuitcompiler) Circuit Compiler

### CLI usage
*The cli still needs some improvements, such as seting input files, etc.*

In this example we will follow the equation example from [Vitalik](https://medium.com/@VitalikButerin/quadratic-arithmetic-programs-from-zero-to-hero-f6d558cea649)'s article: `y = x^3 + x + 5`, where `y==35` and `x==3`. So we want to prove that we know a secret `x` such as the result of the equation is `35`.

#### Compile circuit
Having a circuit file `test.circuit`:
```
func exp3(private a):
	b = a * a
	c = a * b
	return c

func main(private s0, public s1):
	s3 = exp3(s0)
	s4 = s3 + s0
	s5 = s4 + 5
	equals(s1, s5)
	out = 1 * 1
```
And a private inputs file `privateInputs.json`
```
[
	3
]
```
And a public inputs file `publicInputs.json`
```
[
	35
]
```

In the command line, execute:
```
> ./go-snark-cli compile test.circuit
```

This will output the `compiledcircuit.json` file.

#### Trusted Setup
Having the `compiledcircuit.json`, now we can generate the `TrustedSetup`:
```
> ./go-snark-cli trustedsetup
```
This will create the file `trustedsetup.json` with the TrustedSetup data, and also a `toxic.json` file, with the parameters to delete from the `Trusted Setup`.


#### Generate Proofs
Assumming that we have the `compiledcircuit.json`, `trustedsetup.json`, `privateInputs.json` and the `publicInputs.json` we can now generate the `Proofs` with the following command:
```
> ./go-snark-cli genproofs
```

This will store the file `proofs.json`, that contains all the SNARK proofs.

#### Verify Proofs
Having the `proofs.json`, `compiledcircuit.json`, `trustedsetup.json` `publicInputs.json` files, we can now verify the `Pairings` of the proofs, in order to verify the proofs.
```
> ./go-snark-cli verify
```
This will return a `true` if the proofs are verified, or a `false` if the proofs are not verified.


### Library usage

Example:
```go
// compile circuit and get the R1CS
flatCode := `
func exp3(private a):
	b = a * a
	c = a * b
	return c

func main(private s0, public s1):
	s3 = exp3(s0)
	s4 = s3 + s0
	s5 = s4 + 5
	equals(s1, s5)
	out = 1 * 1
`

// parse the code
parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
circuit, err := parser.Parse()
assert.Nil(t, err)
fmt.Println(circuit)


b3 := big.NewInt(int64(3))
privateInputs := []*big.Int{b3}
b35 := big.NewInt(int64(35))
publicSignals := []*big.Int{b35}

// witness
w, err := circuit.CalculateWitness(privateInputs, publicSignals)
assert.Nil(t, err)
fmt.Println("witness", w)

// now we have the witness:
// w = [1 35 3 9 27 30 35 1]

// flat code to R1CS
fmt.Println("generating R1CS from flat code")
a, b, c := circuit.GenerateR1CS()

/*
now we have the R1CS from the circuit:
a: [[0 0 1 0 0 0 0 0] [0 0 1 0 0 0 0 0] [0 0 1 0 1 0 0 0] [5 0 0 0 0 1 0 0] [0 0 0 0 0 0 1 0] [0 1 0 0 0 0 0 0] [1 0 0 0 0 0 0 0]]
b: [[0 0 1 0 0 0 0 0] [0 0 0 1 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0] [1 0 0 0 0 0 0 0]]
c: [[0 0 0 1 0 0 0 0] [0 0 0 0 1 0 0 0] [0 0 0 0 0 1 0 0] [0 0 0 0 0 0 1 0] [0 1 0 0 0 0 0 0] [0 0 0 0 0 0 1 0] [0 0 0 0 0 0 0 1]]
*/


alphas, betas, gammas, _ := snark.Utils.PF.R1CSToQAP(a, b, c)


ax, bx, cx, px := Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

// calculate trusted setup
setup, err := GenerateTrustedSetup(len(w), *circuit, alphas, betas, gammas)

hx := Utils.PF.DivisorPolynomial(px, setup.Pk.Z)

proof, err := GenerateProofs(*circuit, setup, w, px)

b35Verif := big.NewInt(int64(35))
publicSignalsVerif := []*big.Int{b35Verif}
assert.True(t, VerifyProof(*circuit, setup, proof, publicSignalsVerif, true))
```



## Test
```
go test ./... -v
```

## vim/nvim circuit syntax highlighter
For more details and installation instructions see https://github.com/arnaucube/go-snark/tree/master/vim-syntax

---


Thanks to [@jbaylina](https://github.com/jbaylina), [@bellesmarta](https://github.com/bellesmarta), [@adriamb](https://github.com/adriamb) for their explanations that helped to understand this a little bit. Also thanks to [@vbuterin](https://github.com/vbuterin) for all the published articles explaining the zkSNARKs.
