package cryptography

import (
	"crypto/rand"
	"testing"

	"github.com/phantasma-io/phantasma-sdk-go/pkg/encoding/base58"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSeed() []byte {
	return []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32}
}

func TestNewPhantasmaKeys(t *testing.T) {
	kp, err := NewPhantasmaKeys(testSeed())
	require.NoError(t, err)
	assert.Equal(t, kp.Address().String(), "P2KCineFiatZR8fDyU4pSfb3Bq1vtyW3zqBJi268YV5fH9e")
	assert.Equal(t, kp.PrivateKey(), []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 48, 49, 50})
}

func TestPhantasmaKeysReturnDefensiveCopies(t *testing.T) {
	kp, err := NewPhantasmaKeys(testSeed())
	require.NoError(t, err)

	privateKey := kp.PrivateKey()
	expandedPrivateKey := kp.ExpandedPrivateKey()
	publicKey := kp.PublicKey()

	privateKey[0] ^= 0xff
	expandedPrivateKey[0] ^= 0xff
	publicKey[0] ^= 0xff

	assert.Equal(t, testSeed(), kp.PrivateKey())
	assert.Equal(t, "P2KCineFiatZR8fDyU4pSfb3Bq1vtyW3zqBJi268YV5fH9e", kp.Address().String())
}

func TestSign(t *testing.T) {
	kp, err := NewPhantasmaKeys(testSeed())
	require.NoError(t, err)

	msg := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	signature, err := kp.Sign(msg)
	require.NoError(t, err)
	success := signature.Verify(msg, []Address{kp.Address()})
	assert.True(t, success)
	shouldBe := []byte{64, 164, 57, 54, 21, 248, 110, 21, 158, 249, 253, 212, 10, 91, 166, 191, 46, 24, 226, 196, 251, 248, 253, 139, 177, 85, 156, 226, 60, 118, 201, 25, 163, 130, 117, 237, 146, 170, 157, 243, 177, 9, 199, 199, 86, 56, 221, 14, 36, 210, 92, 72, 248, 253, 79, 120, 248, 183, 190, 192, 207, 87, 165, 193, 3}
	assert.Equal(t, shouldBe, signature.Bytes())
}

func TestWif(t *testing.T) {
	kp, err := NewPhantasmaKeys(testSeed())
	require.NoError(t, err)
	assert.Equal(t, "KwFfpDsaF7yxCEQWPjchTihESWYLdG9AnkareK15wKCnhM8BV4QY", kp.WIF())
	kpNew, err := FromWIF(kp.WIF())
	assert.Equal(t, "KwFfpDsaF7yxCEQWPjchTihESWYLdG9AnkareK15wKCnhM8BV4QY", kpNew.WIF())
	assert.Nil(t, err)
}

func TestFromWIFRejectsMalformedPayloads(t *testing.T) {
	tests := []string{
		base58.CheckEncode([]byte{0x80}),
		base58.CheckEncode(append([]byte{0x81}, testSeed()...)),
		base58.CheckEncode(append(append([]byte{0x80}, testSeed()...), 0x02)),
	}

	for _, tc := range tests {
		_, err := FromWIF(tc)
		assert.Error(t, err)
	}
}

func TestGenerate(t *testing.T) {
	kp, err := GeneratePhantasmaKeys()
	require.NoError(t, err)
	kpNew, err := FromWIF(kp.WIF())
	assert.Nil(t, err)
	assert.Equal(t, kpNew.Address(), kp.Address())
}

func TestSignRejectsInvalidKey(t *testing.T) {
	var keys PhantasmaKeys
	signature, err := keys.Sign([]byte{1, 2, 3})
	require.Error(t, err)
	require.Nil(t, signature)
}

func TestSignBig(t *testing.T) {
	kp, err := NewPhantasmaKeys(testSeed())
	require.NoError(t, err)

	msg := make([]byte, 1000000)
	rand.Read(msg)
	signature, err := kp.Sign(msg)
	require.NoError(t, err)
	success := signature.Verify(msg, []Address{kp.Address()})
	assert.True(t, success)
}
