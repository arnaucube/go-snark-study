package circuitcompiler

type Circuit struct {
	NVars       int
	NPublic     int
	NSignals    int
	Inputs      []int
	Witness     []int
	Constraints []Constraint
}
