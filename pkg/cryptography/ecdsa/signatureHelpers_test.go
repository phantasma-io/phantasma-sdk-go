package ecdsa

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignatureDropRecoveryID(t *testing.T) {
	signatureWithRIDBytes, err := hex.DecodeString(k1SignatureRefWithRID)
	if err != nil {
		panic(err)
	}

	result := SignatureDropRecoveryID(signatureWithRIDBytes)
	result2 := SignatureDropRecoveryID(result)

	// Method should not modify input array
	assert.Equal(t, k1SignatureRefWithRID, hex.EncodeToString(signatureWithRIDBytes))
	assert.NotEqual(t, k1SignatureRef, hex.EncodeToString(signatureWithRIDBytes))

	assert.Equal(t, k1SignatureRef, hex.EncodeToString(result))
	assert.Equal(t, result, result2)
}

func TestSignatureToRSConversions(t *testing.T) {
	{
		signatureBytes, err := hex.DecodeString(k1SignatureRef)
		if err != nil {
			panic(err)
		}

		r, s, err := SignatureToRS(signatureBytes)
		if err != nil {
			panic(err)
		}
		signatureBytesRecreated := RSToSignatureWithoutRecoveryID(r, s)
		assert.Equal(t, signatureBytes, signatureBytesRecreated)
	}

	{
		signatureBytes, err := hex.DecodeString(r1SignatureRef)
		if err != nil {
			panic(err)
		}

		r, s, err := SignatureToRS(signatureBytes)
		if err != nil {
			panic(err)
		}
		signatureBytesRecreated := RSToSignatureWithoutRecoveryID(r, s)
		assert.Equal(t, signatureBytes, signatureBytesRecreated)
	}
}

func TestSignatureToRSRejectsMalformedInput(t *testing.T) {
	_, _, err := SignatureToRS([]byte{1, 2, 3})
	assert.Error(t, err)
}
