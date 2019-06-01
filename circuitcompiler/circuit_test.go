package circuitcompiler

import (
	//"fmt"
	////"math/big"
	//"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXor(t *testing.T) {
	assert.Equal(t, false, Xor(true, true))
	assert.Equal(t, true, Xor(true, false))
	assert.Equal(t, true, Xor(false, true))
	assert.Equal(t, false, Xor(false, false))

}
