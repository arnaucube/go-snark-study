# go-snark [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/go-snark)](https://goreportcard.com/report/github.com/arnaucube/go-snark)

zkSNARK library implementation in Go


- `Succinct Non-Interactive Zero Knowledge for a von Neumann Architecture`, Eli Ben-Sasson, Alessandro Chiesa, Eran Tromer, Madars Virza https://eprint.iacr.org/2013/879.pdf
- `Pinocchio: Nearly practical verifiable computation`, Bryan Parno, Craig Gentry, Jon Howell, Mariana Raykova https://eprint.iacr.org/2013/279.pdf

## Caution, Warning
Implementation of the zkSNARK [Pinocchio protocol](https://eprint.iacr.org/2013/279.pdf) from scratch in Go to understand the concepts. Do not use in production.

Not finished, implementing this in my free time to understand it better, so I don't have much time.

Current implementation status:
- [x] Finite Fields (1, 2, 6, 12) operations
- [x] G1 and G2 curve operations
- [x] BN128 Pairing
- [x] circuit code compiler
	- [ ] code to flat code (improve circuit compiler)
	- [x] flat code compiler
- [x] circuit to R1CS
- [x] polynomial operations
- [x] R1CS to QAP
- [x] generate trusted setup
- [x] generate proofs
- [x] verify proofs with BN128 pairing
- [ ] move witness calculation outside the setup phase
- [ ] Groth16
- [ ] multiple optimizations


## Usage
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark?status.svg)](https://godoc.org/github.com/arnaucube/go-snark) zkSnark
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/bn128?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/bn128) bn128 (more details: https://github.com/arnaucube/go-snark/tree/master/bn128)
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/fields?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/fields) Finite Fields operations
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/r1csqap?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/r1csqap) R1CS to QAP (more details: https://github.com/arnaucube/go-snark/tree/master/r1csqap)
- [![GoDoc](https://godoc.org/github.com/arnaucube/go-snark/circuitcompiler?status.svg)](https://godoc.org/github.com/arnaucube/go-snark/circuitcompiler) Circuit Compiler

### Library usage
Warning: not finished.

Example:
```go
// compile circuit and get the R1CS
flatCode := `
func test(x):
	aux = x*x
	y = aux*x
	z = x + y
	out = z + 5
`

// parse the code
parser := circuitcompiler.NewParser(strings.NewReader(flatCode))
circuit, err := parser.Parse()
assert.Nil(t, err)
fmt.Println(circuit)

// witness
b3 := big.NewInt(int64(3))
inputs := []*big.Int{b3}
w := circuit.CalculateWitness(inputs)
fmt.Println("\nwitness", w)
/*
now we have the witness:
w = [1 3 35 9 27 30]
*/

// flat code to R1CS
fmt.Println("generating R1CS from flat code")
a, b, c := circuit.GenerateR1CS()

/*
now we have the R1CS from the circuit:
a == [[0 1 0 0 0 0] [0 0 0 1 0 0] [0 1 0 0 1 0] [5 0 0 0 0 1]]
b == [[0 1 0 0 0 0] [0 1 0 0 0 0] [1 0 0 0 0 0] [1 0 0 0 0 0]]
c == [[0 0 0 1 0 0] [0 0 0 0 1 0] [0 0 0 0 0 1] [0 0 1 0 0 0]]
*/


alphas, betas, gammas, zx := snark.Utils.PF.R1CSToQAP(a, b, c)


ax, bx, cx, px := snark.Utils.PF.CombinePolynomials(w, alphas, betas, gammas)

hx := snark.Utils.PF.DivisorPolinomial(px, zx)

// hx==px/zx so px==hx*zx
assert.Equal(t, px, snark.Utils.PF.Mul(hx, zx))

// p(x) = a(x) * b(x) - c(x) == h(x) * z(x)
abc := snark.Utils.PF.Sub(pf.Mul(ax, bx), cx)
assert.Equal(t, abc, px)
hz := snark.Utils.PF.Mul(hx, zx)
assert.Equal(t, abc, hz)
	
div, rem := snark.Utils.PF.Div(px, zx)
assert.Equal(t, hx, div)
assert.Equal(t, rem, r1csqap.ArrayOfBigZeros(4))

// calculate trusted setup
setup, err := snark.GenerateTrustedSetup(len(w), circuit, alphas, betas, gammas, zx)
assert.Nil(t, err)
fmt.Println("t", setup.Toxic.T)

// piA = g1 * A(t), piB = g2 * B(t), piC = g1 * C(t), piH = g1 * H(t)
proof, err := snark.GenerateProofs(circuit, setup, hx, w)
assert.Nil(t, err)

assert.True(t, snark.VerifyProof(circuit, setup, proof))
```

### CLI usage

#### Compile circuit
Having a circuit file `test.circuit`:
```
func test(x):
	aux = x*x
	y = aux*x
	z = x + y
	out = z + 5
```
And a inputs file `inputs.json`
```
[
	3
]
```

In the command line, execute:
```
> go-snark-cli compile test.circuit
```

This will output the `compiledcircuit.json` file.

#### Trusted Setup
Having the `compiledcircuit.json`, now we can generate the `TrustedSetup`:
```
> go-snark-cli trustedsetup
```
This will create the file `trustedsetup.json` with the TrustedSetup data, and also a `toxic.json` file, with the parameters to delete from the `Trusted Setup`.


#### Generate Proofs
Assumming that we have the `compiledcircuit.json` and the `trustedsetup.json`, we can now generate the `Proofs` with the following command:
```
> go-snark-cli genproofs
```

This will store the file `proofs.json`, that contains all the SNARK proofs.

#### Verify Proofs
Having the `proofs.json`, `compiledcircuit.json`, `trustedsetup.json` files, we can now verify the `Pairings` of the proofs, in order to verify the proofs.
```
> go-snark-cli verify
```
This will return a `true` if the proofs are verified, or a `false` if the proofs are not verified.



## Test
```
go test ./... -v
```

---


Thanks to [@jbaylina](https://github.com/jbaylina), [@bellesmarta](https://github.com/bellesmarta), [@adriamb](https://github.com/adriamb) for their explanations that helped to understand this a little bit. Also thanks to [@vbuterin](https://github.com/vbuterin) for all the published articles explaining the zkSNARKs.
