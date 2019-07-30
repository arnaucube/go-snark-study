package externalVerif

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyFromCircom(t *testing.T) {
	verified, err := VerifyFromCircom("circom-test/verification_key.json", "circom-test/proof.json", "circom-test/public.json")
	assert.Nil(t, err)
	assert.True(t, verified)
}
