package circuitcompiler

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

type factors []*factor

type factor struct {
	typ            Token
	name           string
	invert, negate bool
	multiplicative [2]int
}

func (f factors) Len() int {
	return len(f)
}

func (f factors) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f factors) Less(i, j int) bool {
	if strings.Compare(f[i].String(), f[j].String()) < 0 {
		return false
	}
	return true
}

func (f factor) String() string {
	if f.typ == CONST {
		return fmt.Sprintf("(const fac: %v)", f.multiplicative)
	}
	str := f.name
	if f.invert {
		str += "^-1"
	}
	if f.negate {
		str = "-" + str
	}
	return fmt.Sprintf("(\"%s\"  fac: %v)", str, f.multiplicative)
}

func (f factors) clone() (res factors) {
	res = make(factors, len(f))
	for k, v := range f {
		res[k] = &factor{multiplicative: v.multiplicative, typ: v.typ, name: v.name, invert: v.invert, negate: v.negate}
	}
	return
}

func (f factors) normalizeAll() {
	for i, _ := range f {
		f[i].multiplicative = normalizeFactor(f[i].multiplicative)
	}
}

// find Least Common Multiple (LCM) via GCD
func LCMsmall(a, b int) int {
	result := a * b / GCD(a, b)

	return result
}

func extractFactor(f factors) (factors, [2]int) {

	lcm := f[0].multiplicative[1]

	for i := 1; i < len(f); i++ {
		lcm = LCMsmall(f[i].multiplicative[1], lcm)
	}

	for i := 0; i < len(f); i++ {
		f[i].multiplicative[0] = (lcm / f[i].multiplicative[1]) * f[i].multiplicative[0]
	}

	gcd := f[0].multiplicative[0]
	for i := 1; i < len(f); i++ {
		gcd = GCD(f[i].multiplicative[0], gcd)
	}
	for i := 0; i < len(f); i++ {
		f[i].multiplicative[0] = f[i].multiplicative[0] / gcd
		f[i].multiplicative[1] = 1
	}

	return f, [2]int{gcd, lcm}

}

func factorsSignature(leftFactors, rightFactors factors) (sig MultiplicationGateSignature, extractedLeftFactors, extractedRightFactors factors) {
	leftFactors = leftFactors.clone()
	rightFactors = rightFactors.clone()

	leftFactors.normalizeAll()
	var extractedFacLeft [2]int
	leftFactors, extractedFacLeft = extractFactor(leftFactors)

	sort.Sort(leftFactors)
	hasher.Reset()
	for _, fac := range leftFactors {
		hasher.Write([]byte(fac.String()))
	}
	leftNum := new(big.Int).SetBytes(hasher.Sum(nil))

	rightFactors.normalizeAll()

	var extractedFacRight [2]int
	rightFactors, extractedFacRight = extractFactor(rightFactors)
	sort.Sort(rightFactors)
	hasher.Reset()

	for _, fac := range rightFactors {
		hasher.Write([]byte(fac.String()))
	}
	rightNum := new(big.Int).SetBytes(hasher.Sum(nil))

	//we did all this, because multiplication is commutativ, and we want the signature of a
	//mulitplication gate   factorsSignature(a,b) == factorsSignature(b,a)
	leftNum.Add(leftNum, rightNum)

	res := normalizeFactor(mul2DVector(extractedFacLeft, extractedFacRight))

	return MultiplicationGateSignature{identifier: leftNum.String()[:16], commonExtracted: res}, leftFactors, rightFactors
}

func lengthOfLongestSlice(a, b factors) int {
	if len(a) > len(b) {
		return len(a)
	}
	return len(b)
}

//multiplies factor elements and returns the result
//in case the factors do not hold any constants and all inputs are distinct, the output will be the concatenation of left+right
func mulFactors(leftFactors, rightFactors factors) (result factors) {

	if len(leftFactors) < len(rightFactors) {
		tmp := leftFactors
		leftFactors = rightFactors
		rightFactors = tmp
	}

	for i, left := range leftFactors {

		for _, right := range rightFactors {

			if left.typ == CONST && right.typ == IN {
				leftFactors[i] = &factor{typ: IN, name: right.name, negate: Xor(left.negate, right.negate), invert: right.invert, multiplicative: mul2DVector(right.multiplicative, left.multiplicative)}
				continue
			}

			if right.typ == CONST && left.typ == IN {
				leftFactors[i] = &factor{typ: IN, name: left.name, negate: Xor(left.negate, right.negate), invert: left.invert, multiplicative: mul2DVector(right.multiplicative, left.multiplicative)}
				continue
			}

			if right.typ&left.typ == CONST {
				leftFactors[i] = &factor{typ: CONST, negate: Xor(right.negate, left.negate), multiplicative: mul2DVector(right.multiplicative, left.multiplicative)}
				continue

			}
			//tricky part here
			//this one should only be reached, after a true mgate had its left and right braches computed. here we
			//a factor can appear at most in quadratic form. we reduce terms a*a^-1 here.
			if right.typ&left.typ == IN {
				if left.name == right.name {
					if right.invert != left.invert {
						leftFactors[i] = &factor{typ: CONST, negate: Xor(right.negate, left.negate), multiplicative: mul2DVector(right.multiplicative, left.multiplicative)}
						continue
					}
				}

				//rightFactors[i] = factor{typ: CONST, negate: Xor(facRight.negate, facLeft.negate), multiplicative: mul2DVector(facRight.multiplicative, facLeft.multiplicative)}
				//continue

			}
			panic("unexpected. If this errror is thrown, its probably brcause a true multiplication gate has been skipped and treated as on with constant multiplication or addition ")

		}

	}

	return leftFactors
}

//returns the absolute value of a signed int and a flag telling if the input was positive or not
//this implementation is awesome and fast (see Henry S Warren, Hackers's Delight)
func abs(n int) (val int, positive bool) {
	y := n >> 63
	return (n ^ y) - y, y == 0
}

//adds two factors to one iff they are both are constants or of the same variable
func addFactor(facLeft, facRight factor) (couldAdd bool, sum factor) {
	if facLeft.typ&facRight.typ == CONST {
		var a0, b0 = facLeft.multiplicative[0], facRight.multiplicative[0]
		if facLeft.negate {
			a0 *= -1
		}
		if facRight.negate {
			b0 *= -1
		}
		absValue, positive := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)

		return true, factor{typ: CONST, negate: !positive, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}

	}
	if facLeft.typ&facRight.typ == IN && facLeft.invert == facRight.invert && facLeft.name == facRight.name {
		var a0, b0 = facLeft.multiplicative[0], facRight.multiplicative[0]
		if facLeft.negate {
			a0 *= -1
		}
		if facRight.negate {
			b0 *= -1
		}
		absValue, positive := abs(a0*facRight.multiplicative[1] + facLeft.multiplicative[1]*b0)

		return true, factor{typ: IN, invert: facRight.invert, name: facRight.name, negate: !positive, multiplicative: [2]int{absValue, facLeft.multiplicative[1] * facRight.multiplicative[1]}}

	}
	return false, factor{}

}

//returns the reduced sum of two input factor arrays
//if no reduction was done (worst case), it returns the concatenation of the input arrays
func addFactors(leftFactors, rightFactors factors) factors {
	var found bool
	res := make(factors, 0, len(leftFactors)+len(rightFactors))
	for _, facLeft := range leftFactors {

		found = false
		for i, facRight := range rightFactors {

			var sum factor
			found, sum = addFactor(*facLeft, *facRight)

			if found {
				rightFactors[i] = &sum
				break
			}

		}
		if !found {
			res = append(res, facLeft)
		}
	}

	for _, val := range rightFactors {
		if val.multiplicative[0] != 0 {
			res = append(res, val)
		}
	}

	return res
}

// -4/-5 -> 4/5  ;  5/-7 -> -5/7  ; 6 /2 -> 3 / 1
func normalizeFactor(b [2]int) [2]int {
	resa, signa := abs(b[0])
	resb, signb := abs(b[1])

	gcd := GCD(resa, resb)

	if Xor(signa, signb) {
		resa = -resa
	}
	return [2]int{resa / gcd, resb / gcd}
}

//naive component multiplication
func mul2DVector(a, b [2]int) [2]int {

	return [2]int{a[0] * b[0], a[1] * b[1]}
}

// find Least Common Multiple (LCM) via GCD
func LCM(a, b int, integers ...int) int {
	result := a * b / GCD(a, b)

	for i := 0; i < len(integers); i++ {
		result = LCM(result, integers[i])
	}

	return result
}

//euclidean algo to determine greates common divisor
func GCD(a, b int) int {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}
