package ecdsa

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUncompressedPublicKeyTo65Bytes(t *testing.T) {
	pubKeyBytes, err := hex.DecodeString(k1PubKey)
	if err != nil {
		panic(err)
	}

	result, err := UncompressedPublicKeyTo65Bytes(pubKeyBytes)
	if err != nil {
		panic(err)
	}
	result2, err := UncompressedPublicKeyTo65Bytes(result)
	if err != nil {
		panic(err)
	}

	// Method should not modify input array
	assert.Equal(t, k1PubKey, hex.EncodeToString(pubKeyBytes))
	assert.NotEqual(t, k1PubKey65, hex.EncodeToString(pubKeyBytes))

	assert.Equal(t, k1PubKey65, hex.EncodeToString(result))
	assert.Equal(t, result, result2)
}

func TestPublicKeyHelpersRejectMalformedInput(t *testing.T) {
	_, err := UncompressedPublicKeyTo65Bytes([]byte{1, 2, 3})
	assert.Error(t, err)

	_, err = CompressPublicKey([]byte{1, 2, 3})
	assert.Error(t, err)

	_, err = PublicKeyUnmarshal([]byte{1, 2, 3}, nil)
	assert.Error(t, err)
}
